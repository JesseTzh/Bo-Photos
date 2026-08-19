package asset

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/besscroft/bophotos/backend/internal/imageproc"
	"github.com/besscroft/bophotos/backend/internal/storage"
)

type Enqueuer interface {
	Enqueue(string) error
}

type Service struct {
	repository     *Repository
	storage        *storage.Local
	processor      imageproc.Processor
	queue          Enqueuer
	maxUploadBytes int64
}

type UploadResult struct {
	Asset             Asset
	DuplicateAssetIDs []string
}

func NewService(
	repository *Repository,
	localStorage *storage.Local,
	processor imageproc.Processor,
	queue Enqueuer,
	maxUploadBytes int64,
) *Service {
	return &Service{
		repository:     repository,
		storage:        localStorage,
		processor:      processor,
		queue:          queue,
		maxUploadBytes: maxUploadBytes,
	}
}

func (s *Service) SetQueue(queue Enqueuer) {
	s.queue = queue
}

func (s *Service) Upload(ctx context.Context, originalName string, source io.Reader) (UploadResult, error) {
	id, err := newID()
	if err != nil {
		return UploadResult{}, err
	}
	staged, err := s.storage.WriteStaging(ctx, id, source, s.maxUploadBytes)
	if err != nil {
		return UploadResult{}, err
	}

	stagingFile, err := s.storage.Open(staged.Key)
	if err != nil {
		return UploadResult{}, err
	}
	detection, detectErr := imageproc.DetectMedia(stagingFile)
	_ = stagingFile.Close()
	extension, mimeType := detection.Extension, detection.MIMEType
	isVideo := false
	if detectErr != nil {
		videoFile, openErr := s.storage.Open(staged.Key)
		if openErr != nil {
			_ = s.storage.Purge(staged.Key)
			return UploadResult{}, openErr
		}
		extension, mimeType, detectErr = detectVideoFormat(videoFile, originalName)
		_ = videoFile.Close()
		isVideo = detectErr == nil
	}
	if detectErr != nil {
		_ = s.storage.Purge(staged.Key)
		return UploadResult{}, detectErr
	}
	originalKey, err := s.storage.PromoteOriginal(staged.Key, id, extension)
	if err != nil {
		return UploadResult{}, err
	}

	duplicates, err := s.repository.FindBySHA256(ctx, staged.SHA256)
	if err != nil {
		_ = s.storage.Purge(originalKey)
		return UploadResult{}, err
	}
	now := time.Now().UTC()
	status := StatusProcessing
	if isVideo {
		status = StatusReady
	}
	item := Asset{
		ID:                id,
		Status:            status,
		OriginalName:      filepath.Base(originalName),
		OriginalKey:       originalKey,
		SHA256:            staged.SHA256,
		MIMEType:          mimeType,
		ByteSize:          staged.Bytes,
		Visible:           true,
		ShowOnHomepage:    true,
		DerivativeVersion: 1,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := s.repository.Create(ctx, item); err != nil {
		_ = s.storage.Purge(originalKey)
		return UploadResult{}, err
	}
	if isVideo {
		return UploadResult{Asset: item, DuplicateAssetIDs: duplicates}, nil
	}
	if s.queue == nil {
		_ = s.repository.Transition(ctx, id, StatusFailed, "ASSET_QUEUE_UNAVAILABLE")
		return UploadResult{}, errors.New("asset queue unavailable")
	}
	if err := s.queue.Enqueue(id); err != nil {
		_ = s.repository.Transition(ctx, id, StatusFailed, "ASSET_QUEUE_FULL")
		return UploadResult{}, err
	}
	return UploadResult{Asset: item, DuplicateAssetIDs: duplicates}, nil
}

func detectVideoFormat(reader io.Reader, originalName string) (string, string, error) {
	buffered := bufio.NewReader(reader)
	header, _ := buffered.Peek(32)
	extension := strings.ToLower(filepath.Ext(originalName))
	if len(header) >= 12 && string(header[4:8]) == "ftyp" {
		if extension == ".mov" {
			return ".mov", "video/quicktime", nil
		}
		return ".mp4", "video/mp4", nil
	}
	if len(header) >= 4 && bytes.Equal(header[:4], []byte{0x1a, 0x45, 0xdf, 0xa3}) && extension == ".webm" {
		return ".webm", "video/webm", nil
	}
	return "", "", imageproc.ErrUnsupportedFormat
}

func (s *Service) Process(ctx context.Context, id string) {
	item, err := s.repository.Get(ctx, id)
	if err != nil || item.Status != StatusProcessing {
		return
	}
	originalPath, err := s.storage.Resolve(item.OriginalKey)
	if err != nil {
		_ = s.repository.Transition(ctx, id, StatusFailed, "ASSET_ORIGINAL_MISSING")
		return
	}
	if _, err := os.Stat(originalPath); err != nil {
		_ = s.repository.Transition(ctx, id, StatusFailed, "ASSET_ORIGINAL_MISSING")
		return
	}

	previewKey := fmt.Sprintf("previews/%s-v%d.webp", id, item.DerivativeVersion)
	thumbnailKey := fmt.Sprintf("thumbnails/%s-v%d.webp", id, item.DerivativeVersion)
	previewTempKey := fmt.Sprintf("previews/.%s-v%d.tmp.webp", id, item.DerivativeVersion)
	thumbnailTempKey := fmt.Sprintf("thumbnails/.%s-v%d.tmp.webp", id, item.DerivativeVersion)
	previewPath, err := s.storage.Resolve(previewTempKey)
	if err != nil {
		_ = s.repository.Transition(ctx, id, StatusFailed, "ASSET_STORAGE_ERROR")
		return
	}
	thumbnailPath, err := s.storage.Resolve(thumbnailTempKey)
	if err != nil {
		_ = s.repository.Transition(ctx, id, StatusFailed, "ASSET_STORAGE_ERROR")
		return
	}

	family, kind := detectStoredMedia(originalPath, item.OriginalKey)
	metadata, err := s.processor.Process(ctx, imageproc.Request{
		AssetID:       id,
		OriginalPath:  originalPath,
		PreviewPath:   previewPath,
		ThumbnailPath: thumbnailPath,
		Family:        family,
		Kind:          kind,
	})
	if err != nil {
		_ = s.storage.Purge(previewTempKey, thumbnailTempKey)
		_ = s.repository.Transition(ctx, id, StatusFailed, "ASSET_PROCESSING_FAILED")
		return
	}
	if err := s.storage.Promote(previewTempKey, previewKey); err != nil {
		_ = s.storage.Purge(previewTempKey, thumbnailTempKey)
		_ = s.repository.Transition(ctx, id, StatusFailed, "ASSET_STORAGE_ERROR")
		return
	}
	if err := s.storage.Promote(thumbnailTempKey, thumbnailKey); err != nil {
		_ = s.storage.Purge(previewKey, previewTempKey, thumbnailTempKey)
		_ = s.repository.Transition(ctx, id, StatusFailed, "ASSET_STORAGE_ERROR")
		return
	}
	if err := s.repository.MarkReady(ctx, id, previewKey, thumbnailKey, metadata); err != nil {
		_ = s.storage.Purge(previewKey, thumbnailKey)
	}
}

func (s *Service) Retry(ctx context.Context, id string) error {
	item, err := s.repository.Get(ctx, id)
	if err != nil {
		return err
	}
	if item.Status != StatusReady && item.Status != StatusFailed {
		return ErrInvalidTransition
	}
	if strings.HasPrefix(item.MIMEType, "video/") {
		return ErrInvalidTransition
	}
	originalPath, err := s.storage.Resolve(item.OriginalKey)
	if err != nil {
		return err
	}
	if _, err := os.Stat(originalPath); err != nil {
		if os.IsNotExist(err) {
			return ErrOriginalMissing
		}
		return fmt.Errorf("stat original: %w", err)
	}
	if err := s.repository.Transition(ctx, id, StatusProcessing, ""); err != nil {
		return err
	}
	if err := s.queue.Enqueue(id); err != nil {
		_ = s.repository.Transition(ctx, id, StatusFailed, "ASSET_QUEUE_FULL")
		return err
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repository.Transition(ctx, id, StatusDeleted, "")
}

func (s *Service) Purge(ctx context.Context, id string) error {
	item, err := s.repository.Get(ctx, id)
	if err != nil {
		return err
	}
	if item.Status != StatusDeleted {
		return ErrInvalidTransition
	}
	if err := s.storage.Purge(item.OriginalKey, item.PreviewKey, item.ThumbnailKey); err != nil {
		return err
	}
	return s.repository.Transition(ctx, id, StatusPurged, "")
}

func (s *Service) Restore(ctx context.Context, id string) error {
	return s.repository.Transition(ctx, id, StatusReady, "")
}

func newID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("generate asset ID: %w", err)
	}
	return hex.EncodeToString(value[:]), nil
}

func detectStoredMedia(originalPath, originalKey string) (imageproc.Format, string) {
	file, err := os.Open(originalPath)
	if err == nil {
		detection, detectErr := imageproc.DetectMedia(file)
		_ = file.Close()
		if detectErr == nil {
			return detection.Family, detection.Kind
		}
	}
	if kind := imageproc.RawKindFromExtension(filepath.Ext(originalKey)); kind != "" {
		return imageproc.FormatRAW, kind
	}
	return "", ""
}

package asset

import (
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
	format, detectErr := imageproc.DetectFormat(stagingFile)
	_ = stagingFile.Close()
	if detectErr != nil {
		_ = s.storage.Purge(staged.Key)
		return UploadResult{}, detectErr
	}
	extension, mimeType := formatInfo(format)
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
	item := Asset{
		ID:                id,
		Status:            StatusProcessing,
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

	metadata, err := s.processor.Process(ctx, imageproc.Request{
		OriginalPath: originalPath, PreviewPath: previewPath, ThumbnailPath: thumbnailPath,
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

func formatInfo(format imageproc.Format) (string, string) {
	switch format {
	case imageproc.FormatJPEG:
		return ".jpg", "image/jpeg"
	case imageproc.FormatPNG:
		return ".png", "image/png"
	case imageproc.FormatWebP:
		return ".webp", "image/webp"
	case imageproc.FormatHEIF:
		return ".heic", "image/heif"
	case imageproc.FormatTIFF:
		return ".tiff", "image/tiff"
	default:
		return "." + strings.ToLower(string(format)), "application/octet-stream"
	}
}

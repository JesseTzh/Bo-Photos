package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrUnsafeKey    = errors.New("unsafe storage key")
	ErrFileTooLarge = errors.New("file exceeds upload limit")
)

type Local struct {
	root string
}

type StagingResult struct {
	Key    string
	SHA256 string
	Bytes  int64
}

func NewLocal(root string) (*Local, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve storage root: %w", err)
	}
	for _, directory := range []string{
		"originals",
		"previews",
		"thumbnails",
		"staging",
		"trash",
	} {
		if err := os.MkdirAll(filepath.Join(absolute, directory), 0o755); err != nil {
			return nil, fmt.Errorf("create %s directory: %w", directory, err)
		}
	}
	return &Local{root: absolute}, nil
}

func (l *Local) Root() string {
	return l.root
}

func (l *Local) Resolve(key string) (string, error) {
	if key == "" || strings.ContainsRune(key, '\x00') || filepath.IsAbs(key) {
		return "", ErrUnsafeKey
	}
	clean := filepath.Clean(filepath.FromSlash(key))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", ErrUnsafeKey
	}

	path := filepath.Join(l.root, clean)
	relative, err := filepath.Rel(l.root, path)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", ErrUnsafeKey
	}

	current := l.root
	parts := strings.Split(relative, string(filepath.Separator))
	for index, part := range parts {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			if index < len(parts)-1 {
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					return "", fmt.Errorf("create storage parent: %w", err)
				}
			}
			break
		}
		if err != nil {
			return "", fmt.Errorf("inspect storage path: %w", err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", ErrUnsafeKey
		}
	}
	return path, nil
}

func (l *Local) WriteStaging(
	ctx context.Context,
	assetID string,
	source io.Reader,
	maxBytes int64,
) (StagingResult, error) {
	key := "staging/" + assetID + ".upload"
	path, err := l.Resolve(key)
	if err != nil {
		return StagingResult{}, err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return StagingResult{}, fmt.Errorf("create staging file: %w", err)
	}
	success := false
	defer func() {
		_ = file.Close()
		if !success {
			_ = os.Remove(path)
		}
	}()

	hash := sha256.New()
	reader := io.Reader(source)
	if maxBytes > 0 {
		reader = io.LimitReader(source, maxBytes+1)
	}
	written, err := copyWithContext(ctx, io.MultiWriter(file, hash), reader)
	if err != nil {
		return StagingResult{}, fmt.Errorf("write staging file: %w", err)
	}
	if maxBytes > 0 && written > maxBytes {
		return StagingResult{}, ErrFileTooLarge
	}
	if err := file.Sync(); err != nil {
		return StagingResult{}, fmt.Errorf("sync staging file: %w", err)
	}
	success = true
	return StagingResult{
		Key:    key,
		SHA256: hex.EncodeToString(hash.Sum(nil)),
		Bytes:  written,
	}, nil
}

func (l *Local) PromoteOriginal(stagingKey, assetID, extension string) (string, error) {
	if extension == "" || strings.ContainsAny(extension, `/\`) || !strings.HasPrefix(extension, ".") {
		return "", ErrUnsafeKey
	}
	source, err := l.Resolve(stagingKey)
	if err != nil {
		return "", err
	}
	key := "originals/" + assetID + strings.ToLower(extension)
	target, err := l.Resolve(key)
	if err != nil {
		return "", err
	}
	if err := os.Rename(source, target); err != nil {
		return "", fmt.Errorf("promote original: %w", err)
	}
	return key, nil
}

func (l *Local) Purge(keys ...string) error {
	for _, key := range keys {
		if key == "" {
			continue
		}
		path, err := l.Resolve(key)
		if err != nil {
			return err
		}
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove %s: %w", key, err)
		}
	}
	return nil
}

func (l *Local) Promote(sourceKey, targetKey string) error {
	source, err := l.Resolve(sourceKey)
	if err != nil {
		return err
	}
	target, err := l.Resolve(targetKey)
	if err != nil {
		return err
	}
	if err := os.Rename(source, target); err != nil {
		return fmt.Errorf("promote %s to %s: %w", sourceKey, targetKey, err)
	}
	return nil
}

func (l *Local) Open(key string) (*os.File, error) {
	path, err := l.Resolve(key)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", key, err)
	}
	return file, nil
}

func (l *Local) CleanupStaging(before time.Time) (int, error) {
	directory := filepath.Join(l.root, "staging")
	entries, err := os.ReadDir(directory)
	if err != nil {
		return 0, fmt.Errorf("read staging directory: %w", err)
	}
	removed := 0
	for _, entry := range entries {
		if entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return removed, fmt.Errorf("inspect staging file: %w", err)
		}
		if info.ModTime().Before(before) {
			if err := os.Remove(filepath.Join(directory, entry.Name())); err != nil && !errors.Is(err, os.ErrNotExist) {
				return removed, fmt.Errorf("remove staging file: %w", err)
			}
			removed++
		}
	}
	return removed, nil
}

func copyWithContext(ctx context.Context, target io.Writer, source io.Reader) (int64, error) {
	buffer := make([]byte, 32*1024)
	var written int64
	for {
		if err := ctx.Err(); err != nil {
			return written, err
		}
		count, readErr := source.Read(buffer)
		if count > 0 {
			output, writeErr := target.Write(buffer[:count])
			written += int64(output)
			if writeErr != nil {
				return written, writeErr
			}
			if output != count {
				return written, io.ErrShortWrite
			}
		}
		if errors.Is(readErr, io.EOF) {
			return written, nil
		}
		if readErr != nil {
			return written, readErr
		}
	}
}

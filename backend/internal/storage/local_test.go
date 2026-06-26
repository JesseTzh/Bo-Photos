package storage

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveRejectsUnsafeKeys(t *testing.T) {
	local := newTestLocal(t)

	for _, key := range []string{"", "/etc/passwd", "../escape", "originals/../../escape", "bad\x00key"} {
		if _, err := local.Resolve(key); !errors.Is(err, ErrUnsafeKey) {
			t.Errorf("Resolve(%q) error = %v, want ErrUnsafeKey", key, err)
		}
	}
}

func TestResolveRejectsSymlinkEscape(t *testing.T) {
	local := newTestLocal(t)
	outside := t.TempDir()
	link := filepath.Join(local.Root(), "originals", "link")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	if _, err := local.Resolve("originals/link/escape.jpg"); !errors.Is(err, ErrUnsafeKey) {
		t.Fatalf("Resolve(symlink escape) error = %v, want ErrUnsafeKey", err)
	}
}

func TestWriteStagingComputesHashAndEnforcesLimit(t *testing.T) {
	local := newTestLocal(t)
	result, err := local.WriteStaging(
		context.Background(),
		"asset-1",
		strings.NewReader("hello"),
		5,
	)
	if err != nil {
		t.Fatalf("WriteStaging() error = %v", err)
	}
	if result.Bytes != 5 {
		t.Errorf("Bytes = %d, want 5", result.Bytes)
	}
	if result.SHA256 != "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" {
		t.Errorf("SHA256 = %q", result.SHA256)
	}

	if _, err := local.WriteStaging(
		context.Background(),
		"asset-2",
		bytes.NewReader(make([]byte, 6)),
		5,
	); !errors.Is(err, ErrFileTooLarge) {
		t.Fatalf("oversized WriteStaging() error = %v, want ErrFileTooLarge", err)
	}
}

func TestPromoteAndPurgeAssetFiles(t *testing.T) {
	local := newTestLocal(t)
	staged, err := local.WriteStaging(context.Background(), "asset-1", strings.NewReader("original"), 100)
	if err != nil {
		t.Fatalf("WriteStaging() error = %v", err)
	}

	originalKey, err := local.PromoteOriginal(staged.Key, "asset-1", ".jpg")
	if err != nil {
		t.Fatalf("PromoteOriginal() error = %v", err)
	}
	previewKey := "previews/asset-1-v1.webp"
	thumbnailKey := "thumbnails/asset-1-v1.webp"
	writeKey(t, local, previewKey)
	writeKey(t, local, thumbnailKey)

	if err := local.Purge(originalKey, previewKey, thumbnailKey); err != nil {
		t.Fatalf("Purge() error = %v", err)
	}
	for _, key := range []string{originalKey, previewKey, thumbnailKey} {
		path, err := local.Resolve(key)
		if err != nil {
			t.Fatalf("Resolve(%q) error = %v", key, err)
		}
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("Stat(%q) error = %v, want not exist", key, err)
		}
	}
}

func newTestLocal(t *testing.T) *Local {
	t.Helper()
	local, err := NewLocal(t.TempDir())
	if err != nil {
		t.Fatalf("NewLocal() error = %v", err)
	}
	return local
}

func writeKey(t *testing.T, local *Local, key string) {
	t.Helper()
	path, err := local.Resolve(key)
	if err != nil {
		t.Fatalf("Resolve(%q) error = %v", key, err)
	}
	if err := os.WriteFile(path, []byte("derived"), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", key, err)
	}
}

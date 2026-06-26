package asset

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/besscroft/bophotos/backend/internal/imageproc"
	"github.com/besscroft/bophotos/backend/internal/storage"
)

func TestServiceUploadCreatesProcessingAssetAndDuplicateWarning(t *testing.T) {
	service, repo, _, queue := newTestAssetService(t, imageproc.Metadata{}, nil)
	content := append([]byte{0xff, 0xd8, 0xff, 0xe0}, []byte("same-image")...)

	first, err := service.Upload(context.Background(), "first.jpg", bytes.NewReader(content))
	if err != nil {
		t.Fatalf("first Upload() error = %v", err)
	}
	second, err := service.Upload(context.Background(), "second.jpg", bytes.NewReader(content))
	if err != nil {
		t.Fatalf("second Upload() error = %v", err)
	}
	if first.Asset.Status != StatusProcessing || second.Asset.Status != StatusProcessing {
		t.Fatalf("statuses = %q, %q; want processing", first.Asset.Status, second.Asset.Status)
	}
	if len(second.DuplicateAssetIDs) != 1 || second.DuplicateAssetIDs[0] != first.Asset.ID {
		t.Fatalf("duplicate warnings = %#v, want %q", second.DuplicateAssetIDs, first.Asset.ID)
	}
	if len(queue.ids) != 2 {
		t.Fatalf("queued IDs = %#v, want 2", queue.ids)
	}
	if _, err := repo.Get(context.Background(), second.Asset.ID); err != nil {
		t.Fatalf("Get(second) error = %v", err)
	}
}

func TestServiceProcessTransitionsToReady(t *testing.T) {
	metadata := imageproc.Metadata{Width: 4000, Height: 3000, Camera: "Fujifilm X-T5"}
	service, repo, local, _ := newTestAssetService(t, metadata, nil)
	upload, err := service.Upload(
		context.Background(),
		"photo.jpg",
		bytes.NewReader(append([]byte{0xff, 0xd8, 0xff, 0xe0}, []byte("image")...)),
	)
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}

	service.Process(context.Background(), upload.Asset.ID)

	item, err := repo.Get(context.Background(), upload.Asset.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if item.Status != StatusReady || item.Width != 4000 || item.Height != 3000 || item.Camera != metadata.Camera {
		t.Fatalf("processed item = %#v", item)
	}
	for _, key := range []string{item.PreviewKey, item.ThumbnailKey} {
		path, err := local.Resolve(key)
		if err != nil {
			t.Fatalf("Resolve(%q) error = %v", key, err)
		}
		if _, err := os.Stat(path); err != nil {
			t.Errorf("derived file %q missing: %v", key, err)
		}
	}
}

func TestServiceProcessFailureTransitionsToFailed(t *testing.T) {
	service, repo, _, _ := newTestAssetService(t, imageproc.Metadata{}, errors.New("decode failed"))
	upload, err := service.Upload(
		context.Background(),
		"broken.jpg",
		bytes.NewReader(append([]byte{0xff, 0xd8, 0xff, 0xe0}, []byte("broken")...)),
	)
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}

	service.Process(context.Background(), upload.Asset.ID)

	item, err := repo.Get(context.Background(), upload.Asset.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if item.Status != StatusFailed || item.ErrorCode != "ASSET_PROCESSING_FAILED" {
		t.Fatalf("processed item = %#v", item)
	}
}

func TestServiceProcessingFailureDoesNotPublishPartialDerivative(t *testing.T) {
	service, repo, local, _ := newTestAssetService(t, imageproc.Metadata{}, nil)
	service.processor = partialProcessor{}
	upload, err := service.Upload(
		context.Background(),
		"partial.jpg",
		bytes.NewReader(append([]byte{0xff, 0xd8, 0xff, 0xe0}, []byte("partial")...)),
	)
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}

	service.Process(context.Background(), upload.Asset.ID)

	item, err := repo.Get(context.Background(), upload.Asset.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	finalKey := "previews/" + item.ID + "-v1.webp"
	finalPath, err := local.Resolve(finalKey)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if _, err := os.Stat(finalPath); !os.IsNotExist(err) {
		t.Fatalf("partial derivative published at %q, error = %v", finalPath, err)
	}
}

func TestServiceRetryRequiresOriginal(t *testing.T) {
	service, repo, local, queue := newTestAssetService(t, imageproc.Metadata{}, nil)
	upload, err := service.Upload(
		context.Background(),
		"photo.jpg",
		bytes.NewReader(append([]byte{0xff, 0xd8, 0xff, 0xe0}, []byte("retry")...)),
	)
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if err := repo.Transition(context.Background(), upload.Asset.ID, StatusFailed, "FAILED"); err != nil {
		t.Fatalf("Transition() error = %v", err)
	}
	path, err := local.Resolve(upload.Asset.OriginalKey)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	if err := service.Retry(context.Background(), upload.Asset.ID); !errors.Is(err, ErrOriginalMissing) {
		t.Fatalf("Retry() error = %v, want ErrOriginalMissing", err)
	}
	if len(queue.ids) != 1 {
		t.Fatalf("queue IDs = %#v, retry must not enqueue", queue.ids)
	}
}

func TestServiceRetryIncrementsDerivativeVersion(t *testing.T) {
	service, repo, _, queue := newTestAssetService(t, imageproc.Metadata{}, nil)
	upload, err := service.Upload(
		context.Background(),
		"photo.jpg",
		bytes.NewReader(append([]byte{0xff, 0xd8, 0xff, 0xe0}, []byte("version")...)),
	)
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if err := repo.Transition(context.Background(), upload.Asset.ID, StatusFailed, "FAILED"); err != nil {
		t.Fatalf("Transition() error = %v", err)
	}
	if err := service.Retry(context.Background(), upload.Asset.ID); err != nil {
		t.Fatalf("Retry() error = %v", err)
	}
	item, err := repo.Get(context.Background(), upload.Asset.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if item.DerivativeVersion != 2 {
		t.Fatalf("DerivativeVersion = %d, want 2", item.DerivativeVersion)
	}
	if len(queue.ids) != 2 {
		t.Fatalf("queue IDs = %#v, want initial and retry", queue.ids)
	}
}

func TestServicePurgeDeletesFilesAndMarksPurged(t *testing.T) {
	service, repo, local, _ := newTestAssetService(t, imageproc.Metadata{}, nil)
	upload, err := service.Upload(
		context.Background(),
		"photo.jpg",
		bytes.NewReader(append([]byte{0xff, 0xd8, 0xff, 0xe0}, []byte("purge")...)),
	)
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	service.Process(context.Background(), upload.Asset.ID)
	item, err := repo.Get(context.Background(), upload.Asset.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if err := service.Delete(context.Background(), item.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if err := service.Purge(context.Background(), item.ID); err != nil {
		t.Fatalf("Purge() error = %v", err)
	}

	purged, err := repo.Get(context.Background(), item.ID)
	if err != nil {
		t.Fatalf("Get(purged) error = %v", err)
	}
	if purged.Status != StatusPurged || purged.PurgedAt == nil {
		t.Fatalf("purged item = %#v", purged)
	}
	for _, key := range []string{item.OriginalKey, item.PreviewKey, item.ThumbnailKey} {
		path, err := local.Resolve(key)
		if err != nil {
			t.Fatalf("Resolve(%q) error = %v", key, err)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("file %q still exists or stat error = %v", key, err)
		}
	}
}

func TestServicePurgeRequiresDeletedStatus(t *testing.T) {
	service, _, _, _ := newTestAssetService(t, imageproc.Metadata{}, nil)
	upload, err := service.Upload(
		context.Background(),
		"photo.jpg",
		bytes.NewReader(append([]byte{0xff, 0xd8, 0xff, 0xe0}, []byte("not-deleted")...)),
	)
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if err := service.Purge(context.Background(), upload.Asset.ID); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("Purge() error = %v, want ErrInvalidTransition", err)
	}
}

type fakeQueue struct {
	ids []string
}

func (q *fakeQueue) Enqueue(id string) error {
	q.ids = append(q.ids, id)
	return nil
}

type fakeProcessor struct {
	metadata imageproc.Metadata
	err      error
}

type partialProcessor struct{}

func (partialProcessor) Process(_ context.Context, request imageproc.Request) (imageproc.Metadata, error) {
	if err := os.WriteFile(request.PreviewPath, []byte("partial"), 0o644); err != nil {
		return imageproc.Metadata{}, err
	}
	return imageproc.Metadata{}, errors.New("thumbnail failed")
}

func (p fakeProcessor) Process(_ context.Context, request imageproc.Request) (imageproc.Metadata, error) {
	if p.err != nil {
		return imageproc.Metadata{}, p.err
	}
	for _, path := range []string{request.PreviewPath, request.ThumbnailPath} {
		if err := os.WriteFile(path, []byte("webp"), 0o644); err != nil {
			return imageproc.Metadata{}, err
		}
	}
	return p.metadata, nil
}

func newTestAssetService(
	t *testing.T,
	metadata imageproc.Metadata,
	processErr error,
) (*Service, *Repository, *storage.Local, *fakeQueue) {
	t.Helper()
	repo := newTestRepository(t)
	local, err := storage.NewLocal(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatalf("storage.NewLocal() error = %v", err)
	}
	queue := &fakeQueue{}
	service := NewService(repo, local, fakeProcessor{metadata: metadata, err: processErr}, queue, 10<<20)
	return service, repo, local, queue
}

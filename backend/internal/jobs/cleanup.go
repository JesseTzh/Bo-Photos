package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/besscroft/bophotos/backend/internal/asset"
	"github.com/besscroft/bophotos/backend/internal/storage"
)

type Cleanup struct {
	repository       *asset.Repository
	storage          *storage.Local
	stagingRetention time.Duration
	deletedRetention time.Duration
}

func NewCleanup(
	repository *asset.Repository,
	localStorage *storage.Local,
	stagingRetention time.Duration,
	deletedRetention time.Duration,
) *Cleanup {
	return &Cleanup{
		repository: repository, storage: localStorage,
		stagingRetention: stagingRetention, deletedRetention: deletedRetention,
	}
}

func (c *Cleanup) RunOnce(ctx context.Context) error {
	now := time.Now().UTC()
	if _, err := c.storage.CleanupStaging(now.Add(-c.stagingRetention)); err != nil {
		return err
	}
	if _, err := c.repository.FailStaleProcessing(ctx, now.Add(-time.Hour)); err != nil {
		return err
	}
	items, err := c.repository.DeletedBefore(ctx, now.Add(-c.deletedRetention))
	if err != nil {
		return err
	}
	for _, item := range items {
		if err := c.storage.Purge(item.OriginalKey, item.PreviewKey, item.ThumbnailKey); err != nil {
			return fmt.Errorf("purge asset %s: %w", item.ID, err)
		}
		if err := c.repository.Transition(ctx, item.ID, asset.StatusPurged, ""); err != nil {
			return fmt.Errorf("mark asset %s purged: %w", item.ID, err)
		}
	}
	return nil
}

func (c *Cleanup) Run(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = time.Hour
	}
	_ = c.RunOnce(ctx)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = c.RunOnce(ctx)
		}
	}
}

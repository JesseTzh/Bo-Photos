package guide

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"
)

type Service struct{ repo *Repository }

func NewService(repo *Repository) *Service { return &Service{repo: repo} }

func (s *Service) Create(ctx context.Context, input GuideInput) (Guide, error) {
	item, err := s.normalize(ctx, input)
	if err != nil {
		return Guide{}, err
	}
	item.ID = newID()
	item.CreatedAt, item.UpdatedAt = time.Now(), time.Now()
	if err := s.repo.Create(ctx, item); err != nil {
		return Guide{}, err
	}
	return s.repo.Get(ctx, item.ID, false)
}

func (s *Service) Update(ctx context.Context, id string, input GuideInput) (Guide, error) {
	item, err := s.normalize(ctx, input)
	if err != nil {
		return Guide{}, err
	}
	if err := s.repo.Update(ctx, id, item); err != nil {
		return Guide{}, err
	}
	return s.repo.Get(ctx, id, false)
}

func (s *Service) normalize(ctx context.Context, input GuideInput) (Guide, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return Guide{}, ErrInvalidName
	}
	if input.Days < 0 {
		return Guide{}, ErrInvalidDays
	}
	if input.StartDate != nil && input.EndDate != nil && input.EndDate.Before(*input.StartDate) {
		return Guide{}, ErrInvalidDates
	}
	cover := ""
	if input.CoverAssetID != nil {
		cover = strings.TrimSpace(*input.CoverAssetID)
	}
	available, err := s.repo.CoverAvailable(ctx, cover)
	if err != nil {
		return Guide{}, err
	}
	if !available {
		return Guide{}, ErrInvalidCover
	}
	return Guide{
		Title: title, Country: strings.TrimSpace(input.Country), City: strings.TrimSpace(input.City),
		Days: input.Days, StartDate: input.StartDate, EndDate: input.EndDate,
		CoverAssetID: cover, Published: input.Published, Sort: input.Sort,
	}, nil
}

func newID() string {
	value := make([]byte, 16)
	_, _ = rand.Read(value)
	return hex.EncodeToString(value)
}

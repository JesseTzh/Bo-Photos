package album

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"time"
)

var valuePattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Service struct{ repo *Repository }

func NewService(repo *Repository) *Service { return &Service{repo: repo} }

func (s *Service) Create(ctx context.Context, input Input) (Album, error) {
	item, err := normalize(input)
	if err != nil {
		return Album{}, err
	}
	item.ID = newID()
	item.CreatedAt, item.UpdatedAt = time.Now(), time.Now()
	if err := s.repo.Create(ctx, item); err != nil {
		return Album{}, err
	}
	return s.repo.Get(ctx, item.ID, false)
}

func (s *Service) Update(ctx context.Context, id string, input Input) (Album, error) {
	item, err := normalize(input)
	if err != nil {
		return Album{}, err
	}
	if err := s.repo.Update(ctx, id, item); err != nil {
		return Album{}, err
	}
	return s.repo.Get(ctx, id, false)
}

func normalize(input Input) (Album, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Album{}, ErrInvalidName
	}
	value := strings.Trim(strings.TrimSpace(input.Value), "/")
	if value == "" || !valuePattern.MatchString(value) {
		return Album{}, ErrInvalidValue
	}
	if input.Theme == "" {
		input.Theme = "0"
	}
	if input.Theme != "0" && input.Theme != "1" {
		return Album{}, errors.New("theme must be 0 or 1")
	}
	if input.ImageSorting == 0 {
		input.ImageSorting = 1
	}
	if input.ImageSorting < 1 || input.ImageSorting > 4 {
		return Album{}, errors.New("image_sorting must be between 1 and 4")
	}
	visible, random := true, false
	if input.Visible != nil {
		visible = *input.Visible
	}
	if input.RandomShow != nil {
		random = *input.RandomShow
	}
	return Album{Name: name, Value: value, Detail: input.Detail, Theme: input.Theme,
		Visible: visible, Sort: input.Sort, RandomShow: random, License: input.License,
		CoverAssetID: stringValue(input.CoverAssetID), ImageSorting: input.ImageSorting}, nil
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func newID() string {
	value := make([]byte, 16)
	_, _ = rand.Read(value)
	return hex.EncodeToString(value)
}

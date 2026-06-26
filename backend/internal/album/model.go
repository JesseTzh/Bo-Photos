package album

import (
	"errors"
	"time"
)

var (
	ErrNotFound       = errors.New("album not found")
	ErrInvalidName    = errors.New("album name is required")
	ErrInvalidValue   = errors.New("album value must contain only lowercase letters, numbers, and hyphens")
	ErrDuplicateValue = errors.New("album value already exists")
)

type Album struct {
	ID                    string
	Name                  string
	Value                 string
	Detail                string
	Theme                 string
	Visible               bool
	Sort                  int
	RandomShow            bool
	License               string
	CoverAssetID          string
	EffectiveCoverAssetID string
	ImageSorting          int
	AssetIDs              []string
	AssetCount            int
	CreatedAt             time.Time
	UpdatedAt             time.Time
	DeletedAt             *time.Time
}

type Input struct {
	Name         string  `json:"name"`
	Value        string  `json:"album_value"`
	Detail       string  `json:"detail"`
	Theme        string  `json:"theme"`
	Visible      *bool   `json:"visible"`
	Sort         int     `json:"sort"`
	RandomShow   *bool   `json:"random_show"`
	License      string  `json:"license"`
	CoverAssetID *string `json:"cover_asset_id"`
	ImageSorting int     `json:"image_sorting"`
}

package guide

import (
	"errors"
	"time"
)

var (
	ErrNotFound         = errors.New("guide not found")
	ErrInvalidName      = errors.New("guide title is required")
	ErrInvalidDays      = errors.New("guide days cannot be negative")
	ErrInvalidDates     = errors.New("guide end date cannot be before start date")
	ErrInvalidCover     = errors.New("guide cover asset is unavailable")
	ErrInvalidOwnership = errors.New("items do not belong to the requested parent")
)

type Guide struct {
	ID           string
	Title        string
	Country      string
	City         string
	Days         int
	StartDate    *time.Time
	EndDate      *time.Time
	CoverAssetID string
	Published    bool
	Sort         int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type GuideInput struct {
	Title        string     `json:"title"`
	Country      string     `json:"country"`
	City         string     `json:"city"`
	Days         int        `json:"days"`
	StartDate    *time.Time `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	CoverAssetID *string    `json:"cover_asset_id"`
	Published    bool       `json:"published"`
	Sort         int        `json:"sort"`
}

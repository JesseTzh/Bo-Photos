package asset

import (
	"encoding/json"
	"errors"
	"time"
)

type Status string

const (
	StatusProcessing Status = "processing"
	StatusReady      Status = "ready"
	StatusFailed     Status = "failed"
	StatusDeleted    Status = "deleted"
	StatusPurged     Status = "purged"
)

type SortDirection string

const (
	SortAscending  SortDirection = "asc"
	SortDescending SortDirection = "desc"
)

var (
	ErrInvalidTransition = errors.New("invalid asset status transition")
	ErrNotFound          = errors.New("asset not found")
	ErrOriginalMissing   = errors.New("asset original is missing")
)

type Asset struct {
	ID                string
	Status            Status
	OriginalName      string
	OriginalKey       string
	PreviewKey        string
	ThumbnailKey      string
	SHA256            string
	MIMEType          string
	ByteSize          int64
	Width             int
	Height            int
	Title             string
	Description       string
	Longitude         *float64
	Latitude          *float64
	BlurHash          string
	EXIFJSON          string
	ShootAt           *time.Time
	Camera            string
	Lens              string
	ExposureTime      string
	Aperture          string
	ISO               string
	FocalLength       string
	ErrorCode         string
	Visible           bool
	Private           bool
	ShowOnHomepage    bool
	Featured          bool
	Sort              int
	DerivativeVersion int
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
	PurgedAt          *time.Time
}

type PublicFilter struct {
	Cameras       []string
	Lenses        []string
	Tags          []string
	TagsOperator  string
	Album         string
	Featured      *bool
	ShootTimeSort SortDirection
	HomepageOnly  bool
	Page          int
	PageSize      int
}

type PrivateFilter struct {
	Page     int
	PageSize int
}

type AdminFilter struct {
	Status       Status
	Visible      *bool
	Private      *bool
	Featured     *bool
	Camera       string
	Lens         string
	ExposureTime string
	Aperture     string
	ISO          string
	Album        string
	TagIDs       []string
	TagsOperator string
	Title        string
	Page         int
	PageSize     int
}

type OptionalFloat struct {
	Set   bool
	Value *float64
}

func (value *OptionalFloat) UnmarshalJSON(data []byte) error {
	value.Set = true
	if string(data) == "null" {
		value.Value = nil
		return nil
	}
	var parsed float64
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}
	value.Value = &parsed
	return nil
}

type UpdateInput struct {
	Title          *string       `json:"title"`
	Description    *string       `json:"description"`
	Width          *int          `json:"width"`
	Height         *int          `json:"height"`
	Longitude      OptionalFloat `json:"longitude"`
	Latitude       OptionalFloat `json:"latitude"`
	ShootAt        *string       `json:"shoot_at"`
	Camera         *string       `json:"camera"`
	Lens           *string       `json:"lens"`
	ExposureTime   *string       `json:"exposure_time"`
	Aperture       *string       `json:"aperture"`
	ISO            *string       `json:"iso"`
	FocalLength    *string       `json:"focal_length"`
	EXIFJSON       *string       `json:"exif_json"`
	Visible        *bool         `json:"visible"`
	Private        *bool         `json:"private"`
	ShowOnHomepage *bool         `json:"show_on_homepage"`
	Featured       *bool         `json:"featured"`
	Sort           *int          `json:"sort"`
}

func CanTransition(from, to Status) bool {
	switch from {
	case StatusProcessing:
		return to == StatusReady || to == StatusFailed
	case StatusReady:
		return to == StatusProcessing || to == StatusDeleted
	case StatusFailed:
		return to == StatusProcessing
	case StatusDeleted:
		return to == StatusReady || to == StatusPurged
	default:
		return false
	}
}

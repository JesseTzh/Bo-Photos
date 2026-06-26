package tag

import (
	"errors"
	"time"
)

var (
	ErrNotFound    = errors.New("tag not found")
	ErrInvalidName = errors.New("tag name is required")
	ErrCycle       = errors.New("tag move would create a cycle")
	ErrHasChildren = errors.New("tag has children")
)

type Tag struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Category  string    `json:"category,omitempty"`
	ParentID  string    `json:"parent_id,omitempty"`
	Detail    string    `json:"detail,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Children  []Tag     `json:"children,omitempty"`
}

type CreateInput struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	ParentID string `json:"parent_id"`
	Detail   string `json:"detail"`
}

type UpdateInput struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Detail   string `json:"detail"`
}

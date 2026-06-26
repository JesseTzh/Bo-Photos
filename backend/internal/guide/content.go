package guide

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"strings"
)

type BlockType string
type ModuleTemplate string

const (
	BlockMarkdown BlockType = "markdown"
	BlockImage    BlockType = "image"
	BlockVideo    BlockType = "video"
	BlockLink     BlockType = "link"
	BlockTasks    BlockType = "tasks"
	BlockWarning  BlockType = "warning"
	BlockDivider  BlockType = "divider"
)

const (
	TemplateItinerary ModuleTemplate = "itinerary"
	TemplateExpense   ModuleTemplate = "expense"
	TemplateChecklist ModuleTemplate = "checklist"
	TemplateTransport ModuleTemplate = "transport"
	TemplatePhoto     ModuleTemplate = "photo"
	TemplateTips      ModuleTemplate = "tips"
)

var (
	ErrUnsupportedVersion = errors.New("unsupported content data version")
	ErrInvalidContent     = errors.New("invalid guide content")
)

type markdownData struct {
	Text string `json:"text"`
}
type imageData struct {
	AssetID string `json:"asset_id"`
	Caption string `json:"caption,omitempty"`
}
type videoData struct {
	URL     string `json:"url"`
	Title   string `json:"title,omitempty"`
	Caption string `json:"caption,omitempty"`
}
type linkData struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}
type taskItem struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	Completed bool   `json:"completed"`
}
type tasksData struct {
	Items []taskItem `json:"items"`
}
type warningData struct {
	Level string `json:"level"`
	Title string `json:"title"`
	Text  string `json:"text"`
}
type dividerData struct{}

func ValidateBlockData(blockType BlockType, version int, payload []byte) ([]byte, error) {
	if version != 1 {
		return nil, ErrUnsupportedVersion
	}
	var target any
	switch blockType {
	case BlockMarkdown:
		target = &markdownData{}
	case BlockImage:
		target = &imageData{}
	case BlockVideo:
		target = &videoData{}
	case BlockLink:
		target = &linkData{}
	case BlockTasks:
		target = &tasksData{}
	case BlockWarning:
		target = &warningData{}
	case BlockDivider:
		target = &dividerData{}
	default:
		return nil, ErrInvalidContent
	}
	if err := strictDecode(payload, target); err != nil {
		return nil, ErrInvalidContent
	}
	switch value := target.(type) {
	case *markdownData:
		value.Text = strings.TrimSpace(value.Text)
		if value.Text == "" {
			return nil, ErrInvalidContent
		}
	case *imageData:
		value.AssetID = strings.TrimSpace(value.AssetID)
		if value.AssetID == "" {
			return nil, ErrInvalidContent
		}
	case *videoData:
		if !validURL(value.URL) {
			return nil, ErrInvalidContent
		}
	case *linkData:
		value.Title = strings.TrimSpace(value.Title)
		if !validURL(value.URL) || value.Title == "" {
			return nil, ErrInvalidContent
		}
	case *tasksData:
		for _, item := range value.Items {
			if strings.TrimSpace(item.ID) == "" || strings.TrimSpace(item.Text) == "" {
				return nil, ErrInvalidContent
			}
		}
	case *warningData:
		if value.Level != "info" && value.Level != "warning" {
			return nil, ErrInvalidContent
		}
		if strings.TrimSpace(value.Title) == "" || strings.TrimSpace(value.Text) == "" {
			return nil, ErrInvalidContent
		}
	}
	normalized, err := json.Marshal(target)
	if err != nil {
		return nil, ErrInvalidContent
	}
	return normalized, nil
}

func strictDecode(payload []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return ErrInvalidContent
	}
	return nil
}

func validURL(value string) bool {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(value))
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https")
}

type itineraryItem struct {
	ID          string   `json:"id"`
	Date        string   `json:"date"`
	Title       string   `json:"title"`
	Location    string   `json:"location"`
	Description string   `json:"description"`
	Tips        string   `json:"tips"`
	Highlights  []string `json:"highlights"`
}
type expenseItem struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Category  string  `json:"category"`
	Notes     string  `json:"notes"`
	UnitPrice float64 `json:"unit_price"`
	Quantity  float64 `json:"quantity"`
	Subtotal  float64 `json:"subtotal"`
}
type checklistGroup struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Items []struct {
		ID        string `json:"id"`
		Text      string `json:"text"`
		Completed bool   `json:"completed"`
		Required  bool   `json:"required"`
	} `json:"items"`
}
type transportItem struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Route string `json:"route"`
	Date  string `json:"date"`
	Time  string `json:"time"`
}
type photoItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	FocalLength string `json:"focal_length"`
	BestTime    string `json:"best_time"`
	DronePolicy string `json:"drone_policy"`
	Notes       string `json:"notes"`
}
type tipItem struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Text  string `json:"text"`
	Type  string `json:"type"`
}

func ValidateStructuredData(template ModuleTemplate, version int, payload []byte) ([]byte, error) {
	if version != 1 {
		return nil, ErrUnsupportedVersion
	}
	var target any
	switch template {
	case TemplateItinerary:
		target = &[]itineraryItem{}
	case TemplateExpense:
		target = &[]expenseItem{}
	case TemplateChecklist:
		target = &[]checklistGroup{}
	case TemplateTransport:
		target = &[]transportItem{}
	case TemplatePhoto:
		target = &[]photoItem{}
	case TemplateTips:
		target = &[]tipItem{}
	default:
		return nil, ErrInvalidContent
	}
	if err := strictDecode(payload, target); err != nil {
		return nil, ErrInvalidContent
	}
	valid := func(value string) bool { return strings.TrimSpace(value) != "" }
	switch items := target.(type) {
	case *[]itineraryItem:
		for _, x := range *items {
			if !valid(x.ID) || !valid(x.Title) {
				return nil, ErrInvalidContent
			}
		}
	case *[]expenseItem:
		for _, x := range *items {
			if !valid(x.ID) || !valid(x.Name) || x.Quantity < 0 || x.UnitPrice < 0 || x.Subtotal < 0 {
				return nil, ErrInvalidContent
			}
		}
	case *[]checklistGroup:
		for _, x := range *items {
			if !valid(x.ID) || !valid(x.Name) {
				return nil, ErrInvalidContent
			}
			for _, v := range x.Items {
				if !valid(v.ID) || !valid(v.Text) {
					return nil, ErrInvalidContent
				}
			}
		}
	case *[]transportItem:
		for _, x := range *items {
			if !valid(x.ID) || (x.Type != "flight" && x.Type != "train" && x.Type != "car") {
				return nil, ErrInvalidContent
			}
		}
	case *[]photoItem:
		for _, x := range *items {
			if !valid(x.ID) || !valid(x.Name) || (x.DronePolicy != "allowed" && x.DronePolicy != "forbidden" && x.DronePolicy != "register") {
				return nil, ErrInvalidContent
			}
		}
	case *[]tipItem:
		allowed := map[string]bool{"warning": true, "info": true, "success": true, "weather": true, "emergency": true, "safety": true}
		for _, x := range *items {
			if !valid(x.ID) || !valid(x.Title) || !allowed[x.Type] {
				return nil, ErrInvalidContent
			}
		}
	}
	normalized, err := json.Marshal(target)
	if err != nil {
		return nil, ErrInvalidContent
	}
	if string(normalized) == "null" {
		return []byte("[]"), nil
	}
	return normalized, nil
}

package guide

import "testing"

func TestValidateBlockDataAcceptsSupportedTypes(t *testing.T) {
	tests := []struct {
		blockType BlockType
		payload   string
	}{
		{BlockMarkdown, `{"text":"hello"}`},
		{BlockImage, `{"asset_id":"asset-1","caption":"cover"}`},
		{BlockVideo, `{"url":"https://example.com/video.mp4"}`},
		{BlockLink, `{"url":"https://example.com","title":"Example"}`},
		{BlockTasks, `{"items":[{"id":"one","text":"Pack","completed":false}]}`},
		{BlockWarning, `{"level":"warning","title":"Weather","text":"Bring a coat"}`},
		{BlockDivider, `{}`},
	}
	for _, test := range tests {
		t.Run(string(test.blockType), func(t *testing.T) {
			if _, err := ValidateBlockData(test.blockType, 1, []byte(test.payload)); err != nil {
				t.Fatalf("ValidateBlockData() error = %v", err)
			}
		})
	}
}

func TestValidateBlockDataRejectsUnknownFieldsAndInvalidEnums(t *testing.T) {
	for _, test := range []struct {
		name, payload string
		blockType     BlockType
	}{
		{"unknown", `{"text":"hello","html":"<b>x</b>"}`, BlockMarkdown},
		{"warning-level", `{"level":"danger","title":"x","text":"y"}`, BlockWarning},
		{"missing-link-title", `{"url":"https://example.com"}`, BlockLink},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := ValidateBlockData(test.blockType, 1, []byte(test.payload)); err == nil {
				t.Fatal("ValidateBlockData() accepted invalid payload")
			}
		})
	}
	if _, err := ValidateBlockData(BlockMarkdown, 2, []byte(`{"text":"x"}`)); err != ErrUnsupportedVersion {
		t.Fatalf("version error = %v", err)
	}
}

func TestValidateStructuredDataSupportsSixTemplates(t *testing.T) {
	tests := map[ModuleTemplate]string{
		TemplateItinerary: `[{"id":"1","date":"2026-06-01","title":"Arrive","location":"Tokyo","description":"","tips":"","highlights":[]}]`,
		TemplateExpense:   `[{"id":"1","name":"Hotel","category":"accommodation","unit_price":100,"quantity":2,"subtotal":200,"notes":""}]`,
		TemplateChecklist: `[{"id":"1","name":"Gear","items":[{"id":"i","text":"Camera","completed":false,"required":true}]}]`,
		TemplateTransport: `[{"id":"1","type":"flight","route":"PVG-NRT","date":"2026-06-01","time":"08:00"}]`,
		TemplatePhoto:     `[{"id":"1","name":"Tower","focal_length":"35mm","best_time":"sunset","drone_policy":"forbidden","notes":""}]`,
		TemplateTips:      `[{"id":"1","title":"Weather","text":"Rain","type":"weather"}]`,
	}
	for template, payload := range tests {
		if _, err := ValidateStructuredData(template, 1, []byte(payload)); err != nil {
			t.Fatalf("%s error = %v", template, err)
		}
	}
}

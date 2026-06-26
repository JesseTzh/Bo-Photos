package imageproc

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestCommandProcessorBuildsDerivativesAndParsesExif(t *testing.T) {
	runner := &recordingRunner{
		exifOutput: []byte(`[{
			"ImageWidth": 6000,
			"ImageHeight": 4000,
			"DateTimeOriginal": "2025:06:22 10:11:12",
			"Model": "Fujifilm X-T5",
			"LensModel": "XF 23mm F1.4",
			"ExposureTime": "1/250",
			"FNumber": 2,
			"ISO": 200,
			"FocalLength": 23,
			"GPSLatitude": 31.2,
			"GPSLongitude": 121.5
		}]`),
	}
	processor := NewCommandProcessor(runner, Commands{
		VIPSThumbnail: "vipsthumbnail",
		ExifTool:      "exiftool",
	}, Limits{})
	directory := t.TempDir()
	request := Request{
		OriginalPath:  directory + "/photo.jpg",
		PreviewPath:   directory + "/photo-preview.webp",
		ThumbnailPath: directory + "/photo-thumbnail.webp",
	}

	metadata, err := processor.Process(context.Background(), request)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if metadata.Width != 6000 || metadata.Height != 4000 || metadata.Camera != "Fujifilm X-T5" {
		t.Fatalf("metadata = %#v", metadata)
	}
	if len(runner.calls) != 3 {
		t.Fatalf("calls = %#v, want 3", runner.calls)
	}
	if !strings.Contains(strings.Join(runner.calls[0].args, " "), request.PreviewPath) {
		t.Errorf("preview command = %#v", runner.calls[0])
	}
	if !strings.Contains(strings.Join(runner.calls[1].args, " "), request.ThumbnailPath) {
		t.Errorf("thumbnail command = %#v", runner.calls[1])
	}
}

type commandCall struct {
	name string
	args []string
}

type recordingRunner struct {
	calls      []commandCall
	exifOutput []byte
}

func (r *recordingRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	r.calls = append(r.calls, commandCall{name: name, args: args})
	if name == "exiftool" {
		return r.exifOutput, nil
	}
	for _, arg := range args {
		if strings.Contains(arg, ".webp") {
			path := strings.Split(arg, "[")[0]
			if err := os.WriteFile(path, []byte("webp"), 0o644); err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}

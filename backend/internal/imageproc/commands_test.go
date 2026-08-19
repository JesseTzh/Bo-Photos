package imageproc

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCommandProcessorBuildsDerivativesAndParsesExif(t *testing.T) {
	runner := &scriptedRunner{
		exifJSON: []byte(`[{
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
		Family:        FormatJPEG,
		Kind:          "jpeg",
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
	if strings.Contains(strings.Join(runner.calls[0].args, " "), request.OriginalPath) == false {
		t.Errorf("jpeg preview must use original path: %#v", runner.calls[0])
	}
}

func TestCommandProcessorRAWUsesLibRawThenThumbnails(t *testing.T) {
	runner := &scriptedRunner{
		librawOK: true,
		exifJSON: []byte(`[{"ImageWidth":6048,"ImageHeight":4024,"Model":"ILCE-6400"}]`),
	}
	processor := NewCommandProcessor(runner, Commands{}, Limits{})
	request := newRAWRequest(t)

	metadata, err := processor.Process(context.Background(), request)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if metadata.Camera != "ILCE-6400" {
		t.Fatalf("metadata = %#v", metadata)
	}
	if !runner.saw("simple_dcraw") {
		t.Fatal("expected simple_dcraw")
	}
	if runner.commandWithArg("vipsthumbnail", request.OriginalPath) {
		t.Fatal("vipsthumbnail must not receive the RAW original")
	}
	if !fileExists(request.PreviewPath) || !fileExists(request.ThumbnailPath) {
		t.Fatal("expected webp derivatives")
	}
}

func TestCommandProcessorRAWFallsBackToEmbeddedJPEG(t *testing.T) {
	runner := &scriptedRunner{
		librawErr:    exec.ErrNotFound,
		previewImage: []byte{0xff, 0xd8, 0xff, 0xdb, 1, 2, 3},
		exifJSON:     []byte(`[{"Model":"ILCE-6400"}]`),
		previewTagOK: "PreviewImage",
	}
	processor := NewCommandProcessor(runner, Commands{}, Limits{})
	request := newRAWRequest(t)

	if _, err := processor.Process(context.Background(), request); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if !runner.sawArg("-PreviewImage") {
		t.Fatalf("expected PreviewImage extract, calls=%#v", runner.calls)
	}
	if runner.sawArg(request.OriginalPath) && runner.commandWithArg("vipsthumbnail", request.OriginalPath) {
		t.Fatal("vipsthumbnail must not receive the RAW original")
	}
}

func TestCommandProcessorRAWTriesJpgFromRawAfterEmptyPreview(t *testing.T) {
	runner := &scriptedRunner{
		librawErr:    errors.New("decode failed"),
		previewImage: []byte{0xff, 0xd8, 0xff, 0xe0},
		previewTagOK: "JpgFromRaw",
		exifJSON:     []byte(`[{}]`),
	}
	processor := NewCommandProcessor(runner, Commands{}, Limits{})
	request := newRAWRequest(t)

	if _, err := processor.Process(context.Background(), request); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if !runner.sawArg("-PreviewImage") || !runner.sawArg("-JpgFromRaw") {
		t.Fatalf("expected PreviewImage then JpgFromRaw, calls=%#v", runner.calls)
	}
}

func TestCommandProcessorRAWFailsWhenBothPreviewPathsFail(t *testing.T) {
	runner := &scriptedRunner{librawErr: exec.ErrNotFound}
	processor := NewCommandProcessor(runner, Commands{}, Limits{})
	request := newRAWRequest(t)

	if _, err := processor.Process(context.Background(), request); err == nil {
		t.Fatal("Process() error = nil, want preview failure")
	}
}

func TestCommandProcessorEXIFFailureStillSucceedsAfterPreview(t *testing.T) {
	runner := &scriptedRunner{exifErr: errors.New("exiftool crashed")}
	processor := NewCommandProcessor(runner, Commands{}, Limits{})
	directory := t.TempDir()
	request := Request{
		OriginalPath:  directory + "/photo.jpg",
		PreviewPath:   directory + "/photo-preview.webp",
		ThumbnailPath: directory + "/photo-thumbnail.webp",
		Family:        FormatJPEG,
	}

	metadata, err := processor.Process(context.Background(), request)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if metadata != (Metadata{}) {
		t.Fatalf("metadata = %#v, want empty after EXIF failure", metadata)
	}
	if !fileExists(request.PreviewPath) || !fileExists(request.ThumbnailPath) {
		t.Fatal("preview must remain after EXIF failure")
	}
}

func TestParseExifSonyARWFixtureValues(t *testing.T) {
	metadata, err := parseExif([]byte(`[{
		"ImageWidth": 6048,
		"ImageHeight": 4024,
		"DateTimeOriginal": "2026:05:04 06:40:36",
		"Model": "ILCE-6400"
	}]`))
	if err != nil {
		t.Fatalf("parseExif() error = %v", err)
	}
	if metadata.Camera != "ILCE-6400" {
		t.Fatalf("Camera = %q", metadata.Camera)
	}
	if metadata.ShootAt == nil || !metadata.ShootAt.Equal(time.Date(2026, 5, 4, 6, 40, 36, 0, time.UTC)) {
		t.Fatalf("ShootAt = %v", metadata.ShootAt)
	}
}

func newRAWRequest(t *testing.T) Request {
	t.Helper()
	directory := t.TempDir()
	original := filepath.Join(directory, "photo.arw")
	if err := os.WriteFile(original, []byte("raw-bytes"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return Request{
		AssetID:       "asset-1",
		OriginalPath:  original,
		PreviewPath:   filepath.Join(directory, "preview.webp"),
		ThumbnailPath: filepath.Join(directory, "thumb.webp"),
		Family:        FormatRAW,
		Kind:          "arw",
	}
}

type commandCall struct {
	name string
	args []string
}

type scriptedRunner struct {
	calls        []commandCall
	exifJSON     []byte
	exifErr      error
	librawOK     bool
	librawErr    error
	previewImage []byte
	previewTagOK string
}

func (r *scriptedRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	r.calls = append(r.calls, commandCall{name: name, args: args})
	switch {
	case name == "simple_dcraw" || name == "dcraw_emu":
		if r.librawErr != nil {
			return nil, r.librawErr
		}
		if !r.librawOK {
			return nil, exec.ErrNotFound
		}
		input := args[len(args)-1]
		if err := os.WriteFile(input+".tiff", []byte("decoded"), 0o600); err != nil {
			return nil, err
		}
		return nil, nil
	case name == "exiftool" && containsArg(args, "-b"):
		tag := strings.TrimPrefix(lastDashArg(args), "-")
		if r.previewTagOK != "" && tag == r.previewTagOK {
			return r.previewImage, nil
		}
		return nil, nil
	case name == "exiftool":
		if r.exifErr != nil {
			return nil, r.exifErr
		}
		return r.exifJSON, nil
	case name == "vipsthumbnail":
		for _, arg := range args {
			if strings.Contains(arg, ".webp") {
				path := strings.Split(arg, "[")[0]
				if err := os.WriteFile(path, []byte("webp"), 0o644); err != nil {
					return nil, err
				}
			}
		}
		return nil, nil
	default:
		return nil, exec.ErrNotFound
	}
}

func (r *scriptedRunner) saw(name string) bool {
	for _, call := range r.calls {
		if call.name == name {
			return true
		}
	}
	return false
}

func (r *scriptedRunner) sawArg(part string) bool {
	for _, call := range r.calls {
		for _, arg := range call.args {
			if strings.Contains(arg, part) {
				return true
			}
		}
	}
	return false
}

func (r *scriptedRunner) commandWithArg(name, part string) bool {
	for _, call := range r.calls {
		if call.name != name {
			continue
		}
		for _, arg := range call.args {
			if arg == part {
				return true
			}
		}
	}
	return false
}

func containsArg(args []string, part string) bool {
	for _, arg := range args {
		if arg == part {
			return true
		}
	}
	return false
}

func lastDashArg(args []string) string {
	for i := len(args) - 1; i >= 0; i-- {
		if strings.HasPrefix(args[i], "-") && args[i] != "-b" && args[i] != "-json" && args[i] != "-n" {
			return args[i]
		}
	}
	return ""
}

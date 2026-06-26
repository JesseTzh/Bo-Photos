package imageproc

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"time"
)

type Runner interface {
	Run(context.Context, string, ...string) ([]byte, error)
}

type Commands struct {
	VIPSThumbnail string
	ExifTool      string
}

type Limits struct {
	PreviewMaxWidth int
	ThumbnailSize   int
	PreviewQuality  int
	Timeout         time.Duration
}

type CommandProcessor struct {
	runner   Runner
	commands Commands
	limits   Limits
}

func NewCommandProcessor(runner Runner, commands Commands, limits Limits) *CommandProcessor {
	if runner == nil {
		runner = ExecRunner{}
	}
	if commands.VIPSThumbnail == "" {
		commands.VIPSThumbnail = "vipsthumbnail"
	}
	if commands.ExifTool == "" {
		commands.ExifTool = "exiftool"
	}
	if limits.PreviewMaxWidth == 0 {
		limits.PreviewMaxWidth = 2560
	}
	if limits.ThumbnailSize == 0 {
		limits.ThumbnailSize = 480
	}
	if limits.PreviewQuality == 0 {
		limits.PreviewQuality = 85
	}
	if limits.Timeout == 0 {
		limits.Timeout = 2 * time.Minute
	}
	return &CommandProcessor{runner: runner, commands: commands, limits: limits}
}

func (p *CommandProcessor) Process(ctx context.Context, request Request) (Metadata, error) {
	ctx, cancel := context.WithTimeout(ctx, p.limits.Timeout)
	defer cancel()

	previewOutput := fmt.Sprintf("%s[Q=%d,strip]", request.PreviewPath, p.limits.PreviewQuality)
	if _, err := p.runner.Run(
		ctx,
		p.commands.VIPSThumbnail,
		request.OriginalPath,
		"--size",
		fmt.Sprintf("%dx%d>", p.limits.PreviewMaxWidth, p.limits.PreviewMaxWidth),
		"--rotate",
		"--export-profile",
		"srgb",
		"--output",
		previewOutput,
	); err != nil {
		return Metadata{}, fmt.Errorf("generate preview: %w", err)
	}

	thumbnailOutput := request.ThumbnailPath + "[Q=80,strip]"
	if _, err := p.runner.Run(
		ctx,
		p.commands.VIPSThumbnail,
		request.OriginalPath,
		"--size",
		fmt.Sprintf("%dx%d>", p.limits.ThumbnailSize, p.limits.ThumbnailSize),
		"--rotate",
		"--export-profile",
		"srgb",
		"--output",
		thumbnailOutput,
	); err != nil {
		return Metadata{}, fmt.Errorf("generate thumbnail: %w", err)
	}

	output, err := p.runner.Run(ctx, p.commands.ExifTool, "-json", "-n", request.OriginalPath)
	if err != nil {
		return Metadata{}, fmt.Errorf("extract EXIF: %w", err)
	}
	return parseExif(output)
}

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

type exifRecord struct {
	ImageWidth       int      `json:"ImageWidth"`
	ImageHeight      int      `json:"ImageHeight"`
	DateTimeOriginal string   `json:"DateTimeOriginal"`
	Model            string   `json:"Model"`
	LensModel        string   `json:"LensModel"`
	ExposureTime     any      `json:"ExposureTime"`
	FNumber          any      `json:"FNumber"`
	ISO              any      `json:"ISO"`
	FocalLength      any      `json:"FocalLength"`
	GPSLatitude      *float64 `json:"GPSLatitude"`
	GPSLongitude     *float64 `json:"GPSLongitude"`
}

func parseExif(data []byte) (Metadata, error) {
	var records []exifRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return Metadata{}, fmt.Errorf("decode EXIF JSON: %w", err)
	}
	if len(records) == 0 {
		return Metadata{EXIFJSON: string(data)}, nil
	}
	record := records[0]
	metadata := Metadata{
		Width:        record.ImageWidth,
		Height:       record.ImageHeight,
		EXIFJSON:     string(data),
		Camera:       record.Model,
		Lens:         record.LensModel,
		ExposureTime: valueString(record.ExposureTime),
		Aperture:     valueString(record.FNumber),
		ISO:          valueString(record.ISO),
		FocalLength:  valueString(record.FocalLength),
		Latitude:     record.GPSLatitude,
		Longitude:    record.GPSLongitude,
	}
	if record.DateTimeOriginal != "" {
		if value, err := time.Parse("2006:01:02 15:04:05", record.DateTimeOriginal); err == nil {
			utc := value.UTC()
			metadata.ShootAt = &utc
		}
	}
	return metadata, nil
}

func valueString(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		return fmt.Sprint(typed)
	}
}

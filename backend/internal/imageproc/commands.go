package imageproc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Runner interface {
	Run(context.Context, string, ...string) ([]byte, error)
}

type Commands struct {
	VIPSThumbnail string
	ExifTool      string
	LibRaw        []string
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
	if len(commands.LibRaw) == 0 {
		commands.LibRaw = []string{"simple_dcraw", "dcraw_emu"}
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
	started := time.Now()
	ctx, cancel := context.WithTimeout(ctx, p.limits.Timeout)
	defer cancel()

	sourcePath := request.OriginalPath
	previewSource := "original"
	if request.Family == FormatRAW {
		decoded, method, cleanup, err := p.decodeRAW(ctx, request.OriginalPath)
		if cleanup != nil {
			defer cleanup()
		}
		if err != nil {
			slog.Error("asset preview failed",
				"asset_id", request.AssetID,
				"kind", request.Kind,
				"error", err,
				"duration_ms", time.Since(started).Milliseconds(),
			)
			return Metadata{}, fmt.Errorf("generate preview: %w", err)
		}
		sourcePath = decoded
		previewSource = method
	}

	if err := p.generateDerivatives(ctx, sourcePath, request); err != nil {
		slog.Error("asset preview failed",
			"asset_id", request.AssetID,
			"kind", request.Kind,
			"preview_path", previewSource,
			"error", err,
			"duration_ms", time.Since(started).Milliseconds(),
		)
		return Metadata{}, err
	}

	slog.Info("asset preview generated",
		"asset_id", request.AssetID,
		"kind", request.Kind,
		"preview_path", previewSource,
		"duration_ms", time.Since(started).Milliseconds(),
	)

	output, err := p.runner.Run(ctx, p.commands.ExifTool, "-json", "-n", request.OriginalPath)
	if err != nil {
		slog.Warn("asset EXIF extraction failed",
			"asset_id", request.AssetID,
			"kind", request.Kind,
			"error", err,
		)
		return Metadata{}, nil
	}
	metadata, err := parseExif(output)
	if err != nil {
		slog.Warn("asset EXIF extraction failed",
			"asset_id", request.AssetID,
			"kind", request.Kind,
			"error", err,
		)
		return Metadata{}, nil
	}
	return metadata, nil
}

func (p *CommandProcessor) generateDerivatives(ctx context.Context, source string, request Request) error {
	previewOutput := fmt.Sprintf("%s[Q=%d,strip]", request.PreviewPath, p.limits.PreviewQuality)
	if _, err := p.runner.Run(
		ctx,
		p.commands.VIPSThumbnail,
		source,
		"--size",
		fmt.Sprintf("%dx%d>", p.limits.PreviewMaxWidth, p.limits.PreviewMaxWidth),
		"--rotate",
		"--export-profile",
		"srgb",
		"--output",
		previewOutput,
	); err != nil {
		return fmt.Errorf("generate preview: %w", err)
	}

	thumbnailOutput := request.ThumbnailPath + "[Q=80,strip]"
	if _, err := p.runner.Run(
		ctx,
		p.commands.VIPSThumbnail,
		source,
		"--size",
		fmt.Sprintf("%dx%d>", p.limits.ThumbnailSize, p.limits.ThumbnailSize),
		"--rotate",
		"--export-profile",
		"srgb",
		"--output",
		thumbnailOutput,
	); err != nil {
		return fmt.Errorf("generate thumbnail: %w", err)
	}
	return nil
}

func (p *CommandProcessor) decodeRAW(ctx context.Context, originalPath string) (string, string, func(), error) {
	workDir, err := os.MkdirTemp("", "bophotos-raw-*")
	if err != nil {
		return "", "", nil, fmt.Errorf("create raw work dir: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(workDir) }

	workFile := filepath.Join(workDir, filepath.Base(originalPath))
	if err := copyFile(originalPath, workFile); err != nil {
		cleanup()
		return "", "", nil, fmt.Errorf("stage raw original: %w", err)
	}

	for _, name := range p.commands.LibRaw {
		if _, err := p.runner.Run(ctx, name, "-T", workFile); isNotFound(err) {
			continue
		} else if err != nil {
			continue
		}
		if decoded := firstExistingFile(librawOutputs(workFile)); decoded != "" {
			return decoded, "libraw", cleanup, nil
		}
	}

	for _, tag := range []string{"PreviewImage", "JpgFromRaw", "OtherImage"} {
		data, err := p.runner.Run(ctx, p.commands.ExifTool, "-b", "-"+tag, originalPath)
		if err != nil || !isJPEG(data) {
			continue
		}
		jpegPath := filepath.Join(workDir, "embedded.jpg")
		if err := os.WriteFile(jpegPath, data, 0o600); err != nil {
			continue
		}
		return jpegPath, "embedded_jpeg", cleanup, nil
	}

	cleanup()
	return "", "", nil, errors.New("libraw and embedded jpeg preview failed")
}

func librawOutputs(input string) []string {
	ext := filepath.Ext(input)
	base := strings.TrimSuffix(input, ext)
	return []string{
		input + ".tiff",
		input + ".tif",
		base + ".tiff",
		base + ".tif",
		input + ".ppm",
		base + ".ppm",
	}
}

func firstExistingFile(paths []string) string {
	for _, path := range paths {
		info, err := os.Stat(path)
		if err == nil && info.Size() > 0 {
			return path
		}
	}
	return ""
}

func copyFile(source, target string) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		_ = os.Remove(target)
		return err
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(target)
		return err
	}
	return nil
}

func isJPEG(data []byte) bool {
	return len(data) >= 3 && data[0] == 0xff && data[1] == 0xd8 && data[2] == 0xff
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, exec.ErrNotFound) {
		return true
	}
	var execErr *exec.Error
	return errors.As(err, &execErr) && errors.Is(execErr.Err, exec.ErrNotFound)
}

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return stdout.Bytes(), fmt.Errorf("%s: %w: %s", name, err, strings.TrimSpace(stderr.String()))
		}
		return stdout.Bytes(), fmt.Errorf("%s: %w", name, err)
	}
	return stdout.Bytes(), nil
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

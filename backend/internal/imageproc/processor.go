package imageproc

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"time"
)

var ErrUnsupportedFormat = errors.New("unsupported image format")

type Format string

const (
	FormatJPEG Format = "jpeg"
	FormatPNG  Format = "png"
	FormatWebP Format = "webp"
	FormatHEIF Format = "heif"
	FormatTIFF Format = "tiff"
	FormatRAW  Format = "raw"
)

type Metadata struct {
	Width        int
	Height       int
	EXIFJSON     string
	ShootAt      *time.Time
	Camera       string
	Lens         string
	ExposureTime string
	Aperture     string
	ISO          string
	FocalLength  string
	Longitude    *float64
	Latitude     *float64
}

type Request struct {
	OriginalPath  string
	PreviewPath   string
	ThumbnailPath string
}

type Processor interface {
	Process(context.Context, Request) (Metadata, error)
}

func DetectFormat(reader io.Reader) (Format, error) {
	buffered := bufio.NewReader(reader)
	header, _ := buffered.Peek(32)
	switch {
	case len(header) >= 3 && bytes.Equal(header[:3], []byte{0xff, 0xd8, 0xff}):
		return FormatJPEG, nil
	case len(header) >= 8 && bytes.Equal(header[:8], []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}):
		return FormatPNG, nil
	case len(header) >= 12 && string(header[:4]) == "RIFF" && string(header[8:12]) == "WEBP":
		return FormatWebP, nil
	case len(header) >= 4 && (bytes.Equal(header[:4], []byte{'I', 'I', 42, 0}) || bytes.Equal(header[:4], []byte{'M', 'M', 0, 42})):
		return FormatTIFF, nil
	case len(header) >= 12 && string(header[4:8]) == "ftyp" && isHEIFBrand(string(header[8:12])):
		return FormatHEIF, nil
	case strings.HasPrefix(string(header), "FUJIFILMCCD-RAW"):
		return FormatRAW, nil
	case len(header) >= 12 && string(header[4:8]) == "ftyp" && strings.EqualFold(string(header[8:12]), "crx "):
		return FormatRAW, nil
	case len(header) >= 4 && (bytes.Equal(header[:4], []byte{'I', 'I', 'R', 'O'}) || bytes.Equal(header[:4], []byte{'M', 'M', 'O', 'R'})):
		return FormatRAW, nil
	case len(header) >= 4 && bytes.Equal(header[:4], []byte{'I', 'I', 'U', 0}):
		return FormatRAW, nil
	default:
		return "", ErrUnsupportedFormat
	}
}

func isHEIFBrand(brand string) bool {
	switch strings.ToLower(brand) {
	case "heic", "heix", "hevc", "hevx", "heim", "heis", "mif1", "msf1":
		return true
	default:
		return false
	}
}

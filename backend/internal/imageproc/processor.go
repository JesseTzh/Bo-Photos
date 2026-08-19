package imageproc

import (
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

type Detection struct {
	Family    Format
	Kind      string
	Extension string
	MIMEType  string
}

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
	AssetID       string
	OriginalPath  string
	PreviewPath   string
	ThumbnailPath string
	Family        Format
	Kind          string
}

type Processor interface {
	Process(context.Context, Request) (Metadata, error)
}

func DetectMedia(r io.ReadSeeker) (Detection, error) {
	header := make([]byte, 32)
	n, err := io.ReadFull(r, header)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return Detection{}, err
	}
	header = header[:n]

	switch {
	case len(header) >= 3 && bytes.Equal(header[:3], []byte{0xff, 0xd8, 0xff}):
		return Detection{Family: FormatJPEG, Kind: "jpeg", Extension: ".jpg", MIMEType: "image/jpeg"}, nil
	case len(header) >= 8 && bytes.Equal(header[:8], []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}):
		return Detection{Family: FormatPNG, Kind: "png", Extension: ".png", MIMEType: "image/png"}, nil
	case len(header) >= 12 && string(header[:4]) == "RIFF" && string(header[8:12]) == "WEBP":
		return Detection{Family: FormatWebP, Kind: "webp", Extension: ".webp", MIMEType: "image/webp"}, nil
	case strings.HasPrefix(string(header), "FUJIFILMCCD-RAW"):
		return Detection{Family: FormatRAW, Kind: "raf", Extension: ".raf", MIMEType: "image/x-fuji-raf"}, nil
	case len(header) >= 12 && string(header[4:8]) == "ftyp" && isHEIFBrand(string(header[8:12])):
		return Detection{Family: FormatHEIF, Kind: "heif", Extension: ".heic", MIMEType: "image/heif"}, nil
	case len(header) >= 12 && string(header[4:8]) == "ftyp" && strings.EqualFold(string(header[8:12]), "crx "):
		return Detection{Family: FormatRAW, Kind: "cr3", Extension: ".cr3", MIMEType: "image/x-canon-cr3"}, nil
	case len(header) >= 4 && (bytes.Equal(header[:4], []byte{'I', 'I', 'R', 'O'}) || bytes.Equal(header[:4], []byte{'M', 'M', 'O', 'R'})):
		return Detection{Family: FormatRAW, Kind: "orf", Extension: ".orf", MIMEType: "image/x-olympus-orf"}, nil
	case len(header) >= 4 && bytes.Equal(header[:4], []byte{'I', 'I', 'U', 0}):
		return Detection{Family: FormatRAW, Kind: "rw2", Extension: ".rw2", MIMEType: "image/x-panasonic-rw2"}, nil
	case isTIFFMagic(header):
		if _, err := r.Seek(0, io.SeekStart); err != nil {
			return Detection{}, err
		}
		return detectTIFF(r), nil
	default:
		return Detection{}, ErrUnsupportedFormat
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

func RawKindFromExtension(ext string) string {
	switch strings.ToLower(ext) {
	case ".arw":
		return "arw"
	case ".nef":
		return "nef"
	case ".cr2":
		return "cr2"
	case ".cr3":
		return "cr3"
	case ".raf":
		return "raf"
	case ".dng":
		return "dng"
	case ".orf":
		return "orf"
	case ".rw2":
		return "rw2"
	case ".pef":
		return "pef"
	default:
		return ""
	}
}

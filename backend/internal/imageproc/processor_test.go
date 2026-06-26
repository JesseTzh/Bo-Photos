package imageproc

import (
	"bytes"
	"errors"
	"testing"
)

func TestDetectFormatUsesFileSignature(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want Format
	}{
		{name: "jpeg", data: []byte{0xff, 0xd8, 0xff, 0xe0}, want: FormatJPEG},
		{name: "png", data: []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}, want: FormatPNG},
		{name: "webp", data: append([]byte("RIFF1234WEBP"), 0), want: FormatWebP},
		{name: "tiff little endian", data: []byte{'I', 'I', 42, 0}, want: FormatTIFF},
		{name: "heif", data: append([]byte{0, 0, 0, 24}, []byte("ftypheic")...), want: FormatHEIF},
		{name: "fujifilm raf", data: []byte("FUJIFILMCCD-RAW 0201"), want: FormatRAW},
		{name: "canon cr3", data: append([]byte{0, 0, 0, 24}, []byte("ftypcrx ")...), want: FormatRAW},
		{name: "olympus orf", data: []byte{'I', 'I', 'R', 'O'}, want: FormatRAW},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := DetectFormat(bytes.NewReader(test.data))
			if err != nil {
				t.Fatalf("DetectFormat() error = %v", err)
			}
			if got != test.want {
				t.Errorf("DetectFormat() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestDetectFormatRejectsUnknownContent(t *testing.T) {
	if _, err := DetectFormat(bytes.NewReader([]byte("not an image"))); !errors.Is(err, ErrUnsupportedFormat) {
		t.Fatalf("DetectFormat() error = %v, want ErrUnsupportedFormat", err)
	}
}

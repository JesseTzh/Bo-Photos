package imageproc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDetectMediaUsesFileSignature(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want Detection
	}{
		{
			name: "jpeg",
			data: []byte{0xff, 0xd8, 0xff, 0xe0},
			want: Detection{Family: FormatJPEG, Kind: "jpeg", Extension: ".jpg", MIMEType: "image/jpeg"},
		},
		{
			name: "png",
			data: []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'},
			want: Detection{Family: FormatPNG, Kind: "png", Extension: ".png", MIMEType: "image/png"},
		},
		{
			name: "webp",
			data: append([]byte("RIFF1234WEBP"), 0),
			want: Detection{Family: FormatWebP, Kind: "webp", Extension: ".webp", MIMEType: "image/webp"},
		},
		{
			name: "heif",
			data: append([]byte{0, 0, 0, 24}, []byte("ftypheic")...),
			want: Detection{Family: FormatHEIF, Kind: "heif", Extension: ".heic", MIMEType: "image/heif"},
		},
		{
			name: "fujifilm raf",
			data: []byte("FUJIFILMCCD-RAW 0201"),
			want: Detection{Family: FormatRAW, Kind: "raf", Extension: ".raf", MIMEType: "image/x-fuji-raf"},
		},
		{
			name: "canon cr3",
			data: append([]byte{0, 0, 0, 24}, []byte("ftypcrx ")...),
			want: Detection{Family: FormatRAW, Kind: "cr3", Extension: ".cr3", MIMEType: "image/x-canon-cr3"},
		},
		{
			name: "olympus orf",
			data: []byte{'I', 'I', 'R', 'O'},
			want: Detection{Family: FormatRAW, Kind: "orf", Extension: ".orf", MIMEType: "image/x-olympus-orf"},
		},
		{
			name: "panasonic rw2",
			data: []byte{'I', 'I', 'U', 0},
			want: Detection{Family: FormatRAW, Kind: "rw2", Extension: ".rw2", MIMEType: "image/x-panasonic-rw2"},
		},
		{
			name: "tiff little endian without IFD",
			data: []byte{'I', 'I', 42, 0},
			want: Detection{Family: FormatTIFF, Kind: "tiff", Extension: ".tiff", MIMEType: "image/tiff"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := DetectMedia(bytes.NewReader(test.data))
			if err != nil {
				t.Fatalf("DetectMedia() error = %v", err)
			}
			if got != test.want {
				t.Errorf("DetectMedia() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestDetectMediaRejectsUnknownContent(t *testing.T) {
	if _, err := DetectMedia(bytes.NewReader([]byte("not an image"))); !errors.Is(err, ErrUnsupportedFormat) {
		t.Fatalf("DetectMedia() error = %v, want ErrUnsupportedFormat", err)
	}
}

func TestDetectMediaJPEGRenamedARWIsStillJPEG(t *testing.T) {
	got, err := DetectMedia(bytes.NewReader([]byte{0xff, 0xd8, 0xff, 0xe1, 'J', 'F', 'I', 'F'}))
	if err != nil {
		t.Fatalf("DetectMedia() error = %v", err)
	}
	if got.Family != FormatJPEG || got.Kind != "jpeg" {
		t.Fatalf("DetectMedia() = %#v, want jpeg", got)
	}
}

func TestDetectMediaIdentifiesFixtureARW(t *testing.T) {
	file, err := os.Open(RawARWPath(t))
	if err != nil {
		t.Fatalf("Open(test.ARW) error = %v", err)
	}
	t.Cleanup(func() { _ = file.Close() })

	got, err := DetectMedia(file)
	if err != nil {
		t.Fatalf("DetectMedia(test.ARW) error = %v", err)
	}
	want := Detection{Family: FormatRAW, Kind: "arw", Extension: ".arw", MIMEType: "image/x-sony-arw"}
	if got != want {
		t.Fatalf("DetectMedia(test.ARW) = %#v, want %#v", got, want)
	}
	if got.Family == FormatTIFF || got.Kind == "tiff" {
		t.Fatal("DetectMedia(test.ARW) must not report tiff")
	}
}

func TestDetectMediaClassifiesTIFFContainerRAW(t *testing.T) {
	sony := uint16(compressionSonyARW)
	nikon := uint16(compressionNikonNEF)
	oldJPEG := uint16(6)
	uncompressed := uint16(1)
	lzw := uint16(5)
	deflate := uint16(8)

	tests := []struct {
		name string
		data []byte
		want Detection
	}{
		{
			name: "sony arw",
			data: buildMiniTIFF(miniTIFF{Make: "SONY", Compression: &sony}),
			want: Detection{Family: FormatRAW, Kind: "arw", Extension: ".arw", MIMEType: "image/x-sony-arw"},
		},
		{
			name: "nikon nef",
			data: buildMiniTIFF(miniTIFF{Make: "NIKON CORPORATION", Compression: &nikon}),
			want: Detection{Family: FormatRAW, Kind: "nef", Extension: ".nef", MIMEType: "image/x-nikon-nef"},
		},
		{
			name: "canon cr2",
			data: buildMiniTIFF(miniTIFF{Make: "Canon", Compression: &oldJPEG}),
			want: Detection{Family: FormatRAW, Kind: "cr2", Extension: ".cr2", MIMEType: "image/x-canon-cr2"},
		},
		{
			name: "adobe dng",
			data: buildMiniTIFF(miniTIFF{Make: "Adobe", Compression: &uncompressed, DNGVersion: []byte{1, 4, 0, 0}}),
			want: Detection{Family: FormatRAW, Kind: "dng", Extension: ".dng", MIMEType: "image/x-adobe-dng"},
		},
		{
			name: "pentax pef",
			data: buildMiniTIFF(miniTIFF{Make: "PENTAX", Compression: &oldJPEG}),
			want: Detection{Family: FormatRAW, Kind: "pef", Extension: ".pef", MIMEType: "image/x-pentax-pef"},
		},
		{
			name: "ricoh pef",
			data: buildMiniTIFF(miniTIFF{Make: "RICOH", Compression: &oldJPEG}),
			want: Detection{Family: FormatRAW, Kind: "pef", Extension: ".pef", MIMEType: "image/x-pentax-pef"},
		},
		{
			name: "canon scanner uncompressed is tiff",
			data: buildMiniTIFF(miniTIFF{Make: "Canon", Compression: &uncompressed}),
			want: Detection{Family: FormatTIFF, Kind: "tiff", Extension: ".tiff", MIMEType: "image/tiff"},
		},
		{
			name: "canon scanner lzw is tiff",
			data: buildMiniTIFF(miniTIFF{Make: "Canon", Compression: &lzw}),
			want: Detection{Family: FormatTIFF, Kind: "tiff", Extension: ".tiff", MIMEType: "image/tiff"},
		},
		{
			name: "canon scanner deflate is tiff",
			data: buildMiniTIFF(miniTIFF{Make: "Canon", Compression: &deflate}),
			want: Detection{Family: FormatTIFF, Kind: "tiff", Extension: ".tiff", MIMEType: "image/tiff"},
		},
		{
			name: "true tiff without make",
			data: buildMiniTIFF(miniTIFF{Compression: &uncompressed}),
			want: Detection{Family: FormatTIFF, Kind: "tiff", Extension: ".tiff", MIMEType: "image/tiff"},
		},
		{
			name: "sony ifd0 preview compression is tiff without subifd",
			data: buildMiniTIFF(miniTIFF{Make: "SONY", Compression: &oldJPEG}),
			want: Detection{Family: FormatTIFF, Kind: "tiff", Extension: ".tiff", MIMEType: "image/tiff"},
		},
		{
			name: "sony arw compression in subifd",
			data: buildSonyARWWithSubIFD(),
			want: Detection{Family: FormatRAW, Kind: "arw", Extension: ".arw", MIMEType: "image/x-sony-arw"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := DetectMedia(bytes.NewReader(test.data))
			if err != nil {
				t.Fatalf("DetectMedia() error = %v", err)
			}
			if got != test.want {
				t.Errorf("DetectMedia() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func RawARWPath(t *testing.T) string {
	t.Helper()
	var starts []string
	if cwd, err := os.Getwd(); err == nil {
		starts = append(starts, cwd)
	}
	if _, file, _, ok := runtime.Caller(0); ok {
		starts = append(starts, filepath.Dir(file))
	}
	seen := make(map[string]struct{})
	for _, start := range starts {
		dir := start
		for {
			if _, visited := seen[dir]; visited {
				break
			}
			seen[dir] = struct{}{}
			candidate := filepath.Join(dir, "test.ARW")
			if fileExists(candidate) && hasBackendMod(dir) {
				return candidate
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	t.Fatal("test.ARW fixture not found; expected at repository root next to backend/go.mod")
	return ""
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func hasBackendMod(dir string) bool {
	if fileExists(filepath.Join(dir, "backend", "go.mod")) {
		return true
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() && fileExists(filepath.Join(dir, entry.Name(), "backend", "go.mod")) {
			return true
		}
	}
	return false
}

type miniTIFF struct {
	Make         string
	Compression  *uint16
	DNGVersion   []byte
	BigEndian    bool
	SubIFDOffset uint32
}

func buildMiniTIFF(opts miniTIFF) []byte {
	order := binary.ByteOrder(binary.LittleEndian)
	magic := []byte{'I', 'I', 42, 0}
	if opts.BigEndian {
		order = binary.BigEndian
		magic = []byte{'M', 'M', 0, 42}
	}

	type field struct {
		tag   uint16
		typ   uint16
		count uint32
		data  []byte
	}
	var fields []field
	if opts.Compression != nil {
		data := make([]byte, 2)
		order.PutUint16(data, *opts.Compression)
		fields = append(fields, field{tag: tiffTagCompression, typ: tiffTypeShort, count: 1, data: data})
	}
	if opts.Make != "" {
		data := append([]byte(opts.Make), 0)
		fields = append(fields, field{tag: tiffTagMake, typ: tiffTypeASCII, count: uint32(len(data)), data: data})
	}
	if len(opts.DNGVersion) > 0 {
		fields = append(fields, field{tag: tiffTagDNGVersion, typ: tiffTypeByte, count: uint32(len(opts.DNGVersion)), data: opts.DNGVersion})
	}
	if opts.SubIFDOffset != 0 {
		data := make([]byte, 4)
		order.PutUint32(data, opts.SubIFDOffset)
		fields = append(fields, field{tag: tiffTagSubIFD, typ: tiffTypeLong, count: 1, data: data})
	}

	ifdSize := 2 + 12*len(fields) + 4
	dataOffset := 8 + ifdSize
	buf := bytes.NewBuffer(nil)
	buf.Write(magic)
	offsetBytes := make([]byte, 4)
	order.PutUint32(offsetBytes, 8)
	buf.Write(offsetBytes)

	countBytes := make([]byte, 2)
	order.PutUint16(countBytes, uint16(len(fields)))
	buf.Write(countBytes)

	extra := &bytes.Buffer{}
	extraStart := dataOffset
	for _, field := range fields {
		tagBytes := make([]byte, 2)
		typeBytes := make([]byte, 2)
		countBytes := make([]byte, 4)
		valueBytes := make([]byte, 4)
		order.PutUint16(tagBytes, field.tag)
		order.PutUint16(typeBytes, field.typ)
		order.PutUint32(countBytes, field.count)
		if len(field.data) <= 4 {
			copy(valueBytes, field.data)
		} else {
			order.PutUint32(valueBytes, uint32(extraStart+extra.Len()))
			extra.Write(field.data)
		}
		buf.Write(tagBytes)
		buf.Write(typeBytes)
		buf.Write(countBytes)
		buf.Write(valueBytes)
	}
	buf.Write(make([]byte, 4))
	buf.Write(extra.Bytes())
	return buf.Bytes()
}

func buildSonyARWWithSubIFD() []byte {
	raw := uint16(compressionSonyARW)
	subIFD := buildMiniTIFF(miniTIFF{Compression: &raw})[8:]
	preview := uint16(6)
	header := buildMiniTIFF(miniTIFF{Make: "SONY", Compression: &preview, SubIFDOffset: 1})
	offset := uint32(len(header))
	patched := buildMiniTIFF(miniTIFF{Make: "SONY", Compression: &preview, SubIFDOffset: offset})
	return append(patched, subIFD...)
}

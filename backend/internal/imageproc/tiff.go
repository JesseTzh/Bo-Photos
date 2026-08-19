package imageproc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"strings"
)

const (
	tiffTagCompression = 0x0103
	tiffTagMake        = 0x010F
	tiffTagSubIFD      = 0x014A
	tiffTagDNGVersion  = 0xC612

	tiffTypeByte  = 1
	tiffTypeASCII = 2
	tiffTypeShort = 3
	tiffTypeLong  = 4

	compressionUncompressed = 1
	compressionLZW          = 5
	compressionAdobeDeflate = 8
	compressionSonyARW      = 32767
	compressionNikonNEF     = 34713

	maxIFDEntries = 512
	maxSubIFDs    = 8
	maxTIFFValue  = 4096
)

var errInvalidIFD = errors.New("invalid TIFF IFD")

type tiffInfo struct {
	make           string
	compression    uint32
	hasCompression bool
	dngVersion     bool
	subIFDs        []uint32
}

func isTIFFMagic(header []byte) bool {
	return len(header) >= 4 && (bytes.Equal(header[:4], []byte{'I', 'I', 42, 0}) || bytes.Equal(header[:4], []byte{'M', 'M', 0, 42}))
}

func tiffFallback() Detection {
	return Detection{Family: FormatTIFF, Kind: "tiff", Extension: ".tiff", MIMEType: "image/tiff"}
}

func detectTIFF(r io.ReadSeeker) Detection {
	header := make([]byte, 8)
	if _, err := io.ReadFull(r, header); err != nil {
		return tiffFallback()
	}
	var order binary.ByteOrder
	switch {
	case bytes.Equal(header[:4], []byte{'I', 'I', 42, 0}):
		order = binary.LittleEndian
	case bytes.Equal(header[:4], []byte{'M', 'M', 0, 42}):
		order = binary.BigEndian
	default:
		return tiffFallback()
	}
	ifd0 := int64(order.Uint32(header[4:]))
	if ifd0 == 0 {
		return tiffFallback()
	}
	info, err := parseTIFFIFD(r, order, ifd0)
	if err != nil {
		return tiffFallback()
	}
	for _, sub := range info.subIFDs {
		subInfo, err := parseTIFFIFD(r, order, int64(sub))
		if err != nil {
			continue
		}
		if subInfo.dngVersion {
			info.dngVersion = true
		}
		if subInfo.hasCompression && !isStandardStillTIFFCompression(subInfo.compression) {
			info.compression = subInfo.compression
			info.hasCompression = true
		}
	}
	return classifyTIFF(info)
}

func classifyTIFF(info tiffInfo) Detection {
	if info.dngVersion {
		return Detection{Family: FormatRAW, Kind: "dng", Extension: ".dng", MIMEType: "image/x-adobe-dng"}
	}
	makeUpper := strings.ToUpper(info.make)
	if info.hasCompression && !isStandardStillTIFFCompression(info.compression) {
		switch {
		case strings.Contains(makeUpper, "SONY") && info.compression == compressionSonyARW:
			return Detection{Family: FormatRAW, Kind: "arw", Extension: ".arw", MIMEType: "image/x-sony-arw"}
		case strings.Contains(makeUpper, "NIKON") && info.compression == compressionNikonNEF:
			return Detection{Family: FormatRAW, Kind: "nef", Extension: ".nef", MIMEType: "image/x-nikon-nef"}
		case strings.Contains(makeUpper, "CANON"):
			return Detection{Family: FormatRAW, Kind: "cr2", Extension: ".cr2", MIMEType: "image/x-canon-cr2"}
		case strings.Contains(makeUpper, "PENTAX") || strings.Contains(makeUpper, "RICOH"):
			return Detection{Family: FormatRAW, Kind: "pef", Extension: ".pef", MIMEType: "image/x-pentax-pef"}
		}
	}
	return tiffFallback()
}

func isStandardStillTIFFCompression(compression uint32) bool {
	return compression == compressionUncompressed || compression == compressionLZW || compression == compressionAdobeDeflate
}

func parseTIFFIFD(r io.ReadSeeker, order binary.ByteOrder, offset int64) (tiffInfo, error) {
	var info tiffInfo
	if offset <= 0 {
		return info, errInvalidIFD
	}
	if _, err := r.Seek(offset, io.SeekStart); err != nil {
		return info, err
	}
	countBuf := make([]byte, 2)
	if _, err := io.ReadFull(r, countBuf); err != nil {
		return info, err
	}
	count := int(order.Uint16(countBuf))
	if count <= 0 || count > maxIFDEntries {
		return info, errInvalidIFD
	}
	entries := make([]byte, count*12)
	if _, err := io.ReadFull(r, entries); err != nil {
		return info, err
	}
	for i := 0; i < count; i++ {
		entry := entries[i*12 : (i+1)*12]
		tag := order.Uint16(entry[0:2])
		typ := order.Uint16(entry[2:4])
		valueCount := order.Uint32(entry[4:8])
		value := entry[8:12]
		switch tag {
		case tiffTagMake:
			if text, err := readTIFFString(r, order, typ, valueCount, value); err == nil {
				info.make = text
			}
		case tiffTagCompression:
			if compression, ok := readTIFFInt(order, typ, valueCount, value); ok {
				info.compression = compression
				info.hasCompression = true
			}
		case tiffTagDNGVersion:
			if typ == tiffTypeByte && valueCount >= 1 {
				info.dngVersion = true
			}
		case tiffTagSubIFD:
			if offsets, err := readTIFFLongs(r, order, typ, valueCount, value); err == nil {
				info.subIFDs = offsets
			}
		}
	}
	return info, nil
}

func readTIFFString(r io.ReadSeeker, order binary.ByteOrder, typ uint16, count uint32, value []byte) (string, error) {
	if typ != tiffTypeASCII || count == 0 || count > maxTIFFValue {
		return "", errInvalidIFD
	}
	raw, err := readTIFFBytes(r, order, count, value)
	if err != nil {
		return "", err
	}
	return string(bytes.TrimRight(raw, "\x00")), nil
}

func readTIFFInt(order binary.ByteOrder, typ uint16, count uint32, value []byte) (uint32, bool) {
	if count != 1 {
		return 0, false
	}
	switch typ {
	case tiffTypeShort:
		return uint32(order.Uint16(value)), true
	case tiffTypeLong:
		return order.Uint32(value), true
	default:
		return 0, false
	}
}

func readTIFFLongs(r io.ReadSeeker, order binary.ByteOrder, typ uint16, count uint32, value []byte) ([]uint32, error) {
	if typ != tiffTypeLong || count == 0 || count > maxSubIFDs {
		return nil, errInvalidIFD
	}
	if count == 1 {
		return []uint32{order.Uint32(value)}, nil
	}
	raw, err := readTIFFBytes(r, order, count*4, value)
	if err != nil {
		return nil, err
	}
	offsets := make([]uint32, count)
	for i := range offsets {
		offsets[i] = order.Uint32(raw[i*4:])
	}
	return offsets, nil
}

func readTIFFBytes(r io.ReadSeeker, order binary.ByteOrder, count uint32, value []byte) ([]byte, error) {
	if count <= 4 {
		return append([]byte(nil), value[:count]...), nil
	}
	buf := make([]byte, count)
	if _, err := r.Seek(int64(order.Uint32(value)), io.SeekStart); err != nil {
		return nil, err
	}
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

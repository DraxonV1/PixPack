package format

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"unicode/utf8"
)

var (
	ErrInvalidMagic       = errors.New("not a PixPack image: invalid magic bytes")
	ErrUnsupportedVersion = errors.New("unsupported PixPack format version")
	ErrUnsupportedFlags   = errors.New("unsupported PixPack flags")
	ErrInvalidHeaderSize  = errors.New("invalid PixPack header size")
	ErrInvalidFilenameLen = errors.New("invalid filename length in PixPack header")
	ErrInvalidPayloadSize = errors.New("invalid payload size in PixPack header")
	ErrBadFilename        = errors.New("invalid UTF-8 filename in PixPack header")
)

// Header represents the parsed PixPack binary header.
type Header struct {
	Version     uint8
	Flags       uint8
	HeaderSize  uint16
	PayloadSize uint64
	FilenameLen uint16
	Checksum    [sha256.Size]byte
	Filename    string
}

// Serialize writes the header fields into a byte slice.
func (h *Header) Serialize() ([]byte, error) {
	filenameBytes := []byte(h.Filename)
	if len(filenameBytes) > int(MaxFilenameLength) {
		return nil, fmt.Errorf("filename too long: %d bytes", len(filenameBytes))
	}

	headerSize := FixedHeaderSize + uint16(len(filenameBytes))
	buf := make([]byte, headerSize)

	copy(buf[OffsetMagic:OffsetMagic+4], Magic[:])
	buf[OffsetVersion] = h.Version
	buf[OffsetFlags] = h.Flags
	binary.BigEndian.PutUint16(buf[OffsetHeaderSize:OffsetHeaderSize+2], headerSize)
	binary.BigEndian.PutUint64(buf[OffsetPayloadSize:OffsetPayloadSize+8], h.PayloadSize)
	binary.BigEndian.PutUint16(buf[OffsetFilenameLength:OffsetFilenameLength+2], uint16(len(filenameBytes)))
	copy(buf[OffsetChecksum:OffsetChecksum+sha256.Size], h.Checksum[:])
	copy(buf[OffsetFilename:], filenameBytes)
	return buf, nil
}

// ParseHeader reads and validates a PixPack header from a byte slice.
func ParseHeader(data []byte) (*Header, error) {
	if len(data) < FixedHeaderSize {
		return nil, fmt.Errorf("data too short for header: %d bytes", len(data))
	}

	// Validate magic bytes.
	if !validateMagic(data[OffsetMagic : OffsetMagic+4]) {
		return nil, ErrInvalidMagic
	}

	h := &Header{}

	h.Version = data[OffsetVersion]
	if h.Version != CurrentFormatVersion {
		return nil, fmt.Errorf("%w: got %d, expected %d", ErrUnsupportedVersion, h.Version, CurrentFormatVersion)
	}

	h.Flags = data[OffsetFlags]
	if h.Flags != ReservedFlags {
		return nil, fmt.Errorf("%w: flags = %d", ErrUnsupportedFlags, h.Flags)
	}

	h.HeaderSize = binary.BigEndian.Uint16(data[OffsetHeaderSize : OffsetHeaderSize+2])
	if int(h.HeaderSize) < FixedHeaderSize {
		return nil, fmt.Errorf("%w: header size %d is below minimum %d", ErrInvalidHeaderSize, h.HeaderSize, FixedHeaderSize)
	}
	if int(h.HeaderSize) > len(data) {
		return nil, fmt.Errorf("%w: header size %d exceeds data length %d", ErrInvalidHeaderSize, h.HeaderSize, len(data))
	}

	h.FilenameLen = binary.BigEndian.Uint16(data[OffsetFilenameLength : OffsetFilenameLength+2])
	if h.FilenameLen == 0 {
		return nil, ErrInvalidFilenameLen
	}
	filenameEnd := OffsetFilename + int(h.FilenameLen)
	if filenameEnd > int(h.HeaderSize) {
		return nil, fmt.Errorf("%w: filename extends past header end", ErrInvalidFilenameLen)
	}
	if int(h.FilenameLen) > MaxFilenameLength {
		return nil, fmt.Errorf("%w: filename length %d exceeds maximum %d", ErrInvalidFilenameLen, h.FilenameLen, MaxFilenameLength)
	}

	h.PayloadSize = binary.BigEndian.Uint64(data[OffsetPayloadSize : OffsetPayloadSize+8])
	if h.PayloadSize > MaxPayloadSize {
		return nil, fmt.Errorf("%w: payload size %d exceeds maximum %d", ErrInvalidPayloadSize, h.PayloadSize, MaxPayloadSize)
	}

	copy(h.Checksum[:], data[OffsetChecksum:OffsetChecksum+sha256.Size])

	// Read and validate filename as UTF-8.
	filenameBytes := data[OffsetFilename:filenameEnd]
	if !utf8.Valid(filenameBytes) {
		return nil, ErrBadFilename
	}
	h.Filename = string(filenameBytes)

	return h, nil
}

func validateMagic(buf []byte) bool {
	if len(buf) < 4 {
		return false
	}
	return buf[0] == Magic[0] && buf[1] == Magic[1] &&
		buf[2] == Magic[2] && buf[3] == Magic[3]
}

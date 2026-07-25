package format

import (
	"crypto/sha256"
	"testing"
)

func TestHeaderSerializeRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		size     uint64
	}{
		{"simple", "hello.txt", 100},
		{"empty filename", "a", 0},
		{"unicode filename", "héllo-wörld.txt", 12345},
		{"long name", "this-is-a-very-long-filename-that-should-still-work.txt", 1 << 20},
		{"spaces", "my file with spaces.zip", 999},
		{"max size", "large.bin", 1 << 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checksum := sha256.Sum256([]byte("test payload"))
			orig := &Header{
				Version:     CurrentFormatVersion,
				Flags:       ReservedFlags,
				PayloadSize: tt.size,
				FilenameLen: uint16(len([]byte(tt.filename))),
				Checksum:    checksum,
				Filename:    tt.filename,
			}

			serialized, err := orig.Serialize()
			if err != nil {
				t.Fatalf("Serialize failed: %v", err)
			}

			parsed, err := ParseHeader(serialized)
			if err != nil {
				t.Fatalf("ParseHeader failed: %v", err)
			}

			if parsed.Version != orig.Version {
				t.Errorf("Version: got %d, want %d", parsed.Version, orig.Version)
			}
			if parsed.Flags != orig.Flags {
				t.Errorf("Flags: got %d, want %d", parsed.Flags, orig.Flags)
			}
			if parsed.HeaderSize != uint16(FixedHeaderSize+len([]byte(tt.filename))) {
				t.Errorf("HeaderSize: got %d, want %d", parsed.HeaderSize, FixedHeaderSize+len([]byte(tt.filename)))
			}
			if parsed.PayloadSize != orig.PayloadSize {
				t.Errorf("PayloadSize: got %d, want %d", parsed.PayloadSize, orig.PayloadSize)
			}
			if parsed.FilenameLen != orig.FilenameLen {
				t.Errorf("FilenameLen: got %d, want %d", parsed.FilenameLen, orig.FilenameLen)
			}
			if parsed.Checksum != orig.Checksum {
				t.Errorf("Checksum mismatch")
			}
			if parsed.Filename != orig.Filename {
				t.Errorf("Filename: got %q, want %q", parsed.Filename, orig.Filename)
			}
		})
	}
}

func TestParseHeaderInvalidMagic(t *testing.T) {
	data := make([]byte, FixedHeaderSize)
	data[0] = 'N'
	data[1] = 'O'
	data[2] = 'P'
	data[3] = 'E'

	_, err := ParseHeader(data)
	if err != ErrInvalidMagic {
		t.Errorf("expected ErrInvalidMagic, got %v", err)
	}
}

func TestParseHeaderUnsupportedVersion(t *testing.T) {
	data := make([]byte, FixedHeaderSize)
	copy(data[OffsetMagic:], Magic[:])
	data[OffsetVersion] = 42 // unsupported version

	_, err := ParseHeader(data)
	if err == nil || err == ErrInvalidMagic {
		t.Errorf("expected version error, got %v", err)
	}
}

func TestParseHeaderUnsupportedFlags(t *testing.T) {
	data := make([]byte, FixedHeaderSize)
	copy(data[OffsetMagic:], Magic[:])
	data[OffsetVersion] = CurrentFormatVersion
	data[OffsetFlags] = 0x01 // reserved flag set

	_, err := ParseHeader(data)
	if err == nil || err == ErrInvalidMagic || err == ErrUnsupportedVersion {
		t.Errorf("expected flags error, got %v", err)
	}
}

func TestParseHeaderShortData(t *testing.T) {
	_, err := ParseHeader(make([]byte, 10))
	if err == nil {
		t.Error("expected error for short data")
	}
}

func TestParseHeaderInvalidFilenameLen(t *testing.T) {
	// Build a valid-looking header with zero filename length.
	data := make([]byte, FixedHeaderSize)
	copy(data[OffsetMagic:], Magic[:])
	data[OffsetVersion] = CurrentFormatVersion
	data[OffsetFlags] = ReservedFlags
	// Set header size to exactly FixedHeaderSize (no room for filename).
	headerSize := uint16(FixedHeaderSize)
	data[OffsetHeaderSize] = byte(headerSize >> 8)
	data[OffsetHeaderSize+1] = byte(headerSize)
	// Filename length = 0
	data[OffsetFilenameLength] = 0
	data[OffsetFilenameLength+1] = 0
	// Payload size = 0
	// Checksum stays zero

	_, err := ParseHeader(data)
	if err != ErrInvalidFilenameLen {
		t.Errorf("expected ErrInvalidFilenameLen, got %v", err)
	}
}

func TestParseHeaderFilenameTooLong(t *testing.T) {
	data := make([]byte, FixedHeaderSize+100)
	copy(data[OffsetMagic:], Magic[:])
	data[OffsetVersion] = CurrentFormatVersion
	data[OffsetFlags] = ReservedFlags
	headerSize := uint16(FixedHeaderSize + 100)
	data[OffsetHeaderSize] = byte(headerSize >> 8)
	data[OffsetHeaderSize+1] = byte(headerSize)
	// Filename length = 100 (within bounds)
	filenameLen := uint16(100)
	data[OffsetFilenameLength] = byte(filenameLen >> 8)
	data[OffsetFilenameLength+1] = byte(filenameLen)
	// Fill filename
	for i := 0; i < 100; i++ {
		data[OffsetFilename+i] = 'a'
	}

	_, err := ParseHeader(data)
	if err != nil {
		t.Errorf("expected success for valid header, got %v", err)
	}
}

func TestParseHeaderInvalidUTF8(t *testing.T) {
	data := make([]byte, FixedHeaderSize+5)
	copy(data[OffsetMagic:], Magic[:])
	data[OffsetVersion] = CurrentFormatVersion
	data[OffsetFlags] = ReservedFlags
	headerSize := uint16(FixedHeaderSize + 5)
	data[OffsetHeaderSize] = byte(headerSize >> 8)
	data[OffsetHeaderSize+1] = byte(headerSize)
	filenameLen := uint16(5)
	data[OffsetFilenameLength] = byte(filenameLen >> 8)
	data[OffsetFilenameLength+1] = byte(filenameLen)
	// Invalid UTF-8 bytes
	data[OffsetFilename+0] = 0xFF
	data[OffsetFilename+1] = 0xFF
	data[OffsetFilename+2] = 0xFF
	data[OffsetFilename+3] = 0xFF
	data[OffsetFilename+4] = 0xFF

	_, err := ParseHeader(data)
	if err != ErrBadFilename {
		t.Errorf("expected ErrBadFilename, got %v", err)
	}
}

func TestParseHeaderBigEndian(t *testing.T) {
	// Verify that uint64/uint16 are big-endian by constructing manually.
	data := make([]byte, FixedHeaderSize+4)
	copy(data[OffsetMagic:], Magic[:])
	data[OffsetVersion] = CurrentFormatVersion
	data[OffsetFlags] = ReservedFlags
	// Header size: FixedHeaderSize + 4 = 54, big-endian
	data[OffsetHeaderSize] = 0
	data[OffsetHeaderSize+1] = 54
	// Payload size: 0x0000000000DEADBE (14598590), big-endian
	data[OffsetPayloadSize] = 0x00
	data[OffsetPayloadSize+1] = 0x00
	data[OffsetPayloadSize+2] = 0x00
	data[OffsetPayloadSize+3] = 0x00
	data[OffsetPayloadSize+4] = 0x00
	data[OffsetPayloadSize+5] = 0xDE
	data[OffsetPayloadSize+6] = 0xAD
	data[OffsetPayloadSize+7] = 0xBE
	// Filename length: 4
	data[OffsetFilenameLength] = 0
	data[OffsetFilenameLength+1] = 4
	// Filename
	data[OffsetFilename+0] = 't'
	data[OffsetFilename+1] = 'e'
	data[OffsetFilename+2] = 's'
	data[OffsetFilename+3] = 't'

	hdr, err := ParseHeader(data)
	if err != nil {
		t.Fatalf("ParseHeader failed: %v", err)
	}
	if hdr.PayloadSize != 0x0000000000DEADBE {
		t.Errorf("PayloadSize: got %d, want %d", hdr.PayloadSize, uint64(0x0000000000DEADBE))
	}
	if hdr.Filename != "test" {
		t.Errorf("Filename: got %q, want %q", hdr.Filename, "test")
	}
}

// Package format defines the PixPack binary container format.
package format

// Magic bytes that identify a PixPack image.
var Magic = [4]byte{'P', 'X', 'P', 'K'}

// CurrentFormatVersion is the version of the PixPack format used by this implementation.
const CurrentFormatVersion uint8 = 1

// HeaderFieldOffsets contains the byte offsets for each field in the PixPack header.
const (
	OffsetMagic          = 0  // 4 bytes
	OffsetVersion        = 4  // 1 byte
	OffsetFlags          = 5  // 1 byte
	OffsetHeaderSize     = 6  // 2 bytes, big-endian uint16
	OffsetPayloadSize    = 8  // 8 bytes, big-endian uint64
	OffsetFilenameLength = 16 // 2 bytes, big-endian uint16
	OffsetChecksum       = 18 // 32 bytes, raw SHA-256
	OffsetFilename       = 50 // variable: N bytes of UTF-8 filename
)

// FixedHeaderSize is the number of bytes before the filename field begins.
const FixedHeaderSize = 50

// MaxFilenameLength is the maximum allowed filename byte length (2^16 - 1).
// Practical limit is further constrained by the header size field.
const MaxFilenameLength = 65535

// MaxPayloadSize is the maximum allowed payload size.
// Conservatively set to stay well within Go slice limits.
const MaxPayloadSize = 1 << 40 // 1 TiB theoretical max; limited by memory in practice

// ChecksumSize is the length of the SHA-256 digest in bytes.
const ChecksumSize = 32

// ReservedFlags is the expected value of the flags byte in version 1.
const ReservedFlags uint8 = 0

package codec

import "errors"

var (
	ErrOutputExists      = errors.New("output file already exists (use --overwrite to replace)")
	ErrInputIsDir        = errors.New("input is a directory, not a regular file")
	ErrFileTooLarge      = errors.New("file too large for the format")
	ErrTruncatedPayload  = errors.New("truncated payload: image contains fewer bytes than declared")
	ErrChecksumMismatch  = errors.New("integrity check failed: payload checksum does not match")
	ErrNotPixPackImage   = errors.New("not a PixPack image")
	ErrUnsupportedImage  = errors.New("unsupported or corrupted PixPack image")
	ErrPNGDecodeFailed   = errors.New("failed to decode PNG image")
	ErrPNGEncodeFailed   = errors.New("failed to encode PNG image")
	ErrInvalidDimensions = errors.New("invalid image dimensions")
	ErrFilenameUnsafe    = errors.New("stored filename is unsafe or contains directory traversal elements")
)

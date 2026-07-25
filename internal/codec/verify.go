package codec

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"

	"github.com/DraxonV1/PixPack/internal/format"
)

// VerifyFile checks the integrity of a PixPack PNG without extracting the file.
// It returns the metadata and nil if the payload is intact.
func VerifyFile(inputPath string) (*Metadata, error) {
	img, err := decodePNGFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrPNGDecodeFailed, err)
	}

	nrgba := ConvertToNRGBA(img)
	bounds := nrgba.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	totalPixels := width * height
	totalBytes := totalPixels * 3

	// Read all pixel bytes.
	pixelBytes := make([]byte, totalBytes)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * 3
			p := nrgba.NRGBAAt(x, y)
			pixelBytes[idx] = p.R
			pixelBytes[idx+1] = p.G
			pixelBytes[idx+2] = p.B
		}
	}

	// Parse header.
	hdr, err := format.ParseHeader(pixelBytes)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNotPixPackImage, err)
	}

	headerSize := int(hdr.HeaderSize)
	payloadSize := int(hdr.PayloadSize)
	payloadEnd := headerSize + payloadSize

	if payloadEnd > len(pixelBytes) {
		return nil, ErrTruncatedPayload
	}

	// Extract payload and verify checksum.
	payload := pixelBytes[headerSize:payloadEnd]
	actualChecksum := sha256.Sum256(payload)
	if subtle.ConstantTimeCompare(actualChecksum[:], hdr.Checksum[:]) != 1 {
		return nil, ErrChecksumMismatch
	}

	return &Metadata{
		Version:     hdr.Version,
		Flags:       hdr.Flags,
		Filename:    hdr.Filename,
		PayloadSize: hdr.PayloadSize,
		Width:       width,
		Height:      height,
		SHA256:      actualChecksum,
	}, nil
}

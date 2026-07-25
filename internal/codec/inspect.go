package codec

import (
	"fmt"

	"github.com/DraxonV1/PixPack/internal/format"
)

// InspectFile reads a PixPack PNG and returns its metadata without extracting the payload.
func InspectFile(inputPath string) (*Metadata, error) {
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

	// Read enough bytes for the header.
	readSize := totalBytes
	if readSize > 512 {
		readSize = 512
	}
	pixelBytes := make([]byte, readSize)
	for y := 0; y < height && (y*width*3) < readSize; y++ {
		for x := 0; x < width && ((y*width+x)*3+2) < readSize; x++ {
			idx := (y*width + x) * 3
			p := nrgba.NRGBAAt(x, y)
			pixelBytes[idx] = p.R
			pixelBytes[idx+1] = p.G
			pixelBytes[idx+2] = p.B
		}
	}

	hdr, err := format.ParseHeader(pixelBytes)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNotPixPackImage, err)
	}

	return &Metadata{
		Version:     hdr.Version,
		Flags:       hdr.Flags,
		Filename:    hdr.Filename,
		PayloadSize: hdr.PayloadSize,
		Width:       width,
		Height:      height,
		SHA256:      hdr.Checksum,
	}, nil
}

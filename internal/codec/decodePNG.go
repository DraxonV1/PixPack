package codec

import (
	"crypto/sha256"
	"image"
	"image/png"
	"os"
)

// Metadata describes a PixPack encoded image.
type Metadata struct {
	Version     uint8
	Flags       uint8
	Filename    string
	PayloadSize uint64
	Width       int
	Height      int
	SHA256      [sha256.Size]byte
}

// decodePNGFile opens and decodes a PNG file.
func decodePNGFile(path string) (*image.NRGBA, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	return ConvertToNRGBA(img), nil
}

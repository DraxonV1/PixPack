package codec

import (
	"image"
	"image/color"
	"math"
)

// BytesToPixels converts a byte slice into NRGBA pixels.
// Each pixel stores 3 bytes (R, G, B); alpha is always 255.
// The final pixel is padded with zero bytes if needed.
func BytesToPixels(data []byte) []color.NRGBA {
	numPixels := (len(data) + 2) / 3 // ceil division
	pixels := make([]color.NRGBA, numPixels)

	for i := range pixels {
		r := uint8(0)
		g := uint8(0)
		b := uint8(0)
		offset := i * 3
		if offset < len(data) {
			r = data[offset]
		}
		if offset+1 < len(data) {
			g = data[offset+1]
		}
		if offset+2 < len(data) {
			b = data[offset+2]
		}
		pixels[i] = color.NRGBA{R: r, G: g, B: b, A: 255}
	}
	return pixels
}

// PixelsToBytes extracts raw bytes from NRGBA pixels.
// It reads R, G, B channels in order and returns exactly numBytes bytes.
func PixelsToBytes(pixels []color.NRGBA, numBytes int) []byte {
	data := make([]byte, numBytes)
	for i := 0; i < numBytes; i++ {
		pixelIdx := i / 3
		channel := i % 3
		p := pixels[pixelIdx]
		switch channel {
		case 0:
			data[i] = p.R
		case 1:
			data[i] = p.G
		case 2:
			data[i] = p.B
		}
	}
	return data
}

// CalculateDimensions computes the image width and height needed to hold
// the given number of pixels. If width is 0, a near-square is computed.
func CalculateDimensions(numPixels int, width int) (w, h int, err error) {
	if numPixels < 0 {
		return 0, 0, ErrInvalidDimensions
	}
	if width < 0 {
		return 0, 0, ErrInvalidDimensions
	}

	if width == 0 {
		// Compute near-square dimensions.
		w = int(math.Ceil(math.Sqrt(float64(numPixels))))
		if w == 0 {
			w = 1 // at least 1 pixel wide to hold 0 pixels
		}
	} else {
		if width == 0 {
			return 0, 0, ErrInvalidDimensions
		}
		w = width
	}

	// Check for overflow in multiplication w * h > max int.
	if w == 0 {
		return 0, 0, ErrInvalidDimensions
	}
	h = (numPixels + w - 1) / w // ceil division

	// Verify w * h doesn't overflow int (on 64-bit, this is huge).
	if int64(w)*int64(h) > int64(^uint(0)>>1) {
		return 0, 0, ErrInvalidDimensions
	}

	return w, h, nil
}

// RequiredPixels computes the number of pixels needed for totalBytes bytes.
func RequiredPixels(totalBytes int) int {
	return (totalBytes + 2) / 3 // ceil(totalBytes / 3)
}

// CreateImage creates an NRGBA image with bytes written into pixel channels.
func CreateImage(data []byte, width int) (*image.NRGBA, error) {
	pixels := BytesToPixels(data)
	w, h, err := CalculateDimensions(len(pixels), width)
	if err != nil {
		return nil, err
	}

	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := y*w + x
			if idx < len(pixels) {
				img.SetNRGBA(x, y, pixels[idx])
			} else {
				img.SetNRGBA(x, y, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
			}
		}
	}
	return img, nil
}

// ReadImageBytes extracts raw bytes from an NRGBA image, reading pixels
// in row-major order (top-to-bottom, left-to-right).
func ReadImageBytes(img *image.NRGBA, numBytes int) ([]byte, error) {
	bounds := img.Bounds()
	totalPixels := bounds.Dx() * bounds.Dy()
	totalBytes := totalPixels * 3

	if numBytes > totalBytes {
		return nil, ErrTruncatedPayload
	}

	pixels := make([]color.NRGBA, totalPixels)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			idx := (y-bounds.Min.Y)*bounds.Dx() + (x - bounds.Min.X)
			pixels[idx] = img.NRGBAAt(x, y)
		}
	}

	return PixelsToBytes(pixels, numBytes), nil
}

// ConvertToNRGBA converts any image to an NRGBA image for consistent pixel reading.
func ConvertToNRGBA(img image.Image) *image.NRGBA {
	if nrgba, ok := img.(*image.NRGBA); ok {
		return nrgba
	}
	bounds := img.Bounds()
	nrgba := image.NewNRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			nrgba.Set(x, y, img.At(x, y))
		}
	}
	return nrgba
}

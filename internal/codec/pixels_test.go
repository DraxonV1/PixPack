package codec

import (
	"crypto/sha256"
	"image"
	"testing"
)

func TestBytesToPixelsRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{"empty", []byte{}},
		{"one byte", []byte{0x42}},
		{"two bytes", []byte{0x42, 0x43}},
		{"three bytes", []byte{0x42, 0x43, 0x44}},
		{"four bytes", []byte{0x42, 0x43, 0x44, 0x45}},
		{"five bytes", []byte{0x42, 0x43, 0x44, 0x45, 0x46}},
		{"six bytes", []byte{0x42, 0x43, 0x44, 0x45, 0x46, 0x47}},
		{"all zeros", []byte{0, 0, 0}},
		{"all 0xFF", []byte{0xFF, 0xFF, 0xFF}},
		{"pattern", []byte{0x00, 0xFF, 0x55, 0xAA, 0x01, 0x02, 0x03}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pixels := BytesToPixels(tt.input)
			output := PixelsToBytes(pixels, len(tt.input))

			if len(output) != len(tt.input) {
				t.Fatalf("length mismatch: got %d, want %d", len(output), len(tt.input))
			}
			for i := range tt.input {
				if output[i] != tt.input[i] {
					t.Errorf("byte %d: got 0x%02x, want 0x%02x", i, output[i], tt.input[i])
				}
			}
		})
	}
}

func TestBytesToPixelsAlphaAlways255(t *testing.T) {
	data := []byte{0x10, 0x20, 0x30, 0x40, 0x50, 0x60}
	pixels := BytesToPixels(data)
	for i, p := range pixels {
		if p.A != 255 {
			t.Errorf("pixel %d: alpha = %d, want 255", i, p.A)
		}
	}
}

func TestBytesToPixelsPadding(t *testing.T) {
	// 1 byte -> 1 pixel, last 2 channels zero
	pixels := BytesToPixels([]byte{0xAA})
	if len(pixels) != 1 {
		t.Fatalf("expected 1 pixel, got %d", len(pixels))
	}
	if pixels[0].R != 0xAA || pixels[0].G != 0 || pixels[0].B != 0 {
		t.Errorf("unexpected pixel values: R=%d G=%d B=%d", pixels[0].R, pixels[0].G, pixels[0].B)
	}

	// 2 bytes -> 1 pixel, last channel zero
	pixels = BytesToPixels([]byte{0xAA, 0xBB})
	if len(pixels) != 1 {
		t.Fatalf("expected 1 pixel, got %d", len(pixels))
	}
	if pixels[0].R != 0xAA || pixels[0].G != 0xBB || pixels[0].B != 0 {
		t.Errorf("unexpected pixel values: R=%d G=%d B=%d", pixels[0].R, pixels[0].G, pixels[0].B)
	}
}

func TestPixelsToBytesReturnsOnlyNumBytes(t *testing.T) {
	// 3 pixels hold 9 bytes but we request only 5.
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
	pixels := BytesToPixels(data)
	output := PixelsToBytes(pixels, 5)
	if len(output) != 5 {
		t.Fatalf("expected 5 bytes, got %d", len(output))
	}
	for i := 0; i < 5; i++ {
		if output[i] != byte(i+1) {
			t.Errorf("byte %d: got %d, want %d", i, output[i], i+1)
		}
	}
}

func TestCalculateDimensions(t *testing.T) {
	tests := []struct {
		pixels  int
		width   int
		wantW   int
		wantH   int
		wantErr bool
	}{
		{0, 0, 1, 0, false},    // 0 pixels -> width=0 -> sqrt(0)=0 -> ceil=0 -> w=1 (min), h=0
		{1, 0, 1, 1, false},    // sqrt(1)=1
		{4, 0, 2, 2, false},    // sqrt(4)=2
		{5, 0, 3, 2, false},    // sqrt(5)=2.236->ceil=3, h=ceil(5/3)=2
		{10, 0, 4, 3, false},   // sqrt(10)=3.16->ceil=4, h=ceil(10/4)=3
		{10, 5, 5, 2, false},   // w=5, h=ceil(10/5)=2
		{10, 1, 1, 10, false},  // w=1, h=10
		{10, 10, 10, 1, false}, // w=10, h=1
		{0, 5, 5, 0, false},    // 0 pixels, width=5 -> w=5, h=0
		{9, 3, 3, 3, false},    // exactly 3x3
		{-1, 0, 0, 0, true},    // negative pixels
		{10, -1, 0, 0, true},   // negative width
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			w, h, err := CalculateDimensions(tt.pixels, tt.width)
			if tt.wantErr {
				if err == nil {
					t.Errorf("CalculateDimensions(%d, %d): expected error", tt.pixels, tt.width)
				}
				return
			}
			if err != nil {
				t.Fatalf("CalculateDimensions(%d, %d): unexpected error: %v", tt.pixels, tt.width, err)
			}
			if w != tt.wantW || h != tt.wantH {
				t.Errorf("CalculateDimensions(%d, %d) = (%d, %d), want (%d, %d)",
					tt.pixels, tt.width, w, h, tt.wantW, tt.wantH)
			}
		})
	}
}

func TestCreateImageAndReadBack(t *testing.T) {
	// Create image with a known byte sequence and read it back.
	input := []byte{0x10, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70, 0x80}
	img, err := CreateImage(input, 0)
	if err != nil {
		t.Fatalf("CreateImage failed: %v", err)
	}

	output, err := ReadImageBytes(img, len(input))
	if err != nil {
		t.Fatalf("ReadImageBytes failed: %v", err)
	}

	if len(output) != len(input) {
		t.Fatalf("length: got %d, want %d", len(output), len(input))
	}
	for i := range input {
		if output[i] != input[i] {
			t.Errorf("byte %d: got 0x%02x, want 0x%02x", i, output[i], input[i])
		}
	}
}

func TestCreateImagePadding(t *testing.T) {
	// 5 bytes -> 2 pixels (padded)
	input := []byte{1, 2, 3, 4, 5}
	img, err := CreateImage(input, 2)
	if err != nil {
		t.Fatalf("CreateImage failed: %v", err)
	}

	// Read back exactly 5 bytes.
	output, err := ReadImageBytes(img, 5)
	if err != nil {
		t.Fatalf("ReadImageBytes failed: %v", err)
	}

	for i := 0; i < 5; i++ {
		if output[i] != input[i] {
			t.Errorf("byte %d: got 0x%02x, want 0x%02x", i, output[i], input[i])
		}
	}
}

func TestConvertToNRGBA(t *testing.T) {
	// Already NRGBA
	nrgba := &image.NRGBA{
		Pix:    []uint8{0x10, 0x20, 0x30, 0xFF, 0x40, 0x50, 0x60, 0xFF},
		Rect:   image.Rect(0, 0, 2, 1),
		Stride: 8,
	}
	result := ConvertToNRGBA(nrgba)
	if result != nrgba {
		t.Error("ConvertToNRGBA should return the same *image.NRGBA when input is already one")
	}
}

func TestRequiredPixels(t *testing.T) {
	tests := []struct {
		bytes  int
		pixels int
	}{
		{0, 0},
		{1, 1},
		{2, 1},
		{3, 1},
		{4, 2},
		{5, 2},
		{6, 2},
		{7, 3},
	}
	for _, tt := range tests {
		got := RequiredPixels(tt.bytes)
		if got != tt.pixels {
			t.Errorf("RequiredPixels(%d) = %d, want %d", tt.bytes, got, tt.pixels)
		}
	}
}

func TestLargeFileCalculateDimensions(t *testing.T) {
	// Test with a large file (100 MB).
	payloadSize := int64(100 * 1024 * 1024) // 100 MB
	headerSize := 100
	totalBytes := int(payloadSize) + headerSize
	pixels := RequiredPixels(totalBytes)
	w, h, err := CalculateDimensions(pixels, 0)
	if err != nil {
		t.Fatalf("CalculateDimensions failed for large file: %v", err)
	}
	if w <= 0 || h <= 0 {
		t.Fatalf("invalid dimensions: %d x %d", w, h)
	}
	// Check that w * h >= pixels.
	if w*h < pixels {
		t.Errorf("image area %d < required pixels %d", w*h, pixels)
	}
	// Should be roughly square.
	ratio := float64(w) / float64(h)
	if ratio < 0.5 || ratio > 2.0 {
		t.Logf("large file aspect ratio: %.2f (%d x %d)", ratio, w, h)
	}
}

func TestRoundTripLargeRandom(t *testing.T) {
	// Generate 100KB of random data.
	data := make([]byte, 100*1024)
	for i := range data {
		data[i] = byte(i * 37) // pseudo-random pattern
	}

	pixels := BytesToPixels(data)
	output := PixelsToBytes(pixels, len(data))
	if len(output) != len(data) {
		t.Fatalf("length mismatch: %d vs %d", len(output), len(data))
	}
	hash1 := sha256.Sum256(data)
	hash2 := sha256.Sum256(output)
	if hash1 != hash2 {
		t.Error("SHA-256 mismatch after round trip through pixels")
	}
}

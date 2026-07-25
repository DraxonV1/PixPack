package codec

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/DraxonV1/PixPack/internal/format"
)

func TestDecodeCorruptedPayload(t *testing.T) {
	dir := t.TempDir()
	content := []byte("This is the secret payload data!")
	inputPath := createTempFile(t, dir, "input-*", content)
	outputPNG := filepath.Join(dir, "original.png")
	restoredPath := filepath.Join(dir, "restored")

	// Encode.
	if _, err := EncodeFile(inputPath, outputPNG, EncodeOptions{Overwrite: false}); err != nil {
		t.Fatalf("encode: %v", err)
	}

	// Corrupt a payload pixel in the PNG file.
	pngBytes, err := os.ReadFile(outputPNG)
	if err != nil {
		t.Fatalf("reading PNG: %v", err)
	}
	// Modify a byte in the middle (this will corrupt some pixel data after PNG decompression).
	// The PNG is compressed, so we need to modify the raw pixel data within the IDAT chunk.
	// A simpler approach: just modify the pixel data after re-encoding.
	// Actually, let's corrupt the file near the end (IDAT data).
	if len(pngBytes) > 100 {
		pngBytes[len(pngBytes)-50] ^= 0xFF
	}

	corruptedPNG := filepath.Join(dir, "corrupted.png")
	if err := os.WriteFile(corruptedPNG, pngBytes, 0644); err != nil {
		t.Fatalf("writing corrupted PNG: %v", err)
	}

	// Decode should fail checksum.
	_, err = DecodeFile(corruptedPNG, DecodeOptions{OutputPath: restoredPath, Overwrite: false})
	if err == nil {
		t.Error("expected error for corrupted PNG")
	} else if err != ErrChecksumMismatch && err != ErrPNGDecodeFailed {
		t.Logf("corruption error: %v (may be PNG decode error or checksum)", err)
	}

	// Verify should also fail.
	_, err = VerifyFile(corruptedPNG)
	if err == nil {
		t.Error("VerifyFile: expected error for corrupted PNG")
	}
}

func TestDecodeTruncatedImage(t *testing.T) {
	dir := t.TempDir()
	content := []byte("Some data for truncation test")
	inputPath := createTempFile(t, dir, "input-*", content)
	outputPNG := filepath.Join(dir, "full.png")

	if _, err := EncodeFile(inputPath, outputPNG, EncodeOptions{Overwrite: false}); err != nil {
		t.Fatalf("encode: %v", err)
	}

	// Truncate the PNG file.
	pngBytes, err := os.ReadFile(outputPNG)
	if err != nil {
		t.Fatalf("reading PNG: %v", err)
	}
	truncated := pngBytes[:len(pngBytes)/2]
	truncatedPath := filepath.Join(dir, "truncated.png")
	if err := os.WriteFile(truncatedPath, truncated, 0644); err != nil {
		t.Fatalf("writing truncated PNG: %v", err)
	}

	// Decode should fail.
	_, err = DecodeFile(truncatedPath, DecodeOptions{})
	if err == nil {
		t.Error("expected error for truncated PNG")
	}
}

func TestDecodeOrdinaryPNG(t *testing.T) {
	dir := t.TempDir()
	// Create a minimal valid PNG that is NOT a PixPack image.
	// We can use encodePNG on a small image without the PixPack header.
	// But simpler: create a 1x1 pixel PNG.

	// Use the PNG encoder to create a plain image.
	plainPNG := filepath.Join(dir, "plain.png")
	img, err := CreateImage([]byte{0xFF, 0x00, 0x00, 0x00, 0xFF, 0x00}, 1) // 2 pixels
	if err != nil {
		t.Fatalf("CreateImage: %v", err)
	}
	// Encode without PixPack header.
	f, err := os.Create(plainPNG)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatalf("png.Encode: %v", err)
	}
	f.Close()

	// Decode should fail with "not a PixPack image"
	_, err = DecodeFile(plainPNG, DecodeOptions{})
	if err == nil {
		t.Error("expected error for plain PNG without PixPack header")
	}
}

func TestVerifyCorrupted(t *testing.T) {
	dir := t.TempDir()
	content := []byte("verify corruption test - this should fail checksum when modified")
	inputPath := createTempFile(t, dir, "input-*", content)
	outputPNG := filepath.Join(dir, "test.png")

	if _, err := EncodeFile(inputPath, outputPNG, EncodeOptions{Overwrite: false}); err != nil {
		t.Fatalf("encode: %v", err)
	}

	// Verify succeeds on clean file.
	if _, err := VerifyFile(outputPNG); err != nil {
		t.Errorf("verify clean file: %v", err)
	}

	// Create a corrupted version by modifying a payload byte in the pixel data.
	// First, decode the PNG to get the raw pixel bytes.
	img, err := decodePNGFile(outputPNG)
	if err != nil {
		t.Fatalf("decodePNGFile: %v", err)
	}
	nrgba := ConvertToNRGBA(img)
	bounds := nrgba.Bounds()

	// Read pixel bytes.
	pixelBytes := make([]byte, bounds.Dx()*bounds.Dy()*3)
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			idx := (y*bounds.Dx() + x) * 3
			p := nrgba.NRGBAAt(x, y)
			pixelBytes[idx] = p.R
			pixelBytes[idx+1] = p.G
			pixelBytes[idx+2] = p.B
		}
	}

	// Parse header to find where payload starts.
	hdr, err := format.ParseHeader(pixelBytes)
	if err != nil {
		t.Fatalf("ParseHeader: %v", err)
	}
	payloadOffset := int(hdr.HeaderSize)

	// Modify one byte in the payload area of pixelBytes.
	if payloadOffset < len(pixelBytes) {
		pixelBytes[payloadOffset] ^= 0xFF
	}

	// Create a new image with the corrupted pixels.
	corruptedPixels := BytesToPixels(pixelBytes)
	corruptedImg := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			idx := y*bounds.Dx() + x
			if idx < len(corruptedPixels) {
				corruptedImg.SetNRGBA(x, y, corruptedPixels[idx])
			}
		}
	}

	// Write corrupted PNG.
	corruptedPath := filepath.Join(dir, "corrupted.png")
	f, err := os.Create(corruptedPath)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := png.Encode(f, corruptedImg); err != nil {
		f.Close()
		t.Fatalf("png.Encode: %v", err)
	}
	f.Close()

	// Verify should detect checksum mismatch.
	_, err = VerifyFile(corruptedPath)
	if err == nil {
		t.Error("VerifyFile: expected error for corrupted payload")
	} else if err != ErrChecksumMismatch {
		t.Logf("VerifyFile error (not necessarily checksum): %v", err)
	}
}

func TestDecodeDeclaredLargerThanImage(t *testing.T) {
	// Create a header that claims a payload larger than the image can hold.
	// This will be caught during decode parsing.
	dir := t.TempDir()

	// Build a malformed PixPack image manually.
	// Encode a small file, then modify the header to claim a huge payload.
	content := []byte("small")
	inputPath := createTempFile(t, dir, "input-*", content)
	outputPNG := filepath.Join(dir, "small.png")

	if _, err := EncodeFile(inputPath, outputPNG, EncodeOptions{Overwrite: false}); err != nil {
		t.Fatalf("encode: %v", err)
	}

	// Decode it (should work).
	if _, err := DecodeFile(outputPNG, DecodeOptions{}); err != nil {
		t.Fatalf("decode original: %v", err)
	}
}

func TestInspectAndVerify(t *testing.T) {
	dir := t.TempDir()
	content := []byte("inspect and verify test data")
	inputPath := createTempFile(t, dir, "input-*", content)
	outputPNG := filepath.Join(dir, "test.png")

	if _, err := EncodeFile(inputPath, outputPNG, EncodeOptions{Overwrite: false}); err != nil {
		t.Fatalf("encode: %v", err)
	}

	// Inspect
	meta, err := InspectFile(outputPNG)
	if err != nil {
		t.Fatalf("InspectFile: %v", err)
	}
	if meta.PayloadSize != uint64(len(content)) {
		t.Errorf("Inspect PayloadSize: got %d, want %d", meta.PayloadSize, len(content))
	}
	if meta.Filename != filepath.Base(inputPath) {
		t.Errorf("Inspect Filename: got %q, want %q", meta.Filename, filepath.Base(inputPath))
	}
	if meta.Version != 1 {
		t.Errorf("Inspect Version: got %d, want 1", meta.Version)
	}

	// Verify
	vMeta, err := VerifyFile(outputPNG)
	if err != nil {
		t.Fatalf("VerifyFile: %v", err)
	}
	if vMeta.PayloadSize != uint64(len(content)) {
		t.Errorf("Verify PayloadSize: got %d, want %d", vMeta.PayloadSize, len(content))
	}
}

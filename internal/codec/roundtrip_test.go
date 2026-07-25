package codec

import (
	"bytes"
	"crypto/sha256"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper: create a temporary file with given content.
func createTempFile(t *testing.T, dir, prefix string, content []byte) string {
	t.Helper()
	f, err := os.CreateTemp(dir, prefix)
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	if _, err := f.Write(content); err != nil {
		f.Close()
		t.Fatalf("Write: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	return f.Name()
}

func TestRoundTripBasic(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
	}{
		{"empty", []byte{}},
		{"one byte", []byte{0x42}},
		{"two bytes", []byte{0x42, 0x43}},
		{"three bytes", []byte{0x42, 0x43, 0x44}},
		{"four bytes", []byte{0x42, 0x43, 0x44, 0x45}},
		{"zeros", make([]byte, 100)},
		{"text", []byte("Hello, PixPack!")},
		{"unicode", []byte("日本語テスト\n✓")},
		{"json", []byte(`{"key": "value", "arr": [1,2,3]}`)},
		{"binary pattern", []byte{0x00, 0xFF, 0x55, 0xAA, 0x01, 0x02, 0x03, 0x04}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			inputPath := createTempFile(t, dir, "input-*", tt.content)

			outputPNG := filepath.Join(dir, "output.png")
			restoredPath := filepath.Join(dir, "restored")

			// Encode
			meta, err := EncodeFile(inputPath, outputPNG, EncodeOptions{Overwrite: false})
			if err != nil {
				t.Fatalf("EncodeFile failed: %v", err)
			}
			if meta.PayloadSize != uint64(len(tt.content)) {
				t.Errorf("PayloadSize: got %d, want %d", meta.PayloadSize, len(tt.content))
			}

			// Ensure PNG file was created and is valid.
			if _, err := os.Stat(outputPNG); os.IsNotExist(err) {
				t.Fatal("output PNG not created")
			}

			// Decode
			decodedMeta, err := DecodeFile(outputPNG, DecodeOptions{OutputPath: restoredPath, Overwrite: false})
			if err != nil {
				t.Fatalf("DecodeFile failed: %v", err)
			}
			if decodedMeta.PayloadSize != uint64(len(tt.content)) {
				t.Errorf("Decoded PayloadSize: got %d, want %d", decodedMeta.PayloadSize, len(tt.content))
			}

			// Compare restored file with original.
			restoredBytes, err := os.ReadFile(restoredPath)
			if err != nil {
				t.Fatalf("reading restored file: %v", err)
			}
			if !bytes.Equal(restoredBytes, tt.content) {
				t.Errorf("restored content differs from original")
			}

			// Compare SHA-256.
			origHash := sha256.Sum256(tt.content)
			restoredHash := sha256.Sum256(restoredBytes)
			if origHash != restoredHash {
				t.Error("SHA-256 mismatch between original and restored")
			}
		})
	}
}

func TestRoundTripPaddingCases(t *testing.T) {
	// Test exactly boundary cases for pixel alignment.
	for _, size := range []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15} {
		t.Run("", func(t *testing.T) {
			content := make([]byte, size)
			for i := range content {
				content[i] = byte(i * 7)
			}

			dir := t.TempDir()
			inputPath := createTempFile(t, dir, "input-*", content)
			outputPNG := filepath.Join(dir, "output.png")
			restoredPath := filepath.Join(dir, "restored")

			if _, err := EncodeFile(inputPath, outputPNG, EncodeOptions{Overwrite: false}); err != nil {
				t.Fatalf("EncodeFile(size=%d): %v", size, err)
			}
			if _, err := DecodeFile(outputPNG, DecodeOptions{OutputPath: restoredPath, Overwrite: false}); err != nil {
				t.Fatalf("DecodeFile(size=%d): %v", size, err)
			}

			restored, _ := os.ReadFile(restoredPath)
			if !bytes.Equal(restored, content) {
				t.Errorf("restored differs for size %d", size)
			}
		})
	}
}

func TestRoundTripLargeFile(t *testing.T) {
	// Test with a ~1MB file.
	size := 1 * 1024 * 1024
	content := make([]byte, size)
	for i := range content {
		content[i] = byte(i * 13)
	}

	dir := t.TempDir()
	inputPath := createTempFile(t, dir, "input-*", content)
	outputPNG := filepath.Join(dir, "large.png")
	restoredPath := filepath.Join(dir, "large.bin")

	if _, err := EncodeFile(inputPath, outputPNG, EncodeOptions{Overwrite: false}); err != nil {
		t.Fatalf("EncodeFile large: %v", err)
	}
	if _, err := DecodeFile(outputPNG, DecodeOptions{OutputPath: restoredPath, Overwrite: false}); err != nil {
		t.Fatalf("DecodeFile large: %v", err)
	}

	restored, _ := os.ReadFile(restoredPath)
	if !bytes.Equal(restored, content) {
		t.Error("large file round trip failed")
	}
}

func TestRoundTripZipFile(t *testing.T) {
	dir := t.TempDir()

	// Create a simple ZIP in memory.
	zipContent := []byte("fake-zip-content-that-represents-an-archive")
	inputPath := createTempFile(t, dir, "archive-*", zipContent)
	outputPNG := filepath.Join(dir, "archive.png")
	restoredPath := filepath.Join(dir, "restored.zip")

	if _, err := EncodeFile(inputPath, outputPNG, EncodeOptions{Overwrite: false}); err != nil {
		t.Fatalf("EncodeFile: %v", err)
	}
	if _, err := DecodeFile(outputPNG, DecodeOptions{OutputPath: restoredPath, Overwrite: false}); err != nil {
		t.Fatalf("DecodeFile: %v", err)
	}

	restored, _ := os.ReadFile(restoredPath)
	if !bytes.Equal(restored, zipContent) {
		t.Error("zip round trip failed")
	}
}

func TestRoundTripCustomWidth(t *testing.T) {
	content := []byte("Hello PixPack with custom width!")
	dir := t.TempDir()
	inputPath := createTempFile(t, dir, "input-*", content)
	outputPNG := filepath.Join(dir, "custom.png")
	restoredPath := filepath.Join(dir, "restored")

	if _, err := EncodeFile(inputPath, outputPNG, EncodeOptions{Width: 100, Overwrite: false}); err != nil {
		t.Fatalf("EncodeFile: %v", err)
	}
	meta, err := DecodeFile(outputPNG, DecodeOptions{OutputPath: restoredPath, Overwrite: false})
	if err != nil {
		t.Fatalf("DecodeFile: %v", err)
	}
	if meta.Width != 100 {
		t.Errorf("Width: got %d, want 100", meta.Width)
	}

	restored, _ := os.ReadFile(restoredPath)
	if !bytes.Equal(restored, content) {
		t.Error("custom width round trip failed")
	}
}

func TestOverwriteProtection(t *testing.T) {
	dir := t.TempDir()
	inputPath := createTempFile(t, dir, "input-*", []byte("test"))
	outputPNG := filepath.Join(dir, "output.png")

	// First encode succeeds.
	if _, err := EncodeFile(inputPath, outputPNG, EncodeOptions{Overwrite: false}); err != nil {
		t.Fatalf("first encode: %v", err)
	}

	// Second encode without overwrite fails.
	if _, err := EncodeFile(inputPath, outputPNG, EncodeOptions{Overwrite: false}); err != ErrOutputExists {
		t.Errorf("expected ErrOutputExists, got %v", err)
	}

	// Second encode with overwrite succeeds.
	if _, err := EncodeFile(inputPath, outputPNG, EncodeOptions{Overwrite: true}); err != nil {
		t.Errorf("encode with overwrite: %v", err)
	}
}

func TestDecodeOverwriteProtection(t *testing.T) {
	dir := t.TempDir()
	content := []byte("test data")
	inputPath := createTempFile(t, dir, "input-*", content)
	outputPNG := filepath.Join(dir, "output.png")
	restoredPath := filepath.Join(dir, "restored")

	if _, err := EncodeFile(inputPath, outputPNG, EncodeOptions{Overwrite: false}); err != nil {
		t.Fatalf("encode: %v", err)
	}

	// First decode succeeds.
	if _, err := DecodeFile(outputPNG, DecodeOptions{OutputPath: restoredPath, Overwrite: false}); err != nil {
		t.Fatalf("first decode: %v", err)
	}

	// Second decode without overwrite fails.
	if _, err := DecodeFile(outputPNG, DecodeOptions{OutputPath: restoredPath, Overwrite: false}); err != ErrOutputExists {
		t.Errorf("expected ErrOutputExists, got %v", err)
	}

	// Second decode with overwrite succeeds.
	if _, err := DecodeFile(outputPNG, DecodeOptions{OutputPath: restoredPath, Overwrite: true}); err != nil {
		t.Errorf("decode with overwrite: %v", err)
	}
}

func TestEncodeInputIsDirectory(t *testing.T) {
	dir := t.TempDir()
	_, err := EncodeFile(dir, filepath.Join(dir, "out.png"), EncodeOptions{})
	if err != ErrInputIsDir {
		t.Errorf("expected ErrInputIsDir, got %v", err)
	}
}

func TestDecodeNotPixPack(t *testing.T) {
	dir := t.TempDir()
	// Create a plain valid PNG that is not PixPack.
	plainPNG := filepath.Join(dir, "plain.png")
	plainContent := []byte("not a pixpack png")
	if err := os.WriteFile(plainPNG, plainContent, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Simply not a PNG at all.
	_, err := DecodeFile(plainPNG, DecodeOptions{})
	if err == nil {
		t.Error("expected error for non-PixPack file")
	}
}

func TestEncodeFileTooLarge(t *testing.T) {
	// We can't easily create a file > MaxPayloadSize, but we can check that
	// large-but-valid files work.
	dir := t.TempDir()
	content := make([]byte, 10*1024*1024) // 10MB
	inputPath := createTempFile(t, dir, "big-*", content)
	outputPNG := filepath.Join(dir, "big.png")

	if _, err := EncodeFile(inputPath, outputPNG, EncodeOptions{Overwrite: false}); err != nil {
		t.Fatalf("encoding 10MB file: %v", err)
	}
}

func TestFilenamePreserved(t *testing.T) {
	tests := []struct {
		inputName string
	}{
		{"hello.txt"},
		{"my file with spaces.zip"},
		{"unîcøde-文件名.txt"},
		{"simple"},
	}

	for _, tt := range tests {
		t.Run(tt.inputName, func(t *testing.T) {
			dir := t.TempDir()
			// Put the input file in a subdirectory so the default decode output
			// (which uses just the filename) won't conflict with the input.
			inputDir := filepath.Join(dir, "input")
			if err := os.Mkdir(inputDir, 0755); err != nil {
				t.Fatalf("Mkdir: %v", err)
			}
			inputPath := filepath.Join(inputDir, tt.inputName)
			if err := os.WriteFile(inputPath, []byte("test content"), 0644); err != nil {
				t.Fatalf("WriteFile: %v", err)
			}
			outputPNG := filepath.Join(dir, "output.png")

			meta, err := EncodeFile(inputPath, outputPNG, EncodeOptions{Overwrite: false})
			if err != nil {
				t.Fatalf("EncodeFile: %v", err)
			}
			if meta.Filename != filepath.Base(tt.inputName) {
				t.Errorf("stored filename: got %q, want %q", meta.Filename, filepath.Base(tt.inputName))
			}

			// Decode with explicit output path.
			restoredPath := filepath.Join(dir, tt.inputName)
			decodedMeta, err := DecodeFile(outputPNG, DecodeOptions{OutputPath: restoredPath, Overwrite: false})
			if err != nil {
				t.Fatalf("DecodeFile: %v", err)
			}
			if decodedMeta.Filename != filepath.Base(tt.inputName) {
				t.Errorf("decoded filename: got %q, want %q", decodedMeta.Filename, filepath.Base(tt.inputName))
			}

			// The file should exist.
			if _, err := os.Stat(restoredPath); os.IsNotExist(err) {
				t.Errorf("restored file not found at %s", restoredPath)
			}
		})
	}
}

func TestFilenameSanitization(t *testing.T) {
	tests := []string{
		"../../secret.txt",
		"/etc/passwd",
		"C:\\Windows\\system32\\evil.exe",
		"../parent",
		"..",
		".",
	}

	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			result := sanitizeFilename(name)
			if result != "" {
				t.Logf("sanitizeFilename(%q) = %q (expected empty)", name, result)
			}
			// The function should return empty or a safe base name.
			if strings.Contains(result, "..") {
				t.Errorf("sanitizeFilename(%q) still contains '..': %q", name, result)
			}
			if strings.Contains(result, "/") || strings.Contains(result, "\\") {
				t.Errorf("sanitizeFilename(%q) still contains path separator: %q", name, result)
			}
		})
	}
}

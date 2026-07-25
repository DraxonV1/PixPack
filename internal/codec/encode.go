package codec

import (
	"crypto/sha256"
	"fmt"
	"image/png"
	"os"
	"path/filepath"

	"github.com/DraxonV1/PixPack/internal/format"
)

// EncodeOptions controls the encoding behavior.
type EncodeOptions struct {
	Width     int
	Overwrite bool
}

// EncodeFile reads inputPath, creates a PixPack PNG at outputPath, and returns metadata.
func EncodeFile(inputPath, outputPath string, opts EncodeOptions) (*Metadata, error) {
	// Validate input.
	inputInfo, err := os.Stat(inputPath)
	if err != nil {
		return nil, fmt.Errorf("cannot access input file: %w", err)
	}
	if inputInfo.IsDir() {
		return nil, ErrInputIsDir
	}

	inputSize := inputInfo.Size()
	if inputSize < 0 {
		return nil, ErrFileTooLarge
	}

	// Read the entire input file.
	inputBytes, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read input file: %w", err)
	}
	if int64(len(inputBytes)) != inputSize {
		return nil, fmt.Errorf("input file size changed during read")
	}

	// Extract base filename.
	filename := filepath.Base(inputPath)
	if filename == "." || filename == "" {
		filename = "file"
	}
	filenameBytes := []byte(filename)
	if len(filenameBytes) > format.MaxFilenameLength {
		return nil, fmt.Errorf("filename too long: %d bytes (max %d)", len(filenameBytes), format.MaxFilenameLength)
	}

	// Calculate SHA-256 of payload.
	checksum := sha256.Sum256(inputBytes)

	// Build header.
	hdr := &format.Header{
		Version:     format.CurrentFormatVersion,
		Flags:       format.ReservedFlags,
		PayloadSize: uint64(len(inputBytes)),
		FilenameLen: uint16(len(filenameBytes)),
		Checksum:    checksum,
		Filename:    filename,
	}
	headerBytes, err := hdr.Serialize()
	if err != nil {
		return nil, fmt.Errorf("header serialization failed: %w", err)
	}

	// Combine header + payload + padding.
	totalBytes := len(headerBytes) + len(inputBytes)
	padding := (3 - totalBytes%3) % 3
	combined := make([]byte, totalBytes+padding)
	copy(combined, headerBytes)
	copy(combined[len(headerBytes):], inputBytes)
	// padding bytes are already zero from make()

	// Check output file.
	if !opts.Overwrite {
		if _, err := os.Stat(outputPath); err == nil {
			return nil, ErrOutputExists
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("cannot stat output path: %w", err)
		}
	}

	// Create image and encode PNG.
	img, err := CreateImage(combined, opts.Width)
	if err != nil {
		return nil, fmt.Errorf("cannot create image: %w", err)
	}

	// Write to a temporary file first.
	tmpFile, err := os.CreateTemp(filepath.Dir(outputPath), "pixpack-*.png")
	if err != nil {
		return nil, fmt.Errorf("cannot create temporary file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Clean up temp file on any error.
	success := false
	defer func() {
		if !success {
			os.Remove(tmpPath)
		}
	}()

	if err := png.Encode(tmpFile, img); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("%w: %w", ErrPNGEncodeFailed, err)
	}
	if err := tmpFile.Close(); err != nil {
		return nil, fmt.Errorf("cannot close temporary file: %w", err)
	}

	// Rename temporary file to final output path.
	if err := os.Rename(tmpPath, outputPath); err != nil {
		// On some platforms, rename across devices fails; try copy+delete.
		if err := os.Remove(outputPath); err != nil && !os.IsNotExist(err) {
			os.Remove(tmpPath)
			return nil, fmt.Errorf("cannot replace output file: %w", err)
		}
		if err := os.Rename(tmpPath, outputPath); err != nil {
			os.Remove(tmpPath)
			return nil, fmt.Errorf("cannot rename output file: %w", err)
		}
	}

	success = true

	meta := &Metadata{
		Version:     hdr.Version,
		Flags:       hdr.Flags,
		Filename:    hdr.Filename,
		PayloadSize: hdr.PayloadSize,
		Width:       img.Bounds().Dx(),
		Height:      img.Bounds().Dy(),
		SHA256:      checksum,
	}
	return meta, nil
}

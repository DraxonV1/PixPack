package codec

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DraxonV1/PixPack/internal/format"
)

// DecodeOptions controls the decoding behavior.
type DecodeOptions struct {
	OutputPath string
	Overwrite  bool
}

// DecodeFile reads a PixPack PNG and restores the original file.
func DecodeFile(inputPath string, opts DecodeOptions) (*Metadata, error) {
	// Open and decode the PNG.
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

	// Parse header from the first bytes.
	hdr, err := format.ParseHeader(pixelBytes)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNotPixPackImage, err)
	}

	headerSize := int(hdr.HeaderSize)
	if headerSize > len(pixelBytes) {
		return nil, fmt.Errorf("%w: header declares %d bytes but image only has %d", ErrUnsupportedImage, headerSize, len(pixelBytes))
	}

	payloadSize := int(hdr.PayloadSize)
	if payloadSize < 0 {
		return nil, fmt.Errorf("%w: negative payload size", ErrUnsupportedImage)
	}
	payloadEnd := headerSize + payloadSize
	if payloadEnd > len(pixelBytes) {
		return nil, ErrTruncatedPayload
	}

	// Extract payload bytes.
	payload := pixelBytes[headerSize : headerSize+payloadSize]

	// Verify checksum.
	actualChecksum := sha256.Sum256(payload)
	if subtle.ConstantTimeCompare(actualChecksum[:], hdr.Checksum[:]) != 1 {
		return nil, ErrChecksumMismatch
	}

	// Determine safe output filename.
	outputFilename := sanitizeFilename(hdr.Filename)
	if outputFilename == "" {
		return nil, ErrFilenameUnsafe
	}

	// Determine output path.
	outputPath := opts.OutputPath
	if outputPath == "" {
		outputPath = outputFilename
	}

	// Check if output exists.
	if !opts.Overwrite {
		if _, err := os.Stat(outputPath); err == nil {
			return nil, ErrOutputExists
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("cannot stat output path: %w", err)
		}
	}

	// Write to temp file first, then rename.
	tmpFile, err := os.CreateTemp(filepath.Dir(outputPath), "pixpack-*"+filepath.Ext(outputPath))
	if err != nil {
		return nil, fmt.Errorf("cannot create temporary file: %w", err)
	}
	tmpPath := tmpFile.Name()

	success := false
	defer func() {
		if !success {
			os.Remove(tmpPath)
		}
	}()

	if _, err := tmpFile.Write(payload); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("cannot write output: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return nil, fmt.Errorf("cannot close output: %w", err)
	}

	// Rename.
	if err := os.Rename(tmpPath, outputPath); err != nil {
		// Try remove + rename.
		if removeErr := os.Remove(outputPath); removeErr != nil && !os.IsNotExist(removeErr) {
			os.Remove(tmpPath)
			return nil, fmt.Errorf("cannot replace output file: %w", err)
		}
		if err := os.Rename(tmpPath, outputPath); err != nil {
			os.Remove(tmpPath)
			return nil, fmt.Errorf("cannot finalize output file: %w", err)
		}
	}

	success = true

	meta := &Metadata{
		Version:     hdr.Version,
		Flags:       hdr.Flags,
		Filename:    hdr.Filename,
		PayloadSize: hdr.PayloadSize,
		Width:       width,
		Height:      height,
		SHA256:      actualChecksum,
	}
	return meta, nil
}

// sanitizeFilename ensures the filename is safe for writing.
func sanitizeFilename(filename string) string {
	// Use filepath.Base to get only the final component.
	base := filepath.Base(filename)
	if base == "." || base == ".." || base == "" {
		return ""
	}
	// Reject if filepath.Base changed the name (indicates directory traversal).
	if base != filename {
		// Only allow if the original was a simple name (no path separators).
		if strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
			return ""
		}
	}
	// Reject any remaining dangerous characters.
	if strings.ContainsAny(base, "<>:\"|?*") {
		return ""
	}
	return base
}

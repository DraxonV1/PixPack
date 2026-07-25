package codec

import (
	"os"
	"testing"

	"github.com/DraxonV1/PixPack/internal/format"
)

func FuzzParseHeader(f *testing.F) {
	// Seed with valid data.
	hdr := &format.Header{
		Version:     format.CurrentFormatVersion,
		Flags:       format.ReservedFlags,
		PayloadSize: 100,
		FilenameLen: 4,
		Filename:    "test",
	}
	seed, err := hdr.Serialize()
	if err != nil {
		f.Fatal(err)
	}
	f.Add(seed)

	f.Fuzz(func(t *testing.T, data []byte) {
		// Must never panic.
		_, _ = format.ParseHeader(data)
	})
}

func FuzzDecodeBytes(f *testing.F) {
	// Seed with a valid small PixPack PNG.
	dir := f.TempDir()
	inputPath := dir + "/input"
	outputPath := dir + "/out.png"
	if err := os.WriteFile(inputPath, []byte("fuzz test"), 0644); err != nil {
		f.Fatal(err)
	}
	if _, err := EncodeFile(inputPath, outputPath, EncodeOptions{Overwrite: false}); err != nil {
		f.Fatal(err)
	}
	pngData, err := os.ReadFile(outputPath)
	if err != nil {
		f.Fatal(err)
	}
	f.Add(pngData)

	f.Fuzz(func(t *testing.T, data []byte) {
		// Write to temp file and try to decode.
		tmpDir := t.TempDir()
		tmpPath := tmpDir + "/fuzz.png"
		if err := os.WriteFile(tmpPath, data, 0644); err != nil {
			return
		}
		// Must never panic.
		_, _ = DecodeFile(tmpPath, DecodeOptions{})
		_, _ = InspectFile(tmpPath)
		_, _ = VerifyFile(tmpPath)
	})
}

package codec

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/DraxonV1/PixPack/internal/format"
)

func makeBenchData(size int) []byte {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i * 7)
	}
	return data
}

func BenchmarkHeaderConstruction(b *testing.B) {
	data := makeBenchData(1024)
	checksum := sha256.Sum256(data)
	filename := "benchmark-test-file.bin"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hdr := &format.Header{
			Version:     format.CurrentFormatVersion,
			Flags:       format.ReservedFlags,
			PayloadSize: uint64(len(data)),
			FilenameLen: uint16(len(filename)),
			Checksum:    checksum,
			Filename:    filename,
		}
		if _, err := hdr.Serialize(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBytesToPixels(b *testing.B) {
	sizes := []int{1024, 100 * 1024, 1024 * 1024}
	for _, size := range sizes {
		data := makeBenchData(size)
		b.Run(fmt.Sprintf("%dKB", size/1024), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = BytesToPixels(data)
			}
		})
	}
}

func BenchmarkPixelsToBytes(b *testing.B) {
	sizes := []int{1024, 100 * 1024, 1024 * 1024}
	for _, size := range sizes {
		data := makeBenchData(size)
		pixels := BytesToPixels(data)
		b.Run(fmt.Sprintf("%dKB", size/1024), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = PixelsToBytes(pixels, len(data))
			}
		})
	}
}

func BenchmarkEncodeFile(b *testing.B) {
	sizes := []int{1024, 100 * 1024}
	for _, size := range sizes {
		data := makeBenchData(size)
		b.Run(fmt.Sprintf("%dKB", size/1024), func(b *testing.B) {
			dir := b.TempDir()
			inputPath := filepath.Join(dir, "input")
			if err := os.WriteFile(inputPath, data, 0644); err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				outputPath := filepath.Join(dir, fmt.Sprintf("out-%d.png", i))
				if _, err := EncodeFile(inputPath, outputPath, EncodeOptions{Overwrite: false}); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkDecodeFile(b *testing.B) {
	sizes := []int{1024, 100 * 1024}
	for _, size := range sizes {
		data := makeBenchData(size)
		b.Run(fmt.Sprintf("%dKB", size/1024), func(b *testing.B) {
			dir := b.TempDir()
			inputPath := filepath.Join(dir, "input")
			if err := os.WriteFile(inputPath, data, 0644); err != nil {
				b.Fatal(err)
			}
			pngPath := filepath.Join(dir, "encoded.png")
			if _, err := EncodeFile(inputPath, pngPath, EncodeOptions{Overwrite: false}); err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				restorePath := filepath.Join(dir, fmt.Sprintf("restored-%d", i))
				if _, err := DecodeFile(pngPath, DecodeOptions{OutputPath: restorePath, Overwrite: false}); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

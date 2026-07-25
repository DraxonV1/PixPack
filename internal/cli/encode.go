package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/DraxonV1/PixPack/internal/codec"
)

func runEncode(args []string) int {
	fs := flag.NewFlagSet("encode", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: pixpack encode [--width pixels] [--overwrite] <input-file> <output.png>\n")
		fs.PrintDefaults()
	}

	width := fs.Int("width", 0, "Image width in pixels (auto-calculated if not specified)")
	overwrite := fs.Bool("overwrite", false, "Overwrite existing output file")

	if err := fs.Parse(args); err != nil {
		return ExitUsage
	}

	if fs.NArg() != 2 {
		fs.Usage()
		return ExitUsage
	}

	inputPath := fs.Arg(0)
	outputPath := fs.Arg(1)

	opts := codec.EncodeOptions{
		Width:     *width,
		Overwrite: *overwrite,
	}

	meta, err := codec.EncodeFile(inputPath, outputPath, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		switch {
		case errors.Is(err, codec.ErrInputIsDir),
			errors.Is(err, codec.ErrFileTooLarge),
			errors.Is(err, codec.ErrOutputExists):
			return ExitError
		default:
			return ExitError
		}
	}

	fmt.Printf("Encoded successfully\n\n")
	fmt.Printf("Input:        %s\n", inputPath)
	fmt.Printf("Input size:   %s\n", formatBytes(meta.PayloadSize))
	fmt.Printf("Output:       %s\n", outputPath)
	fmt.Printf("Image size:   %d x %d\n", meta.Width, meta.Height)
	fmt.Printf("SHA-256:      %s\n", formatHash(meta.SHA256))

	return ExitSuccess
}

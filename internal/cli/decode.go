package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/DraxonV1/PixPack/internal/codec"
)

func runDecode(args []string) int {
	fs := flag.NewFlagSet("decode", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: pixpack decode [--output file] [--overwrite] <input.png>\n")
		fs.PrintDefaults()
	}

	outputPath := fs.String("output", "", "Output file path (defaults to stored filename)")
	overwrite := fs.Bool("overwrite", false, "Overwrite existing output file")

	if err := fs.Parse(args); err != nil {
		return ExitUsage
	}

	if fs.NArg() != 1 {
		fs.Usage()
		return ExitUsage
	}

	inputPath := fs.Arg(0)

	opts := codec.DecodeOptions{
		OutputPath: *outputPath,
		Overwrite:  *overwrite,
	}

	meta, err := codec.DecodeFile(inputPath, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		switch {
		case errors.Is(err, codec.ErrChecksumMismatch):
			return ExitCorrupt
		case errors.Is(err, codec.ErrNotPixPackImage),
			errors.Is(err, codec.ErrUnsupportedImage),
			errors.Is(err, codec.ErrTruncatedPayload),
			errors.Is(err, codec.ErrFilenameUnsafe),
			errors.Is(err, codec.ErrPNGDecodeFailed):
			return ExitInvalid
		default:
			return ExitError
		}
	}

	outputDisplay := *outputPath
	if outputDisplay == "" {
		outputDisplay = meta.Filename
	}

	fmt.Printf("Decoded successfully\n\n")
	fmt.Printf("Input:        %s\n", inputPath)
	fmt.Printf("Output:       %s\n", outputDisplay)
	fmt.Printf("Output size:  %s\n", formatBytes(meta.PayloadSize))
	fmt.Printf("Integrity:    valid\n")
	fmt.Printf("SHA-256:      %s\n", formatHash(meta.SHA256))

	return ExitSuccess
}

package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/DraxonV1/PixPack/internal/codec"
)

func runInspect(args []string) int {
	fs := flag.NewFlagSet("inspect", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: pixpack inspect <input.png>\n")
	}

	if err := fs.Parse(args); err != nil {
		return ExitUsage
	}

	if fs.NArg() != 1 {
		fs.Usage()
		return ExitUsage
	}

	inputPath := fs.Arg(0)

	meta, err := codec.InspectFile(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		switch {
		case errors.Is(err, codec.ErrNotPixPackImage),
			errors.Is(err, codec.ErrUnsupportedImage),
			errors.Is(err, codec.ErrPNGDecodeFailed):
			return ExitInvalid
		default:
			return ExitError
		}
	}

	fmt.Println("PixPack image")
	fmt.Println()
	fmt.Printf("Format version:  %d\n", meta.Version)
	fmt.Printf("Filename:        %s\n", meta.Filename)
	fmt.Printf("Payload size:    %s\n", formatBytes(meta.PayloadSize))
	fmt.Printf("Image size:      %d x %d\n", meta.Width, meta.Height)
	fmt.Printf("Checksum:        SHA-256\n")
	fmt.Printf("SHA-256:         %s\n", formatHash(meta.SHA256))

	return ExitSuccess
}

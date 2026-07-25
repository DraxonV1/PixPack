package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/DraxonV1/PixPack/internal/codec"
)

func runVerify(args []string) int {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: pixpack verify <input.png>\n")
	}

	if err := fs.Parse(args); err != nil {
		return ExitUsage
	}

	if fs.NArg() != 1 {
		fs.Usage()
		return ExitUsage
	}

	inputPath := fs.Arg(0)

	_, err := codec.VerifyFile(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		switch {
		case errors.Is(err, codec.ErrChecksumMismatch):
			return ExitCorrupt
		case errors.Is(err, codec.ErrNotPixPackImage),
			errors.Is(err, codec.ErrUnsupportedImage),
			errors.Is(err, codec.ErrTruncatedPayload),
			errors.Is(err, codec.ErrPNGDecodeFailed):
			return ExitInvalid
		default:
			return ExitError
		}
	}

	fmt.Println("Integrity check passed.")
	return ExitSuccess
}

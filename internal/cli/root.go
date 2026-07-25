package cli

import (
	"fmt"
	"os"
	"strings"
)

const Version = "0.1.0"

// Exit codes.
const (
	ExitSuccess = 0
	ExitError   = 1
	ExitUsage   = 2
	ExitInvalid = 3
	ExitCorrupt = 4
)

// Run dispatches CLI commands.
func Run(args []string) int {
	if len(args) < 2 {
		printHelp()
		return ExitUsage
	}

	cmd := args[1]

	switch cmd {
	case "encode":
		return runEncode(args[2:])
	case "decode":
		return runDecode(args[2:])
	case "inspect":
		return runInspect(args[2:])
	case "verify":
		return runVerify(args[2:])
	case "--help", "-h", "help":
		printHelp()
		return ExitSuccess
	case "--version", "-v", "version":
		fmt.Printf("pixpack version %s\n", Version)
		return ExitSuccess
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		printHelp()
		return ExitUsage
	}
}

func printHelp() {
	fmt.Print(`PixPack — Convert any file into a lossless PNG image and restore it byte-for-byte.

Usage:
  pixpack <command> [arguments]

Commands:
  encode <input-file> <output.png>    Encode a file into a PixPack PNG
  decode <input.png> [--output file]  Decode a PixPack PNG back to the original file
  inspect <input.png>                 Show PixPack image metadata without extracting
  verify <input.png>                  Verify payload integrity without extracting

Flags:
  --help, -h     Show this help
  --version, -v  Show version

Use "pixpack <command> --help" for more information about a command.
`)
}

func printUsage(cmd string, usage string) {
	fmt.Fprintf(os.Stderr, "Usage: pixpack %s\n", usage)
}

func formatBytes(n uint64) string {
	const unit = 1000
	if n < unit {
		return fmt.Sprintf("%d bytes", n)
	}
	div, exp := uint64(unit), 0
	for n >= div*unit && exp < 5 {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "kMGTPE"[exp])
}

func formatHash(checksum [32]byte) string {
	return fmt.Sprintf("%x", checksum)
}

// isFlag checks if a string is a flag (starts with -).
func isFlag(s string) bool {
	return strings.HasPrefix(s, "-")
}

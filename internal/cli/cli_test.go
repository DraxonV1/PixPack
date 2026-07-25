package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCLIEncodeDecode(t *testing.T) {
	dir := t.TempDir()

	// Create input file.
	inputPath := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(inputPath, []byte("Hello PixPack CLI!"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	outputPNG := filepath.Join(dir, "test.png")
	restoredPath := filepath.Join(dir, "restored.txt")

	// Encode.
	exitCode := Run([]string{"pixpack", "encode", inputPath, outputPNG})
	if exitCode != ExitSuccess {
		t.Fatalf("encode exit code: %d", exitCode)
	}

	// Inspect.
	exitCode = Run([]string{"pixpack", "inspect", outputPNG})
	if exitCode != ExitSuccess {
		t.Errorf("inspect exit code: %d", exitCode)
	}

	// Verify.
	exitCode = Run([]string{"pixpack", "verify", outputPNG})
	if exitCode != ExitSuccess {
		t.Errorf("verify exit code: %d", exitCode)
	}

	// Decode.
	exitCode = Run([]string{"pixpack", "decode", "--output", restoredPath, outputPNG})
	if exitCode != ExitSuccess {
		t.Fatalf("decode exit code: %d", exitCode)
	}

	// Verify restored content.
	restored, err := os.ReadFile(restoredPath)
	if err != nil {
		t.Fatalf("reading restored: %v", err)
	}
	if string(restored) != "Hello PixPack CLI!" {
		t.Errorf("restored content: got %q", string(restored))
	}
}

func TestCLIMissingArgs(t *testing.T) {
	tests := []struct {
		args []string
	}{
		{[]string{"pixpack"}},
		{[]string{"pixpack", "encode"}},
		{[]string{"pixpack", "encode", "only-one-arg"}},
		{[]string{"pixpack", "decode"}},
		{[]string{"pixpack", "inspect"}},
		{[]string{"pixpack", "verify"}},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			exitCode := Run(tt.args)
			if exitCode != ExitUsage {
				t.Errorf("args %v: expected exit code %d, got %d", tt.args, ExitUsage, exitCode)
			}
		})
	}
}

func TestCLIInvalidCommand(t *testing.T) {
	exitCode := Run([]string{"pixpack", "invalid-command"})
	if exitCode != ExitUsage {
		t.Errorf("expected exit code %d, got %d", ExitUsage, exitCode)
	}
}

func TestCLIVersion(t *testing.T) {
	exitCode := Run([]string{"pixpack", "--version"})
	if exitCode != ExitSuccess {
		t.Errorf("expected exit code %d, got %d", ExitSuccess, exitCode)
	}
}

func TestCLIHelp(t *testing.T) {
	exitCode := Run([]string{"pixpack", "--help"})
	if exitCode != ExitSuccess {
		t.Errorf("expected exit code %d, got %d", ExitSuccess, exitCode)
	}
}

func TestCLIEncodeOverwrite(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(inputPath, []byte("content"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	outputPNG := filepath.Join(dir, "test.png")

	// First encode succeeds.
	exitCode := Run([]string{"pixpack", "encode", inputPath, outputPNG})
	if exitCode != ExitSuccess {
		t.Fatalf("first encode: %d", exitCode)
	}

	// Second encode without overwrite fails.
	exitCode = Run([]string{"pixpack", "encode", inputPath, outputPNG})
	if exitCode != ExitError {
		t.Errorf("expected exit %d, got %d", ExitError, exitCode)
	}

	// With --overwrite succeeds.
	exitCode = Run([]string{"pixpack", "encode", "--overwrite", inputPath, outputPNG})
	if exitCode != ExitSuccess {
		t.Errorf("encode with overwrite: %d", exitCode)
	}
}

func TestCLIDecodeOverwrite(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(inputPath, []byte("content"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	outputPNG := filepath.Join(dir, "test.png")
	restoredPath := filepath.Join(dir, "restored.txt")

	Run([]string{"pixpack", "encode", inputPath, outputPNG})

	// First decode succeeds.
	exitCode := Run([]string{"pixpack", "decode", "--output", restoredPath, outputPNG})
	if exitCode != ExitSuccess {
		t.Fatalf("first decode: %d", exitCode)
	}

	// Second decode without overwrite fails.
	exitCode = Run([]string{"pixpack", "decode", "--output", restoredPath, outputPNG})
	if exitCode != ExitError {
		t.Errorf("expected exit %d, got %d", ExitError, exitCode)
	}

	// With --overwrite succeeds.
	exitCode = Run([]string{"pixpack", "decode", "--overwrite", "--output", restoredPath, outputPNG})
	if exitCode != ExitSuccess {
		t.Errorf("decode with overwrite: %d", exitCode)
	}
}

func TestCLIEncodeCustomWidth(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(inputPath, []byte("width test data here"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	outputPNG := filepath.Join(dir, "width.png")

	exitCode := Run([]string{"pixpack", "encode", "--width", "50", inputPath, outputPNG})
	if exitCode != ExitSuccess {
		t.Errorf("encode with width: %d", exitCode)
	}
}

func TestCLIDecodeNotPixPack(t *testing.T) {
	dir := t.TempDir()
	// Create a non-PixPack PNG.
	plainPNG := filepath.Join(dir, "plain.png")
	// Just write something that isn't a PixPack PNG.
	img, _ := os.Create(plainPNG)
	img.Write([]byte("not a png"))
	img.Close()

	exitCode := Run([]string{"pixpack", "decode", plainPNG})
	if exitCode != ExitInvalid && exitCode != ExitError {
		t.Errorf("expected exit %d or %d, got %d", ExitInvalid, ExitError, exitCode)
	}
}

# PixPack

<img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="MIT License">

**PixPack converts any file into a lossless PNG image and restores it byte-for-byte using a small, documented, open container format.**

```bash
pixpack encode project.zip project.png
pixpack decode project.png restored.zip
```

The resulting PNG looks like colored noise, but every byte of the original file is preserved in the RGB pixel channels. PNG's own lossless DEFLATE compression is applied to the image data.

---

## Features

- **Encode** any file (ZIP, PDF, image, audio, video, executable, source code, text, etc.) into a PNG
- **Decode** the PNG back to the original file, byte-for-byte
- **Inspect** metadata without extracting the payload
- **Verify** payload integrity without extracting
- **Cross-platform**: Windows, Linux, macOS
- **No external dependencies**: single, statically-compiled binary
- **SHA-256 integrity verification** with constant-time comparison
- **Safe filename handling**: directory traversal and unsafe names are rejected
- **Overwrite protection**: refuses to overwrite existing files unless `--overwrite` is specified

## How It Works

PixPack reads the input file as raw bytes and stores them in the RGB channels of each pixel:

- Red channel = byte 1
- Green channel = byte 2
- Blue channel = byte 3

The byte stream is prefixed with a binary header containing the original filename, payload length, and SHA-256 checksum. The combined data is converted to pixels, then encoded as a lossless PNG.

For a detailed technical specification, see [docs/FORMAT.md](docs/FORMAT.md).

## Installation

### From source (requires Go 1.24+)

```bash
git clone https://github.com/DraxonV1/PixPack.git
cd pixpack
go build -o pixpack ./cmd/pixpack
```

### Pre-built binaries

Download the latest release from the [Releases](https://github.com/DraxonV1/PixPack/releases) page:

| Platform | Architecture | Archive |
|----------|-------------|---------|
| Linux    | amd64       | `pixpack_<version>_linux_amd64.tar.gz` |
| Linux    | arm64       | `pixpack_<version>_linux_arm64.tar.gz` |
| Windows  | amd64       | `pixpack_<version>_windows_amd64.zip` |
| Windows  | arm64       | `pixpack_<version>_windows_arm64.zip` |
| macOS    | amd64       | `pixpack_<version>_darwin_amd64.tar.gz` |
| macOS    | arm64       | `pixpack_<version>_darwin_arm64.tar.gz` |

Each release includes a SHA-256 checksum file for verification.

## Usage

### Encode a file

```bash
pixpack encode secret.zip secret.png
```

With custom width:

```bash
pixpack encode secret.zip secret.png --width 512
```

Overwrite existing output:

```bash
pixpack encode secret.zip secret.png --overwrite
```

### Decode a PNG

```bash
pixpack decode secret.png
```

This restores the file using the original filename. To specify a different output path:

```bash
pixpack decode secret.png --output restored-secret.zip
```

### Inspect metadata

```bash
pixpack inspect secret.png
```

Example output:

```
PixPack image

Format version:  1
Filename:        secret.zip
Payload size:    174,080 bytes
Image size:      241 x 241
Checksum:        SHA-256
SHA-256:         3c88...
```

### Verify integrity

```bash
pixpack verify secret.png
```

Output for a valid file:

```
Integrity check passed.
```

### Get help

```bash
pixpack --help
pixpack --version
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0    | Success |
| 1    | General error |
| 2    | Invalid CLI usage |
| 3    | Invalid or unsupported PixPack image |
| 4    | Checksum verification failure |

## Supported File Types

PixPack supports **any file** representable as bytes — there are no format restrictions:

- ZIP, TAR, GZIP archives
- Source code (any language)
- PDF documents
- Images (PNG, JPEG, GIF, WebP, etc.)
- Audio (MP3, FLAC, WAV, etc.)
- Video (MP4, MKV, AVI, etc.)
- Executables (PE, ELF, Mach-O)
- JSON, XML, YAML
- Text files
- Databases
- Any other file

## Limitations

1. **PixPack is NOT encryption.** The data is stored as-is in PNG pixels. Anyone with PixPack can restore the file. Encrypt sensitive data separately before encoding.
2. **PixPack is NOT steganography.** The image looks like colored noise and clearly contains structured data. It is not designed to hide the presence of data.
3. **Resizing the PNG will destroy the payload.** Pixel dimensions must remain unchanged.
4. **Cropping or editing pixels will likely cause checksum failure.**
5. **Converting to JPEG or any lossy format will destroy the payload.** Only lossless PNG is supported.
6. **Some websites and chat applications may recompress uploaded images.** Uploading the PNG as a document/file is safer than uploading it as a photo.
7. **The resulting PNG may be larger than the original file.** PNG's lossless compression adds overhead, especially for already-compressed data (ZIP, video, encrypted data).
8. **Huge files produce huge image dimensions** and require substantial memory. A 100 MB file may produce an image around 6000x6000 pixels (~36 megapixels), requiring significant RAM during encoding and decoding.

## Development

### Prerequisites

- Go 1.24 or later

### Commands

```bash
go build -o pixpack ./cmd/pixpack   # Build
go test ./...                        # Run tests
go test -race ./...                  # Run tests with race detector
go vet ./...                         # Static analysis
go fmt ./...                         # Format code
go test -bench=. ./...               # Run benchmarks
```

### End-to-end test

```bash
go build -o pixpack ./cmd/pixpack
echo "Hello, PixPack!" > testdata/hello.txt

./pixpack encode testdata/hello.txt hello.png
./pixpack inspect hello.png
./pixpack verify hello.png
./pixpack decode hello.png --output restored.txt

# Compare (Unix)
cmp testdata/hello.txt restored.txt

# Compare SHA-256 (Unix)
sha256sum testdata/hello.txt restored.txt
```

## License

MIT License — see [LICENSE](LICENSE).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Security

See [SECURITY.md](SECURITY.md) for reporting vulnerabilities.

# PixPack — Implementation Plan

## 1. Project Overview

Build an open-source command-line tool named **PixPack** that converts any file into a valid PNG image and can later restore the original file byte-for-byte.

Example:

```bash
pixpack encode project.zip project.png
pixpack decode project.png restored.zip
```

PixPack must support arbitrary binary files, not only text files.

Supported examples include:

* ZIP archives
* Source code
* PDFs
* Images
* Audio
* Video
* Executables
* JSON
* Text files
* Any other file represented as bytes

The initial version should be powerful and production-quality without becoming unnecessarily complicated.

---

## 2. Core Principle

PixPack reads the input file as raw bytes and stores those bytes inside RGB pixel channels.

Each RGB pixel stores three bytes:

```text
Red channel   = byte 1
Green channel = byte 2
Blue channel  = byte 3
```

Encoding:

```text
Original file
    ↓
Read bytes
    ↓
Create PixPack header
    ↓
Header + payload
    ↓
Convert every 3 bytes into one RGB pixel
    ↓
Write lossless PNG
```

Decoding:

```text
PixPack PNG
    ↓
Read RGB pixels
    ↓
Convert RGB channels back into bytes
    ↓
Parse PixPack header
    ↓
Extract exact payload length
    ↓
Verify checksum
    ↓
Write original file
```

PNG must be used because it preserves exact pixel values.

JPEG and other lossy formats must not be supported for encoded output.

---

## 3. Technology

Use:

```text
Language: Go
Minimum Go version: Go 1.24
License: MIT
```

Prefer the Go standard library wherever practical.

Useful standard-library packages include:

```go
image
image/color
image/png
crypto/sha256
encoding/binary
encoding/hex
flag
os
io
path/filepath
math
bytes
fmt
errors
```

Avoid adding dependencies unless they provide clear value.

The final CLI should compile into a single executable with no external runtime requirements.

---

## 4. Main Goals

The first stable release must provide:

1. Encode any file into PNG.
2. Decode a PixPack PNG into the original file.
3. Preserve the original filename.
4. Preserve the original file extension.
5. Restore the file byte-for-byte.
6. Detect corrupted payloads.
7. Reject ordinary PNG files that are not PixPack files.
8. Inspect metadata without extracting.
9. Verify payload integrity without extracting.
10. Work on Windows, Linux, and macOS.
11. Handle large files without unnecessary memory duplication.
12. Include proper tests and documentation.

---

## 5. Non-Goals for Version 1

Do not initially implement:

* Custom cryptography
* Steganography inside normal photographs
* Web interfaces
* Desktop GUI
* Cloud storage
* Accounts or authentication
* Networking
* Multiple files inside one PNG
* Automatic folder archiving
* Plugins
* Custom PNG metadata chunks
* JPEG support
* Mobile applications

These can be considered after the core format is stable.

Version 1 should encode exactly one file into one PNG.

Users can already place folders into ZIP or TAR archives before encoding.

---

## 6. CLI Design

The executable name should be:

```bash
pixpack
```

### Encode

```bash
pixpack encode <input-file> <output.png>
```

Example:

```bash
pixpack encode project.zip project.png
```

Optional flags:

```bash
pixpack encode project.zip project.png --width 512
pixpack encode project.zip project.png --overwrite
```

Behavior:

* Refuse to overwrite an existing output file unless `--overwrite` is used.
* Validate that the input is a regular file.
* Automatically calculate image dimensions unless width is specified.
* Print useful result information.

Example output:

```text
Encoded successfully

Input:        project.zip
Input size:   174,080 bytes
Output:       project.png
Image size:   241 × 241
SHA-256:      3c88...
```

### Decode

```bash
pixpack decode <input.png>
```

By default, restore the original filename into the current directory.

Example:

```bash
pixpack decode project.png
```

Optional output path:

```bash
pixpack decode project.png --output restored-project.zip
```

Additional option:

```bash
pixpack decode project.png --overwrite
```

Behavior:

* Confirm that the PNG contains a PixPack header.
* Read the stored filename and payload length.
* Extract exactly the stored number of payload bytes.
* Verify the SHA-256 checksum before writing the final output.
* Refuse to overwrite an existing file unless requested.
* Never leave a corrupted partial output file.

Example output:

```text
Decoded successfully

Input:        project.png
Output:       project.zip
Output size:  174,080 bytes
Integrity:    valid
SHA-256:      3c88...
```

### Inspect

```bash
pixpack inspect <input.png>
```

This command must not extract the file.

Example output:

```text
PixPack image

Format version:  1
Filename:        project.zip
Payload size:    174,080 bytes
Image size:      241 × 241
Checksum:        SHA-256
SHA-256:         3c88...
```

### Verify

```bash
pixpack verify <input.png>
```

This command should:

1. Parse the PixPack header.
2. Extract the payload in memory or through a stream.
3. Calculate its SHA-256 hash.
4. Compare it with the stored hash.
5. Exit successfully only when the payload is valid.

Example:

```text
Integrity check passed.
```

Corrupted file:

```text
Integrity check failed: payload checksum does not match.
```

---

## 7. Binary Format Specification

Create and document a stable PixPack version 1 binary format.

Before converting bytes into pixels, construct this byte sequence:

```text
[Fixed header]
[UTF-8 filename]
[Payload bytes]
[Zero padding]
```

Use big-endian byte order for all integers.

### Version 1 Header

|    Offset |         Size | Field                    |
| --------: | -----------: | ------------------------ |
|         0 |      4 bytes | Magic bytes              |
|         4 |       1 byte | Format version           |
|         5 |       1 byte | Flags                    |
|         6 |      2 bytes | Header size              |
|         8 |      8 bytes | Original payload size    |
|        16 |      2 bytes | Filename byte length     |
|        18 |     32 bytes | SHA-256 payload checksum |
|        50 |      N bytes | UTF-8 filename           |
|    50 + N | Payload size | Original file bytes      |
| Remaining |    0–2 bytes | Zero padding             |

Magic bytes:

```text
PXPK
```

Hex representation:

```text
50 58 50 4B
```

Version:

```text
1
```

Flags for version 1:

```text
0
```

Reserved flag bits must remain zero.

Header size:

```text
50 + filename byte length
```

Payload size:

```text
Unsigned 64-bit integer
```

Filename length:

```text
Unsigned 16-bit integer
```

Checksum:

```text
Raw 32-byte SHA-256 digest
```

Do not store the checksum as a hexadecimal string inside the binary format.

---

## 8. Filename Handling

The filename must be stored as UTF-8.

Store only the base filename, not the full original path.

Example:

```text
/home/user/projects/secret.zip
```

Stored filename:

```text
secret.zip
```

When decoding:

* Reject empty filenames.
* Reject absolute paths.
* Remove directory traversal components.
* Never allow the stored filename to write outside the requested destination.
* Reject or sanitize names such as:

```text
../../secret.txt
C:\Windows\file.txt
/etc/passwd
```

Use `filepath.Base`, but also validate behavior across platforms.

The decoder must treat filenames from encoded images as untrusted input.

---

## 9. Image Dimension Calculation

Each RGB pixel stores three bytes.

Calculate:

```text
totalBytes = headerSize + payloadSize
requiredPixels = ceil(totalBytes / 3)
```

When no width is specified, create a near-square image:

```text
width  = ceil(sqrt(requiredPixels))
height = ceil(requiredPixels / width)
```

Example for approximately 170 KB:

```text
required pixels ≈ 58,000
dimensions ≈ 241 × 241
```

When the user supplies a width:

```text
height = ceil(requiredPixels / width)
```

Validate:

* Width must be greater than zero.
* Width and height must fit into Go `int`.
* Dimensions must be accepted by the PNG encoder.
* Multiplication must not overflow.
* Reject impossible or unreasonable dimensions cleanly.

Unused RGB channels at the end must contain zero bytes.

---

## 10. Pixel Mapping

Pixels must be written row by row:

```text
top to bottom
left to right
```

Byte mapping:

```text
byte 0 → pixel 0 red
byte 1 → pixel 0 green
byte 2 → pixel 0 blue

byte 3 → pixel 1 red
byte 4 → pixel 1 green
byte 5 → pixel 1 blue
```

Use an opaque image.

Alpha must always be:

```text
255
```

Recommended Go image type:

```go
image.NewNRGBA(...)
```

Only the RGB channels are part of the PixPack byte stream.

Alpha must not contain payload data in format version 1.

---

## 11. Encoding Requirements

The encoder must:

1. Validate the input path.
2. Reject directories.
3. Read file metadata.
4. Reject files too large for the current platform or format.
5. Extract the base filename.
6. Validate the filename byte length.
7. Calculate SHA-256 over the original file bytes.
8. Build the PixPack header.
9. Calculate image dimensions safely.
10. Write the header and payload into RGB channels.
11. Add zero padding when required.
12. Save the image using lossless PNG encoding.
13. Close files correctly.
14. Delete incomplete output files when encoding fails.
15. Return descriptive errors.

The hash must cover only the original payload bytes.

It must not cover:

* Header
* Filename
* Padding
* PNG structure

---

## 12. Decoding Requirements

The decoder must:

1. Open and decode the PNG.
2. Convert the image into a predictable RGB-readable representation.
3. Read RGB channels in row-major order.
4. Verify the magic bytes.
5. Verify the supported format version.
6. Reject unsupported flags.
7. Parse all numeric fields safely.
8. Validate header size.
9. Validate filename length.
10. Validate payload size.
11. Ensure the image contains enough bytes.
12. Extract only the declared payload length.
13. Ignore final padding bytes.
14. Calculate SHA-256 over the extracted payload.
15. Compare hashes securely.
16. Determine a safe output filename.
17. Write to a temporary file.
18. Rename the temporary file only after verification succeeds.
19. Delete the temporary file after any failure.

Never trust lengths stored in the image without bounds checks.

Malformed files must return errors instead of panicking.

---

## 13. Integrity and Corruption Detection

SHA-256 must be used to verify byte-for-byte restoration.

The following guarantee should hold:

```text
SHA256(original file) == SHA256(decoded file)
```

Changing even one payload channel should normally make verification fail.

The tool must distinguish errors where practical:

```text
Not a PixPack image
Unsupported PixPack version
Unsupported flags
Malformed PixPack header
Invalid filename
Truncated payload
Checksum mismatch
PNG decoding failed
Output file already exists
```

Define sentinel errors or typed errors inside the core package where useful.

---

## 14. Repository Structure

Use a clean Go project structure without unnecessary packages.

```text
pixpack/
├── cmd/
│   └── pixpack/
│       └── main.go
│
├── internal/
│   ├── codec/
│   │   ├── encode.go
│   │   ├── decode.go
│   │   ├── inspect.go
│   │   ├── verify.go
│   │   ├── pixels.go
│   │   └── errors.go
│   │
│   ├── format/
│   │   ├── header.go
│   │   └── constants.go
│   │
│   └── cli/
│       ├── root.go
│       ├── encode.go
│       ├── decode.go
│       ├── inspect.go
│       └── verify.go
│
├── testdata/
│   ├── hello.txt
│   ├── random.bin
│   └── sample.zip
│
├── docs/
│   └── FORMAT.md
│
├── .github/
│   └── workflows/
│       ├── test.yml
│       └── release.yml
│
├── .gitignore
├── CHANGELOG.md
├── CONTRIBUTING.md
├── LICENSE
├── PLAN.md
├── README.md
├── SECURITY.md
├── go.mod
└── go.sum
```

Do not create dozens of tiny abstractions.

Keep the format code independent from CLI printing.

---

## 15. Core API

Even though the main product is a CLI, keep the codec logic reusable.

Suggested API:

```go
package codec

type EncodeOptions struct {
    Width     int
    Overwrite bool
}

type DecodeOptions struct {
    OutputPath string
    Overwrite  bool
}

type Metadata struct {
    Version     uint8
    Flags       uint8
    Filename    string
    PayloadSize uint64
    Width       int
    Height      int
    SHA256      [32]byte
}

func EncodeFile(inputPath, outputPath string, options EncodeOptions) (*Metadata, error)

func DecodeFile(inputPath string, options DecodeOptions) (*Metadata, error)

func InspectFile(inputPath string) (*Metadata, error)

func VerifyFile(inputPath string) (*Metadata, error)
```

The exact API may be improved while implementing, but keep it small and understandable.

---

## 16. CLI Implementation

Prefer Go's standard `flag` package or a small custom parser.

Do not add a large CLI framework solely for four commands.

Required behavior:

```bash
pixpack --help
pixpack --version
pixpack encode --help
pixpack decode --help
pixpack inspect --help
pixpack verify --help
```

Use meaningful exit codes:

```text
0 = success
1 = general error
2 = invalid CLI usage
3 = invalid or unsupported PixPack image
4 = checksum verification failure
```

Do not print stack traces during normal CLI errors.

Errors should be sent to stderr.

Normal output should be sent to stdout.

---

## 17. Progress and Memory

The first implementation may hold the complete encoded byte stream in memory for simplicity, but avoid unnecessary copies.

Prefer:

* Reading file bytes once
* Preallocating buffers when sizes are known
* Avoiding repeated `append` growth
* Using `io.Reader` and `io.Writer` where practical

Do not implement complex streaming pixel encoders before the basic implementation is correct.

After the working MVP, optimize large-file behavior only when benchmarks show a need.

The project should comfortably handle files of at least:

```text
100 MB
```

on an ordinary modern computer, assuming enough memory is available.

Document that PNG encoding can use significantly more memory than the original file.

---

## 18. Tests

Testing is mandatory.

Use Go's built-in testing package.

### Round-trip tests

For each test input:

```text
decode(encode(input)) == input
```

Compare exact bytes and SHA-256 hashes.

Test:

* Empty file
* One-byte file
* Two-byte file
* Three-byte file
* Four-byte file
* Plain text
* Unicode text
* JSON
* Source code
* ZIP archive
* Random binary bytes
* File with zero bytes
* Filename containing spaces
* Unicode filename
* Large file
* Payload requiring one padding byte
* Payload requiring two padding bytes
* Payload requiring no padding

### Header tests

Test:

* Correct serialization
* Correct deserialization
* Big-endian values
* Maximum filename length
* Filename too long
* Invalid magic bytes
* Unsupported version
* Unsupported flags
* Invalid header size
* Impossible payload size

### Corruption tests

Test:

* Modified header pixel
* Modified payload pixel
* Truncated image
* Ordinary PNG
* Invalid PNG
* Incorrect checksum
* Declared payload larger than image capacity
* Invalid UTF-8 filename
* Dangerous filename
* Zero-length filename

### CLI tests

Test:

* Successful encode
* Successful decode
* Inspect command
* Verify command
* Missing arguments
* Invalid command
* Existing output without overwrite
* Existing output with overwrite
* Proper exit codes

### Fuzzing

Add Go fuzz tests for header parsing and PNG decoding input.

Examples:

```go
func FuzzParseHeader(f *testing.F)
func FuzzDecodeBytes(f *testing.F)
```

The decoder must never panic for arbitrary data.

---

## 19. Benchmarks

Add basic benchmarks for:

```text
1 KB
100 KB
1 MB
10 MB
```

Benchmark:

* Header construction
* Byte-to-pixel conversion
* Pixel-to-byte conversion
* Full encoding
* Full decoding

Do not optimize blindly.

Correctness comes first.

---

## 20. Documentation

### README.md

The README must include:

1. One-sentence project description.
2. Example encode command.
3. Example decode command.
4. Installation instructions.
5. Supported file types.
6. Explanation of how it works.
7. Security and privacy clarification.
8. Known limitations.
9. Development instructions.
10. License.
11. Example screenshots or generated PNG later.

Suggested introduction:

```text
PixPack converts any file into a lossless PNG image and restores it
byte-for-byte using a small, documented, open container format.
```

### FORMAT.md

Create a proper technical specification containing:

* Pixel ordering
* Channel ordering
* Header structure
* Integer endianness
* Filename encoding
* Checksum behavior
* Padding behavior
* Decoder validation requirements
* Versioning rules
* Compatibility expectations

The specification should be sufficient for another developer to create a compatible implementation in Python, JavaScript, Rust, or another language.

### SECURITY.md

Explain:

* Encoding is not encryption.
* Anyone with PixPack can restore the file.
* Sensitive data must be encrypted separately.
* PixPack images may be corrupted by resizing or conversion.
* Social media platforms may alter uploaded images.
* Decoded filenames are treated as untrusted.
* Security issues should be privately reported.

### CONTRIBUTING.md

Include:

* How to run tests
* How to build
* How to submit changes
* Code formatting expectations
* Requirement to update tests for format changes

---

## 21. Limitations to Document

Clearly explain:

1. PixPack is not encryption.
2. PixPack is not steganography.
3. The image may look like colored noise.
4. Resizing the PNG will destroy the payload.
5. Cropping the PNG will destroy the payload.
6. Converting the PNG to JPEG will destroy the payload.
7. Editing pixels will likely cause checksum failure.
8. Some websites may recompress images.
9. Uploading the PNG as a document/file is safer than uploading it as a photo.
10. Encoding a ZIP usually does not meaningfully compress it again.
11. The resulting PNG may be slightly larger than the original file.
12. Huge files produce huge image dimensions and require substantial memory.

---

## 22. Compression

Do not add custom compression in version 1.

PNG already applies lossless DEFLATE compression to image data.

For ordinary source code or text, PNG may compress repeated byte patterns.

For ZIP, GZIP, video, encrypted data, and other already-compressed input, the output PNG will usually be slightly larger than the original.

Later versions may offer optional preprocessing compression, but it must be clearly recorded in flags and the documented format.

Do not silently compress the payload with an additional algorithm in version 1.

---

## 23. Encryption

Do not implement encryption in the first release.

Encoding is not encryption.

Never invent a custom encryption algorithm.

A later version may optionally use:

```text
Argon2id or scrypt for key derivation
AES-256-GCM or XChaCha20-Poly1305 for authenticated encryption
```

Any encryption feature must receive a separate design and security review.

Reserve a flag bit for future encrypted payload support, but reject it in the version 1 decoder.

---

## 24. Format Versioning

The first format version is:

```text
1
```

Rules:

* Never silently reinterpret an existing version.
* Breaking changes require a new format version.
* Unknown versions must be rejected.
* Reserved fields and flags must remain zero.
* Minor CLI changes do not require a format-version change.
* The format specification is independent of the software release version.

Example:

```text
PixPack CLI v0.1.0
PixPack format v1
```

---

## 25. GitHub Actions

Create a test workflow that runs on:

```text
Ubuntu
Windows
macOS
```

The workflow must:

```bash
go fmt check
go vet ./...
go test ./...
go test -race ./...
go build ./cmd/pixpack
```

A release workflow should build binaries for:

```text
linux/amd64
linux/arm64
windows/amd64
windows/arm64
darwin/amd64
darwin/arm64
```

Release archive names:

```text
pixpack_<version>_linux_amd64.tar.gz
pixpack_<version>_linux_arm64.tar.gz
pixpack_<version>_windows_amd64.zip
pixpack_<version>_windows_arm64.zip
pixpack_<version>_darwin_amd64.tar.gz
pixpack_<version>_darwin_arm64.tar.gz
```

Generate SHA-256 checksum files for release artifacts.

---

## 26. Developer Commands

Document commands such as:

```bash
go build -o pixpack ./cmd/pixpack
go test ./...
go test -race ./...
go vet ./...
go fmt ./...
```

Example development usage:

```bash
go run ./cmd/pixpack encode testdata/sample.zip sample.png
go run ./cmd/pixpack inspect sample.png
go run ./cmd/pixpack verify sample.png
go run ./cmd/pixpack decode sample.png --output restored.zip
```

Then compare:

```bash
sha256sum testdata/sample.zip restored.zip
```

---

## 27. Implementation Milestones

### Milestone 1 — Repository Foundation

Create:

* Go module
* License
* README skeleton
* Project directories
* Basic CLI command dispatch
* CI workflow
* Version constant

Acceptance criteria:

```bash
go test ./...
go build ./cmd/pixpack
pixpack --help
```

All must work.

### Milestone 2 — Binary Header

Implement:

* Constants
* Header structure
* Header serialization
* Header parsing
* Bounds validation
* Unit tests
* Fuzz tests

Acceptance criteria:

```text
parse(serialize(header)) equals original header
```

Malformed input must not panic.

### Milestone 3 — Pixel Codec

Implement:

* Bytes to RGB pixels
* RGB pixels to bytes
* Dimension calculation
* Padding
* Overflow protection
* Unit tests

Acceptance criteria:

```text
pixelsToBytes(bytesToPixels(data)) starts with exact original data
```

Test every padding case.

### Milestone 4 — Encoding

Implement:

* File reading
* Filename handling
* SHA-256
* Header construction
* PNG generation
* Safe output handling
* Encode CLI command

Acceptance criteria:

```bash
pixpack encode testdata/sample.zip sample.png
```

Produces a valid PNG readable by ordinary image viewers.

### Milestone 5 — Decoding

Implement:

* PNG reading
* Header parsing
* Payload extraction
* Checksum verification
* Safe filename handling
* Temporary file writes
* Decode CLI command

Acceptance criteria:

```bash
pixpack decode sample.png --output restored.zip
```

The restored ZIP must be byte-for-byte identical.

### Milestone 6 — Inspect and Verify

Implement:

* Inspect command
* Verify command
* Metadata formatting
* Error exit codes

Acceptance criteria:

```bash
pixpack inspect sample.png
pixpack verify sample.png
```

Both must work without creating extracted files.

### Milestone 7 — Hardening

Add:

* Corruption tests
* Ordinary PNG rejection
* Fuzz testing
* Large-file tests
* Cross-platform filename tests
* Overflow tests
* Race detector fixes
* Better error messages

Acceptance criteria:

```bash
go test ./...
go test -race ./...
go vet ./...
```

All pass.

### Milestone 8 — Documentation and Release

Complete:

* README
* FORMAT.md
* SECURITY.md
* CONTRIBUTING.md
* CHANGELOG.md
* Release workflow
* Cross-platform build artifacts

Create the first tagged release:

```text
v0.1.0
```

---

## 28. Definition of Done

The project is complete for version `0.1.0` when:

1. Any normal file can be encoded.
2. Encoded output is a valid PNG.
3. The original file can be restored exactly.
4. Filename and extension are preserved.
5. Corruption is detected using SHA-256.
6. Ordinary PNG files are rejected clearly.
7. The decoder handles malformed input without panicking.
8. Encode, decode, inspect, and verify commands work.
9. Tests pass on Linux, Windows, and macOS.
10. The binary compiles with no runtime dependencies.
11. The format is documented.
12. Release binaries are generated automatically.
13. The README contains complete installation and usage instructions.
14. No unnecessary frameworks or abstractions were introduced.

---

## 29. Coding Standards

Follow these rules:

* Write idiomatic Go.
* Run `gofmt`.
* Return errors instead of panicking.
* Wrap errors with useful context.
* Keep functions focused.
* Avoid global mutable state.
* Avoid unnecessary interfaces.
* Avoid premature abstraction.
* Avoid unnecessary dependencies.
* Add comments for exported declarations.
* Add comments explaining binary-format decisions.
* Use table-driven tests.
* Validate all untrusted values.
* Never ignore file-close or write errors when they matter.
* Never silently overwrite files.
* Never produce a decoded file before checksum verification.
* Keep the CLI thin and the codec logic reusable.

---

## 30. Instructions for the Coding Agent

Implement the complete project described above.

Work milestone by milestone, but continue through all milestones without waiting for confirmation.

Do not stop after creating scaffolding.

Do not leave placeholder functions, TODO-only files, fake tests, or nonfunctional commands.

After each major milestone:

1. Run formatting.
2. Run tests.
3. Run static analysis.
4. Fix all discovered failures.
5. Continue to the next milestone.

Use real round-trip testing with generated binary files.

Before considering the work finished, perform this end-to-end test:

```bash
go build -o pixpack ./cmd/pixpack

./pixpack encode testdata/sample.zip sample.png
./pixpack inspect sample.png
./pixpack verify sample.png
./pixpack decode sample.png --output restored.zip
```

Then verify that the original and restored files are identical.

On Unix:

```bash
cmp testdata/sample.zip restored.zip
```

Also compare SHA-256 hashes.

Corrupt one payload pixel and confirm:

```bash
pixpack verify corrupted.png
```

fails with a checksum mismatch.

Create a polished README with copy-paste installation and usage examples.

The final response should summarize:

* What was implemented
* Repository structure
* Format specification
* CLI commands
* Test results
* Build results
* Any remaining noncritical limitations

Do not add encryption, folder archiving, a GUI, networking, or steganography to version 1.

Focus on delivering a small, reliable, documented, cross-platform open-source utility.

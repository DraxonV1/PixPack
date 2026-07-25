# Changelog

All notable changes to PixPack are documented in this file.

## [0.1.0] - 2025-07-25

### Added

- Initial release
- `encode` command: convert any file into a lossless PNG image
- `decode` command: restore original file from a PixPack PNG
- `inspect` command: view metadata without extracting
- `verify` command: check payload integrity without extracting
- PixPack binary format v1 with SHA-256 integrity verification
- Safe filename handling (directory traversal prevention)
- Overwrite protection
- Custom image width support
- Cross-platform support (Windows, Linux, macOS)
- Comprehensive test suite with round-trip, corruption, and fuzz tests
- Benchmarks for encoding and decoding operations
- Full documentation (README, FORMAT.md, SECURITY.md, CONTRIBUTING.md)
- GitHub Actions CI (test workflow for Ubuntu, Windows, macOS)
- GitHub Actions release workflow (builds for 6 platforms/architectures)
- MIT License

### Fixed

- CLI exit code mapping: invalid/unsupported PixPack images now exit with code 3; checksum mismatches exit with code 4; usage errors exit with code 2; general errors exit with code 1
- Unicode multiplication sign (×) replaced with ASCII 'x' in output for better terminal compatibility
- Error matching updated to use `errors.Is` for proper wrapped error detection
- Module path updated to `github.com/DraxonV1/PixPack`

### Known Limitations

- No encryption — encode is not encryption
- No steganography features
- Large files require significant memory for PNG encoding
- PNG output may be larger than input for compressed data (ZIP, video, etc.)

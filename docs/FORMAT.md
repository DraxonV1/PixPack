# PixPack Binary Format Specification v1

This document defines the PixPack version 1 binary container format for storing arbitrary file data inside lossless PNG images.

## Format Overview

PixPack converts any file into a sequence of bytes, prepends a binary header, converts the combined byte stream into RGB pixels, and encodes the result as a lossless PNG image.

## Pixel Ordering and Channel Assignment

- Pixels are arranged in **row-major** order: top row first, left to right within each row.
- Each pixel stores **three bytes**: Red, Green, Blue.
- Alpha channel is always **255** (opaque) and does not carry payload data in version 1.
- Byte mapping:

```
byte 0 → pixel 0 Red
byte 1 → pixel 0 Green
byte 2 → pixel 0 Blue

byte 3 → pixel 1 Red
byte 4 → pixel 1 Green
byte 5 → pixel 1 Blue

...
```

- If the total byte count is not a multiple of 3, the remaining channels in the final pixel are filled with **zero bytes**.
- These padding zero bytes are part of the pixel data but are not part of the payload. The decoder uses the declared payload length (from the header) to extract exactly the right number of bytes.

## Byte Order

All multi-byte integers in the PixPack format use **big-endian** (network) byte order.

## Header Structure

The PixPack header is placed at the beginning of the byte stream, before the payload. The header contains metadata needed for decoding and verification.

### Version 1 Header Layout

| Offset | Size | Field | Description |
|--------|------|-------|-------------|
| 0 | 4 bytes | Magic | Magic bytes: `50 58 50 4B` (`PXPK`) |
| 4 | 1 byte | Version | Format version number: `1` |
| 5 | 1 byte | Flags | Bit flags (must be `0` for v1) |
| 6 | 2 bytes | Header Size | Total header size including filename (big-endian uint16) |
| 8 | 8 bytes | Payload Size | Original payload length in bytes (big-endian uint64) |
| 16 | 2 bytes | Filename Length | Byte length of the filename (big-endian uint16) |
| 18 | 32 bytes | Checksum | Raw SHA-256 digest of the payload (not hex-encoded) |
| 50 | N bytes | Filename | UTF-8 encoded filename |
| 50+N | variable | Payload | Original file bytes |
| variable | 0–2 bytes | Padding | Zero bytes to align to 3-byte pixel boundary |

### Field Details

**Magic (4 bytes)**
- Identifies the byte stream as PixPack format.
- Hex: `50 58 50 4B`
- ASCII: `PXPK`

**Version (1 byte)**
- Current version: `1`
- Unknown versions must be rejected by decoders.

**Flags (1 byte)**
- All bits are reserved for future use.
- Version 1 decoders MUST reject any non-zero flags value.
- Bit 0 is reserved for future encrypted payload support; version 1 decoders reject it.

**Header Size (2 bytes)**
- Total number of bytes consumed by the header, from the start of Magic to the end of the filename.
- Computed as: `50 + filename_byte_length`
- Minimum value: 50 (when the filename is empty, but empty filenames are rejected).
- Decoders MUST validate that this value is at least 50 and does not exceed the available data.

**Payload Size (8 bytes)**
- Length of the original file in bytes.
- Unsigned 64-bit integer in big-endian byte order.
- This is the exact number of bytes to extract from the pixel stream after the header.
- Padding bytes after the payload are NOT included.

**Filename Length (2 bytes)**
- Length of the following filename in bytes (not characters).
- Unsigned 16-bit integer in big-endian byte order.
- Must be at least 1 (filenames cannot be empty).
- Decoders MUST validate this field against the header size.

**Checksum (32 bytes)**
- Raw SHA-256 cryptographic hash of the payload bytes only.
- The checksum does NOT cover the header, filename, or padding bytes.
- Stored as raw binary, not as a hexadecimal string.
- Decoders MUST perform constant-time comparison to prevent timing attacks.

**Filename (N bytes)**
- The base filename of the original file (without directory path).
- Encoded as UTF-8.
- Length is specified by the Filename Length field.
- Maximum length: 65,535 bytes (practically limited by the header size field).
- Decoders MUST validate UTF-8 well-formedness.
- Decoders MUST reject empty filenames, absolute paths, and names containing directory traversal components (`..`, `/`, `\`).
- The decoder uses `filepath.Base` to extract the final component and validates the result.

## Byte Stream Construction

The complete byte stream before pixel conversion is:

```
[Header] [Payload] [Zero Padding]
```

Where:
- **Header**: 50 + filename length bytes (as defined above)
- **Payload**: N bytes of the original file
- **Zero Padding**: 0–2 zero bytes added so the total length is a multiple of 3

## Pixel Conversion

Each consecutive group of 3 bytes from the byte stream becomes one RGB pixel as described above. The image is created with dimensions sufficient to hold all pixels plus any unused pixels filled with `(0, 0, 0, 255)`.

## Image Dimension Calculation

Given `totalBytes = headerSize + payloadSize`:

```
requiredPixels = ceil(totalBytes / 3)
```

When no width is specified:

```
width  = ceil(sqrt(requiredPixels))
height = ceil(requiredPixels / width)
```

When a width is specified:

```
height = ceil(requiredPixels / width)
```

## Decoder Validation Requirements

A compliant decoder MUST perform the following validations:

1. Verify the PNG is valid and can be decoded.
2. Verify the first 4 bytes of the pixel stream match the magic bytes `PXPK`.
3. Verify the version byte equals `1`.
4. Verify the flags byte equals `0`.
5. Validate the header size is ≥ 50 and within the available pixel data.
6. Validate the filename length ≥ 1 and consistent with header size.
7. Validate the payload size is non-negative and fits within the available pixel data.
8. Validate the filename is valid UTF-8.
9. Sanitize the decoded filename, rejecting unsafe paths.
10. Extract exactly `payloadSize` bytes from the pixel stream after the header.
11. Compute SHA-256 of the extracted payload.
12. Compare the computed hash with the stored checksum using constant-time comparison.
13. Never write output before checksum verification passes.
14. Use temporary files for output, renaming only after successful verification.

## Versioning Rules

1. The format version is independent of the software release version.
2. Breaking changes to the binary format require a new format version number.
3. Decoders must reject unknown format versions.
4. Minor additions that do not break backward compatibility may use flag bits.
5. Reserved fields and flags must remain zero in this version.
6. The format specification is the authoritative reference; implementations must follow it exactly.

## Compatibility

- PixPack v1 images created by this implementation can be decoded by any v1-compliant decoder.
- The format is designed for portability across platforms (Windows, Linux, macOS).
- Integer sizes, endianness, and UTF-8 encoding are platform-independent.
- PNG encoding may vary slightly between libraries (e.g., compression level), but the decoded pixel data must be identical.

## Security Considerations

- The checksum uses SHA-256 for integrity verification, not for authentication.
- There is no encryption in version 1. The payload bytes are stored as-is in pixel channels.
- Any image processing operation (resize, crop, re-encode, format conversion) will destroy the payload.
- Decoded filenames must be treated as untrusted input to prevent directory traversal attacks.

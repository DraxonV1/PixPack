# Security Policy

## Encoding Is Not Encryption

PixPack does **not** encrypt data. The payload bytes are stored directly in PNG pixel channels without any obfuscation or cryptographic protection.

Anyone with PixPack (or any tool that reads RGB pixels from a PNG) can restore the original file. The SHA-256 checksum provides integrity verification only — it does not provide confidentiality.

If you need to protect sensitive data:

1. **Encrypt separately** using a strong, well-vetted encryption tool before encoding with PixPack.
2. Use established tools like GnuPG, OpenSSL, age, or a password manager.
3. Never rely on PixPack alone to keep data private.

## Data Integrity

PixPack uses SHA-256 to detect accidental corruption or modification of the payload:

- The checksum covers only the payload bytes (not the header, filename, or padding).
- Checksums are compared using a constant-time comparison to prevent timing side-channel attacks.
- If the checksum does not match, the decoded file is never written to disk.

## Image Manipulation Risks

PixPack images are **fragile by design**. Any of the following operations will destroy the payload:

- Resizing the PNG
- Cropping the PNG
- Converting to JPEG or any lossy format
- Editing pixels (even a single channel change)
- Re-compressing with a different PNG encoder
- Uploading to social media that re-encodes images
- Using image processing filters

**Uploading the PNG as a file/document** (e.g., as an email attachment, cloud storage file, or software download) preserves the payload. **Uploading as a photo** to social media, messaging apps, or image hosts may recompress or re-encode the image, destroying the data.

## Safe Filename Handling

PixPack treats decoded filenames as **untrusted input**:

- Only the base filename is stored (the directory portion is stripped during encoding).
- During decoding, the filename is sanitized using `filepath.Base` and checked for directory traversal.
- Filenames containing `..`, `/`, `\`, or Windows reserved characters are rejected.
- Absolute paths and empty filenames are rejected.
- The decoded file is written to a temporary file first and renamed only after checksum verification succeeds.

## Memory Safety

PixPack is written in Go, a memory-safe language. The decoder validates all lengths and offsets before use to prevent out-of-bounds access. Malformed or malicious PixPack images should never cause unsafe memory access.

## Reporting a Vulnerability

If you discover a security vulnerability in PixPack:

1. **Do not** open a public GitHub issue.
2. Send a private email to the maintainers (see the GitHub repository for contact information).
3. Include a detailed description of the issue, steps to reproduce, and any relevant files.
4. Allow reasonable time for a fix before public disclosure.

We take all security reports seriously and will respond as quickly as possible.

## Scope

The following are **not** considered security vulnerabilities:

- PixPack not being encryption (by design)
- The PNG image looking like colored noise (by design)
- Large files consuming significant memory (documented limitation)
- SHA-256 not being perfect forward secrecy (PixPack is not a communications protocol)

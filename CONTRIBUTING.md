# Contributing to PixPack

Thanks for your interest in contributing!

## How to Contribute

1. **Fork** the repository.
2. **Create a branch** for your changes.
3. **Make your changes** following the guidelines below.
4. **Run tests** to ensure nothing is broken.
5. **Submit a pull request** with a clear description of the changes.

## Development Setup

```bash
git clone https://github.com/DraxonV1/PixPack.git
cd pixpack
go build ./cmd/pixpack
go test ./...
```

## Coding Standards

- Write **idiomatic Go** code.
- Run `go fmt ./...` before committing.
- Run `go vet ./...` and fix all warnings.
- Follow the [Go Proverbs](https://go-proverbs.github.io/).
- Use **table-driven tests** where appropriate.
- Avoid **unnecessary dependencies**. Prefer the Go standard library.
- Avoid **premature abstraction**. Don't create interfaces or types until they're needed.

## Testing

- All code must have test coverage.
- Run `go test ./...` to run all tests.
- Run `go test -race ./...` to check for race conditions.
- Run `go test -bench=. ./...` to run benchmarks.
- Include **round-trip tests** for encoding and decoding.
- Include **corruption tests** for malformed inputs.
- Add **fuzz tests** for header parsing and PNG decoding.
- The decoder must never panic on arbitrary input.

## Pull Request Process

1. Ensure all tests pass on your branch.
2. Update tests if your changes affect behavior.
3. Update documentation (README, FORMAT.md) if your changes affect the format or CLI.
4. Update CHANGELOG.md with a description of your changes.
5. Add yourself to the contributors list if desired.

## Code of Conduct

Be respectful and constructive. We strive to maintain a welcoming community for all contributors.

## License

By contributing, you agree that your contributions will be licensed under the MIT License (see [LICENSE](LICENSE)).

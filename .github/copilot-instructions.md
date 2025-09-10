# Copilot Instructions for Picocrypt

## Project Overview
Picocrypt is a single-file Go application for secure file encryption, focused on simplicity, reliability, and strong cryptography. It uses XChaCha20 and Argon2id by default, with optional Serpent cascade and Reed-Solomon error correction for advanced use cases. The project is archived and read-only, but remains stable and secure.

## Architecture & Key Files
- Main logic is in `src/Picocrypt.go` (UI, cryptography, file handling, options)
- Project metadata and build info: `VERSION`, `README.md`, `Internals.md`
- Images/icons: `images/`
- Go dependencies: `src/go.mod`, `src/go.sum`
- Build workflows: `.github/workflows/` (see `build-windows.yml` for Windows specifics)

## Build & Developer Workflow
- **Build from source:**
  - Windows: `go build -ldflags="-s -w -H=windowsgui -extldflags=-static" Picocrypt.go`
  - macOS/Linux: `go build -ldflags="-s -w" Picocrypt.go`
  - Set `CGO_ENABLED=1` for all platforms
  - Output is a portable executable (`Picocrypt.exe`/`Picocrypt`)
- **Windows CI:**
  - Workflow adds icons, manifest, and version info using Resource Hacker
  - Final binary compressed with UPX
  - See `.github/workflows/build-windows.yml` for full steps
- **Dependencies:**
  - Uses Picocrypt-maintained forks for all major libraries (see `go.mod`)
  - All cryptography from `golang.org/x/crypto`

## Cryptography & Data Flow
- **Encryption:**
  - XChaCha20 (default), Argon2id KDF
  - Paranoid mode: XChaCha20 + Serpent cascade, HMAC-SHA3, stronger Argon2id params
  - Reed-Solomon ECC for headers and optionally data
  - HKDF-SHA3 for subkey derivation
- **Keyfiles:**
  - Supports multiple keyfiles, with optional order enforcement
  - Keyfile hashes combined via XOR (unordered) or concatenation (ordered)
- **Volume Format:**
  - Custom header (see `Internals.md` for offsets and encoding)
  - Deniable volumes omit header, appear as random data

## UI & Features
- All UI logic is in `Picocrypt.go` using `giu` and `imgui-go`
- Features: password generator, comments, keyfiles, paranoid mode, Reed-Solomon, force decrypt, chunking, compression, deniability, recursive file handling
- See `README.md` for feature details and usage patterns

## Security Practices
- Assumes trusted host environment; does not protect against side-channel attacks
- All dependencies are forked and reviewed before updates
- No telemetry, no file shredding (deletes only)
- GPL-3.0-only license for all original code

## Conventions & Patterns
- Single source file for all logic; UI and cryptography intermixed
- Advanced options exposed via UI modals and flags
- All build and release automation is in `.github/workflows/`
- No test suite; security and correctness are documented in `Internals.md` and `README.md`

## Example: Adding a Feature
- Edit `src/Picocrypt.go` directly
- Follow UI patterns using `giu`/`imgui-go` for new dialogs or options
- Use cryptographic primitives from `golang.org/x/crypto` or Picocrypt forks
- Update documentation in `README.md` and `Internals.md` as needed

---
For unclear or missing conventions, review `README.md`, `Internals.md`, and workflow files. Ask for feedback if any section is ambiguous or incomplete.

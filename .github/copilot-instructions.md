# Copilot Instructions for HadesCrypt

## Project Overview
HadesCrypt is a cross-platform file and folder encryption tool built in Go, featuring a Fyne GUI. It supports AES-256-GCM encryption, Argon2id key derivation, password generation, and integrates with GnuPG/OpenPGP for industry-standard encryption. The codebase is modular, with core logic in `internal/` and the GUI in `main.go`.

## Architecture & Key Components
- **main.go**: Entry point, GUI logic, user interaction.
- **internal/cryptoengine/**: Core encryption/decryption (AES-256-GCM, Argon2id, secure RNG, chunking, file format logic).
- **internal/archiver/**: Folder archiving (tar.gz) for recursive encryption.
- **internal/config/**: Persistent configuration (JSON at `~/.hadescrypt/config.json`).
- **internal/password/**: Password generation and strength meter.
- **internal/ui/**: UI formatting utilities.
- **internal/gnupg/**: GnuPG/OpenPGP integration (external process, standard .gpg files, symmetric/asymmetric modes).

## File Formats
- **Encrypted files**: `.hadescrypt` extension, custom binary format (see README for header structure).
- **Encrypted folders**: Compressed as tar.gz, then encrypted.
- **GnuPG files**: Standard `.gpg` format, compatible with other PGP tools.

## Developer Workflows
- **Build**: `go build -o HadesCrypt.exe` (Windows), use `GOOS`/`GOARCH` for cross-compilation.
- **Run**: Launch the built binary; GUI is the main interface.
- **Test**: No standard Go test files; manual testing via GUI and sample files (see `test_*.txt`).
- **GnuPG Integration**: Requires GnuPG installed and available in PATH. Use `gnupg.IsAvailable()` and `gnupg.GetGPGInfo()` for diagnostics.

## Patterns & Conventions
- **Chunked encryption**: Large files are processed in chunks for memory efficiency.
- **Progress reporting**: All long-running operations provide real-time progress (UI callbacks).
- **Secure memory handling**: Sensitive data is zeroed after use.
- **Configurable algorithms**: Most cryptographic parameters (cipher, compression, KDF) are user-configurable via UI and config.
- **Error handling**: User-facing errors are descriptive; internal errors are logged.
- **Planned features**: Paranoid mode (multi-cipher), keyfiles, ECC, deniability, etc. (see README for roadmap).

## Integration Points
- **Fyne GUI**: All user interaction via Fyne widgets.
- **GnuPG**: External process, supports both symmetric and asymmetric encryption. Keyring integration for public keys.
- **Config file**: Persistent settings, profiles, history.

## Examples
- To encrypt a file: drag-and-drop in GUI, enter password, click "Encrypt".
- To use GnuPG: select "GnuPG/OpenPGP" mode, configure options, ensure GnuPG is installed.
- To build for Linux: `GOOS=linux GOARCH=amd64 go build -o HadesCrypt-linux`

## References
- See `README.md` (project root) and `internal/gnupg/README.md` for detailed format specs and integration notes.
- Key modules: `internal/cryptoengine/crypto.go`, `internal/gnupg/gnupg.go`, `internal/config/config.go`

---
For unclear workflows or missing conventions, ask the user for clarification or examples before proceeding.

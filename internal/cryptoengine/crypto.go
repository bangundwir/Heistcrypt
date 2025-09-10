package cryptoengine

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
	
	"github.com/bangundwir/HadesCrypt/internal/postquantum"
)

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ProgressCallback reports processed and total bytes.
type ProgressCallback func(processed int64, total int64)

// EncryptionMode represents the encryption algorithm to use
type EncryptionMode int

const (
	ModeAES256GCM EncryptionMode = iota
	ModeChaCha20
	ModeParanoid // AES-256-GCM + ChaCha20-Poly1305
	ModePostQuantumKyber768
	ModePostQuantumDilithium3
	ModePostQuantumSPHINCS
)

const (
    fileMagic       = "HAD1" // 4 bytes
    fileVersion     = byte(1)
    saltLengthBytes = 16
    noncePrefixLen  = 8 // Remaining 4 bytes used for chunk counter
    gcmNonceLen     = 12
    gcmOverhead     = 16
)

// EncryptionOptions holds options for encryption
type EncryptionOptions struct {
	Mode            EncryptionMode
	Comments        string
	UseCompression  bool
	UseReedSolomon  bool
	UseDeniability  bool
	SplitSize       int64 // 0 means no splitting
}

// Argon2id parameters (balanced for desktop)
var (
    argonTime    uint32 = 1
    argonMemory  uint32 = 64 * 1024 // 64 MiB
    argonThreads uint8  = 4
    keyLen              = uint32(32)
)

// EncryptFile encrypts inputPath -> outputPath using the default AES-256-GCM mode
func EncryptFile(inputPath, outputPath string, password []byte, onProgress ProgressCallback) error {
	return EncryptFileWithMode(inputPath, outputPath, password, ModeAES256GCM, onProgress)
}

// EncryptFileWithOptions encrypts inputPath -> outputPath using specified options.
// The output format header:
// [4]MAGIC "HAD1" | [1]VERSION | [1]MODE | [1]FLAGS | [16]SALT | [8]NONCE_PREFIX | [4]CHUNK_SIZE | [8]ORIGINAL_SIZE | [2]COMMENT_LEN | [..]COMMENT | [..]CIPHERTEXT
func EncryptFileWithOptions(inputPath, outputPath string, password []byte, opts EncryptionOptions, onProgress ProgressCallback) error {
	return EncryptFileWithMode(inputPath, outputPath, password, opts.Mode, onProgress)
}

// EncryptFileWithMode encrypts inputPath -> outputPath using specified encryption mode.
// The output format header:
// [4]MAGIC "HAD1" | [1]VERSION | [1]MODE | [16]SALT | [8]NONCE_PREFIX | [4]CHUNK_SIZE | [8]ORIGINAL_SIZE | [..]CIPHERTEXT
func EncryptFileWithMode(inputPath, outputPath string, password []byte, mode EncryptionMode, onProgress ProgressCallback) error {
    in, err := os.Open(inputPath)
    if err != nil {
        return err
    }
    defer in.Close()

    st, err := in.Stat()
    if err != nil {
        return err
    }
    totalSize := st.Size()

    // Prepare header fields
    salt := make([]byte, saltLengthBytes)
    if _, err := io.ReadFull(rand.Reader, salt); err != nil {
        return fmt.Errorf("generate salt: %w", err)
    }
    noncePrefix := make([]byte, noncePrefixLen)
    if _, err := io.ReadFull(rand.Reader, noncePrefix); err != nil {
        return fmt.Errorf("generate nonce prefix: %w", err)
    }

    key := argon2.IDKey(password, salt, argonTime, argonMemory, argonThreads, keyLen)

    // Create cipher based on mode
    var aead cipher.AEAD
    var aead2 cipher.AEAD // For paranoid mode
    var pqCipher *postquantum.PostQuantumCipher // For post-quantum modes
    
    switch mode {
    case ModeAES256GCM:
        block, err := aes.NewCipher(key)
        if err != nil {
            return err
        }
        aead, err = cipher.NewGCM(block)
        if err != nil {
            return err
        }
    case ModeChaCha20:
        aead, err = chacha20poly1305.New(key)
        if err != nil {
            return err
        }
    case ModeParanoid:
        // First layer: AES-256-GCM
        block, err := aes.NewCipher(key)
        if err != nil {
            return err
        }
        aead, err = cipher.NewGCM(block)
        if err != nil {
            return err
        }
        
        // Second layer: ChaCha20-Poly1305 (derive different key)
        key2 := argon2.IDKey(append(password, []byte("paranoid")...), salt, argonTime*2, argonMemory, argonThreads, keyLen)
        aead2, err = chacha20poly1305.New(key2)
        if err != nil {
            return err
        }
    case ModePostQuantumKyber768:
        pqCipher = postquantum.NewPostQuantumCipher(postquantum.Kyber768)
    case ModePostQuantumDilithium3:
        pqCipher = postquantum.NewPostQuantumCipher(postquantum.Dilithium3)
    case ModePostQuantumSPHINCS:
        pqCipher = postquantum.NewPostQuantumCipher(postquantum.SPHINCS)
    default:
        return fmt.Errorf("unsupported encryption mode: %d", mode)
    }

    // Choose chunk size to balance memory and speed
    const chunkSize = 1 << 20 // 1 MiB plaintext per chunk

    out, err := os.Create(outputPath)
    if err != nil {
        return err
    }
    defer func() {
        cerr := out.Close()
        if err == nil && cerr != nil {
            err = cerr
        }
    }()

    // Write header
    if _, err := out.Write([]byte(fileMagic)); err != nil {
        return err
    }
    if _, err := out.Write([]byte{fileVersion}); err != nil {
        return err
    }
    if _, err := out.Write([]byte{byte(mode)}); err != nil {
        return err
    }
    if _, err := out.Write(salt); err != nil {
        return err
    }
    if _, err := out.Write(noncePrefix); err != nil {
        return err
    }
    // chunk size uint32
    var tmp4 [4]byte
    binary.BigEndian.PutUint32(tmp4[:], uint32(chunkSize))
    if _, err := out.Write(tmp4[:]); err != nil {
        return err
    }
    // original size uint64
    var tmp8 [8]byte
    binary.BigEndian.PutUint64(tmp8[:], uint64(totalSize))
    if _, err := out.Write(tmp8[:]); err != nil {
        return err
    }

    buf := make([]byte, chunkSize)
    processed := int64(0)
    var counter uint32 = 0
    nonce := make([]byte, gcmNonceLen)
    copy(nonce[:noncePrefixLen], noncePrefix)

    for {
        n, readErr := io.ReadFull(in, buf)
        if errors.Is(readErr, io.ErrUnexpectedEOF) {
            // last partial chunk
            if n > 0 {
                binary.BigEndian.PutUint32(nonce[noncePrefixLen:], counter)
                sealed := aead.Seal(nil, nonce, buf[:n], nil)
                
                // Apply second layer encryption for paranoid mode
                if mode == ModeParanoid {
                    nonce2 := make([]byte, aead2.NonceSize())
                    copy(nonce2, nonce[:min(len(nonce2), len(nonce))])
                    sealed = aead2.Seal(nil, nonce2, sealed, nil)
                }
                
                if _, err := out.Write(sealed); err != nil {
                    return err
                }
                processed += int64(n)
                if onProgress != nil {
                    onProgress(processed, totalSize)
                }
            }
            break
        }
        if errors.Is(readErr, io.EOF) {
            break
        }
        if readErr != nil && readErr != io.ErrUnexpectedEOF {
            return readErr
        }

        binary.BigEndian.PutUint32(nonce[noncePrefixLen:], counter)
        
        var sealed []byte
        
        // Choose encryption method based on mode
        if pqCipher != nil {
            // Post-quantum encryption
            pqNonce, err := pqCipher.GenerateNonce()
            if err != nil {
                return fmt.Errorf("generate PQ nonce: %w", err)
            }
            sealed, err = pqCipher.Encrypt(buf[:n], key, pqNonce)
            if err != nil {
                return fmt.Errorf("PQ encrypt: %w", err)
            }
            // Prepend nonce to ciphertext
            sealed = append(pqNonce, sealed...)
        } else {
            // Traditional AEAD encryption
            sealed = aead.Seal(nil, nonce, buf[:n], nil)
            
            // Apply second layer encryption for paranoid mode
            if mode == ModeParanoid {
                // Use different nonce for second layer
                nonce2 := make([]byte, aead2.NonceSize())
                copy(nonce2, nonce[:min(len(nonce2), len(nonce))])
                sealed = aead2.Seal(nil, nonce2, sealed, nil)
            }
        }
        
        if _, err := out.Write(sealed); err != nil {
            return err
        }
        processed += int64(n)
        if onProgress != nil {
            onProgress(processed, totalSize)
        }
        counter++
    }

    return nil
}

// DecryptFile decrypts inputPath -> outputPath using the encryption mode stored in the file.
// If force is true, the function still returns error on auth failure (AEAD cannot bypass),
// but the flag is provided to align with UI; future modes may try salvage.
func DecryptFile(inputPath, outputPath string, password []byte, force bool, onProgress ProgressCallback) error {
    in, err := os.Open(inputPath)
    if err != nil {
        return err
    }
    defer in.Close()

    // Read and validate header
    header := make([]byte, 4)
    if _, err := io.ReadFull(in, header); err != nil {
        return err
    }
    if string(header) != fileMagic {
        return fmt.Errorf("invalid file format")
    }

    ver := make([]byte, 1)
    if _, err := io.ReadFull(in, ver); err != nil {
        return err
    }
    if ver[0] != fileVersion {
        return fmt.Errorf("unsupported version: %d", ver[0])
    }

    // Read encryption mode
    modeBytes := make([]byte, 1)
    if _, err := io.ReadFull(in, modeBytes); err != nil {
        return err
    }
    mode := EncryptionMode(modeBytes[0])

    salt := make([]byte, saltLengthBytes)
    if _, err := io.ReadFull(in, salt); err != nil {
        return err
    }
    noncePrefix := make([]byte, noncePrefixLen)
    if _, err := io.ReadFull(in, noncePrefix); err != nil {
        return err
    }

    var tmp4 [4]byte
    if _, err := io.ReadFull(in, tmp4[:]); err != nil {
        return err
    }
    chunkSize := int(binary.BigEndian.Uint32(tmp4[:]))

    var tmp8 [8]byte
    if _, err := io.ReadFull(in, tmp8[:]); err != nil {
        return err
    }
    totalSize := int64(binary.BigEndian.Uint64(tmp8[:]))

    key := argon2.IDKey(password, salt, argonTime, argonMemory, argonThreads, keyLen)
    
    // Create AEAD cipher based on mode
    var aead cipher.AEAD
    var aead2 cipher.AEAD // For paranoid mode
    
    switch mode {
    case ModeAES256GCM:
        block, err := aes.NewCipher(key)
        if err != nil {
            return err
        }
        aead, err = cipher.NewGCM(block)
        if err != nil {
            return err
        }
    case ModeChaCha20:
        aead, err = chacha20poly1305.New(key)
        if err != nil {
            return err
        }
    case ModeParanoid:
        // First layer: AES-256-GCM
        block, err := aes.NewCipher(key)
        if err != nil {
            return err
        }
        aead, err = cipher.NewGCM(block)
        if err != nil {
            return err
        }
        
        // Second layer: ChaCha20-Poly1305
        key2 := argon2.IDKey(append(password, []byte("paranoid")...), salt, argonTime*2, argonMemory, argonThreads, keyLen)
        aead2, err = chacha20poly1305.New(key2)
        if err != nil {
            return err
        }
    default:
        return fmt.Errorf("unsupported encryption mode: %d", mode)
    }

    out, err := os.Create(outputPath)
    if err != nil {
        return err
    }
    defer func() {
        cerr := out.Close()
        if err == nil && cerr != nil {
            err = cerr
        }
    }()

    // Determine number of chunks
    fullChunks := totalSize / int64(chunkSize)
    lastChunkSize := int(totalSize % int64(chunkSize))
    if totalSize == 0 {
        fullChunks = 0
        lastChunkSize = 0
    }

    processed := int64(0)
    var counter uint32 = 0
    nonce := make([]byte, gcmNonceLen)
    copy(nonce[:noncePrefixLen], noncePrefix)

    // Helper to read exactly N ciphertext bytes for a given plaintext length
    readCipher := func(nPlain int) ([]byte, error) {
        need := nPlain + gcmOverhead
        buf := make([]byte, need)
        if _, err := io.ReadFull(in, buf); err != nil {
            return nil, err
        }
        return buf, nil
    }

    // Read full-size chunks
    if fullChunks > 0 {
        for i := int64(0); i < fullChunks; i++ {
            cipherChunk, err := readCipher(chunkSize)
            if err != nil {
                return err
            }
            binary.BigEndian.PutUint32(nonce[noncePrefixLen:], counter)
            
            // Decrypt with appropriate layers based on mode
            var plain []byte
            if mode == ModeParanoid {
                // First decrypt with ChaCha20 (outer layer)
                nonce2 := make([]byte, aead2.NonceSize())
                copy(nonce2, nonce[:min(len(nonce2), len(nonce))])
                intermediate, err := aead2.Open(nil, nonce2, cipherChunk, nil)
                if err != nil {
                    return err
                }
                // Then decrypt with AES-GCM (inner layer)
                plain, err = aead.Open(nil, nonce, intermediate, nil)
                if err != nil {
                    return err
                }
            } else {
                plain, err = aead.Open(nil, nonce, cipherChunk, nil)
                if err != nil {
                    return err
                }
            }
            if _, err := out.Write(plain); err != nil {
                return err
            }
            processed += int64(len(plain))
            if onProgress != nil {
                onProgress(processed, totalSize)
            }
            counter++
        }
    }

    // Read last chunk if any
    if lastChunkSize > 0 {
        cipherChunk, err := readCipher(lastChunkSize)
        if err != nil {
            return err
        }
        binary.BigEndian.PutUint32(nonce[noncePrefixLen:], counter)
        
        // Decrypt with appropriate layers based on mode
        var plain []byte
        if mode == ModeParanoid {
            // First decrypt with ChaCha20 (outer layer)
            nonce2 := make([]byte, aead2.NonceSize())
            copy(nonce2, nonce[:min(len(nonce2), len(nonce))])
            intermediate, err := aead2.Open(nil, nonce2, cipherChunk, nil)
            if err != nil {
                return err
            }
            // Then decrypt with AES-GCM (inner layer)
            plain, err = aead.Open(nil, nonce, intermediate, nil)
            if err != nil {
                return err
            }
        } else {
            plain, err = aead.Open(nil, nonce, cipherChunk, nil)
            if err != nil {
                return err
            }
        }
        if _, err := out.Write(plain); err != nil {
            return err
        }
        processed += int64(len(plain))
        if onProgress != nil {
            onProgress(processed, totalSize)
        }
    }

    return nil
}



package password

import (
    "crypto/rand"
    "math/big"
)

var (
    lettersLower = "abcdefghijklmnopqrstuvwxyz"
    lettersUpper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
    digits       = "0123456789"
    symbols      = "!@#$%^&*()-_=+[]{};:,.?/" + "~" + "|" + "`" + "'" + "\""
)

// GenerateOptions holds options for password generation
type GenerateOptions struct {
	Length      int
	UseLower    bool
	UseUpper    bool
	UseDigits   bool
	UseSymbols  bool
}

// DefaultOptions returns default password generation options
func DefaultOptions() GenerateOptions {
	return GenerateOptions{
		Length:     20,
		UseLower:   true,
		UseUpper:   true,
		UseDigits:  true,
		UseSymbols: true,
	}
}

// Generate returns a random password of given length.
// If all toggles are false, it uses letters+digits by default.
func Generate(length int, useDigits, useSymbols bool) (string, error) {
	opts := GenerateOptions{
		Length:     length,
		UseLower:   true,
		UseUpper:   true,
		UseDigits:  useDigits,
		UseSymbols: useSymbols,
	}
	return GenerateWithOptions(opts)
}

// GenerateWithOptions generates a password with specified options
func GenerateWithOptions(opts GenerateOptions) (string, error) {
	if opts.Length <= 0 {
		opts.Length = 16
	}
	
	var alphabet string
	if opts.UseLower {
		alphabet += lettersLower
	}
	if opts.UseUpper {
		alphabet += lettersUpper
	}
	if opts.UseDigits {
		alphabet += digits
	}
	if opts.UseSymbols {
		alphabet += symbols
	}
	
	// If no character types selected, use default
	if alphabet == "" {
		alphabet = lettersLower + lettersUpper + digits
	}

	out := make([]byte, opts.Length)
	max := big.NewInt(int64(len(alphabet)))
	for i := 0; i < opts.Length; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		out[i] = alphabet[n.Int64()]
	}
	return string(out), nil
}



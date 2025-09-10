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

// Generate returns a random password of given length.
// If all toggles are false, it uses letters+digits by default.
func Generate(length int, useDigits, useSymbols bool) (string, error) {
    if length <= 0 {
        length = 16
    }
    alphabet := lettersLower + lettersUpper
    if useDigits {
        alphabet += digits
    }
    if useSymbols {
        alphabet += symbols
    }

    out := make([]byte, length)
    max := big.NewInt(int64(len(alphabet)))
    for i := 0; i < length; i++ {
        n, err := rand.Int(rand.Reader, max)
        if err != nil {
            return "", err
        }
        out[i] = alphabet[n.Int64()]
    }
    return string(out), nil
}



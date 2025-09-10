package password

import "unicode"

// StrengthScore returns a score in [0,1] and a label.
func StrengthScore(pw string) (float64, string) {
    if len(pw) == 0 {
        return 0.0, "Empty"
    }
    var hasLower, hasUpper, hasDigit, hasSymbol bool
    for _, r := range pw {
        switch {
        case unicode.IsLower(r):
            hasLower = true
        case unicode.IsUpper(r):
            hasUpper = true
        case unicode.IsDigit(r):
            hasDigit = true
        default:
            hasSymbol = true
        }
    }
    variety := 0
    if hasLower {
        variety++
    }
    if hasUpper {
        variety++
    }
    if hasDigit {
        variety++
    }
    if hasSymbol {
        variety++
    }

    lengthScore := float64(len(pw)) / 20.0
    if lengthScore > 1 {
        lengthScore = 1
    }
    varietyScore := float64(variety) / 4.0
    score := 0.6*lengthScore + 0.4*varietyScore

    label := "Weak"
    switch {
    case score >= 0.85:
        label = "Very Strong"
    case score >= 0.7:
        label = "Strong"
    case score >= 0.5:
        label = "Medium"
    default:
        label = "Weak"
    }
    return score, label
}



package ui

import "fmt"

func HumanBytes(n int64) string {
    const unit = 1024
    if n < unit {
        return fmt.Sprintf("%d B", n)
    }
    div, exp := int64(unit), 0
    for m := n / unit; m >= unit; m /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}



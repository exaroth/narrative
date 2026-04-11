package kitten

import (
	"strings"
)

// Remove trailing hyphen from each phoneme if present
func removeTrailingHyphens(input string) string {
	parts := strings.Split(input, " ")
	for idx, p := range parts {
		parts[idx] = strings.TrimLeft(p, "'")
	}
	return strings.Join(parts, " ")
}

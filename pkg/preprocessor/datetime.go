package preprocessor

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	DECADES = map[uint8]string{
		0: "hundreds", 1: "tens", 2: "twenties", 3: "thirties", 4: "forties",
		5: "fifties", 6: "sixties", 7: "seventies", 8: "eighties", 9: "nineties",
	}
	DECADES_RE = regexp.MustCompile(`\b(\d{1,3})0s\b`)
)

// Expand decades into words, eg 90s -> nineties
func expandDecades(input string) (string, error) {
	for _, g := range DECADES_RE.FindAllStringSubmatch(input, -1) {
		base, err := strconv.Atoi(g[1])
		if err != nil {
			return "", fmt.Errorf("Err expanding decades for %s, %w", input, err)
		}
		var decade_w string
		if w, ok := DECADES[uint8(base%10)]; ok {
			decade_w = w
		}
		if base < 10 {
			input = strings.ReplaceAll(input, g[0], decade_w)
			continue
		}
		century_part := base / 10
		input = strings.ReplaceAll(
			input,
			g[0],
			fmt.Sprintf("%s %s", numberToWords(century_part), decade_w),
		)
	}
	return input, nil
}

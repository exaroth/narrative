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
	// todo ignore case
	TIME_RE = regexp.MustCompile(`\b(\d{1,2}):(\d{2})(?::(\d{2}))?\s*(am|pm)?\b`)
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

func expandTime(input string) (string, error) {
	for _, g := range TIME_RE.FindAllStringSubmatch(input, -1) {
		hours, err := strconv.Atoi(g[1])
		mins, err := strconv.Atoi(g[2])
		if err != nil {
			return "", fmt.Errorf("Err expanding time for %s, %w", input, err)
		}
		hours_s := numberToWords(hours)
		mins_s := numberToWords(mins)
		suffix := strings.ToLower(strings.Trim(g[4], " "))
		if mins == 0 {
			if len(suffix) > 0 {
				input = strings.ReplaceAll(
					input,
					g[0],
					fmt.Sprintf("%s %s", hours_s, suffix),
				)
			} else {
				input = strings.ReplaceAll(
					input,
					g[0],
					fmt.Sprintf("%s hundred", hours_s),
				)
			}
		} else if mins < 10 {
			input = strings.ReplaceAll(
				input,
				g[0],
				fmt.Sprintf("%s oh %s %s", hours_s, mins_s, suffix),
			)
		} else {
			input = strings.ReplaceAll(
				input,
				g[0],
				fmt.Sprintf("%s %s %s", hours_s, mins_s, suffix),
			)
		}
	}
	return input, nil
}

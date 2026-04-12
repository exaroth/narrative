package preprocessor

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	UNIT_MAP = map[string]string{
		"km": "kilometers", "kg": "kilograms", "mg": "milligrams",
		"ml": "milliliters", "gb": "gigabytes", "mb": "megabytes",
		"kb": "kilobytes", "tb": "terabytes",
		"hz": "hertz", "khz": "kilohertz", "mhz": "megahertz", "ghz": "gigahertz",
		"mph": "miles per hour", "kph": "kilometers per hour",
		"ms": "milliseconds", "ns": "nanoseconds", "µs": "microseconds",
		"°c": "degrees Celsius", "c°": "degrees Celsius",
		"°f": "degrees Fahrenheit", "f°": "degrees Fahrenheit",
	}
	CURRENCY_SYMBOLS = map[string]string{
		"$": "dollar", "€": "euro", "£": "pound", "¥": "yen",
		"₹": "rupee", "₩": "won", "₿": "bitcoin",
	}
	SCALE_MAP = map[string]string{
		"k": "thousand", "m": "million", "b": "billion", "t": "trillion",
	}

	PERCENT_RE  = regexp.MustCompile(`(-?[\d,]+(?:\.\d+)?)\s*%`)
	UNIT_RE     = regexp.MustCompile(`(\d+(?:\.\d+)?)\s*(km|kg|mg|ml|GB|gb|MB|mb|KB|kb|TB|tb|hz|khz|mhz|ghz|mph|kph|°[cCfF]|[cCfF]°|ms|ns|µs)\b`)
	CURRENCY_RE = regexp.MustCompile(`([$€£¥₹₩₿])\s*([\d,]+(?:\.\d+)?)\s*([KMBTkmbt])?`)
)

// Expand percentage values into words.
func expandPercentages(input string) (string, error) {
	for _, g := range PERCENT_RE.FindAllStringSubmatch(input, -1) {
		if strings.Contains(g[0], ".") {
			parts := strings.Split(g[1], ".")
			if len(parts) != 2 {
				return "", fmt.Errorf("Bar percentage %s passed in %s", g[0], input)
			}
			base, err := strconv.Atoi(parts[0])
			if err != nil {
				return "", fmt.Errorf("Bad perc %s in %s, err: %w", g[0], input, err)
			}
			input = strings.ReplaceAll(
				input,
				g[0],
				floatToWords(base, parts[1])+" percent",
			)
			continue
		}
		num, err := strconv.Atoi(g[1])
		if err != nil {
			return "", fmt.Errorf("Bad perc %s in %s, err: %w", g[0], input, err)
		}
		input = strings.ReplaceAll(
			input,
			g[0],
			numberToWords(num)+" percent",
		)
	}
	return input, nil
}

// Expand common units of measurement eg. km, mb, ml etc.
// into words.
func expandUnits(input string) (string, error) {
	var unit string
	var val string
	for _, g := range UNIT_RE.FindAllStringSubmatch(input, -1) {
		if u, ok := UNIT_MAP[strings.ToLower(g[2])]; ok {
			unit = u
		} else {
			unit = g[2]
		}
		if strings.Contains(g[0], ".") {
			parts := strings.Split(g[1], ".")
			if len(parts) != 2 {
				return "", fmt.Errorf("Bad unit string %s passed in %s", g[0], input)
			}
			base, err := strconv.Atoi(parts[0])
			if err != nil {
				return "", fmt.Errorf("Bad unit val %s in %s, err: %w", g[0], input, err)
			}
			val = floatToWords(base, parts[1])
		} else {
			num, err := strconv.Atoi(g[1])
			if err != nil {
				return "", fmt.Errorf("Bad unit val %s in %s, err: %w", g[0], input, err)
			}
			val = numberToWords(num)
		}
		input = strings.ReplaceAll(
			input,
			g[0],
			fmt.Sprintf("%s %s", val, unit),
		)
	}
	return input, nil
}

// Expand currency (eg $20 ) to words.
func expandCurrency(input string) (string, error) {
	var cur string
	var val string
	var exp string
	for _, g := range CURRENCY_RE.FindAllStringSubmatch(input, -1) {
		if u, ok := CURRENCY_SYMBOLS[g[1]]; ok {
			cur = u
		} else {
			return "", fmt.Errorf("Fixme, invalid currency passed %s in %s", g[0], input)
		}
		if len(g[3]) > 0 {
			if u, ok := SCALE_MAP[strings.ToLower(g[3])]; ok {
				exp = u
			}
		}
		if strings.Contains(g[2], ".") {
			parts := strings.Split(g[2], ".")
			if len(parts) != 2 {
				return "", fmt.Errorf("Bad currency str %s passed in %s", g[0], input)
			}
			base, err := strconv.Atoi(parts[0])
			if err != nil {
				return "", fmt.Errorf("Bad currency str %s in %s, err: %w", g[0], input, err)
			}
			val = floatToWords(base, parts[1])
			if base != 1 && base != -1 {
				cur = cur + "s"
			}
		} else {
			num, err := strconv.Atoi(g[2])
			if err != nil {
				return "", fmt.Errorf("Bad currency str %s in %s, err: %w", g[0], input, err)
			}
			val = numberToWords(num)
			if num != 1 && num != -1 {
				cur = cur + "s"
			}
		}

		if len(exp) > 0 {
			input = strings.ReplaceAll(
				input,
				g[0],
				fmt.Sprintf("%s %s %s", val, exp, cur),
			)
		} else {
			input = strings.ReplaceAll(
				input,
				g[0],
				fmt.Sprintf("%s %s ", val, cur),
			)
		}
	}
	return input, nil
}

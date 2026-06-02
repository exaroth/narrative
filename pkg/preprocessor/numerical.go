package preprocessor

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var (
	ONES = []string{
		"", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine",
		"ten", "eleven", "twelve", "thirteen", "fourteen", "fifteen", "sixteen",
		"seventeen", "eighteen", "nineteen",
	}
	TENS = []string{
		"", "", "twenty", "thirty", "forty", "fifty",
		"sixty", "seventy", "eighty", "ninety",
	}
	SCALE = []string{"", "thousand", "million", "billion", "trillion"}

	ORDINAL_EXCEPTIONS = map[string]string{
		"one": "first", "two": "second", "three": "third", "four": "fourth",
		"five": "fifth", "six": "sixth", "seven": "seventh", "eight": "eighth",
		"nine": "ninth", "twelve": "twelfth",
	}

	ROMAN_NUM_MAP = map[int]string{
		1000: "M", 900: "CM", 500: "D",
		400: "CD", 100: "C", 90: "XC",
		50: "L", 40: "XL", 10: "X",
		9: "IX", 5: "V", 4: "IV",
		1: "I",
	}

	NUMBER_RE      = regexp.MustCompile(`([-])?[\d]+(?:\.\d+)?`)
	FRACTION_RE    = regexp.MustCompile(`\b(\d+)\s*/\s*(\d+)\b`)
	LEADING_DEC_RE = regexp.MustCompile(`(^|\s+)(-)?\.([\d]+)+\b`)
	ORDINAL_RE     = regexp.MustCompile(`\b(\d+)(st|nd|rd|th)\b`)
	ROMAN_RE       = regexp.MustCompile(`\b(M{0,4})(CM|CD|D?C{0,3})(XC|XL|L?X{0,3})(IX|IV|V?I{0,3})\b`)
	TITLE_WORDS    = `(war|chapter|part|volume|act|scene|book|section|article|` +
		`king|queen|pope|louis|henry|edward|george|william|james|` +
		"phase|round|level|stage|class|type|version|episode|season|letter)"
	TITLE_WORDS_RE = regexp.MustCompile(`(?i)` + TITLE_WORDS)
)

// Expand ordinal numbers eg. 1st -> first.
func expandOrdinals(input string) (string, error) {
	for _, g := range ORDINAL_RE.FindAllStringSubmatch(input, -1) {
		ord, err := strconv.Atoi(g[1])
		if err != nil {
			return "", fmt.Errorf("Error expanding ordinal for %s, %w", input, err)
		}
		input = strings.ReplaceAll(input, g[0], ordinalSuffix(ord))
	}
	return input, nil
}

// Expand fractions into words, eg. 1/2 -> one half,
// 2/3 -> two thirds.
func expandFractions(input string) (string, error) {
	for _, g := range FRACTION_RE.FindAllStringSubmatch(input, -1) {
		first, err := strconv.Atoi(g[1])
		second, err := strconv.Atoi(g[2])
		if err != nil {
			return "", fmt.Errorf("Error converting fractions for %s %w", input, err)
		}
		numerator := numberToWords(first)
		var denom string
		switch second {
		case 2:
			if first == 1 {
				denom = "half"
			} else {
				denom = "halves"
			}
		case 4:
			if first == 1 {
				denom = "quarter"
			} else {
				denom = "quarters"
			}
		default:
			denom = ordinalSuffix(second)
			if first != 1 {
				denom = denom + "s"
			}

		}
		input = strings.ReplaceAll(input, g[0], fmt.Sprintf("%s %s", numerator, denom))
	}
	return input, nil
}

// Convert roman numerals to human readable format.
func expandRomanNumerals(input string) (string, error) {
	s_r := []rune(input)
	for _, g := range ROMAN_RE.FindAllStringIndex(input, -1) {
		if g[1]-g[0] > 0 {
			if g[1]-g[0] == 1 {
				// Check if roman numeral is prefixed with title.
				title_s := string(s_r[max(0, g[0]-30):g[0]])
				if len(TITLE_WORDS_RE.FindStringIndex(title_s)) == 0 {
					return input, nil
				}
			}
			r_s := string(s_r[g[0]:g[1]])
			as_int := romanToInt(r_s)

			// we assume same roman numeral wont be repeated
			// more than once.
			input = strings.ReplaceAll(input, r_s, numberToWords(as_int))
		}
	}
	return input, nil
}

// Expand decimals without leading zero (eg .8) to make it easier
// to process further down the pipeline.
func expandLeadingDecimals(input string) (string, error) {
	for _, g := range LEADING_DEC_RE.FindAllStringSubmatch(input, -1) {
		parts := []string{g[1]}
		if len(g[2]) > 0 {
			parts = append(parts, "-")
		}
		parts = append(parts, "0.")
		parts = append(parts, g[3])
		input = strings.ReplaceAll(input, g[0], strings.Join(parts, ""))
	}
	return input, nil
}

// Replace all numeric values (float and integers) with words,
// this function should be run at the end of processing pipeline.
func replaceNumbers(input string) (string, error) {

	if len(NUMBER_RE.FindStringIndex(input)) == 0 {
		return input, nil
	}
	for _, g := range NUMBER_RE.FindAllStringSubmatch(input, -1) {
		if strings.Contains(g[0], ".") {
			f_parts := strings.Split(g[0], ".")
			if len(f_parts) != 2 {
				return "", fmt.Errorf("Err replacing numbers for float %s @ %s", input, g[0])
			}
			base, err := strconv.Atoi(f_parts[0])
			if err != nil {
				return "", fmt.Errorf("Err replacing numbers %s @ %s, %w", input, g[0], err)
			}
			input = strings.ReplaceAll(input, g[0], floatToWords(base, f_parts[1]))
		} else {
			num, err := strconv.Atoi(g[0])
			if err != nil {
				return "", fmt.Errorf("Err replacing numbers %s @ %s, %w", input, g[0], err)
			}
			input = strings.ReplaceAll(input, g[0], numberToWords(num))
		}
	}

	return input, nil
}

// Convert integers up to 999 to words.
func threeDigitsToWords(num int) string {
	if num == 0 {
		return ""
	}
	parts := []string{}
	hundreds := num / 100
	remainder := num % 100
	if hundreds > 0 {
		parts = append(parts, fmt.Sprintf("%s hundred", ONES[hundreds]))
	}
	if remainder > 0 && remainder < 20 {
		parts = append(parts, ONES[remainder])
	} else {
		tens_w := TENS[remainder/10]
		ones_w := ONES[remainder%10]
		if len(ones_w) > 0 {
			parts = append(parts, fmt.Sprintf("%s %s", tens_w, ones_w))
		} else {
			parts = append(parts, tens_w)
		}
	}
	return strings.Join(parts, " ")
}

// Convert integer to words.
func numberToWords(num int) string {
	if num == 0 {
		return "zero"
	}
	if num < 0 {
		return fmt.Sprintf("minus %s", numberToWords(-num))
	}

	if 100 <= num && num <= 9999 && num%100 == 0 && num%1000 != 0 {
		hundreds := num / 100
		if hundreds < 20 {
			return fmt.Sprintf("%s hundred", ONES[hundreds])
		}
	}
	parts := []string{}
	for _, scale := range SCALE {
		chunk := num % 1000
		if chunk > 0 {
			c_words := threeDigitsToWords(chunk)
			if len(scale) > 0 {
				parts = append(parts, fmt.Sprintf("%s %s", c_words, scale))
			} else {
				parts = append(parts, c_words)
			}
		}
		num = num / 1000
		if num == 0 {
			break
		}
	}
	slices.Reverse(parts)
	return strings.Join(parts, " ")
}

// Convert float to words.
func floatToWords(base int, rest string) string {
	base_s := numberToWords(base)
	if len(rest) == 0 {
		return base_s
	}
	if base < 0 {
		return fmt.Sprintf("negative %s", floatToWords(-base, rest))
	}
	d_m := []string{"zero"}
	d_m = append(d_m, ONES[1:]...)
	digits := []string{}
	var err error
	var ci int
	for _, c := range rest {
		ci, err = strconv.Atoi(string(c))
		if err != nil {
			// todo
			continue
		}
		digits = append(digits, d_m[ci])
	}
	return fmt.Sprintf("%s point %s", base_s, strings.Join(digits, " "))
}

// Convert number to ordinal suffix
// eg 1 - first.
func ordinalSuffix(num int) string {
	words := strings.Split(numberToWords(num), " ")
	var prefix, last, last_ord string
	if len(words) == 2 {
		prefix, last = words[0], words[1]
	} else {
		prefix, last = "", words[0]
	}
	if _, ok := ORDINAL_EXCEPTIONS[last]; ok {
		last_ord = ORDINAL_EXCEPTIONS[last]
	} else {
		switch last[len(last)-1] {
		case 't':
			last_ord = last + "h"
		case 'e':
			last_ord = last[0:len(last)-1] + "th"
		default:
			last_ord = last + "th"
		}
	}
	if len(prefix) > 0 {
		return fmt.Sprintf("%s %s", prefix, last_ord)
	}
	return last_ord
}

// Reverse input string order.
func reverseString(s string) (result string) {
	for _, v := range s {
		result = string(v) + result
	}
	return
}

// Convert roman numeral to integer.
func romanToInt(in string) int {
	roman_map := map[rune]int{
		'I': 1, 'V': 5, 'X': 10, 'L': 50,
		'C': 100, 'D': 500, 'M': 1000,
	}
	var result, prev int
	for _, ch := range reverseString(in) {
		if val, ok := roman_map[ch]; ok {
			if val >= prev {
				result += val
			} else {
				result -= val
			}
			prev = val
		} else {
			return result
		}
	}
	return result
}

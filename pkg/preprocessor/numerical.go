package preprocessor

import (
	"fmt"
	"slices"
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
)

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

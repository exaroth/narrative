package preprocessor

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/k0kubun/pp"
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

	NUMBER_RE      = regexp.MustCompile(`([-])?[\d]+(?:\.\d+)?`)
	LEADING_DEC_RE = regexp.MustCompile(`(^|\s+)(-)?\.([\d]+)+\b`)
)

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

func replaceNumbers(input string) (string, error) {

	if len(NUMBER_RE.FindStringIndex(input)) == 0 {
		return input, nil
	}
	for _, g := range NUMBER_RE.FindAllStringSubmatch(input, -1) {
		fmt.Println(">>>>>>>>>>>> g")
		pp.Println(g)
		fmt.Println("<<<<<<<<<<<<")
		if strings.Contains(g[0], ".") {
			f_parts := strings.Split(g[0], ".")
			if len(f_parts) != 2 {
				// todo
				fmt.Println("unsupported string1: ", g[0])
				return input, nil
			}
			base, err := strconv.Atoi(f_parts[0])
			if err != nil {
				// todo
				fmt.Println("unsupported string2: ", g[0])
				return input, nil
			}
			input = strings.ReplaceAll(input, g[0], floatToWords(base, f_parts[1]))
		} else {
			num, err := strconv.Atoi(g[0])
			if err != nil {
				// todo
				fmt.Println("unsupported string: ", g[0])
				return input, nil
			}
			input = strings.ReplaceAll(input, g[0], numberToWords(num))
		}
	}

	return input, nil
}

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
			//todo
			panic(err)
		}
		digits = append(digits, d_m[ci])
	}
	return fmt.Sprintf("%s point %s", base_s, strings.Join(digits, " "))
}

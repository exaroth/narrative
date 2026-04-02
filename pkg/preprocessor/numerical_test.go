package preprocessor

import (
	"fmt"
	"testing"
)

func TestExpandingLeadingDecimals(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    ".22",
			expected: "0.22",
		},
		{
			input:    "percentage is .99%",
			expected: "percentage is 0.99%",
		},
		{
			input:    "malformed.99 0.11 *.22",
			expected: "malformed.99 0.11 *.22",
		},
	}
	for idx, test := range tests {
		testname := fmt.Sprintf("expand leading dec: %d", idx)
		t.Run(testname, func(t *testing.T) {

			result, _ := expandLeadingDecimals(test.input)
			if result != test.expected {
				t.Errorf("got %s, want %s", result, test.expected)
			}
		})
	}
}

func TestConvertingNumericalValues(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{
			input:    34,
			expected: "thirty four",
		},
		{
			input:    0,
			expected: "zero",
		},
		{
			input:    128,
			expected: "one hundred twenty eight",
		},
		{
			input:    1999,
			expected: "one thousand nine hundred ninety nine",
		},
		{
			input:    -8,
			expected: "minus eight",
		},
		{
			input:    20_000_001,
			expected: "twenty million one",
		},
	}
	for idx, test := range tests {
		testname := fmt.Sprintf("number to string: %d", idx)
		t.Run(testname, func(t *testing.T) {

			result := numberToWords(test.input)
			if result != test.expected {
				t.Errorf("got %s, want %s", result, test.expected)
			}
		})
	}
}

func TestConvertingFloatsToStrings(t *testing.T) {
	tests := []struct {
		base     int
		rest     string
		expected string
	}{
		{
			base:     3,
			rest:     "14",
			expected: "three point one four",
		},
		{
			base:     0,
			rest:     "29",
			expected: "zero point two nine",
		},
		{
			base:     -11,
			rest:     "01",
			expected: "negative eleven point zero one",
		},
		{
			base:     1,
			rest:     "2103012",
			expected: "one point two one zero three zero one two",
		},
		{
			base:     21993,
			rest:     "",
			expected: "twenty one thousand nine hundred ninety three",
		},
	}
	for idx, test := range tests {
		testname := fmt.Sprintf("float to string: %d", idx)
		t.Run(testname, func(t *testing.T) {

			result := floatToWords(test.base, test.rest)
			if result != test.expected {
				t.Errorf("got %s, want %s", result, test.expected)
			}
		})
	}
}

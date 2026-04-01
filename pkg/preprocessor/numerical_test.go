package preprocessor

import (
	"fmt"
	"testing"
)

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

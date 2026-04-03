package preprocessor

import (
	"fmt"
	"testing"
)

func TestExpandingPercentages(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "50% off",
			expected: "fifty percent off",
		},
		{
			input:    "2.99% price drop-off",
			expected: "two point nine nine percent price drop-off",
		},
		{
			input:    "-9% decrease",
			expected: "minus nine percent decrease",
		},
	}
	for idx, test := range tests {
		testname := fmt.Sprintf("expanding percentage: %d", idx)
		t.Run(testname, func(t *testing.T) {
			result, _ := expandPercentages(test.input)
			if result != test.expected {
				t.Errorf("got %s, want %s", result, test.expected)
			}
		})
	}
}

func TestExpandingiUnits(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "120km",
			expected: "one hundred twenty kilometers",
		},
		{
			input:    "9.2°C",
			expected: "nine point two degrees Celsius",
		},
		{
			input:    "2401 MB",
			expected: "two thousand four hundred one megabytes",
		},
	}
	for idx, test := range tests {
		testname := fmt.Sprintf("expanding units: %d", idx)
		t.Run(testname, func(t *testing.T) {
			result, _ := expandUnits(test.input)
			if result != test.expected {
				t.Errorf("got %s, want %s", result, test.expected)
			}
		})
	}
}

func TestExpandingCurrency(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "$100",
			expected: "one hundred dollars",
		},
		{
			input:    "£9.99",
			expected: "nine point nine nine pounds",
		},
		{
			input:    "$2.5M",
			expected: "five million dollars",
		},
		{
			input:    "€1",
			expected: "one euro",
		},
	}
	for idx, test := range tests {
		testname := fmt.Sprintf("expanding currency: %d", idx)
		t.Run(testname, func(t *testing.T) {
			result, _ := expandCurrency(test.input)
			if result != test.expected {
				t.Errorf("got %s, want %s", result, test.expected)
			}
		})
	}
}

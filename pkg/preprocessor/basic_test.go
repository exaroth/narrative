package preprocessor

import (
	"fmt"
	"testing"
)

func TestReplacingQuotes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "«this» “should” ’normalized’",
			expected: "\"this\" \"should\" 'normalized'",
		},
	}
	for idx, test := range tests {
		testname := fmt.Sprintf("Normalize quotes: %d", idx)
		t.Run(testname, func(t *testing.T) {

			result, _ := normalizeQuotes(test.input)
			if result != test.expected {
				t.Errorf("got %s, want %s", result, test.input)
			}
		})
	}
}

func TestSplittingHyphenizedWords(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "half-life -rocks-",
			expected: "half life - rocks -",
		},
		{
			input:    "thirty-five tokens",
			expected: "thirty five tokens",
		},
	}
	for idx, test := range tests {
		testname := fmt.Sprintf("Split hyphens: %d", idx)
		t.Run(testname, func(t *testing.T) {

			result, _ := splitHyphenizedWords(test.input)
			if result != test.expected {
				t.Errorf("got %s, want %s", result, test.input)
			}
		})
	}
}

func TestRemovingTrailingApostrophes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "people's rights' are important",
			expected: "peoples rights are important",
		},
	}
	for idx, test := range tests {
		testname := fmt.Sprintf("Remove trailing apostrophes: %d", idx)
		t.Run(testname, func(t *testing.T) {

			result, _ := removeTrailingApostrophes(test.input)
			if result != test.expected {
				t.Errorf("got %s, want %s", result, test.input)
			}
		})
	}
}

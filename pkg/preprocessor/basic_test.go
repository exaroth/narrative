package preprocessor

import (
	"fmt"
	"testing"
)

func testNormalizingPunctuation(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "«this» “should” ’normalized’",
			expected: "\"this\" \"should\" 'normalized'",
		},
		{
			input:    "half—life",
			expected: "half-life",
		},
	}
	for idx, test := range tests {
		testname := fmt.Sprintf("Normalize quotes: %d", idx)
		t.Run(testname, func(t *testing.T) {

			result, _ := normalizePunctuation(test.input)
			if result != test.expected {
				t.Errorf("got %s, want %s", result, test.expected)
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
				t.Errorf("got %s, want %s", result, test.expected)
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
			input:    "people's rights' must've been important",
			expected: "peoples rights must've been important",
		},
	}
	for idx, test := range tests {
		testname := fmt.Sprintf("Remove trailing apostrophes: %d", idx)
		t.Run(testname, func(t *testing.T) {

			result, _ := removeTrailingApostrophes(test.input)
			if result != test.expected {
				t.Errorf("got %s, want %s", result, test.expected)
			}
		})
	}
}

func TestRemovingWhitespace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "   This \r should     be \t normalized   \n",
			expected: "This should be normalized",
		},
	}
	for idx, test := range tests {
		testname := fmt.Sprintf("Normalize whitespace: %d", idx)
		t.Run(testname, func(t *testing.T) {

			result, _ := normalizeWhitespace(test.input)
			if result != test.expected {
				t.Errorf("got %s, want %s", result, test.expected)
			}
		})
	}
}

func TestRemovingPunctuation(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "testing, #@(){}[]#%$ '\" punctuation;  removal!== @",
			expected: "testing,                punctuation;  removal!    ",
		},
	}
	for idx, test := range tests {
		testname := fmt.Sprintf("Remove punctuation: %d", idx)
		t.Run(testname, func(t *testing.T) {

			result, _ := removeNonProsodicPunctuation(test.input)
			if result != test.expected {
				t.Errorf("got %s, want %s", result, test.expected)
			}
		})
	}
}

func TestRemovingUnusableTextParts(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "https://www.test.com?test=1 https://www.google.com",
			expected: " ",
		},
		{
			input:    "test@test.com",
			expected: "",
		},
		{
			input:    "<p>test</p>",
			expected: "test",
		},
		{
			input:    "@test",
			expected: "",
		},
		{
			input:    "#test",
			expected: "",
		},
	}
	for idx, test := range tests {
		testname := fmt.Sprintf("Cleanup trash: %d", idx)
		t.Run(testname, func(t *testing.T) {

			result, _ := cleanupUnusableTextParts(test.input)
			if result != test.expected {
				t.Errorf("got |%s|, want |%s|", result, test.expected)
			}
		})
	}
}

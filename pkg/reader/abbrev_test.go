package reader

import (
	"fmt"
	"testing"
)

func TestExpandingAbbreviations(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "asst. to dr. Kleimer was Donald jr.",
			expected: "assistant to doctor Kleimer was Donald junior",
		},
		{
			input:    "50 c. is not 50 C.",
			expected: "50 cent is not 50 C.",
		},
		{
			input:    "rev Johnson was born on first of jan.",
			expected: "reverend Johnson was born on first of January",
		},
		{
			input:    "No, v. 1.2.3 was the last no. that mattered",
			expected: "No, version 1.2.3 was the last number that mattered",
		},
		{
			input:    "Claude.md is located at ./cmd/claude.md",
			expected: "Claude.md is located at ./cmd/claude.md",
		},
	}
	for idx, test := range tests {
		testname := fmt.Sprintf("expanding abbrev: %d", idx)
		t.Run(testname, func(t *testing.T) {
			result := ExpandAbbreviations(test.input)
			if result != test.expected {
				t.Errorf("got %s, want %s", result, test.expected)
			}
		})
	}
}

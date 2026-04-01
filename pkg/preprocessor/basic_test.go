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

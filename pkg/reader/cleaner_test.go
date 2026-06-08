package reader

import (
	"fmt"
	"testing"
)

func TestCleaningExcerpts(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "[201] [20] [1] Text",
			expected: "   Text",
		},
		{
			input:    "[notanumber]",
			expected: "[notanumber]",
		},
	}
	for idx, test := range tests {
		testname := fmt.Sprintf("clean excerpt: %d", idx)
		t.Run(testname, func(t *testing.T) {

			result := cleanExcerpts(test.input)
			if result != test.expected {
				t.Errorf("got %s, want %s", result, test.expected)
			}
		})
	}
}

package preprocessor

import (
	"fmt"
	"testing"
)

func TestExpandingContractions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "He'll be, I haven't any idea",
			expected: "He will be, I haven't any idea",
		},
		{
			input:    "I could've done more but i won't",
			expected: "I could have done more but i won't",
		},
		{
			input:    "I can't",
			expected: "I can't",
		},
		{
			input:    "I'd like some",
			expected: "I would like some",
		},
		{
			input:    "Haven't done",
			expected: "Haven't done",
		},
	}
	for idx, test := range tests {
		testname := fmt.Sprintf("expand contractions: %d", idx)
		t.Run(testname, func(t *testing.T) {

			result, _ := expandContractions(test.input)
			if result != test.expected {
				t.Errorf("got %s, want %s", result, test.expected)
			}
		})
	}
}

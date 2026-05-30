package preprocessor

import (
	"fmt"
	"testing"
)

func TestPreprocessingSentences(t *testing.T) {

	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "half—life could've been at 50% Capacity, but will decrease by 5 in year 2092",
			expected: "half life could have been at fifty percent capacity, but will decrease by five in year two thousand ninety two.",
		},
		{
			input:    ".5 percent of people use taxes.website.co.uk website.",
			expected: "zero point five percent of people use taxes dot website dot co dot uk website.",
		},
		{
			input:    "$5 is too much",
			expected: "five dollars is too much.",
		},
		{
			input:    "Chapter VI was 0.00002% longer.",
			expected: "chapter six was zero point zero zero zero zero two percent longer.",
		},
	}
	for idx, test := range tests {
		testname := fmt.Sprintf("Preprocessor test: %d", idx)
		proc := NewPreprocessor()
		t.Run(testname, func(t *testing.T) {
			result := proc.ProcessSentence(test.input)
			if result != test.expected {
				t.Errorf("got %s, want %s", result, test.expected)
			}
		})
	}
}

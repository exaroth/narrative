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
		// {
		// 	input:    "half—life could've been at 50% Capacity, but will decrease by 5 in year 2092",
		// 	expected: "half life could have been at fifty percent, but will decrease by five in year twenty ninenty two",
		// },
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

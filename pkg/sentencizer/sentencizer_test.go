package sentencizer

import (
	"fmt"
	"testing"
)

func TestSentencizer(t *testing.T) {

	tests := []struct {
		input    string
		expected []string
	}{
		{
			input: ` First sentence.
			Second sentence.
			`,
			expected: []string{"First sentence.", "Second sentence."},
		},
		{
			input: `Testing.

			Empty lines.
			`,
			expected: []string{"Testing.", "Empty lines."},
		},
		{
			input:    `"This is sentence wrapped in quotes. Sometimes used to denote dialog. It should be split into multiple sentences."`,
			expected: []string{"This is sentence wrapped in quotes.", "Sometimes used to denote dialog.", "It should be split into multiple sentences."},
		},
	}

	for idx, test := range tests {
		testname := fmt.Sprintf("Test sentences: %d", idx)
		t.Run(testname, func(t *testing.T) {
			out := Sentencize([]byte(test.input))
			if len(out) != len(test.expected) {
				t.Errorf("Invalid len, got %d, want %d, out %v", len(out), len(test.expected), out)
			} else {
				for i, s := range out {
					if s != test.expected[i] {
						t.Errorf("Invalid sentence at idx %d, got %s, want %s", i, s, test.expected[i])
					}
				}
			}
		})
	}
}

func TestSplittingLongSentences(t *testing.T) {

	tests := []struct {
		input     string
		desired_l int
		range_l   int
		expected  []string
		delims    []rune
	}{
		{
			input:     `String 1, String 2`,
			expected:  []string{"String 1,", "String 2"},
			desired_l: 8,
			range_l:   2,
			delims:    []rune{',', '.'},
		},
		{
			input:     `String 1 String 2 String 3`,
			expected:  []string{"String 1", "String 2", "String 3"},
			desired_l: 6,
			range_l:   2,
			delims:    []rune{',', '.'},
		},
	}

	for idx, test := range tests {
		testname := fmt.Sprintf("Test splitting long sentences: %d", idx)
		t.Run(testname, func(t *testing.T) {
			out := SplitLongSentence(test.input, test.desired_l, test.range_l, test.delims)
			if len(out) != len(test.expected) {
				t.Errorf("Invalid len, got %d, want %d, out %v", len(out), len(test.expected), out)
			} else {
				for i, s := range out {
					if s != test.expected[i] {
						t.Errorf("Invalid sentence at idx %d, got %s, want %s", i, s, test.expected[i])
					}
				}
			}
		})
	}
}

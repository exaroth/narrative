package phonemizer

import (
	"fmt"
	"testing"
)

func TestSplittingPunctuation(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
		marks    []*Mark
	}{
		{
			input:    " ; ",
			expected: []string{""},
			marks: []*Mark{
				{
					char:  ";",
					pad:   uint8(0),
					index: 0,
				},
			},
		},
		{
			input:    "doesn't",
			expected: []string{"doesn't"},
			marks:    []*Mark{},
		},
		{
			input:    "Multi — byte",
			expected: []string{"Multi", "", "byte"},
			marks: []*Mark{
				{
					char:  "—",
					pad:   uint8(0),
					index: 1,
				},
			},
		},
		{
			input:    "\"Working, but could be better\"",
			expected: []string{"Working", "but", "could", "be", "better"},
			marks: []*Mark{
				{
					char:  "\"",
					pad:   uint8(1),
					index: 0,
				},
				{
					char:  ",",
					pad:   uint8(2),
					index: 0,
				},
				{
					char:  "\"",
					pad:   uint8(2),
					index: 4,
				},
			},
		},
		{
			input:    "Test : 1",
			expected: []string{"Test", "", "1"},
			marks: []*Mark{
				{
					char:  ":",
					pad:   uint8(0),
					index: 1,
				},
			},
		},
		{
			input:    "[test]!",
			expected: []string{"test"},
			marks: []*Mark{
				{
					char:  "[",
					pad:   uint8(1),
					index: 0,
				},
				{
					char:  "]",
					pad:   uint8(2),
					index: 0,
				},
				{
					char:  "!",
					pad:   uint8(2),
					index: 0,
				},
			},
		},
	}
	for idx, test := range tests {
		testname := fmt.Sprintf("%d,%s", idx, test.input)
		len_check := t.Run(testname, func(t *testing.T) {
			split, marks := SplitPunctuation(test.input)
			if len(split) != len(test.expected) {
				t.Errorf("Invalid length for punctuation sanitized array , got %d, want %d, t: %d",
					len(split),
					len(test.expected),
					idx,
				)
			}
			if len(marks) != len(test.marks) {
				t.Errorf("Invalid length for punct marks len, got %d, want %d, t: %d",
					len(marks),
					len(test.marks),
					idx,
				)
			}
		})
		if len_check {
			t.Run(testname, func(t *testing.T) {
				split, _ := SplitPunctuation(test.input)
				for i := range split {
					if split[i] != test.expected[i] {
						t.Errorf("Invalid value returned at index %d, got %s, want %s, t: %d",
							i,
							split[i],
							test.expected[i],
							idx,
						)
					}
				}
			})
			t.Run(testname, func(t *testing.T) {
				_, marks := SplitPunctuation(test.input)
				for i := range marks {
					if marks[i].char != test.marks[i].char {
						t.Errorf("Invalid mark char returned at index %d, got %s, want %s, t: %d",
							i,
							marks[i].char,
							test.marks[i].char,
							idx,
						)
					}
				}
			})
		}
	}
}

func TestCombiningPunctuation(t *testing.T) {

	tests := []struct {
		input_text  []string
		input_marks []*Mark
		expected    string
	}{
		{
			input_text: []string{"Test", "string"},
			input_marks: []*Mark{
				{
					char:  ",",
					pad:   uint8(2),
					index: 0,
				},
				{
					char:  "!",
					pad:   uint8(2),
					index: 1,
				},
			},
			expected: "Test, string!",
		},
		{
			input_text: []string{"test"},
			input_marks: []*Mark{
				{
					char:  "[",
					pad:   uint8(1),
					index: 0,
				},
				{
					char:  "]",
					pad:   uint8(2),
					index: 0,
				},
				{
					char:  "!",
					pad:   uint8(2),
					index: 0,
				},
			},
			expected: "[test]!",
		},
		{
			input_text: []string{"test", "", "string"},
			input_marks: []*Mark{
				{
					char:  ":",
					pad:   uint8(0),
					index: 1,
				},
				{
					char:  ":",
					pad:   uint8(0),
					index: 1,
				},
			},
			expected: "test :: string",
		},
	}

	for idx, test := range tests {
		testname := fmt.Sprintf("Punctuation combining, %d", idx)
		t.Run(testname, func(t *testing.T) {
			result := CompactPunctuation(test.input_text, test.input_marks)
			if result != test.expected {
				t.Errorf("Invalid compacted string, got %s, want %s, t: %d", result, test.expected, idx)
			}

		})
	}
}

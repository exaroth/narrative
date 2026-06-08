package preprocessor

import (
	"fmt"
	"testing"
)

func TestExpandingYears(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "In january 2020",
			expected: "In january twenty twenty",
		},
		{
			input:    "From 1905 to march 1990",
			expected: "From nineteen oh five to march nineteen ninety",
		},
		{
			input:    "2998 is a great number",
			expected: "2998 is a great number",
		},
		{
			input:    "May of 1410 was a great month",
			expected: "May of fourteen ten was a great month",
		},
	}
	for idx, test := range tests {
		testname := fmt.Sprintf("expand years: %d", idx)
		t.Run(testname, func(t *testing.T) {

			result, _ := expandYears(test.input)
			if result != test.expected {
				t.Errorf("got %s, want %s", result, test.expected)
			}
		})
	}
}

func TestExpandingTime(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "3:30pm",
			expected: "three thirty pm",
		},
		{
			input:    "11:05",
			expected: "eleven oh five ",
		},
		{
			input:    "14:00",
			expected: "fourteen hundred",
		},
		{
			input:    "12:00pm",
			expected: "twelve pm",
		},
	}
	for idx, test := range tests {
		testname := fmt.Sprintf("expand time: %d", idx)
		t.Run(testname, func(t *testing.T) {

			result, _ := expandTime(test.input)
			if result != test.expected {
				t.Errorf("got %s, want %s", result, test.expected)
			}
		})
	}
}

func TestExpandingDecades(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "80s",
			expected: "eighties",
		},
		{
			input:    "1980s",
			expected: "nineteen eighties",
		},
		{
			input:    "90s",
			expected: "nineties",
		},
		{
			input:    "2020s",
			expected: "twenty twenties",
		},
	}
	for idx, test := range tests {
		testname := fmt.Sprintf("expand decades: %d", idx)
		t.Run(testname, func(t *testing.T) {

			result, _ := expandDecades(test.input)
			if result != test.expected {
				t.Errorf("got %s, want %s", result, test.expected)
			}
		})
	}
}

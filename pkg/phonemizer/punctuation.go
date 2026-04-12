package phonemizer

import (
	"regexp"
	"slices"
	"strings"
)

const PUNCTUATION = ";:,.!?¡¿—…\"«»“”(){}\\[\\]'"

var DEFAULT_MARKS_RE = regexp.MustCompile("[" + PUNCTUATION + "]")
var LETTERS_RE = regexp.MustCompile("[a-zA-Z]")

// Mark represents data associated with
// single punctuation character.
type Mark struct {
	// Punctuation character
	char string
	// Index of the word
	index int
	// Whether punctuation belongs before
	// the word (1) after the word (2)
	// or is separate (eg " - ") (0).
	// Punctuation in the middle of words
	// is not processed.
	pad uint8
}

type Punctuation []*Mark

// Convert punctuation data to array containing pre- and post-
// prefixes for the sentence words.
func (p Punctuation) AsArr(text_len int) [][2]string {

	p_a := make([][2]string, text_len)

	for _, mark := range p {
		switch mark.pad {
		case 0:
			fallthrough
		case 1:
			p_a[mark.index][0] = p_a[mark.index][0] + mark.char
		case 2:
			p_a[mark.index][1] = p_a[mark.index][1] + mark.char
		default:
			// TODO
			panic("Invalid pad value")
		}
	}
	return p_a

}

// Strip punctuation from given sentence, returning punctuation
// free slice of words and punctuation marks.
func SplitPunctuation(text string) ([]string, Punctuation) {
	text = strings.Trim(text, " ")
	t_a := strings.Split(text, " ")
	m_a := []*Mark{}
	for t_idx, word := range t_a {
		if len(word) == 0 {
			continue
		}
		punctuation_marks := DEFAULT_MARKS_RE.FindAllStringIndex(word, -1)
		if len(punctuation_marks) == 0 {
			continue
		}
		var pad uint8 = 0
		l_res := LETTERS_RE.FindAllStringIndex(word, -1)
		for _, m := range punctuation_marks {
			if len(l_res) > 0 {
				if m[0] < l_res[0][0] {
					pad = 1
				} else if m[0] > l_res[len(l_res)-1][0] {
					pad = 2
				} else {
					// don't process punctuation in the middle of the word
					continue
				}
			}
			m_a = append(m_a, &Mark{
				char:  string(word[m[0]:m[1]]),
				pad:   pad,
				index: t_idx,
			})
		}
	}

	for _, mark := range m_a {
		if mark.pad == 0 {
			t_a[mark.index] = ""
			continue
		}

		ss := strings.Split(t_a[mark.index], "")
		idx := slices.Index(ss, mark.char)

		if idx == -1 {
			continue
		}

		if mark.pad == 1 {
			ss = ss[idx+1:]
		} else {
			ss = ss[:idx]
		}
		t_a[mark.index] = strings.Join(ss, "")
	}
	return t_a, m_a
}

// Recombine punctuation back into the sentence.
func CompactPunctuation(text []string, punctuation Punctuation) string {
	p_a := punctuation.AsArr(len(text))
	result := make([]string, len(text))

	for i, t := range text {
		p_m := p_a[i]
		if len(p_m[0]) > 0 {
			t = p_m[0] + t
		}
		if len(p_m[1]) > 0 {
			t = t + p_m[1]
		}
		result[i] = t
	}
	return strings.Join(result, " ")
}

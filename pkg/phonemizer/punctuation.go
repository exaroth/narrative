package phonemizer

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/k0kubun/pp"
)

const PUNCTUATION = ";:,.!?¡¿—…\"«»“”(){}\\[\\]'"

var DEFAULT_MARKS_RE = regexp.MustCompile("[" + PUNCTUATION + "]")
var LETTERS_RE = regexp.MustCompile("[a-zA-Z]")

type Mark struct {
	char  string
	index int
	pad   uint8
}

func SplitPunctuation(text string) ([]string, []*Mark) {
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

func CompactPunctuation(text []string, punctuation []*Mark) string {
	p_a := make([][2]string, len(text))
	result := make([]string, len(text))

	for _, mark := range punctuation {
		fmt.Println(">>>>>>>>>>>> mark")
		pp.Println(mark)
		fmt.Println("<<<<<<<<<<<<")
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

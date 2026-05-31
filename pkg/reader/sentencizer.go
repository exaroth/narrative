package reader

import (
	"bufio"
	"bytes"
	"regexp"
	"slices"
	"strings"

	"github.com/sentencizer/sentencizer"
)

var MULTISPACE_STRIP_RE = regexp.MustCompile(`\s\s+`)

// max number of tokens that can be passed to KittenTTS model.
// excludes spaces?
const MAX_SENTENCE_LENGTH int = 398

var (
	SENTENCE_SPLIT_DELIMITERS = []rune{',', ';', '.', '?', '!'}
	SENTENCE_WRAP_CHARS       = []string{"\"", "'"}
)

const (
	SENTENCE_DESIRED_L = 299
	SENTENCE_DESIRED_R = 50
)

// Split input string into sentence slice.
func Sentencize(text []byte) []string {
	chunks := []string{}

	t := string(text)
	t = MULTISPACE_STRIP_RE.ReplaceAllString(t, " ")
	// We want to expand abbreviations as soon as possible
	// to avoid having issues with splitting sentences.
	t = ExpandAbbreviations(t)

	segmenter := sentencizer.NewSegmenter("en")
	scanner := bufio.NewScanner(strings.NewReader(t))
	var (
		t_s string
		t_c []string
		s_a []rune
	)
	for scanner.Scan() {
		sentences := segmenter.Segment(scanner.Text())
	outer:
		for _, sentence := range sentences {
			t_c = []string{}
			sentence = strings.Trim(sentence, "\n\t\r ")
			if sentence == "" {
				continue
			}
			if len(sentence) < 3 {
				continue
			}
			// Remove lines which dont contain
			// any readable characters.
			var empty bool = true
			for _, c := range sentence {
				if isAlNum(c) {
					empty = false
					break
				}
			}
			if empty {
				continue
			}
			s_a = []rune(sentence)
			sentence = strings.ToTitle(string(s_a[0])) + string(s_a[1:])
			for _, c := range SENTENCE_WRAP_CHARS {
				if string(sentence[0]) == c && string(sentence[len(sentence)-1]) == c {
					t_s = sentence[1 : len(sentence)-1]
					t_c = segmenter.Segment(t_s)
					if len(t_c) > 1 {
						chunks = append(chunks, t_c...)
						continue outer
					}
				}
			}

			if len([]rune(sentence))-strings.Count(sentence, " ") > MAX_SENTENCE_LENGTH {
				chunks = append(chunks,
					SplitLongSentence(sentence,
						SENTENCE_DESIRED_L,
						SENTENCE_DESIRED_R,
						SENTENCE_SPLIT_DELIMITERS)...)
			} else {
				chunks = append(chunks, sentence)
			}
		}
	}
	return chunks
}

// Split long sentence, necessary in order not to exceed maximum token len allowable by
// the kitten tts model.
func SplitLongSentence(text string, desired_l int, range_l int, delimiters []rune) []string {
	var chunks []string
	current := new(bytes.Buffer)

	for _, char := range text {
		is_delimiter := slices.Contains(delimiters, char)

		if char == ' ' && current.Len() >= desired_l+range_l {
			chunks = append(chunks, current.String())
			current.Reset()
			continue
		}

		if is_delimiter && current.Len() >= desired_l-range_l {
			current.WriteRune(char)
			chunks = append(chunks, strings.Trim(current.String(), " "))
			current.Reset()
			continue
		}

		current.WriteRune(char)
	}
	rest := strings.Trim(current.String(), " ")
	if len(rest) > range_l {
		chunks = append(chunks, rest)
	} else {
		chunks[len(chunks)-1] += " " + rest
	}

	return chunks
}

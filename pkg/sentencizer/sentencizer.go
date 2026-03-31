package sentencizer

import (
	"bufio"
	"strings"

	"github.com/sentencizer/sentencizer"
)

var SENTENCE_WRAP_CHARS = []string{"\"", "'"}

func Sentencize(text []byte) []string {
	chunks := []string{}
	segmenter := sentencizer.NewSegmenter("en")
	scanner := bufio.NewScanner(strings.NewReader(string(text)))
	var t_s string
	var t_c []string
	for scanner.Scan() {
		sentences := segmenter.Segment(scanner.Text())
	outer:
		for _, sentence := range sentences {
			t_c = []string{}
			sentence = strings.Trim(sentence, "\n\t\r ")
			if sentence == "" {
				continue
			}
			if len(sentence) > 3 {
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
			}
			chunks = append(chunks, sentence)

		}
	}
	return chunks
}

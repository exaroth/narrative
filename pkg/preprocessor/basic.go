package preprocessor

import (
	"regexp"
	"strings"
)

var DOUBLE_QUOTES_REP_RE = regexp.MustCompile("[«»“”]")
var SINGLE_QUOTES_REP_RE = regexp.MustCompile("[’]")

// Replace custom quotes with normalized version
// to make text easier to process later on.
func normalizeQuotes(input string) (string, error) {
	input = DOUBLE_QUOTES_REP_RE.ReplaceAllString(input, "\"")
	input = SINGLE_QUOTES_REP_RE.ReplaceAllString(input, "'")
	return input, nil
}

// Split each hyphenized word into 2 separate words
// eg. half-life -> half life. This assumes that multiple
// instances of the hyphen has been removed.
func splitHyphenizedWords(input string) (string, error) {
	result := []string{}
	for word := range strings.SplitSeq(input, " ") {
		if len(word) == 1 || !strings.ContainsAny(word, "-") {
			result = append(result, word)
			continue
		}
		chunks := []string{}
		if word[0] == 45 {
			chunks = append(chunks, "-")
		}
		chunks = append(chunks, strings.Split(strings.Trim(word, "-"), "-")...)

		if word[len(word)-1] == 45 {
			chunks = append(chunks, "-")
		}
		result = append(result, chunks...)
	}
	return strings.Join(result, " "), nil
}

// Remove trailing apostrophes from words,
// eg peoples' -> peoples or disciple's -> disciples.
func removeTrailingApostrophes(input string) (string, error) {

	result := []string{}

	for word := range strings.SplitSeq(input, " ") {
		if len(word) < 3 || !strings.ContainsAny(word, "'") {
			result = append(result, word)
			continue
		}
		if word[len(word)-1] == 39 && word[0] != 39 {
			result = append(result, strings.TrimRight(word, "'"))
			continue
		}
		// trim hyphen from 's
		if word[len(word)-1] == 115 && word[len(word)-2] == 39 {
			result = append(result, string(word[0:len(word)-2])+"s")
		}
	}
	return strings.Join(result, " "), nil
}

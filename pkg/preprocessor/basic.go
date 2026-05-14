package preprocessor

import (
	"regexp"
	"slices"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	PUNCT     = []rune{'.', ',', '?', '!', ';', ':', '\'', '"'}
	SPACES_RE = regexp.MustCompile(`\s+`)
	// todo - recheck
	PUNCT_RE   = regexp.MustCompile(`[^\w\s.,?!;:'"-]`)
	URL_RE     = regexp.MustCompile(`https?://\S+|www\.\S+`)
	EMAIL_RE   = regexp.MustCompile(`\b[\w.+-]+@[\w-]+\.[a-z]{2,}\b`)
	HASHTAG_RE = regexp.MustCompile(`#\w+`)
	MENTION_RE = regexp.MustCompile(`@\w+`)
	HTML_RE    = regexp.MustCompile(`<[^>]+>`)
)

// Array containing words which should be excluded when
// removing apostrophes from the words.
var APOSTROPHE_REMOVAL_EXCLUSIONS = []string{
	"he's", "it's", "there's", "that's", "here's", "where's",
	"what's",
}

// Contains patterns and replacement rules to be
// used when normalizing characters in the input string.
var PUNCT_REPLACEMENT_MAP = []struct {
	re   *regexp.Regexp
	repl string
}{
	{
		re:   regexp.MustCompile("[«»“”]"),
		repl: "\"",
	},
	{
		re:   regexp.MustCompile("[’]"),
		repl: "'",
	},
	{
		re:   regexp.MustCompile("[—]"),
		repl: "-",
	},
	{
		re:   regexp.MustCompile("[_]"),
		repl: "",
	},
}

// Normalize unicode characters into their english diactric
// equivalents.
func normalizeUnicode(input string) (string, error) {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, err := transform.String(t, input)
	return result, err
}

// Normalize whitespace removing multiple occurences
// and trimming the sentence.
func normalizeWhitespace(input string) (string, error) {
	input = SPACES_RE.ReplaceAllString(input, " ")
	return strings.Trim(input, " "), nil
}

// Narmalize punctuation across the sentence
// for easier processing.
func normalizePunctuation(input string) (string, error) {
	for _, r := range PUNCT_REPLACEMENT_MAP {
		input = r.re.ReplaceAllString(input, r.repl)
	}
	return input, nil
}

// This function removes all parts of sentence that
// are unusable for phonetization, eg. email, urls etc.
func cleanupUnusableTextParts(input string) (string, error) {
	input = HTML_RE.ReplaceAllString(input, "")
	input = URL_RE.ReplaceAllString(input, "")
	input = EMAIL_RE.ReplaceAllString(input, "")
	input = HASHTAG_RE.ReplaceAllString(input, "")
	input = MENTION_RE.ReplaceAllString(input, "")
	return input, nil
}

// Split each hyphenized word into 2 separate words
// eg. half-life -> half life. This assumes that multiple
// instances of the hyphen has been removed.
func splitHyphenizedWords(input string) (string, error) {
	result := []string{}
	for word := range strings.SplitSeq(input, " ") {
		if len(word) < 2 || !strings.ContainsAny(word, "-") {
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
		if slices.Index(APOSTROPHE_REMOVAL_EXCLUSIONS, strings.ToLower(word)) > -1 {
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
			continue
		}
		result = append(result, word)

	}
	return strings.Join(result, " "), nil
}

// Remove non prosodic punctuation keeping only punctuation
// that affects intonation.
func removeNonProsodicPunctuation(input string) (string, error) {
	return PUNCT_RE.ReplaceAllString(input, " "), nil
}

// Trim wrapping quotes from the sentence,
// as kittenTTS doesn't like it.
func trimSentenceQuotes(input string) (string, error) {
	if input[0] == '"' && input[len(input)-1] == '"' {
		return trimSentenceQuotes(strings.Trim(input, "\""))
	}
	if input[0] == '\'' && input[len(input)-1] == '\'' {
		return trimSentenceQuotes(strings.Trim(input, "'"))
	}
	return input, nil
}

// Process trailing and leading dashed splitting them from
// the word and replace them with ; for better phonemization
// result.
func processDashes(input string) (string, error) {

	result := []string{}
	for idx, word := range strings.Split(input, " ") {
		if len(word) == 0 {
			continue
		}
		// dont process dashes for first word
		// as these might indicate dialogue
		if idx == 0 {
			result = append(result, word)
			continue
		}
		if word == "-" {
			result = append(result, ";")
			continue
		}
		if word[0] == '-' {
			word = "; " + word[1:]
		}
		if word[len(word)-1] == '-' {
			word = word[:len(word)-1] + " ;"
		}
		result = append(result, word)
	}
	return strings.Join(result, " "), nil
}

// Add trailing period if there's no valid sentence termination.
func appendPeriod(input string) (string, error) {
	if len(input) == 0 {
		return input, nil
	}
	s := strings.Split(input, " ")
	last := s[len(s)-1]
	llast := rune(last[len(last)-1])
	if !slices.Contains(PUNCT, llast) {
		input = input + "."
	}
	return input, nil
}

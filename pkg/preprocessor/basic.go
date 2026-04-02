package preprocessor

import (
	"regexp"
	"slices"
	"strings"
)

var (
	SPACES_RE = regexp.MustCompile(`\s+`)
	// todo - recheck
	PUNCT_RE   = regexp.MustCompile(`[^\w\s.,?!;:'"-]`)
	URL_RE     = regexp.MustCompile(`https?://\S+|www\.\S+`)
	EMAIL_RE   = regexp.MustCompile(`\b[\w.+-]+@[\w-]+\.[a-z]{2,}\b`)
	HASHTAG_RE = regexp.MustCompile(`#\w+`)
	MENTION_RE = regexp.MustCompile(`@\w+`)
	HTML_RE    = regexp.MustCompile(`<[^>]+>`)
)

var APOSTROPHE_REMOVAL_EXCLUSIONS = []string{"he's", "it's"}

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

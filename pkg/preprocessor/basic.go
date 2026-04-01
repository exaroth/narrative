package preprocessor

import "regexp"

var DOUBLE_QUOTES_REP_RE = regexp.MustCompile("[«»“”]")
var SINGLE_QUOTES_REP_RE = regexp.MustCompile("[’]")

func normalizeQuotes(input string) (string, error) {
	processed := DOUBLE_QUOTES_REP_RE.ReplaceAllString(input, "\"")
	processed = SINGLE_QUOTES_REP_RE.ReplaceAllString(processed, "'")
	return processed, nil
}

package reader

import (
	"regexp"
	"strings"
)

var EXCERPT_RE = regexp.MustCompile(`([\[]\d+[\]])`)

// Clean excerpts in a form [<num>].
func cleanExcerpts(input string) string {
	for _, g := range EXCERPT_RE.FindAllStringSubmatch(input, -1) {
		input = strings.ReplaceAll(input, g[0], "")
	}
	return input
}

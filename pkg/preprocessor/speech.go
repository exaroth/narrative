package preprocessor

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

var CONTRACTION_MAP = []struct {
	re         *regexp.Regexp
	repl       string
	exclusions []string
}{
	{
		re:         regexp.MustCompile(`\b(\w+)'m\b`),
		repl:       "am",
		exclusions: []string{"i'm"},
	},
	{
		re:         regexp.MustCompile(`\b(\w+)'d\b`),
		repl:       "would",
		exclusions: []string{},
	},
	{
		re:         regexp.MustCompile(`\b(\w+)'ll\b`),
		repl:       "will",
		exclusions: []string{},
	},
	{
		re:         regexp.MustCompile(`\b(\w+)'ve\b`),
		repl:       "have",
		exclusions: []string{},
	},
	{
		re:         regexp.MustCompile(`\b(\w+)'re\b`),
		repl:       "are",
		exclusions: []string{"you're", "we're", "You're"},
	},
	{
		re:         regexp.MustCompile(`\b(\w+)n't\b`),
		repl:       "not",
		exclusions: []string{"won't", "can't", "hasn't", "haven't", "don't", "ain't", "isn't"},
	},
}

func expandContractions(input string) (string, error) {
	for _, c := range CONTRACTION_MAP {
		if len(c.re.FindStringIndex(input)) == 0 {
			continue
		}
		groups := c.re.FindAllStringSubmatch(input, -1)
		for _, g := range groups {
			if slices.Index(c.exclusions, strings.ToLower(g[0])) > -1 {
				continue
			}
			input = strings.ReplaceAll(input, g[0], fmt.Sprintf("%s %s", g[1], c.repl))
		}
	}
	return input, nil
}

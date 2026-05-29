package reader

import (
	"regexp"
	"strings"
	"unicode"
)

type whitespaceFn func(string) string

var (
	reRemoveSurroundLF = regexp.MustCompile("(?m)[\t\r ]*\n[\t\r ]*")
	reTransformSpace   = regexp.MustCompile(" +")
)

var escapePattern = regexp.MustCompile(`(\[[a-zA-Z0-9_,;: \-\."#]+\[*)\]`)

func Escape(text string) string {
	text = strings.ReplaceAll(text, "\u00a0", "")
	text = strings.ReplaceAll(text, "\ufeff", "")
	// Kindle unpack sometimes inserts garbage into xhtml
	// so we ought to clean it
	text = strings.ReplaceAll(text, "­", "")
	return escapePattern.ReplaceAllString(text, "$1[]")
}

func processWhitespace(text string) string {
	for _, fn := range []whitespaceFn{
		wsRemoveSurroundLF,
		// wsTransformLF,
		wsTransformTab,
		wsTransformSpace,
	} {
		text = fn(text)
	}

	return text
}

func wsRemoveSurroundLF(text string) string {
	return reRemoveSurroundLF.ReplaceAllString(text, "\n")
}

func wsTransformLF(text string) string {
	return strings.ReplaceAll(text, "\n", " ")
}

func wsTransformTab(text string) string {
	return strings.ReplaceAll(text, "\t", " ")
}

func wsTransformSpace(text string) string {
	return reTransformSpace.ReplaceAllString(text, " ")
}

func hasText(text string) bool {
	if len(text) > 0 {
		for _, r := range text {
			if !unicode.IsSpace(r) {
				return true
			}
		}
	}

	return false
}

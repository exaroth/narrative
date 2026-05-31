package phonemizer

import (
	"encoding/json"
	"unicode"
)

// Based on the Goruut implementation by neurlang
// https://github.com/neurlang/goruut
// MIT Licence

// This struct contains all language settings for
// goruut dictionaries.
type Language struct {
	Mapping        map[string][]string `json:"Map"`
	SrcMulti       []string            `json:"SrcMulti"`
	DstMulti       []string            `json:"DstMulti"`
	SrcMultiSuffix []string            `json:"SrcMultiSuffix"`
	DstMultiSuffix []string            `json:"DstMultiSuffix"`
	DropLast       []string            `json:"DropLast"`
	SrcDuplicate   [][]string          `json:"SrcDuplicate"`
	//Histogram         []string            `json:"Histogram"`
	mapTokenizer      map[uint32]map[[2]uint32]string
	mapSrcMultiLen    int
	mapSrcMultiSufLen int
	mapSrcMulti       map[string]struct{}
	mapDstMulti       map[string]struct{}
	mapSrcMultiSuffix map[string]struct{}
	mapDstMultiSuffix map[string]struct{}
	mapLetters        map[string]struct{}
	mapDropLast       map[string]struct{}
}

func NewLanguage(raw []byte) (*Language, error) {

	var lang Language
	if err := json.Unmarshal(raw, &lang); err != nil {
		return nil, err
	}
	lang.mapize()
	lang.srcdst()
	lang.letters()
	return &lang, nil
}

// Convert slize []T into map[T]struct{}.
func mapize(arr []string) (out map[string]struct{}) {
	out = make(map[string]struct{})
	for _, v := range arr {
		out[v] = struct{}{}
	}
	return
}

// Convert language options into a map
func (l *Language) mapize() {
	l.mapSrcMulti = mapize(l.SrcMulti)
	l.mapDstMulti = mapize(l.DstMulti)
	l.mapSrcMultiSuffix = mapize(l.SrcMultiSuffix)
	l.mapDstMultiSuffix = mapize(l.DstMultiSuffix)
	l.mapDropLast = mapize(l.DropLast)
	l.SrcMulti = nil
	l.DstMulti = nil
	l.SrcMultiSuffix = nil
	l.DstMultiSuffix = nil
	l.DropLast = nil
}

func (l *Language) srcdst() {
	for k, v := range l.Mapping {
		if len(v) == 0 {
			continue
		}
		if len([]rune(k)) > 1 {
			l.mapSrcMulti[k] = struct{}{}
		}
		for _, w := range v {
			if len([]rune(w)) > 1 {
				l.mapDstMulti[w] = struct{}{}
			}
		}
	}
	for k := range l.mapSrcMulti {
		if len([]rune(k)) > l.mapSrcMultiLen {
			l.mapSrcMultiLen = len([]rune(k))
		}
	}
	for k := range l.mapSrcMultiSuffix {
		if len([]rune(k)) > l.mapSrcMultiSufLen {
			l.mapSrcMultiSufLen = len([]rune(k))
		}
	}
}

func (l *Language) letters() {
	l.mapLetters = make(map[string]struct{})
	for k := range l.Mapping {
		addLetters(k, l.mapLetters)
	}
	for _, rule := range l.SrcDuplicate {
		for _, v := range rule {
			addLetters(v, l.mapLetters)
		}
	}
}

func isCombining(r uint32) bool {
	return unicode.Is(unicode.Mn, rune(r)) || unicode.Is(unicode.Mc, rune(r))
}

func addLetters(word string, mapping map[string]struct{}) {
	if mapping == nil {
		return
	}

	runes := []rune(word)
	n := len(runes)
	var str string
	var baseLetterFound bool

	// Process characters in reverse order
	for i := n - 1; i >= 0; i-- {
		r := runes[i]

		if isCombining(uint32(r)) {
			// Accumulate combining characters
			str = string(r) + str
		} else {
			// Found a base letter, prepend accumulated combiners
			fullLetter := string(r) + str
			// log.Now().Debugf("Adding to letters: %s", fullLetter)
			mapping[fullLetter] = struct{}{}

			// Reset for next letter
			str = ""
			baseLetterFound = true
		}
	}

	// Edge case: if only combiners were present at the start
	if str != "" && !baseLetterFound {
		// log.Now().Debugf("Adding standalone combiners: %s", str)
		mapping[str] = struct{}{}
	}
}

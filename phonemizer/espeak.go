package phonemizer

import (
	"strings"
)

type flavorMap map[string]string

func (f flavorMap) maxLen() int {
	var result int = 1
	for k := range f {
		if len(k) > result {
			result = len(k)
		}
	}
	return result
}

var espeak_map flavorMap = map[string]string{
	"ː":  ":",
	"æ":  "{",
	"ɐ":  "6",
	"ɑ":  "a",
	"ɒ":  "0",
	"ʙ":  "B",
	"ɓ":  "b^",
	"ð":  "D",
	"ʤ":  "dZ",
	"ɖ":  "d^",
	"ə":  "@",
	"ɛ":  "E",
	"ɚ":  "3",
	"ɜ":  "3",
	"ɞ":  "3",
	"ɤ":  "7",
	"ɡ":  "g",
	"ɦ":  "h",
	"ɪ":  "I",
	"ɨ":  "1",
	"ɟ":  "J",
	"ɟʝ": "J",
	"ɫ":  "l",
	"ɮ":  "L",
	"ʎ":  "l^",
	"ɱ":  "m^",
	"ɲ":  "n^",
	"ɳ":  "n^",
	"ŋ":  "N",
	"ø":  "2",
	"œ":  "9",
	"ɶ":  "&",
	"ɔ":  "o",
	"ɸ":  "P",
	"r̝": "R^",
	"ɹ":  "r",
	"ɽ":  "r^",
	"ʃ":  "S",
	"ʧ":  "tS",
	"ʈ":  "t^",
	"ʉ":  "}",
	"ʊ":  "U",
	"ʋ":  "v^",
	"ʌ":  "V",
	"ʍ":  "W",
	"ʒ":  "Z",
	"ʔ":  "?",
	"θ":  "T",
	"χ":  "X",
	"ɣ":  "G",
	"ç":  "C",
	"ʁ":  "R",
	"ɾ":  "4",
	"ħ":  "H",
	"ʕ":  "H",
	"n̩": "n",
	"ã":  "a",
	"ɛ̃": "I",
	"œ̃": "U",
	"ɐ̯": "",
	"i̯": "I",
	"ˈ":  "'",
	"ˌ":  ",",
	"{":  "a",
}

func ApplyEspeak(word string) string {
	maxLen := espeak_map.maxLen()
	for i := maxLen; i > 0; i-- {
		for k, v := range espeak_map {
			if len(k) != i {
				continue
			}

			word = strings.ReplaceAll(word, k, v)
		}
	}
	return word
}

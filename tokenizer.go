package main

import "fmt"

const (
	pad = '$'
)

var chars_punctuation = []rune{';', ':', ',', '.', '!', '?', '¡', '¿', '—', '…', '"', '«', '»', '"', '"', ' '}
var chars_letter = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
var chars_ipa = []rune("ɑɐɒæɓʙβɔɕçɗɖðʤəɘɚɛɜɝɞɟʄɡɠɢʛɦɧħɥʜɨɪʝɭɬɫɮʟɱɯɰŋɳɲɴøɵɸθœɶʘɹɺɾɻʀʁɽʂʃʈʧʉʊʋⱱʌɣɤʍχʎʏʑʐʒʔʡʕʢǀǁǂǃˈˌːˑʼʴʰʱʲʷˠˤ˞↓↑→↗↘'̩'ᵻ")

type TokenMap map[rune]int64

func (t TokenMap) tokenize(char rune) int64 {
	if val, ok := t[char]; ok {
		return val
	}
	fmt.Printf("char %s not found in token map", string(char))
	return -1
}

func (t TokenMap) TokenizeWord(word string) []int64 {
	result := []int64{}
	for _, char := range word {
		result = append(result, t.tokenize(char))
	}
	return result
}

func buildTokenMap() TokenMap {
	token_arr := []rune{pad}
	token_arr = append(token_arr, chars_punctuation...)
	token_arr = append(token_arr, chars_letter...)
	token_arr = append(token_arr, chars_ipa...)

	var result = make(map[rune]int64)
	for idx, c := range token_arr {
		result[c] = int64(idx)
	}
	return result
}

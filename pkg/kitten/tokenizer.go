package kitten

const (
	PAD = '$'
)

var CHARS_PUNCTUATION = []rune{';', ':', ',', '.', '!', '?', '¡', '¿', '—', '…', '"', '«', '»', '"', '"', ' '}
var CHARS_LETTER = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
var CHARS_IPA = []rune("ɑɐɒæɓʙβɔɕçɗɖðʤəɘɚɛɜɝɞɟʄɡɠɢʛɦɧħɥʜɨɪʝɭɬɫɮʟɱɯɰŋɳɲɴøɵɸθœɶʘɹɺɾɻʀʁɽʂʃʈʧʉʊʋⱱʌɣɤʍχʎʏʑʐʒʔʡʕʢǀǁǂǃˈˌːˑʼʴʰʱʲʷˠˤ˞↓↑→↗↘'̩'ᵻ")

// Representation of mapping
// <phoneme char> -> <token int>
// Token value is arbitrary, based
// on KittenTTS model.
type TokenMap map[rune]int64

// Tokenize single rune into a token,
// return -1 if not found.
func (t TokenMap) tokenize(char rune) int64 {
	if val, ok := t[char]; ok {
		return val
	}
	return -1
}

// Tokenize single word.returning slice of tokens
// ready to be passed into the model.
func (t TokenMap) TokenizeWord(word string) []int64 {
	result := []int64{}
	var token int64
	for _, char := range word {
		token = t.tokenize(char)
		if token == -1 {
			continue
		}
		result = append(result, token)
	}
	return result
}

// Initialize new token map
func BuildTokenMap() TokenMap {
	token_arr := []rune{PAD}
	token_arr = append(token_arr, CHARS_PUNCTUATION...)
	token_arr = append(token_arr, CHARS_LETTER...)
	token_arr = append(token_arr, CHARS_IPA...)

	var result = make(map[rune]int64)
	for idx, c := range token_arr {
		result[c] = int64(idx)
	}
	return result
}

package phonemizer

import (
	"strings"
)

var DICT_FALLBACK_PREFIXES = map[string]string{
	"de":   "dɪ",
	"un":   "ˌʌn",
	"dis":  "dɪsˈ",
	"mis":  "mˈɪs",
	"be":   "bə",
	"pre":  "pɹiː",
	"im":   "ɪm",
	"in":   "ɪnˈ",
	"en":   "ənˈ",
	"anti": "æntaɪ",
	"a":    "ə",
}

var DICT_FALLBACK_SUFFIXES = map[string]string{
	"ed":     "ɪd",
	"d":      "t",
	"ly":     "li",
	"y":      "i",
	"s":      "z",
	"es":     "ɪz",
	"ing":    "ɪŋ",
	"ish":    "ɪʃ",
	"le":     "əl",
	"ness":   "nəs",
	"al":     "ɔl",
	"ation":  "eɪʃən",
	"ations": "eɪʃənz",
	"less":   "lˈɛss",
	"able":   "əbəl",
	"ity":    "ˈɪti",
}

// Compact ps fallback result into repository result.
// This will also update tags for given word.
func compactRepositoryFallbackResult(
	repo *PhonemizerRepository,
	d_result map[string]uint32,
	prefix, suffix, base string,
) map[string]uint32 {

	if d_result == nil {
		return nil
	}

	result := map[string]uint32{}

	orig := prefix + base + suffix
	result[orig+" "] = 0

	(*repo.langWords)[orig] = make(map[string]uint32)

	var tags []string
	for p, tk := range d_result {
		if tk == 0 {
			continue
		}
		tags = repo.LookupTags(base, p)
		tags = append(tags, "ps-fallback")
		tagK, tagJ, _ := serializeTags(addTags(make(map[uint32]string), tags...))

		if len(prefix) > 0 {
			p = DICT_FALLBACK_PREFIXES[prefix] + p
		}
		if len(suffix) > 0 {
			p = p + DICT_FALLBACK_SUFFIXES[suffix]
		}
		(*repo.langWords)[orig][p] = tagK
		(*repo.langTags)[tagK] = tagJ
		(*repo.wordTags)[[2]string{orig, p}] = tagK
		result[p] = tagK
	}

	return result
}

// Attempt to try fallback repository result by splitting common prefixes and suffixes
// from the input word.
func PrefixSuffixFallbackCheck(repo *PhonemizerRepository, word string) map[string]uint32 {

	var prefix_a = [][2]string{}
	var d_result_raw []map[string]uint32
	var d_result map[string]uint32
	var prefix, suffix, base_string string

	for p := range DICT_FALLBACK_PREFIXES {
		if len(p) >= len(word)+1 {
			continue
		}
		if base, ok := strings.CutPrefix(word, p); ok {
			d_result_raw = repo.LookupWords(base)
			if len(d_result_raw) > 0 {
				d_result = d_result_raw[0]
				prefix = p
				base_string = base
				break
			}
			prefix_a = append(prefix_a, [2]string{p, base})
		}
	}

	if d_result != nil {
		return compactRepositoryFallbackResult(repo, d_result, prefix, suffix, base_string)
	}

	for s := range DICT_FALLBACK_SUFFIXES {
		if len(s) >= len(word)+1 {
			continue
		}
		if base, ok := strings.CutSuffix(word, s); ok {
			d_result_raw = repo.LookupWords(base)
			if len(d_result_raw) > 0 {
				d_result = d_result_raw[0]
				suffix = s
				base_string = base
				break
			}
		}
		for _, p_e := range prefix_a {
			if len(s) >= len(p_e[1])+1 {
				continue
			}
			if base, ok := strings.CutSuffix(p_e[1], s); ok {
				d_result_raw = repo.LookupWords(base)
				if len(d_result_raw) > 0 {
					d_result = d_result_raw[0]
					prefix = p_e[0]
					suffix = s
					base_string = base
					break
				}
			}
		}
	}
	return compactRepositoryFallbackResult(repo, d_result, prefix, suffix, base_string)
}

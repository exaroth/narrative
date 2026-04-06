package phonemizer

import (
	"fmt"
	"slices"
	"strings"

	"github.com/neurlang/classifier/hash"
)

type PhonemeOptions struct {
	Tags         *[]map[string][]string
	WordOrigins  *[]string
	DictOpts     *[]*[2]string
	PrefOpts     *[]*[2]string
	OverrideOpts *[]*[2]string
}

func NewPhonemeOpts(l int) *PhonemeOptions {
	tags := make([]map[string][]string, l)
	word_orig := make([]string, l)
	dict_m := make([]*[2]string, l)
	pref_m := make([]*[2]string, l)
	override_m := make([]*[2]string, l)
	return &PhonemeOptions{
		Tags:         &tags,
		WordOrigins:  &word_orig,
		DictOpts:     &dict_m,
		PrefOpts:     &pref_m,
		OverrideOpts: &override_m,
	}
}

type Phonemizer struct {
	repository *PhonemizerRepository
	hashtron   *HashtronPhonemizer
	selector   *PhonemeSelector
	cache      *WordCache
}

func (p *Phonemizer) Phonemize(sentence string) (string, error) {

	words, punct := SplitPunctuation(sentence)
	phonemes := make([]map[string]uint32, len(words))
	for i, w := range words {
		phonemized_w, err := p.PhonemizeWord(w)
		if err != nil {
			return "", err
		}
		if phonemized_w == nil {
			phonemes[i] = map[string]uint32{w: 0}
		} else {
			phonemes[i] = phonemized_w
		}
	}
	phoneme_opts := p.GetPhonemeOptions(phonemes)

	selected := p.SelectPhonemes(phonemes, phoneme_opts)

	phoneme_a := []string{}
	for _, p := range selected {
		phoneme_a = append(phoneme_a, p[1])
	}

	return CompactPunctuation(phoneme_a, punct), nil

}

func (p *Phonemizer) GetPhonemeOptions(sentence []map[string]uint32) *PhonemeOptions {

	opts := NewPhonemeOpts(len(sentence))

	var input []map[string][2]uint32

	for i, phoneme_map := range sentence {
		var orig string
		for word, k := range phoneme_map {
			if k == 0 {
				orig = strings.TrimRight(word, " ")
				(*opts.WordOrigins)[i] = orig
				break
			}
		}
		var inputmap = make(map[string][2]uint32)
		inputmap[orig+" "] = [2]uint32{0, 0}

		var tags []string
		var is_dict, is_override bool
		for phoneme, k := range phoneme_map {
			if k == 0 {
				continue
			}
			tags = p.repository.LookupTags(orig, phoneme)
			if (*opts.Tags)[i] == nil {
				(*opts.Tags)[i] = make(map[string][]string)
			}
			(*opts.Tags)[i][phoneme] = tags

			is_dict = slices.Contains(tags, "dict")
			is_override = slices.Contains(tags, "override") || slices.Contains(tags, "override-ext")
			if is_dict || is_override {
				inputmap[phoneme] = [2]uint32{k, 0}
				if is_override {
					prev := (*opts.Tags)[i][phoneme]
					// dont set override phoneme if one was already set
					// from external dict.
					if !slices.Contains(prev, "override-ext") {
						(*opts.OverrideOpts)[i] = &[2]string{orig, phoneme}
					}
				} else {
					if (*opts.DictOpts)[i] == nil {
						(*opts.DictOpts)[i] = &[2]string{orig, phoneme}
						continue
					}
					prev_l := len((*opts.Tags)[i][phoneme])
					if len(tags) > prev_l {
						(*opts.DictOpts)[i] = &[2]string{orig, phoneme}
					}
				}
			}
		}
		input = append(input, inputmap)
	}

	var preferred = p.selector.Select(input)

	for i, words := range sentence {
		var last_preferred, hash_preferred uint32
		for _, row := range preferred {
			if row[0] != uint32(i) {
				continue
			}
			last_preferred = row[2]
			hash_preferred = row[1]
			break
		}

		for word, k := range words {
			if k == 0 {
				continue
			}
			if last_preferred != 0 && last_preferred == k || hash_preferred == hash.StringHash(0, word) {
				(*opts.PrefOpts)[i] = &[2]string{(*opts.WordOrigins)[i], word}
				break
			}
		}
	}
	return opts
}

func (p *Phonemizer) SelectPhonemes(
	sentence []map[string]uint32,
	opts *PhonemeOptions,
) [][2]string {

	result := [][2]string{}

	for idx, words := range sentence {
		if (*opts.OverrideOpts)[idx] != nil {
			result = append(result, *(*opts.OverrideOpts)[idx])
			continue
		}
		if (*opts.PrefOpts)[idx] != nil {
			result = append(result, *(*opts.PrefOpts)[idx])
			continue
		}
		if (*opts.DictOpts)[idx] != nil {
			result = append(result, *(*opts.DictOpts)[idx])
			continue
		}
		for word, k := range words {
			if k == 0 {
				if word == "" {
					result = append(result, [2]string{"", ""})
					break
				}
				continue
			}
			result = append(result, [2]string{(*opts.WordOrigins)[idx], word})
			break
		}
	}

	// TODO
	// trim left hyphen as it doesnt play well with kitten tts
	// for i, p := range result {
	// 	result[i][1] = strings.TrimLeft(p[1], "'ˈ")
	// }

	return result
}

func (p *Phonemizer) PhonemizeWord(word string) (map[string]uint32, error) {

	if len(word) == 0 {
		return nil, nil
	}

	hash := p.cache.HashWord(word)
	cached := p.cache.LoadWord(word, hash)

	if cached != nil {
		return cached, nil
	}

	repo_result := p.repository.LookupWords(word)
	if len(repo_result) > 0 {
		// TODO
		if len(repo_result) > 1 {
			fmt.Printf("Repository returned more that one result for word %s: %v", word, repo_result)
		}
		r := repo_result[0]
		p.cache.StoreWord(r, hash)
		return r, nil
	}

	p_result, err := p.hashtron.PhonemizeWord(word)
	if err != nil {
		return nil, err
	}
	if len(p_result) == 0 {
		// todo
		fmt.Printf("No results returned from phonemizer for word %s", word)
		return nil, nil

	}
	if len(p_result) > 1 {
		// todo
		if len(repo_result) > 1 {
			fmt.Printf("Phonemizer returned more that one result for word %s: %v", word, p_result)
		}

	}
	r := p_result[0]
	p.cache.StoreWord(r, hash)
	return r, nil

}

func NewPhonemizer() (*Phonemizer, error) {
	repo := NewPhonemizerRepository()
	pho := NewHashtronPhonemizer(nil, false)
	selector := NewPhonemeSelector(nil)

	if err := repo.LoadLanguage(); err != nil {
		return nil, err
	}
	if err := pho.LoadLanguage(); err != nil {
		return nil, err
	}
	if err := selector.LoadLanguage(); err != nil {
		return nil, err
	}

	cache, err := NewWordCache()
	if err != nil {
		return nil, err
	}
	return &Phonemizer{
		hashtron:   pho,
		repository: repo,
		selector:   selector,
		cache:      cache,
	}, nil
}

package phonemizer

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/k0kubun/pp"
	"github.com/neurlang/classifier/hash"
)

type Phonemizer struct {
	repository *PhonemizerRepository
	hashtron   *HashtronPhonemizer
	selector   *PhonemeSelector
	cache      *WordCache
}

func (p *Phonemizer) Phonemize(sentence string) (string, error) {
	// TODO lowercase
	words, punct := SplitPunctuation(sentence)
	phonemes := make([]map[string]uint32, len(words))
	for i, w := range words {
		phonemized_w, err := p.phonemizeWord(w)
		if err != nil {
			return "", err
		}
		if phonemized_w == nil {
			phonemes[i] = map[string]uint32{w: 0}
		} else {
			phonemes[i] = phonemized_w
		}
	}

	selected := p.selectPhonemes(phonemes)
	phoneme_a := []string{}
	for _, p := range selected {
		phoneme_a = append(phoneme_a, p[1])
	}

	return CompactPunctuation(phoneme_a, punct), nil

}

func (p *Phonemizer) selectPhonemes(sentence []map[string]uint32) [][2]string {

	fmt.Println(">>>>>>>>>>>> input")
	fmt.Println(sentence)
	fmt.Println("<<<<<<<<<<<<")
	result := [][2]string{}
	dict_m := make([]*[2]string, len(sentence))
	pref_m := make([]*[2]string, len(sentence))
	word_orig := make([]string, len(sentence))

	var input []map[string][2]uint32

	for i, words := range sentence {
		var orig string
		for word, k := range words {
			if k == 0 {
				orig = strings.TrimRight(word, " ")
				word_orig[i] = orig
				break
			}
		}
		var inputmap = make(map[string][2]uint32)
		inputmap[orig+" "] = [2]uint32{0, 0}
		for word, k := range words {
			if k == 0 {
				continue
			}
			var tags = p.repository.LookupTags(orig, word)
			json_tags := []string{}
			err := json.Unmarshal([]byte(tags), &json_tags)
			if err != nil {
				// todo
				panic(err)
			}
			if slices.Contains(json_tags, "dict") {
				inputmap[word] = [2]uint32{k, 0}
				dict_m[i] = &[2]string{orig, word}

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
				pref_m[i] = &[2]string{word_orig[i], word}
				break
			}
		}
	}
	fmt.Println(">>>>>>>>>>>> selection")
	fmt.Println("Dict")
	pp.Println(dict_m)
	fmt.Println("Pref")
	pp.Println(pref_m)
	fmt.Println("<<<<<<<<<<<< <`1`>")

	for idx, words := range sentence {
		if pref_m[idx] != nil {
			result = append(result, *pref_m[idx])
			continue
		}
		if dict_m[idx] != nil {
			result = append(result, *dict_m[idx])
			continue
		}
		for word, k := range words {
			if k == 0 {
				continue
			}
			result = append(result, [2]string{word_orig[idx], word})
			break
		}

	}

	return result
}

func (p *Phonemizer) phonemizeWord(word string) (map[string]uint32, error) {

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
	repo := NewPhonemizerRepository(nil, false)
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

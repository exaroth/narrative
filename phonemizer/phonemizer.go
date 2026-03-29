package phonemizer

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/neurlang/classifier/hash"
)

type Phonemizer struct {
	repository *PhonemizerRepository
	hashtron   *HashtronPhonemizer
	wselector  *WordSelector
	cache      *WordCache
}

func (p *Phonemizer) Phonemize(sentence string) (string, error) {
	words, punct := SplitPunctuation(sentence)
	phonemes := make([]string, len(words))
	for i, w := range words {
		if len(w) == 0 {
			phonemes[i] = w
			continue
		}
		p, err := p.phonemizeWord(w)
		if err != nil {
			return "", err
		}
		if len(p) == 0 {
			// TODO
			panic(fmt.Sprintf("No result returned for word %s", w))
			// continue
		}
		pthis := p[0]
		for k := range pthis {
			if k == w {
				continue
			}
			phonemes[i] = k
			break
		}
	}
	return CompactPunctuation(phonemes, punct), nil

}

func (p *Phonemizer) SelectWords(sentence []map[string]uint32) [][2]string {

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
			var json_tags []string
			err := json.Unmarshal([]byte(tags), json_tags)
			if err != nil {
				// todo
				panic(err)
			}

			for _, tag := range json_tags {
				if tag == "dict" {
					inputmap[word] = [2]uint32{k, 0}
					dict_m[i] = &[2]string{orig, word}
				}
			}
		}
		input = append(input, inputmap)
	}

	var preferred = p.wselector.Select(input)

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

func (p *Phonemizer) phonemizeWord(word string) ([]map[string]uint32, error) {
	var result []map[string]uint32

	result = p.repository.LookupWords(word)
	if result != nil {
		fmt.Printf("repo for word %s found::: %s\n", word, result)
		return result, nil
	}
	fmt.Println("NOT FOUND::: ", word)

	cached := p.cache.LoadWord(word)
	if cached != nil {
		result = []map[string]uint32{}
		result = append(result, cached)
		return result, nil
	}

	result, err := p.hashtron.PhonemizeWord(word)
	if err != nil {
		return nil, err
	}
	// TODO handle multiple results
	for _, w := range result {
		p.cache.StoreWord(w)
	}
	return result, nil

}

func NewPhonemizer() (*Phonemizer, error) {
	repo := NewPhonemizerRepository(nil, false)
	pho := NewHashtronPhonemizer(nil, false)
	hselector := NewHomonymSelector(nil)

	if err := repo.LoadLanguage(); err != nil {
		return nil, err
	}
	if err := pho.LoadLanguage(); err != nil {
		return nil, err
	}
	if err := hselector.LoadLanguage(); err != nil {
		return nil, err
	}

	cache, err := NewWordCache()
	if err != nil {
		return nil, err
	}
	return &Phonemizer{
		hashtron:   pho,
		repository: repo,
		hselector:  hselector,
		cache:      cache,
	}, nil
}

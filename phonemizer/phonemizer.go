package phonemizer

type Phonemizer struct {
	repository *PhonemizerRepository
	hashtron   *HashtronPhonemizer
	cache      *WordCache
}

func (p *Phonemizer) Phonemize(word string) ([]map[string]uint32, error) {
	var result []map[string]uint32

	result = p.repository.LookupWords(word)
	if result != nil {
		return result, nil
	}

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
	// TODO
	for _, w := range result {
		p.cache.StoreWord(w)
	}
	return result, nil

}

func NewPhonemizer() (*Phonemizer, error) {
	repo := NewPhonemizerRepository(nil, false)
	pho := NewHashtronPhonemizer(nil, false)

	if err := repo.LoadLanguage(); err != nil {
		return nil, err
	}
	if err := pho.LoadLanguage(); err != nil {
		return nil, err
	}

	cache, err := NewWordCache()
	if err != nil {
		return nil, err
	}
	return &Phonemizer{
		hashtron:   pho,
		repository: repo,
		cache:      cache,
	}, nil
}

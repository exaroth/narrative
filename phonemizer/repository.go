package phonemizer

import (
	"bytes"
	"compress/zlib"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	dict "github.com/exaroth/narrative/dictionary"
	"github.com/neurlang/classifier/hash"
)

const MISSING_F_NAME = "missing.all.zlib"

type PhonemizerRepository struct {
	reverse    bool
	dict_path  *string
	lang_words *map[string]map[string]uint32
	lang_tags  *map[uint32]string
	words_tags *map[[2]string]uint32
	mut        *sync.RWMutex
}

func (r *PhonemizerRepository) LoadLanguage() error {

	// TODO. if path set load from custom

	r.mut.Lock()
	defer r.mut.Unlock()

	f_contents, err := dict.Language.ReadFile(MISSING_F_NAME)
	if err != nil {
		return err
	}
	zlib_reader, err := zlib.NewReader(bytes.NewReader(f_contents))

	if err != nil {
		return err
	}
	var reader = csv.NewReader(zlib_reader)

	reader.Comma = '\t'

	recs, err := reader.ReadAll()
	if err != nil {
		return err
	}

	for _, v := range recs {

		for i := range v {
			v[i] = strings.ReplaceAll(v[i], " ", "")
		}
		var src, dst, tagstr string
		if len(v) == 2 {
			if r.reverse {
				src = v[1]
				dst = v[0]
			} else {
				src = v[0]
				dst = v[1]
			}
			tagstr = "[]"
		} else if len(v) == 3 {
			if r.reverse {
				src = v[1]
				dst = v[0]
			} else {
				src = v[0]
				dst = v[1]
			}
			tagstr = v[2]
		} else {
			return fmt.Errorf("Language %s has wrong number of columns: %d", src, len(v))
		}
		var tagkey, tagjson, err = serializeTags(addTags(parseTags(tagstr), "dict"))

		if err != nil {
			return err
		}

		if (*r.lang_words)[src] == nil {

			(*r.lang_words)[src] = make(map[string]uint32)

		} else if m, ok := (*r.lang_words)[src][dst]; ok {

			existingTags := parseTags((*r.lang_tags)[m])
			var existing []string
			for _, tag := range existingTags {
				existing = append(existing, tag)
			}
			tagkey, tagjson, err = serializeTags(addTags(parseTags(tagstr), existing...))
			if err != nil {
				return err
			}
		}

		(*r.lang_words)[src][dst] = tagkey
		(*r.lang_tags)[tagkey] = tagjson
		(*r.words_tags)[[2]string{src, dst}] = tagkey
	}
	// fmt.Println(r.lang_tags)
	// pp.Print(r.words_tags)
	return nil
}

func (r *PhonemizerRepository) LookupWords(word string) (ret []map[string]uint32) {

	r.mut.RLock()
	found := (*r.lang_words)[word]
	var foundCopy map[string]uint32
	if len(found) > 0 {
		foundCopy = make(map[string]uint32)
		for k, v := range found {
			foundCopy[k] = v
		}
	}
	r.mut.RUnlock()

	if len(foundCopy) == 0 {
		return nil
	}
	var m = make(map[string]uint32)
	for k, v := range foundCopy {
		m[k] = v
	}
	m[word+" "] = 0
	ret = append(ret, m)
	return
}

func (r *PhonemizerRepository) LookupTags(word1, word2 string) string {
	r.mut.RLock()
	// Copy the result while holding the mutex
	tagKey := (*r.words_tags)[[2]string{word1, word2}]
	found := (*r.lang_tags)[tagKey]
	r.mut.RUnlock()
	if found != "" {
		return found
	}
	return "[]"
}

func NewPhonemizerRepository(dict_path *string, reverse bool) *PhonemizerRepository {
	mapping := make(map[string]map[string]uint32)
	mapping2 := make(map[uint32]string)
	mapping3 := make(map[[2]string]uint32)

	return &PhonemizerRepository{
		dict_path:  dict_path,
		reverse:    reverse,
		lang_words: &mapping,
		lang_tags:  &mapping2,
		words_tags: &mapping3,
		mut:        &sync.RWMutex{},
	}
}

func addTags(bag map[uint32]string, tags ...string) map[uint32]string {
	for _, v := range tags {
		bag[hash.StringHash(0, v)] = v
	}
	return bag
}

func parseTags(cell string) (ret map[uint32]string) {
	ret = make(map[uint32]string)
	if cell == "" {
		return
	}
	var tags []string
	err := json.Unmarshal([]byte(cell), &tags)
	if err != nil {

		// todo
		fmt.Errorf("Cell tag: %s, Error: %v", cell, err)
	}
	for _, v := range tags {
		ret[hash.StringHash(0, v)] = v
	}
	return
}

func serializeTags(tags map[uint32]string) (key uint32, ret string, err error) {
	var tagstrings = []string{}
	for k, v := range tags {
		key ^= k
		tagstrings = append(tagstrings, v)
	}
	sort.Strings(tagstrings)
	data, err := json.Marshal(tagstrings)
	if err != nil {
		return 0, "", err
	}
	if len(data) > 0 {
		ret = string(data)
	} else {
		ret = "[]"
	}
	return
}

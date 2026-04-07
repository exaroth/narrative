package phonemizer

import (
	"bytes"
	"compress/zlib"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"
	"sync"

	dict "github.com/exaroth/narrative/dictionary"
	"github.com/neurlang/classifier/hash"
)

const DICT_F_NAME = "missing.all.zlib"
const AUX_DICT_F_NAME = "aux_dict.csv"

type PhonemizerRepository struct {
	langWords        *map[string]map[string]uint32
	langTags         *map[uint32]string
	wordTags         *map[[2]string]uint32
	externalDictPath string
	mut              *sync.RWMutex
}

func (r *PhonemizerRepository) LoadLanguage() error {
	err := r.loadMainDict()
	err = r.loadAuxDict("")
	if r.externalDictPath != "" {
		err = r.loadAuxDict(r.externalDictPath)
	}
	return err
}

func (r *PhonemizerRepository) loadAuxDict(fpath string) error {
	var f_reader fs.File
	var err error
	var tagkey uint32
	var tagjson string
	if len(fpath) > 0 {
		tagkey, tagjson, err = serializeTags(parseTags("[\"override-ext\"]"))
		f_reader, err = os.Open(fpath)
	} else {
		tagkey, tagjson, err = serializeTags(parseTags("[\"override\"]"))
		f_reader, err = dict.Language.Open(AUX_DICT_F_NAME)
	}
	if err != nil {
		return err
	}
	var reader = csv.NewReader(f_reader)
	reader.Comma = ' '

	recs, err := reader.ReadAll()
	if err != nil {
		return err
	}

	(*r.langTags)[tagkey] = tagjson

	var src, dst string
	for _, rec := range recs {
		if len(rec) != 2 {
			return fmt.Errorf("Invalid number of columns returned from aux dict, %v", rec)
		}
		src = rec[0]
		dst = rec[1]
		if (*r.langWords)[src] == nil {
			(*r.langWords)[src] = make(map[string]uint32)
		}
		(*r.langWords)[src][dst] = tagkey
		(*r.wordTags)[[2]string{src, dst}] = tagkey
	}
	return nil
}

func (r *PhonemizerRepository) loadMainDict() error {
	r.mut.Lock()
	defer r.mut.Unlock()

	f_contents, err := dict.Language.ReadFile(DICT_F_NAME)
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
			src = v[1]
			dst = v[0]
			tagstr = "[]"
		} else if len(v) == 3 {
			src = v[0]
			dst = v[1]
			tagstr = v[2]
		} else {
			return fmt.Errorf("Language %s has wrong number of columns: %d", src, len(v))
		}

		var tagkey, tagjson, err = serializeTags(addTags(parseTags(tagstr), "dict"))

		if err != nil {
			return err
		}

		if (*r.langWords)[src] == nil {

			(*r.langWords)[src] = make(map[string]uint32)

		} else if m, ok := (*r.langWords)[src][dst]; ok {

			existingTags := parseTags((*r.langTags)[m])
			var existing []string
			for _, tag := range existingTags {
				existing = append(existing, tag)
			}
			tagkey, tagjson, err = serializeTags(addTags(parseTags(tagstr), existing...))
			if err != nil {
				return err
			}
		}

		(*r.langWords)[src][dst] = tagkey
		(*r.langTags)[tagkey] = tagjson
		(*r.wordTags)[[2]string{src, dst}] = tagkey
	}
	return nil
}

func (r *PhonemizerRepository) LookupWords(word string) (ret []map[string]uint32) {

	r.mut.RLock()
	found := (*r.langWords)[word]
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

func (r *PhonemizerRepository) LookupTags(word1, word2 string) []string {
	r.mut.RLock()
	// Copy the result while holding the mutex
	tagKey := (*r.wordTags)[[2]string{word1, word2}]
	found := (*r.langTags)[tagKey]
	r.mut.RUnlock()

	if found == "" {
		return []string{}
	}

	var json_tags []string
	err := json.Unmarshal([]byte(found), &json_tags)
	if err != nil {
		// todo
		panic(err)
	}
	return json_tags
}

func NewPhonemizerRepository(external_dict_path string) *PhonemizerRepository {
	langWords := make(map[string]map[string]uint32)
	langTags := make(map[uint32]string)
	word_tags := make(map[[2]string]uint32)

	return &PhonemizerRepository{
		langWords:        &langWords,
		langTags:         &langTags,
		wordTags:         &word_tags,
		externalDictPath: external_dict_path,
		mut:              &sync.RWMutex{},
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

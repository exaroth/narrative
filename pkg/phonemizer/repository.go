package phonemizer

import (
	"bytes"
	"compress/zlib"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

type DictionaryReloadRequest int

const (
	ReloadRequestDict = iota
	ReloadRequestAux
	ReloadRequestExt
)

// Main dictionary repository.
type PhonemizerRepository struct {
	langWords                     *map[string]map[string]uint32
	langTags                      *map[uint32]string
	wordTags                      *map[[2]string]uint32
	externalDictPath              string
	mainDictR, auxDictR, extDictR *bytes.Reader
	mut                           *sync.RWMutex
}

// Initialize new phonemizer repository
func NewPhonemizerRepository(
	mainDict, auxDict *bytes.Reader,
	external_dict_path string,
) *PhonemizerRepository {
	lang_words := make(map[string]map[string]uint32)
	lang_tags := make(map[uint32]string)
	word_tags := make(map[[2]string]uint32)

	return &PhonemizerRepository{
		langWords:        &lang_words,
		langTags:         &lang_tags,
		wordTags:         &word_tags,
		externalDictPath: external_dict_path,
		mainDictR:        mainDict,
		auxDictR:         auxDict,
		mut:              &sync.RWMutex{},
	}
}

func (r *PhonemizerRepository) LoadLanguage() error {
	err := r.loadMainDict()
	err = r.loadAuxDict(false)
	if r.externalDictPath != "" {
		f_reader, err := os.Open(r.externalDictPath)
		if err != nil {
			return err
		}
		contents, err := io.ReadAll(f_reader)
		if err != nil {
			return err
		}
		r.extDictR = bytes.NewReader(contents)
		err = r.loadAuxDict(true)
	}
	return err
}

// Initialize auxiliary dictionary, be it built in or
// external.
func (r *PhonemizerRepository) loadAuxDict(ext bool) error {
	if r.extDictR == nil && ext {
		logrus.Warning("Skip loading ext dict.")
		return nil
	}
	if r.auxDictR == nil && !ext {
		logrus.Warning("Skip loading aux dict.")
		return nil
	}
	var f_reader *bytes.Reader
	var err error
	var tagkey uint32
	var tagjson string
	if ext {
		tagkey, tagjson, err = serializeTags(parseTags("[\"override-ext\"]"))
		f_reader = r.extDictR
	} else {
		tagkey, tagjson, err = serializeTags(parseTags("[\"override\"]"))
		f_reader = r.auxDictR
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

// Initialize main dictionary for the repository.
func (r *PhonemizerRepository) loadMainDict() error {
	if r.mainDictR == nil {
		panic("No dictionary loaded")
	}
	r.mut.Lock()
	defer r.mut.Unlock()

	zlib_reader, err := zlib.NewReader(r.mainDictR)

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

// Retrieve phoneme options for given word.
func (r *PhonemizerRepository) LookupWords(word string) (ret []map[string]uint32) {

	r.mut.RLock()
	found := (*r.langWords)[word]
	var foundCopy map[string]uint32
	if len(found) > 0 {
		foundCopy = make(map[string]uint32)
		maps.Copy(foundCopy, found)
	}
	r.mut.RUnlock()

	if len(foundCopy) == 0 {
		return nil
	}

	foundCopy[word+" "] = 0
	ret = append(ret, foundCopy)
	return
}

// Retrieve tags for given word-phoneme pair.
func (r *PhonemizerRepository) LookupTags(orig, phoneme string) []string {
	r.mut.RLock()
	tagKey := (*r.wordTags)[[2]string{orig, phoneme}]
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

// TODO: Fixme
// DO not use: bugged.
func (r *PhonemizerRepository) Reload(request DictionaryReloadRequest) error {

	lang_words := make(map[string]map[string]uint32)
	lang_tags := make(map[uint32]string)
	word_tags := make(map[[2]string]uint32)
	r.langWords = &lang_words
	r.langTags = &lang_tags
	r.wordTags = &word_tags

	var dict_r, aux_r, ext_r bool
	var err error
	switch request {
	case ReloadRequestDict:
		dict_r = true
	case ReloadRequestAux:
		dict_r = true
		aux_r = true
	case ReloadRequestExt:
		dict_r = true
		aux_r = true
		ext_r = true
	}

	if dict_r {
		if err = r.loadMainDict(); err != nil {
			return err
		}
	}
	if aux_r {
		if err = r.loadAuxDict(false); err != nil {
			return err
		}
	}
	if ext_r {
		if err = r.loadAuxDict(true); err != nil {
			return err
		}
	}
	return nil
}

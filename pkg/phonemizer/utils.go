package phonemizer

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/neurlang/classifier/hash"
)

// Add tags to existing map of tags.returning updated version.
func addTags(bag map[uint32]string, tags ...string) map[uint32]string {
	for _, v := range tags {
		bag[hash.StringHash(0, v)] = v
	}
	return bag
}

// Parse tags from tag string (json arr.).
func parseTags(cell string) (ret map[uint32]string) {
	ret = make(map[uint32]string)
	if cell == "" {
		return
	}
	var tags []string
	err := json.Unmarshal([]byte(cell), &tags)
	if err != nil {

		// todo
		// fmt.Errorf("Cell tag: %s, Error: %v", cell, err)
		fmt.Println(err)
	}
	for _, v := range tags {
		ret[hash.StringHash(0, v)] = v
	}
	return
}

// Serialize tags into json array returning both resulting string
// and tag key.
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

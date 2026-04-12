package phonemizer

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/gob"
	"fmt"

	"github.com/maypok86/otter"
	"github.com/neurlang/classifier/hash"
)

// Cache controller for handling cache operations.
// Cache uses gob format for storing data.
type WordCache struct {
	seed  uint32
	cache otter.Cache[uint32, string]
}

// Retrieve word with given hash, returns nil if not found.
func (c *WordCache) LoadWord(word string, hash uint32) map[string]uint32 {
	value, _ := c.cache.Get(hash)

	if value == "" {
		return nil
	}
	var result map[string]uint32
	if err := gob.NewDecoder(bytes.NewReader([]byte(value))).Decode(&result); err != nil {
		// todo
		fmt.Printf("Error decoding cache value  %+v", err)
		return nil
	}
	return result
}

// Store new word in cache, words are stored as map in a form
// map[phoneme]phonemeHash.
func (c *WordCache) StoreWord(value map[string]uint32, hash uint32) {

	buf := bytes.Buffer{}

	if err := gob.NewEncoder(&buf).Encode(value); err != nil {
		fmt.Printf("Error encoding word %+v", value)
		return
	}
	c.cache.Set(hash, buf.String())
}

// Generate unique hash for given word.
func (c *WordCache) HashWord(word string) uint32 {
	return hash.StringHash(c.seed, word+"\x00")
}

// Initialize new cache client.
func NewWordCache() (*WordCache, error) {

	var buf [4]byte
	rand.Read(buf[:])
	seed := binary.LittleEndian.Uint32(buf[:])

	cache, err := otter.MustBuilder[uint32, string](10_000).CollectStats().Cost(
		func(key uint32, value string) uint32 {
			return 1
		}).Build()

	if err != nil {
		return nil, err
	}

	return &WordCache{
		seed:  seed,
		cache: cache,
	}, nil
}

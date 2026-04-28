package narrative

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"

	"github.com/exaroth/narrative/pkg/kitten"
	"github.com/exaroth/narrative/pkg/phonemizer"
	"github.com/exaroth/narrative/pkg/preprocessor"
	"github.com/exaroth/narrative/pkg/sentencizer"
)

// Represents single text source
type Source struct {
	pproc      *preprocessor.Preprocessor
	phonemizer *phonemizer.Phonemizer
	ttsClient  *kitten.Kitten
	// Raw sentence data
	data []string
	// Current sentence number
	sentenceNum int
	// Id of the text source
	id string
	// Sentence waveform data cache
	cache_buf map[int][]float32
	// Buffer size
	buf_size int
}

// Initialize new source instance.
func InitSource(
	path string, sentence_n, buf_size int,
	pproc *preprocessor.Preprocessor,
	phonemizer *phonemizer.Phonemizer,
	ttsClient *kitten.Kitten,
) (*Source, error) {
	data, err := LoadTextSource(path)
	if err != nil {
		return nil, err
	}

	return &Source{
		data:        data.Data,
		id:          data.Id,
		pproc:       pproc,
		phonemizer:  phonemizer,
		sentenceNum: sentence_n,
		cache_buf:   make(map[int][]float32),
	}, nil
}

// Retrieve raw sentence data.
func (s *Source) getRawSentence(n int) string {
	if n < 0 || n > len(s.data)-1 {
		// should never happen
		panic(fmt.Sprintf("Invalid sentence number passed %d", n))
	}
	return s.data[n]
}

// Retrieve raw sentence data for current sentence num.
func (s *Source) getCurrentRawSentence() string {
	return s.getRawSentence(s.sentenceNum)
}

// Retrieve phonemized version of sentence at index n.
func (s *Source) getSentence(n int) (string, error) {
	p_sentence := s.pproc.ProcessSentence(s.getCurrentRawSentence())
	return s.phonemizer.Phonemize(p_sentence)
}

// Retrieve phonemized version of sentence stored
// at idx == sentenceNum.
func (s *Source) getCurrentSentence() (string, error) {
	return s.getSentence(s.sentenceNum)

}

// Increment current sentence number returning updated
// value, returns -1 if number cannot be incremented.
func (s *Source) incrementSentenceNum() int {
	if s.sentenceNum > len(s.data)-1 {
		return -1
	}
	s.sentenceNum += 1
	return s.sentenceNum
}

// Decrement sentence num returning updated value,
// returns -1 if value cannot be decremented.
func (s *Source) decrementSentenceNum() int {
	if s.sentenceNum <= 0 {
		return -1
	}
	s.sentenceNum -= 1
	return s.sentenceNum
}

// Representation of source as saved on disk.
type SourceData struct {
	Id   string
	Data []string
}

// Save text source as gob file.
func SaveTextSource(path, id string, data []byte) (string, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(SourceData{
		Data: sentencizer.Sentencize(data),
		Id:   id,
	})
	if err != nil {
		return "", err
	}
	p := filepath.Join(path, fmt.Sprintf("%s.gob", id))
	return p, os.WriteFile(p, buf.Bytes(), 0644)
}

// Load raw source from file.
func LoadTextSource(path string) (*SourceData, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0644)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	var result SourceData
	err = gob.NewDecoder(f).Decode(&result)
	return &result, err
}

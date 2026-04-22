package narrative

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"

	"github.com/exaroth/narrative/pkg/preprocessor"
	"github.com/exaroth/narrative/pkg/sentencizer"
)

// Represents single text source
type Source struct {
	pproc           *preprocessor.Preprocessor
	data            []string
	currentSentence int
}

// Initialize new source instance.
func InitSource(path string, sentence_n int, pproc *preprocessor.Preprocessor) (*Source, error) {
	data, err := LoadTextSource(path)
	if err != nil {
		return nil, err
	}

	return &Source{
		data:            sentencizer.Sentencize(data.Data),
		pproc:           pproc,
		currentSentence: sentence_n,
	}, nil
}

type SourceData struct {
	Id   string
	Data []byte
}

// Save text source as gob file.
func SaveTextSource(path, id string, data []byte) (string, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(SourceData{
		Data: data,
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

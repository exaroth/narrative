package reader

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// Basic text reader used for reading raw text files.
type TextReader struct {
	data  []string
	id    string
	title string
}

// Read data from given source
func (r TextReader) Read(source string) (SourceReader, error) {
	data, err := os.ReadFile(source)
	if err != nil {
		return nil, err
	}
	r.data = Sentencize(data)
	r.id = uuid.New().String()
	filename := filepath.Base(source)
	r.title = strings.TrimSuffix(filename, filepath.Ext(filename))
	return &r, nil
}

// For text file returns empty string
// as theres no metadata associated with the file.
func (r TextReader) Author() string {
	return ""
}

// Returns title, for text files we use filename.
func (r TextReader) Title() string {
	return r.title
}

// Return raw text data.
func (r TextReader) Data() []string {
	return r.data
}

// Return random uuid.
func (r TextReader) Id() string {
	return r.id
}

// We dont support chapters in bare text.
func (r TextReader) Chapters() []int {
	return []int{}
}

// We dont return any metadata associated with text files.
func (r TextReader) Metadata() string {
	return ""
}

func (r TextReader) Type() SourceType {
	return SourceTypeText
}

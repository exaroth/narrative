package reader

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// Basic text reader used for reading raw text files.
type TextReader struct {
	data  []byte
	id    string
	title string
}

// Read text file from local path
func (r *TextReader) readLocal(path string) error {
	var err error
	r.data, err = os.ReadFile(path)
	r.id = uuid.New().String()
	filename := filepath.Base(path)
	r.title = strings.TrimSuffix(filename, filepath.Ext(filename))
	return err
}

// Read data from given source
func (r TextReader) Read(source string) (SourceReader, error) {
	err := r.readLocal(source)
	return &r, err
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
func (r TextReader) Data() []byte {
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

func (r TextReader) Type() SourceType {
	return SourceTypeText
}

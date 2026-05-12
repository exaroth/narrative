package reader

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-shiori/go-readability"
	"github.com/google/uuid"
)

// Basic text reader used for reading raw text files.
type HtmlReader struct {
	data   []byte
	id     string
	title  string
	author string
}

// Read data from given source
func (r HtmlReader) Read(source string) (SourceReader, error) {
	var err error

	f, err := os.Open(source)
	if err != nil {
		return nil, fmt.Errorf("Error reading file: %w", err)
	}
	parsed, err := readability.FromReader(f, nil)
	if err != nil {
		return nil, fmt.Errorf("Error parsing html: %w", err)

	}
	if len(parsed.Title) > 0 {
		r.title = parsed.Title
	} else {
		r.title = filepath.Base(source)
	}

	// TODO: might do lookup in html for author tags.
	if len(parsed.SiteName) > 0 {
		r.author = parsed.SiteName
	}

	r.id = uuid.New().String()
	r.data = []byte(parsed.TextContent)
	return &r, nil
}

// For html we return site name as author atm.
func (r HtmlReader) Author() string {
	return r.author
}

func (r HtmlReader) Title() string {
	return r.title
}

// Return raw text data.
func (r HtmlReader) Data() []byte {
	return r.data
}

// Return random uuid.
func (r HtmlReader) Id() string {
	return r.id
}

// TODO: add h1 chapters.
func (r HtmlReader) Chapters() []int {
	return []int{}
}

func (r HtmlReader) Type() SourceType {
	return SourceTypeHTML
}

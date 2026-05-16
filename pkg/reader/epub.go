package reader

import (
	"fmt"

	"github.com/taylorskalyo/goreader/epub"
)

// Reader for handling epub formatted ebooks.
type EpubReader struct {
	data     []string
	id       string
	title    string
	author   string
	chapters []int
	updateCh chan<- string
}

// Read data from given source
func (r EpubReader) Read(source string) (SourceReader, error) {
	book_d, err := epub.OpenReader(source)
	if err != nil {
		panic(err)
	}
	if len(book_d.Rootfiles) == 0 {
		return nil, fmt.Errorf("Book does not seem to contain any data.")
	}
	book := book_d.Rootfiles[0]

	r.title = book.Title
	r.author = book.Creator
	r.chapters = []int{}

	if len(book.Spine.Itemrefs) == 0 {
		return nil, fmt.Errorf("Book does not contain any text.")
	}

	proc := NewEpubProcessor(&book.Package, book.Spine.Itemrefs)
	s, c, err := proc.ProcessBookContents(r.updateCh)
	if err != nil {
		return nil, err
	}
	r.data = s
	r.chapters = c

	return &r, nil
}

func (r EpubReader) Author() string {
	return r.author
}

func (r EpubReader) Title() string {
	return r.title
}

func (r EpubReader) Data() []string {
	return r.data
}

func (r EpubReader) Id() string {
	return r.id
}

func (r EpubReader) Chapters() []int {
	return r.chapters
}

func (r EpubReader) Metadata() string {
	return ""
}

func (r EpubReader) Type() SourceType {
	return SourceTypeEpub
}

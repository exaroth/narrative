package reader

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Interface to implement by readers of different
// types of text files.
type SourceReader interface {
	Read(string) (SourceReader, error)
	Data() []string
	Title() string
	Author() string
	Chapters() []int
	Id() string
	Type() SourceType
}

// SourceType represents unique type of text source
// passed to narrative.
type SourceType int

func (t SourceType) String() string { return SourceTypeName[t] }

const (
	SourceTypeText SourceType = iota
	SourceTypeEpub
	SourceTypeMobi
	SourceTypeAzw3
	SourceTypeHTML
	SourceTypeMarkdown
	SourceTypeUnsupported
)

// Returns human readable name for source type.
var SourceTypeName = map[SourceType]string{
	SourceTypeText:        "text",
	SourceTypeEpub:        "epub",
	SourceTypeMobi:        "mobi",
	SourceTypeAzw3:        "azw3",
	SourceTypeHTML:        "html",
	SourceTypeMarkdown:    "markdown",
	SourceTypeUnsupported: "unsupported",
}

// Retrieve source type and indicator whether data is remote
// and should be downloaded before processing.
func getSourceType(input string) SourceType {
	ext := filepath.Ext(strings.ToLower(input))
	switch ext {
	case ".txt":
		return SourceTypeText
	case ".epub":
		return SourceTypeEpub
	case ".mobi":
		return SourceTypeMobi
	case ".md", ".mkd", ".markdown":
		return SourceTypeMarkdown
	case ".azw3":
		return SourceTypeAzw3
	case ".html":
		return SourceTypeHTML
	default:
		return SourceTypeUnsupported
	}
}

// Retrieve reader for file at given path.
func GetReaderForContent(f_path string, update_ch chan<- string) (SourceReader, error) {
	var r SourceReader
	var err error
	st := getSourceType(f_path)
	switch st {
	case SourceTypeText:
		r, err = TextReader{}.Read(f_path)
	case SourceTypeHTML:
		r, err = HtmlReader{}.Read(f_path)
	case SourceTypeEpub:
		r, err = EpubReader{update_ch: update_ch}.Read(f_path)
	default:
		return nil, fmt.Errorf("Provided file type is not supported by Narrative.")
	}
	if err != nil {
		return nil, fmt.Errorf("Error retrieving reader for data: %w", err)
	}
	return r, nil
}

package reader

import (
	"bytes"
	"fmt"
	"os"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/google/uuid"
)

var MARKDOWN_EXTENSIONS = parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
var MARKDOWN_HTML_FLAGS = html.CommonFlags | html.HrefTargetBlank

// Reader for markdown files.
type MarkdownReader struct {
	data     []string
	chapters []int
	id       string
	title    string
}

// Convert markdown file to html, then sentencize the output.
func (r MarkdownReader) Read(source string) (SourceReader, error) {
	data, err := os.ReadFile(source)
	if err != nil {
		return nil, fmt.Errorf("Error reading markdown file: %w", err)
	}
	p := parser.NewWithExtensions(MARKDOWN_EXTENSIONS)

	html_ast := p.Parse(data)

	opts := html.RendererOptions{Flags: MARKDOWN_HTML_FLAGS}
	renderer := html.NewRenderer(opts)

	html_r := markdown.Render(html_ast, renderer)
	html_b := bytes.NewReader(html_r)
	proc, err := NewHTMLProcessor(html_b, SourceTypeMarkdown)
	if err != nil {

		return nil, fmt.Errorf("Error initializing html processor: %w", err)
	}
	s, c, err := proc.ProcessBookContents(nil)
	if err != nil {
		return nil, err
	}
	r.data = s
	r.chapters = c
	r.id = uuid.New().String()
	r.title = "Test"
	return &r, nil
}

// We dont return author for markdown files.
func (r MarkdownReader) Author() string {
	return ""
}

func (r MarkdownReader) Title() string {
	return r.title
}

func (r MarkdownReader) Data() []string {
	return r.data
}

func (r MarkdownReader) Id() string {
	return r.id
}

// For markdown files we mark h1/h2 headers
// as chapters
func (r MarkdownReader) Chapters() []int {
	return r.chapters
}

func (r MarkdownReader) Metadata() string {
	return ""
}

func (r MarkdownReader) Type() SourceType {
	return SourceTypeMarkdown
}

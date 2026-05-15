package reader

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path"
	"regexp"
	"strings"

	"github.com/taylorskalyo/goreader/epub"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
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

	var buf bytes.Buffer
	proc := NewHTMLProcessor(&book.Package)

	refs := book.Spine.Itemrefs
	if len(refs) == 0 {
		return nil, fmt.Errorf("Book does not contain any text.")
	}

	if len(refs) == 1 {
		err := proc.ProcessChapter(context.TODO(), 0, &buf)
		if err != nil {
			return nil, fmt.Errorf("Error processing chapter, %w", err)
		}
		r.data = Sentencize(buf.Bytes())
		return &r, nil
	}

	r.data = []string{}
	ch_l := 0
	for idx, _ := range refs {
		if r.updateCh != nil {
			r.updateCh <- fmt.Sprintf(
				"Processing chapter %d (%3.0f%%)",
				idx,
				float64(idx+1)/float64(len(refs))*100,
			)
		}
		if idx < len(refs)-1 {
			r.chapters = append(r.chapters, ch_l)
		}
		err := proc.ProcessChapter(context.TODO(), idx, &buf)
		if err != nil {
			return nil, fmt.Errorf("Error processing chapter %d, %w", idx+1, err)
		}
		sentences := Sentencize(buf.Bytes())
		ch_l += len(sentences)
		r.data = append(r.data, sentences...)
	}

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

type htmlParser struct {
	tagStack  []atom.Atom
	rowStack  [][]string
	cellStack []strings.Builder

	tokenizer *html.Tokenizer
	newlines  int
	writer    *sentenceWriter
	basepath  string
}

func (p *htmlParser) ensureNewlines(n int) {
	if p.newlines >= n {
		return
	}
	p.newlines = n
}

type HTMLProcessor struct {
	content *epub.Package
	parser  htmlParser
}

func NewHTMLProcessor(content *epub.Package) HTMLProcessor {
	return HTMLProcessor{
		content: content,
	}
}

func (r *HTMLProcessor) ProcessChapter(ctx context.Context, chapter int, w io.Writer) error {
	item := r.content.Spine.Itemrefs[chapter]
	doc, err := item.Open()
	if err != nil {
		return err
	}

	r.parser = htmlParser{
		tokenizer: html.NewTokenizer(doc),
		writer:    newSentenceWriter(w),
		basepath:  path.Dir(item.HREF),
	}

	return r.process(ctx)
}

func (r *HTMLProcessor) process(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if err := r.handleToken(); err == io.EOF {
			r.parser.writer.Flush()
			return nil
		} else if err == io.EOF {
			return err
		}
	}
}

// handleToken is triggered when an HTML token is parsed.
func (r *HTMLProcessor) handleToken() error {
	tokenType := r.parser.tokenizer.Next()
	token := r.parser.tokenizer.Token()
	switch tokenType {
	case html.ErrorToken:
		return r.parser.tokenizer.Err()
	case html.StartTagToken:
		r.parser.tagStack = append(r.parser.tagStack, token.DataAtom) // push element
		return r.handleStartTag(token)
	case html.SelfClosingTagToken:
		return r.handleStartTag(token)
	case html.TextToken:
		return r.handleText(token)
	case html.EndTagToken:
		r.parser.tagStack = r.parser.tagStack[:len(r.parser.tagStack)-1] // pop element
		return nil
	}

	return nil
}

var escapePattern = regexp.MustCompile(`(\[[a-zA-Z0-9_,;: \-\."#]+\[*)\]`)

func Escape(text string) string {
	return escapePattern.ReplaceAllString(text, "$1[]")
}

// appendText appends text to the underlying writer.
func (r *HTMLProcessor) appendText(text string) error {
	if !hasText(text) {
		return nil
	}

	text = Escape(text)

	pendingLines := strings.Repeat("\n", r.parser.newlines)
	text = fmt.Sprintf("%s%s", pendingLines, text)

	r.parser.newlines = 0

	_, err := io.WriteString(r.parser.writer, text)

	return err
}

func (r *HTMLProcessor) handleText(token html.Token) error {
	// Skip style tags
	if len(r.parser.tagStack) > 0 && r.parser.tagStack[len(r.parser.tagStack)-1] == atom.Style {
		return nil
	}

	text := processWhitespace(token.Data)
	return r.appendText(string(text))
}

func (r *HTMLProcessor) handleStartTag(token html.Token) (err error) {
	switch token.DataAtom {
	case atom.Br:
		r.parser.newlines++
	case atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6, atom.Div:
		r.parser.ensureNewlines(2)
	case atom.P:
		r.parser.ensureNewlines(2)
	}

	return err
}

type sentenceWriter struct {
	w      io.Writer
	buffer strings.Builder
}

func newSentenceWriter(w io.Writer) *sentenceWriter {
	return &sentenceWriter{
		w: w,
	}
}

func (w *sentenceWriter) Write(p []byte) (n int, err error) {
	w.buffer.Write(p)
	lines := strings.Split(w.buffer.String(), "\n")

	for i, line := range lines {
		if i == len(lines)-1 {
			w.buffer.Reset()
			w.buffer.WriteString(line)
			break
		}

		nLine, err := w.w.Write([]byte(line + "\n"))
		if err != nil {
			return n, err
		}
		n += nLine
	}

	return len(p), nil
}

// Flush writes any lines remaining in the buffer.
func (w *sentenceWriter) Flush() error {
	if w.buffer.Len() > 0 {
		_, err := w.w.Write([]byte(w.buffer.String()))
		w.buffer.Reset()

		return err
	}

	return nil
}

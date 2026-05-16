package reader

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path"
	"slices"
	"strings"

	"github.com/taylorskalyo/goreader/epub"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var excludedAtoms = []atom.Atom{
	atom.Style, atom.Head,
	atom.Header, atom.Footer,
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

type EpubProcessor struct {
	content *epub.Package
	refs    []epub.Itemref
	parser  htmlParser
}

func NewEpubProcessor(content *epub.Package, refs []epub.Itemref) EpubProcessor {
	return EpubProcessor{
		content: content,
		refs:    refs,
	}
}

// Process all chapters of the book, returning sentence list and chapter list
func (r *EpubProcessor) ProcessBookContents(updateCh chan<- string) (s []string, c []int, err error) {
	var buf bytes.Buffer

	if len(r.refs) == 1 {
		if e := r.ProcessChapter(context.TODO(), 0, &buf); e != nil {
			err = fmt.Errorf("Error processing chapter, %w", e)
			return
		}
		s = Sentencize(buf.Bytes())
		return
	}

	ch_l := 0
	for idx, _ := range r.refs {
		if updateCh != nil {
			updateCh <- fmt.Sprintf(
				"Processing chapter %d (%3.0f%%)",
				idx,
				float64(idx+1)/float64(len(r.refs))*100,
			)
		}
		if idx < len(r.refs)-1 {
			c = append(c, ch_l)
		}
		if e := r.ProcessChapter(context.TODO(), idx, &buf); e != nil {
			err = fmt.Errorf("Error processing chapter %d, %w", idx+1, err)
			return
		}
		sentences := Sentencize(buf.Bytes())
		ch_l += len(sentences)
		s = append(s, sentences...)
	}
	return
}

func (r *EpubProcessor) ProcessChapter(ctx context.Context, chapter int, w io.Writer) error {
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

func (r *EpubProcessor) process(ctx context.Context) error {
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

func (r *EpubProcessor) handleToken() error {
	tokenType := r.parser.tokenizer.Next()
	token := r.parser.tokenizer.Token()
	switch tokenType {
	case html.ErrorToken:
		return r.parser.tokenizer.Err()
	case html.StartTagToken:
		r.parser.tagStack = append(r.parser.tagStack, token.DataAtom)
		return r.handleStartTag(token)
	case html.SelfClosingTagToken:
		return r.handleStartTag(token)
	case html.TextToken:
		return r.handleText(token)
	case html.EndTagToken:
		r.parser.tagStack = r.parser.tagStack[:len(r.parser.tagStack)-1]
		return nil
	}

	return nil
}

// appendText appends text to the underlying writer.
func (r *EpubProcessor) appendText(text string) error {
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

func (r *EpubProcessor) handleText(token html.Token) error {
	// Skip style tags

	for _, t := range r.parser.tagStack {
		if slices.Index(excludedAtoms, t) > -1 {
			return nil
		}
	}

	text := processWhitespace(token.Data)
	return r.appendText(string(text))
}

func (r *EpubProcessor) handleStartTag(token html.Token) (err error) {

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

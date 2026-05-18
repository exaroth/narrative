package reader

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path"
	"slices"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var htmlExcludedAtoms = []atom.Atom{
	atom.Style, atom.Head,
	atom.Header, atom.Footer,
	atom.Table, atom.Tbody, atom.Td,
	atom.Tr, atom.Code, atom.Pre,
}

type HTMLProcessor struct {
	source    io.Reader
	sentences [][]string
	parser    htmlParser
	writer    *bytes.Buffer
	updateCh  chan<- string
	sourceT   SourceType
}

func NewHTMLProcessor(source io.Reader, t SourceType) (*HTMLProcessor, error) {
	var buf bytes.Buffer

	return &HTMLProcessor{
		source:    source,
		writer:    &buf,
		sentences: [][]string{},
		sourceT:   t,
	}, nil
}

// Process all chapters of the html book, returning sentence list and chapter list.
func (r *HTMLProcessor) ProcessBookContents(updateCh chan<- string) ([]string, []int, error) {
	r.updateCh = updateCh
	r.parser = htmlParser{
		tokenizer: html.NewTokenizer(r.source),
		writer:    newSentenceWriter(r.writer),
		basepath:  path.Dir("/"),
	}

	err := r.process(context.TODO())
	if err != nil {
		return nil, nil, err
	}
	r.sentences = append(r.sentences, Sentencize(r.writer.Bytes()))

	if len(r.sentences) == 1 {
		return r.sentences[0], []int{}, nil
	}

	var cur_ch_offset int
	chapters := []int{}
	result := []string{}
	for idx, chapter_sentences := range r.sentences {
		result = append(result, chapter_sentences...)
		if idx < len(r.sentences)-1 {
			cur_ch_offset += len(chapter_sentences)
			chapters = append(chapters, cur_ch_offset)

		}
	}

	return result, chapters, nil
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
		} else if err != nil {
			return err
		}
	}
}

func (r *HTMLProcessor) handleToken() error {
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
	for _, t := range r.parser.tagStack {
		if slices.Index(htmlExcludedAtoms, t) > -1 {
			return nil
		}
	}

	text := processWhitespace(token.Data)
	return r.appendText(string(text))
}

// Detect new chapter, these are arbitrary breakpoints, dependent
// on type of html source.
func (r *HTMLProcessor) detectChapter(token html.Token) bool {
	if r.sourceT == SourceTypeMobi && token.DataAtom == atom.A {
		attrs := token.Attr
		for _, a := range attrs {
			if a.Key == "id" && strings.HasPrefix(a.Val, "filepos") {
				return true
			}
		}
	}
	if r.sourceT == SourceTypeMarkdown {
		if token.DataAtom == atom.H1 || token.DataAtom == atom.H2 {
			return true
		}
	}
	return false
}

// Write current chapter sentences and reset buffer.
func (r *HTMLProcessor) updateChapter() {
	if r.updateCh != nil {
		r.updateCh <- fmt.Sprintf("Processing chapter %d", len(r.sentences)+1)
	}
	r.sentences = append(r.sentences, Sentencize(r.writer.Bytes()))
	r.writer.Reset()
}

func (r *HTMLProcessor) handleStartTag(token html.Token) (err error) {
	if r.detectChapter(token) {
		r.updateChapter()
	}
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

package reader

import (
	"io"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type htmlParser struct {
	tagStack []atom.Atom

	tokenizer *html.Tokenizer
	writer    *sentenceWriter
	newlines  int
	basepath  string
}

func (p *htmlParser) ensureNewlines(n int) {
	if p.newlines >= n {
		return
	}
	p.newlines = n
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

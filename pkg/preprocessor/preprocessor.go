package preprocessor

import (
	"regexp"
	"strings"
)

var (
	LETTERS_RE = regexp.MustCompile("[a-zA-Z]")
)

type PreprocessorFunc func(input string) (string, error)

type Preprocessor struct {
	funcs []PreprocessorFunc
}

func (p *Preprocessor) ProcessSentence(input string) string {
	var err error
	for _, f := range p.funcs {
		input, err = f(input)
		if err != nil {
			// todo
			panic(err)
		}
	}
	return strings.ToLower(input)
}

func (p *Preprocessor) RegisterFunc(f PreprocessorFunc) {
	p.funcs = append(p.funcs, f)
}

func NewPreprocessor() *Preprocessor {
	proc := &Preprocessor{
		funcs: []PreprocessorFunc{},
	}

	proc.RegisterFunc(cleanupUnusableTextParts)
	proc.RegisterFunc(normalizePunctuation)
	proc.RegisterFunc(splitHyphenizedWords)
	proc.RegisterFunc(removeTrailingApostrophes)

	proc.RegisterFunc(removeNonProsodicPunctuation)
	proc.RegisterFunc(normalizeWhitespace)

	return proc
}

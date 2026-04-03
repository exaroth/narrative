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
	var proc string
	for _, f := range p.funcs {
		proc, err = f(input)
		if err != nil {
			// todo, log
			continue
		}
		input = proc
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
	proc.RegisterFunc(trimSentenceQuotes)
	proc.RegisterFunc(expandContractions)
	proc.RegisterFunc(expandLeadingDecimals)
	proc.RegisterFunc(expandCurrency)
	proc.RegisterFunc(splitHyphenizedWords)
	proc.RegisterFunc(removeTrailingApostrophes)

	proc.RegisterFunc(expandPercentages)
	proc.RegisterFunc(expandOrdinals)
	proc.RegisterFunc(expandUnits)
	proc.RegisterFunc(expandTime)
	proc.RegisterFunc(expandDecades)
	proc.RegisterFunc(expandFractions)

	proc.RegisterFunc(replaceNumbers)

	proc.RegisterFunc(removeNonProsodicPunctuation)
	proc.RegisterFunc(normalizeWhitespace)

	return proc
}

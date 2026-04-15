package preprocessor

import (
	"regexp"
	"strings"
)

var (
	LETTERS_RE = regexp.MustCompile("[a-zA-Z]")
)

// This represents basic preprocessor function,
// which accepts input and returns modified input or error.
type PreprocessorFunc func(input string) (string, error)

// Preprocessor represents controller
// struct for handling preprocessing
type Preprocessor struct {
	funcs []PreprocessorFunc
}

// Run processing pipeline for the input string.
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

// Register new preprocessor function, should be called in order.
func (p *Preprocessor) RegisterFunc(f PreprocessorFunc) {
	p.funcs = append(p.funcs, f)
}

// Initialize new preprocessor.
func NewPreprocessor() *Preprocessor {
	proc := &Preprocessor{
		funcs: []PreprocessorFunc{},
	}

	proc.RegisterFunc(cleanupUnusableTextParts)
	proc.RegisterFunc(normalizePunctuation)
	proc.RegisterFunc(trimSentenceQuotes)
	proc.RegisterFunc(processDashes)
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

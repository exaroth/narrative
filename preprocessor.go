package main

// type PreprocessorFun(input string) (string, error) func

type Preprocessor struct {
	source string
}

func (p *Preprocessor) ProcessSentence(input string) string {
	return input
}

func NewPreprocessor() *Preprocessor {
	return &Preprocessor{}
}

package main

type Preprocessor struct {
	source string
}

func (p *Preprocessor) ProcessSentence(input string) string {
	return input
}

func NewPreprocessor() *Preprocessor {
	return &Preprocessor{}
}

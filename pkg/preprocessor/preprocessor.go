package preprocessor

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
	return input
}

func (p *Preprocessor) RegisterFunc(f PreprocessorFunc) {
	p.funcs = append(p.funcs, f)
}

func NewPreprocessor() *Preprocessor {
	return &Preprocessor{
		funcs: []PreprocessorFunc{},
	}
}

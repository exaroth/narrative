package narrative

import "github.com/alexflint/go-arg"

// Holds arguments used by Narrative.
type NarrativeArgs struct {
	Source     string `arg:"positional"`
	ListVoices bool   `arg:"--list-voices"`
	Voice      string
	Help       bool
}

// Parse command line arguments.
func ParseArgs() (*NarrativeArgs, error) {
	var args NarrativeArgs
	err := arg.Parse(&args)
	return &args, err
}

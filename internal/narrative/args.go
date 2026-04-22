package narrative

import "github.com/alexflint/go-arg"

// Holds arguments used by Narrative.
type NarrativeArgs struct {
	Source string `arg:"positional"`
	Help   bool
}

// Parse command line arguments.
func ParseArgs() *NarrativeArgs {
	var args NarrativeArgs
	arg.MustParse(&args)
	return &args
}

package debugger

import (
	"github.com/exaroth/narrative/internal/config"
)

type Config struct {
	mainConfig *config.Config
	// Full relative or absolute path
	// to directory to be used for storing
	// debugger data.
	debuggerDirPathName string
	// Filename to be used for external dictionary
	externalDictFName string
	// Filename to be used for missing words dictionary.
	missingDictFName string
}

func DefaultConfig() *Config {
	use_sel_inf := true
	remove_word_hyp := false
	c := config.DefaultConfig()
	c.UsePhonemeSelectionInference = &use_sel_inf
	c.RemoveWordTrailingHyphens = &remove_word_hyp
	c.DebuggerMode = true
	return &Config{
		mainConfig:          c,
		debuggerDirPathName: "./narrative-debugger",
		externalDictFName:   "aux_dict.csv",
		missingDictFName:    "missing_dict.csv",
	}
}

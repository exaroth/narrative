package debugger

import (
	"github.com/exaroth/narrative/internal/config"
	log "github.com/sirupsen/logrus"
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
	return &Config{
		mainConfig: &config.Config{
			LogLevel:                     log.InfoLevel,
			DebuggerMode:                 true,
			UsePhonemeSelectionInference: true,
			RemoveWordTrailingHyphens:    false,
		},
		debuggerDirPathName: "./narrative-debugger",
		externalDictFName:   "aux_dict.csv",
		missingDictFName:    "missing_dict.csv",
	}
}

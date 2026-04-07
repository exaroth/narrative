package config

import (
	log "github.com/sirupsen/logrus"
)

// Config stores all configuration
// used for Narrative operation.
type Config struct {
	LogLevel                     log.Level
	DebuggerMode                 bool
	UsePhonemeSelectionInference bool
	// remove trailing hyphenes from phonemes,
	// ass kitten tts does not like those
	RemoveWordTrailingHyphens bool
}

// Initialize new config based on the options provided.
func NewConfig(
	log_level log.Level,
	debug_mode bool,
	use_selection_inference bool,
	remove_word_trailing_hyphens bool,
) *Config {
	return &Config{
		DebuggerMode:                 debug_mode,
		LogLevel:                     log_level,
		UsePhonemeSelectionInference: use_selection_inference,
		RemoveWordTrailingHyphens:    remove_word_trailing_hyphens,
	}
}

// Initialize new config with default options.
func DefaultConfig() *Config {
	return NewConfig(
		log.ErrorLevel,
		false,
		true,
		true,
	)
}

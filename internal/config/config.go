package config

import (
	log "github.com/sirupsen/logrus"
)

// Config stores all configuration
// used for Narrative operation.
type Config struct {
	DebuggerMode                 bool
	LogLevel                     log.Level
	UsePhonemeSelectionInference bool
}

// Initialize new config based on the options provided.
func NewConfig(log_level log.Level, debug_mode bool, use_selection_inference bool) *Config {
	return &Config{
		DebuggerMode:                 debug_mode,
		LogLevel:                     log_level,
		UsePhonemeSelectionInference: use_selection_inference,
	}
}

// Initialize new config with default options.
func DefaultConfig() *Config {
	return NewConfig(
		log.ErrorLevel,
		false,
		true,
	)
}

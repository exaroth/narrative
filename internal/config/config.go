package config

import (
	log "github.com/sirupsen/logrus"
)

// Config stores all configuration
// used for Narrative operation.
type Config struct {
	DebuggerMode bool
	LogLevel     log.Level
}

// Initialize new config based on the options provided.
func NewConfig(log_level log.Level, debug_mode bool) *Config {
	return &Config{
		DebuggerMode: debug_mode,
		LogLevel:     log_level,
	}
}

// Initialize new config with default options.
func DefaultConfig() *Config {
	return NewConfig(
		log.ErrorLevel,
		false,
	)
}

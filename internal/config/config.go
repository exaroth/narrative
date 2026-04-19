package config

import (
	log "github.com/sirupsen/logrus"
)

const (
	// Speech speed
	DEFAULT_TTS_SPEED float64 = 1.0
	// Pause between sentences
	DEFAULT_TTS_PAUSE = 0
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
	Paths                     *NarrativePaths
	Speed                     float64
	Pause                     float64
}

// Initialize new config based on the options provided.
func NewConfig(
	log_level log.Level,
	debug_mode bool,
	use_selection_inference bool,
	remove_word_trailing_hyphens bool,
	speed float64,
	pause float64,
) *Config {
	paths := InitPaths()

	MakePath(paths.ConfigDir)
	MakePath(paths.DataDir)
	MakePath(paths.ModelPath)
	MakePath(paths.LibPath)

	return &Config{
		DebuggerMode:                 debug_mode,
		LogLevel:                     log_level,
		UsePhonemeSelectionInference: use_selection_inference,
		RemoveWordTrailingHyphens:    remove_word_trailing_hyphens,
		Speed:                        speed,
		Pause:                        pause,
		Paths:                        paths,
	}
}

// Initialize new config with default options.
func DefaultConfig() *Config {
	return NewConfig(
		log.ErrorLevel,
		false,
		true,
		true,
		DEFAULT_TTS_SPEED,
		DEFAULT_TTS_PAUSE,
	)
}

package config

import (
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	// Speech speed
	DEFAULT_TTS_SPEED float64 = 1.0
	// Pause between sentences
	DEFAULT_TTS_PAUSE = 0
	// Number of sentneces to buffer during playback
	DEFAULT_BUFFER_SIZE = 5
)

// Config stores all configuration
// used for Narrative operation.
type Config struct {
	// LogLevel                     log.Level
	DebuggerMode                 bool `yaml:"debugger_mode"`
	UsePhonemeSelectionInference bool `yaml:"use_phoneme_selection"`
	// remove trailing hyphenes from phonemes,
	// ass kitten tts does not like those
	RemoveWordTrailingHyphens bool `yaml:"remove_trailing_hyphens"`
	// Speed of playback
	Speed float64 `yaml:"speed"`
	// Pause between sentences
	Pause float64 `yaml:"pause"`
	// Number of sentences to cache to keep in cache.
	BufferSize int `yaml:"buffer_size"`
}

// Save config as yaml file.
func (c *Config) Save(path string) error {
	m, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, m, 0644)
}

// Load Config from yaml file.
func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	bytes, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(bytes, &cfg)
	return &cfg, err
}

// Initialize new config based on the options provided.
func NewConfig(
	debug_mode bool,
	use_selection_inference bool,
	remove_word_trailing_hyphens bool,
	speed float64,
	pause float64,
	buffer_size int,
) *Config {

	return &Config{
		DebuggerMode:                 debug_mode,
		UsePhonemeSelectionInference: use_selection_inference,
		RemoveWordTrailingHyphens:    remove_word_trailing_hyphens,
		Speed:                        speed,
		Pause:                        pause,
		BufferSize:                   buffer_size,
	}
}

// Initialize new config with default options.
func DefaultConfig() *Config {
	return NewConfig(
		false,
		true,
		true,
		DEFAULT_TTS_SPEED,
		DEFAULT_TTS_PAUSE,
		DEFAULT_BUFFER_SIZE,
	)
}

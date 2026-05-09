package kitten

// Main configuration
// for kittenTTS runner
type KittenConfig struct {

	// Whether or not to remove leading hyphens
	// in words. This is desired as kitten model
	// does not render words properly with those in place
	RemoveLeadingHyphens bool

	// name of the voice to use during inference.
	Voice string

	// Replace soft g's with hard ones, input dicts
	// for some reason contain only former ones
	ReplaceSoftG bool

	// Path to voice npz file.
	VoiceFilePath string

	// Path to Kitten TTS onnx model file.
	ModelFilePath string

	// Path to ONNX library file.
	LibraryFilePath string

	// Playback speed
	Speed float32
}

// Return kitten config with default values.
func DefaultConfig() *KittenConfig {
	return &KittenConfig{
		RemoveLeadingHyphens: true,
		Voice:                DEFAULT_VOICE,
		ReplaceSoftG:         true,
		Speed:                DEFAULT_SPEED,
	}
}

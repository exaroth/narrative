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
}

// Return kitten config with default values.
func DefaultConfig() *KittenConfig {
	return &KittenConfig{
		RemoveLeadingHyphens: true,
		Voice:                "Luna",
		ReplaceSoftG:         true,
	}
}

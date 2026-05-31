package phonemizer

import "bytes"

// Main config for the phonemizer.
type PhonemizerConfig struct {
	// Contains optional path to external dictionary
	// for overriding phonemes found in default dicts.
	ExternalDictPath string
	// Whether or not to use phoneme selection service
	// when picking phoneme for given word.
	UseSelectionInference bool
	// Contains information about given language.
	Language *Language
	// Contains weights used for inferring phonemes
	// that are not found in dictionary.
	InferenceWeights *bytes.Reader
	// Contains weights used in phoneme selection
	// service.
	HomonymWeights *bytes.Reader
	// Contains main dictionary which is zlib'ed
	// csv file in a format
	// word phoneme [tags...]
	Dictionary *bytes.Reader
	// Contains auxiliary dictionary (csv) which contains
	// overrides for phonemes found in default
	// dictionary, in format:
	// word phoneme.
	AuxDictionary *bytes.Reader
}

package common

import (
	"bytes"

	dict "github.com/exaroth/narrative/dictionary"
	"github.com/exaroth/narrative/pkg/phonemizer"
)

const (
	LANGUAGE_FILE_FNAME             = "language.json"
	PHONEME_INFERENCE_WEIGHTS_FNAME = "weights6.json.zlib"
	HOMONYM_WEIGHTS_FNAME           = "weights7.json.zlib"
	DICT_F_NAME                     = "missing.all.zlib"
	AUX_DICT_F_NAME                 = "aux_dict.csv"
)

// Initialize default instance of phonemizer for Narrative.
func NewPhonemizer(
	externalDictPath string,
	useSelectionInference bool,
) (*phonemizer.Phonemizer, error) {
	var err error
	inf_contents, err := dict.Language.ReadFile(PHONEME_INFERENCE_WEIGHTS_FNAME)
	if err != nil {
		return nil, err
	}

	main_dict, err := dict.Language.ReadFile(DICT_F_NAME)
	if err != nil {
		return nil, err
	}

	sel_weights, err := dict.Language.ReadFile(HOMONYM_WEIGHTS_FNAME)
	if err != nil {
		return nil, err
	}

	aux_dict, err := dict.Language.ReadFile(AUX_DICT_F_NAME)
	if err != nil {
		return nil, err
	}

	lang_raw, err := dict.Language.ReadFile(LANGUAGE_FILE_FNAME)
	if err != nil {
		return nil, err
	}

	lang, err := phonemizer.NewLanguage(lang_raw)
	if err != nil {
		return nil, err
	}

	p_config := &phonemizer.PhonemizerConfig{
		ExternalDictPath:      externalDictPath,
		UseSelectionInference: useSelectionInference,
		Language:              lang,
		InferenceWeights:      bytes.NewReader(inf_contents),
		Dictionary:            bytes.NewReader(main_dict),
		AuxDictionary:         bytes.NewReader(aux_dict),
		HomonymWeights:        bytes.NewReader(sel_weights),
	}
	return phonemizer.NewPhonemizer(p_config)
}

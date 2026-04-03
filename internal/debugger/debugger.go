package debugger

import (
	"fmt"
	"os"

	"github.com/exaroth/narrative/internal/config"
	"github.com/exaroth/narrative/pkg/kitten"
	"github.com/exaroth/narrative/pkg/phonemizer"
	"github.com/exaroth/narrative/pkg/preprocessor"
	"github.com/exaroth/narrative/pkg/sentencizer"
	"github.com/k0kubun/pp"
)

type Debugger struct {
	source       []string
	config       *config.Config
	ttsClient    *kitten.Kitten
	phonemizer   *phonemizer.Phonemizer
	preprocessor *preprocessor.Preprocessor
}

func NewDebugger(fpath string) (*Debugger, error) {
	config := config.DefaultConfig()
	config.DebuggerMode = true

	input, err := os.ReadFile(fpath)
	if err != nil {
		return nil, fmt.Errorf("debugger init err; invalid input %s: %w", fpath, err)
	}

	phonemizer, err := phonemizer.NewPhonemizer()
	if err != nil {
		return nil, fmt.Errorf("debugger init err; phonemizer init: %w", err)
	}

	preprocessor := preprocessor.NewPreprocessor()
	kitten := kitten.NewKitten(nil)

	return &Debugger{
		config:       config,
		ttsClient:    kitten,
		phonemizer:   phonemizer,
		preprocessor: preprocessor,
		source:       sentencizer.Sentencize(input),
	}, nil
}

func (d *Debugger) Deinit() {
	d.ttsClient.Deinit()
}

func (d *Debugger) Run() error {
	for _, sentence := range d.source {
		if err := d.processSentence(sentence); err != nil {
			return err
		}
	}
	return nil
}

func (d *Debugger) processSentence(sentence string) error {
	fmt.Println("Original: ", sentence)
	p_sentence := d.preprocessor.ProcessSentence(sentence)
	fmt.Println("Processed: ", p_sentence)
	words, punct := phonemizer.SplitPunctuation(p_sentence)

	fmt.Println("Words:")
	pp.Println(words)
	fmt.Println("Punct:")
	pp.Println(punct)

	phonemes, phoneme_opts, err := d.getPhonemeOptionsForSentence(words)
	if err != nil {
		return fmt.Errorf("Err retrieving phoneme opts for %s: %w", sentence, err)
	}

	selected := d.phonemizer.SelectPhonemes(phonemes, phoneme_opts)
	fmt.Println("Selected: ", selected)

	phoneme_a := []string{}
	for _, p := range selected {
		phoneme_a = append(phoneme_a, p[1])
	}

	compacted := phonemizer.CompactPunctuation(phoneme_a, punct)
	fmt.Println("Phonemized: ", compacted)

	_, err = d.ttsClient.RunInference(compacted)
	if err != nil {
		return err
	}

	return nil
}

func (d *Debugger) getPhonemeOptionsForSentence(sentence []string) (
	[]map[string]uint32, *phonemizer.PhonemeOptions, error,
) {

	phonemes := make([]map[string]uint32, len(sentence))
	for i, w := range sentence {
		phonemized_w, err := d.phonemizer.PhonemizeWord(w)
		if err != nil {
			return nil, nil, err
		}
		if phonemized_w == nil {
			phonemes[i] = map[string]uint32{w: 0}
		} else {
			phonemes[i] = phonemized_w
		}
	}
	phoneme_opts := d.phonemizer.GetPhonemeOptions(phonemes)

	return phonemes, phoneme_opts, nil
}

package debugger

import (
	"fmt"
	"os"

	"github.com/exaroth/narrative/internal/config"
	"github.com/exaroth/narrative/pkg/kitten"
	"github.com/exaroth/narrative/pkg/phonemizer"
	"github.com/exaroth/narrative/pkg/preprocessor"
	"github.com/exaroth/narrative/pkg/sentencizer"

	tea "charm.land/bubbletea/v2"
)

type sentenceData struct {
	// all available options for each word in sentence,
	// including tags
	opts *phonemizer.PhonemeOptions
	// phonemes selected to be used
	selectedPhonemes [][2]string
	// array containing all punctuation in the sentence
	punctuation []*phonemizer.Mark
	// phonemized output
	phonemized string
}

type Debugger struct {
	// all sentences from the input
	source []string
	// index of the current sentence
	currentSentence int
	// phonemization data for each sentence
	sentenceData map[int]*sentenceData
	model        tea.Model
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

	debugger := &Debugger{
		config:          config,
		ttsClient:       kitten,
		phonemizer:      phonemizer,
		preprocessor:    preprocessor,
		model:           nil,
		source:          sentencizer.Sentencize(input),
		currentSentence: 0,
		sentenceData:    make(map[int]*sentenceData),
	}

	return debugger, nil
}

func (d *Debugger) Deinit() {
	d.ttsClient.Deinit()
}

func (d *Debugger) play(sentence_n int) {
	// todo
	// _, err = d.ttsClient.RunInference(compacted)
	// if err != nil {
	// 	return err
	// }
}

func (d *Debugger) getSentenceData(sentence_num uint) (*sentenceData, error) {
	n := int(sentence_num)

	if int(n) > len(d.source) {
		return nil, fmt.Errorf("Invalid sentence idx: %d", n)
	}
	if d.sentenceData[n] != nil {
		return d.sentenceData[n], nil
	}
	sentence := d.source[n]

	fmt.Println("Original: ", sentence)
	p_sentence := d.preprocessor.ProcessSentence(sentence)
	fmt.Println("Processed: ", p_sentence)
	words, punct := phonemizer.SplitPunctuation(p_sentence)

	phonemes, opts, err := d.getPhonemeOptionsForSentence(words)
	if err != nil {
		return nil, err
	}

	selected := d.phonemizer.SelectPhonemes(phonemes, opts)

	phoneme_a := []string{}
	for _, p := range selected {
		phoneme_a = append(phoneme_a, p[1])
	}

	compacted := phonemizer.CompactPunctuation(phoneme_a, punct)
	data := &sentenceData{
		opts:             opts,
		selectedPhonemes: selected,
		punctuation:      punct,
		phonemized:       compacted,
	}
	d.sentenceData[n] = data

	return data, nil
}

func (d *Debugger) nextSentenceData() (*sentenceData, error) {
	if d.currentSentence == len(d.source) {
		return nil, nil
	}
	data, err := d.getSentenceData(uint(d.currentSentence + 1))
	if err != nil {
		return nil, err
	}
	d.currentSentence += 1
	return data, nil
}

func (d *Debugger) prevSentenceData() (*sentenceData, error) {
	if d.currentSentence == 0 {
		return nil, nil
	}
	data, err := d.getSentenceData(uint(d.currentSentence - 1))
	if err != nil {
		return nil, err
	}
	d.currentSentence -= 1
	return data, nil
}

func (d *Debugger) Run() error {
	// for _, sentence := range d.source {
	// 	if err := d.processSentence(sentence); err != nil {
	// 		return err
	// 	}
	// }
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

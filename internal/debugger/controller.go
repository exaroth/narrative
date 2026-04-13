package debugger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/exaroth/narrative/pkg/kitten"
	"github.com/exaroth/narrative/pkg/phonemizer"
	"github.com/exaroth/narrative/pkg/player"
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
	punctuation phonemizer.Punctuation
	// phonemized output
	phonemized string
}

type Debugger struct {
	// all sentences from the input
	source []string
	// phonemization data for each sentence
	sentenceData map[int]*sentenceData
	model        tea.Model
	config       *Config
	extDictPath  string
	extDict      *CsvDict
	ttsClient    *kitten.Kitten
	phonemizer   *phonemizer.Phonemizer
	preprocessor *preprocessor.Preprocessor

	player *player.Player
	// whether we are plating sentence/phoneme sample atmj
	isPlaying bool
}

// Initialize debugger directory along with
// dictionaries inside.
func initDebuggerDir(config *Config) ([]string, error) {
	var err error
	d_path := config.debuggerDirPathName
	if !filepath.IsAbs(d_path) {
		d_path, err = filepath.Abs(config.debuggerDirPathName)
		if err != nil {
			return nil, err
		}
	}
	if _, err = os.Stat(d_path); err != nil {
		err = os.Mkdir(d_path, 0755)
		if err != nil {
			return nil, err
		}
	}

	d_filenames := []string{
		config.externalDictFName,
		config.missingDictFName,
	}
	var result []string
	for _, f := range d_filenames {
		f_path := filepath.Join(d_path, f)
		if _, err = os.Stat(f_path); err != nil {
			_, err = os.Create(f_path)
			if err != nil {
				return nil, err
			}
		}
		result = append(result, f_path)
	}

	return result, nil
}

func NewDebugger(input_fpath string) (*Debugger, error) {
	config := DefaultConfig()

	input, err := os.ReadFile(input_fpath)
	if err != nil {
		return nil, fmt.Errorf("debugger init err; invalid input %s: %w", input_fpath, err)
	}
	paths, err := initDebuggerDir(config)
	if err != nil {
		return nil, fmt.Errorf("err initializing debugger dir: %w", err)
	}

	phonemizer, err := phonemizer.NewPhonemizer(paths[0])
	if err != nil {
		return nil, fmt.Errorf("debugger init err; phonemizer init: %w", err)
	}

	csv_dict, err := LoadCsvDict(paths[0])
	if err != nil {
		return nil, err
	}

	preprocessor := preprocessor.NewPreprocessor()
	kitten := kitten.NewKitten(kitten.DefaultConfig())

	debugger := &Debugger{
		config:       config,
		ttsClient:    kitten,
		phonemizer:   phonemizer,
		preprocessor: preprocessor,
		extDictPath:  paths[0],
		extDict:      csv_dict,
		model:        nil,
		source:       sentencizer.Sentencize(input),
		sentenceData: make(map[int]*sentenceData),
		player:       player.InitPlayer(),
	}

	model := NewDebuggerModel(debugger, 0)
	debugger.model = model

	return debugger, nil
}

func (d *Debugger) Deinit() {
	d.ttsClient.Deinit()
}

func (d *Debugger) Run() {
	p := tea.NewProgram(d.model)
	if _, err := p.Run(); err != nil {
		panic(err)
	}

}

func (d *Debugger) play(input, suffix string) {
	s_data, err := d.ttsClient.RunInference(input + suffix)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	d.player.AddSample(s_data)
	d.player.Play()
}

func (d *Debugger) updateExtDict(word, phoneme string) error {
	d.extDict.Update(word, phoneme)
	if err := d.extDict.Save(); err != nil {
		return err
	}
	// TODO fix reload dict funct
	// return d.phonemizer.ReloadDictionaries(phonemizer.ReloadRequestExt)
	phonemizer, err := phonemizer.NewPhonemizer(d.extDictPath)
	if err != nil {
		return err
	}
	d.phonemizer = phonemizer
	return nil
}

func (d *Debugger) clearCache(sentence_n int) {
	delete(d.sentenceData, sentence_n)
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

	p_sentence := d.preprocessor.ProcessSentence(sentence)
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

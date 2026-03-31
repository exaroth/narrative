package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/exaroth/narrative/pkg/kitten"
	"github.com/exaroth/narrative/pkg/phonemizer"
	"github.com/exaroth/narrative/pkg/preprocessor"
	"github.com/exaroth/narrative/pkg/sentencizer"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	log.SetLevel(log.WarnLevel)
}

func main() {

	input, err := os.ReadFile("./scratch/kafka-on-the-shore.txt")
	if err != nil {
		panic(err)
	}

	token_map := kitten.BuildTokenMap()

	phonemizer, err := phonemizer.NewPhonemizer()
	if err != nil {
		panic(err)
	}

	preprocessor := preprocessor.NewPreprocessor()
	kitten := kitten.NewKitten(nil)

	defer kitten.Deinit()

	var waveform_data []float32
	var phonemized string
	for _, sentence := range sentencizer.Sentencize(input) {

		_, _ = os.Create("./.tts-processing.lock")

		sentence = preprocessor.ProcessSentence(sentence)
		phonemized, err = phonemizer.Phonemize(sentence)

		fmt.Println("Original: ", sentence)
		fmt.Println("Phonemized: ", phonemized)

		if err != nil {
			panic(err)
		}

		waveform_data, err = kitten.RunInference(token_map.TokenizeWord(phonemized))
		if err != nil {
			log.Fatalf("%+v", err)
		}

		fname := "out.bin"
		file, _ := os.Create(fname)

		for _, sample := range waveform_data {
			var buf [8]byte
			binary.LittleEndian.PutUint32(buf[:], math.Float32bits(float32(sample)))
			_, err := file.Write(buf[:])
			if err != nil {
				panic(err)
			}
		}

		_ = os.Remove("./.tts-processing.lock")
		time.Sleep(1 * time.Second)

		for {
			if f, _ := os.Stat("./.tts-playback.lock"); f != nil {
				time.Sleep(1 * time.Second)
			} else {
				break
			}
		}
	}
}

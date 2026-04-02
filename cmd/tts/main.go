package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"strings"

	"github.com/exaroth/narrative/pkg/kitten"
	"github.com/exaroth/narrative/pkg/phonemizer"
	"github.com/exaroth/narrative/pkg/preprocessor"
)

func main() {
	var err error
	if len(os.Args) < 2 {
		panic(errors.New("No input provided"))
	}

	input := strings.Join(os.Args[1:], " ")

	token_map := kitten.BuildTokenMap()

	repo := phonemizer.NewPhonemizerRepository()

	if err := repo.LoadLanguage(); err != nil {
		panic(err)
	}

	phonemizer, err := phonemizer.NewPhonemizer()
	if err != nil {
		panic(err)
	}
	preprocessor := preprocessor.NewPreprocessor()

	kitten := kitten.NewKitten(nil)
	defer kitten.Deinit()

	var waveform_data []float32
	var phonemized string

	input_p := preprocessor.ProcessSentence(input)
	phonemized, err = phonemizer.Phonemize(input_p)
	if err != nil {
		panic(err)
	}

	fmt.Println("Original: ", input)
	fmt.Println("Processed: ", input_p)
	fmt.Println("Phonemized: ", phonemized)

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

}

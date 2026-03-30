package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"

	"github.com/exaroth/narrative/phonemizer"
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

	token_map := buildTokenMap()

	phonemizer, err := phonemizer.NewPhonemizer()
	if err != nil {
		panic(err)
	}
	preprocessor := NewPreprocessor()

	kitten := NewKitten(nil)

	defer kitten.Deinit()

	for _, sentence := range Sentencize(input) {

		sentence = preprocessor.ProcessSentence(sentence)
		phonemized, err := phonemizer.Phonemize(sentence)

		fmt.Println("Original: ", sentence)
		fmt.Println("Phonemized: ", phonemized)

		if err != nil {
			panic(err)
		}

		tokens := token_map.TokenizeWord(phonemized)
		outputData, err := kitten.RunInference(tokens)
		if err != nil {
			log.Fatalf("%+v", err)
		}

		fname := "out.bin"
		file, _ := os.Create(fname)

		for _, sample := range outputData {
			var buf [8]byte
			binary.LittleEndian.PutUint32(buf[:], math.Float32bits(float32(sample)))
			_, err := file.Write(buf[:])
			if err != nil {
				panic(err)
			}
		}
	}
}

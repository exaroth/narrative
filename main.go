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

const testText = "When I first tackled this, I made the mistake of not aligning the library versions, which caused unexpected runtime errors. So let’s get it right from the start."

func main() {

	token_map := buildTokenMap()
	phonemizer, err := phonemizer.NewPhonemizer()
	if err != nil {
		panic(err)
	}
	word := "unexpected"
	r, _ := phonemizer.Phonemize(word)
	i := r[0]
	fmt.Println(i)
	var w string
	for k, v := range i {
		if v == 0 {
			continue
		}
		fmt.Println("k: ", k)
		w = k
	}

	tokens := token_map.TokenizeWord(w)
	fmt.Println(tokens)

	fmt.Println("err")
	fmt.Println(err)

	kitten := NewKitten()

	defer kitten.Deinit()

	outputData, err := kitten.RunInference(tokens)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	fname := "out.bin"
	file, _ := os.Create(fname)
	// decayfac := math.Pow(end/start, 1.0/float64(nsamps))
	for _, sample := range outputData {
		var buf [8]byte
		binary.LittleEndian.PutUint32(buf[:], math.Float32bits(float32(sample)))
		_, err := file.Write(buf[:])
		if err != nil {
			panic(err)
		}
	}
}

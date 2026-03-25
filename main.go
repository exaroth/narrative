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

	input := ""
	if len(os.Args) > 1 {
		input = os.Args[1]
	}
	fmt.Println("input", input)

	token_map := buildTokenMap()
	phonemizer, err := phonemizer.NewPhonemizer()
	if err != nil {
		panic(err)
	}
	word := "mighty"
	r, _ := phonemizer.Phonemize(word)

	i := r[0]
	var w string
	for k, v := range i {
		if v == 0 {
			continue
		}
		w = k
	}

	tokens := token_map.TokenizeWord(w)
	fmt.Println(tokens)

	fmt.Println("err")
	fmt.Println(err)

	kitten := NewKitten(nil)

	defer kitten.Deinit()

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

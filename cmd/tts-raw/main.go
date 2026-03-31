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
)

func main() {
	var err error
	if len(os.Args) < 2 {
		panic(errors.New("No input provided"))
	}

	input := strings.Join(os.Args[1:], " ")

	token_map := kitten.BuildTokenMap()

	kitten := kitten.NewKitten(nil)

	defer kitten.Deinit()

	var waveform_data []float32

	fmt.Println("Input: ", input)

	waveform_data, err = kitten.RunInference(token_map.TokenizeWord(input))
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

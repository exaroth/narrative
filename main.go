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
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.WarnLevel)
}

const testText = "When I first tackled this, I made the mistake of not aligning the library versions, which caused unexpected runtime errors. So let’s get it right from the start."

func main() {

	// token_map := buildTokenMap()

	phonemizer, err := phonemizer.NewPhonemizer()
	if err != nil {
		panic(err)
	}
	r, _ := phonemizer.Phonemize("deprofundis")
	fmt.Println(r)

	inputData := []int64{0, 46, 47, 58, 60, 57, 48, 63, 56, 46, 51, 61, 16, 45, 54, 43, 55, 57, 16, 43, 46, 16, 62, 47, 16, 46, 57, 55, 51, 56, 47, 10, 0}

	fmt.Println("err")
	fmt.Println(err)

	kitten := NewKitten()

	defer kitten.Deinit()

	outputData, err := kitten.RunInference(inputData)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	// fmt.Println(outputData)

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

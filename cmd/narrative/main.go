package main

import (
	"os"

	"github.com/exaroth/narrative/internal/narrative"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)
}

func main() {

	ctrl, err := narrative.NewCtrl()
	if err != nil {
		panic(err)
	}
	err = ctrl.Run()
	if err != nil {
		panic(err)
	}
	// input, err := os.ReadFile("./dump/kafka-on-the-shore.txt")
	// if err != nil {
	// 	panic(err)
	// }

	// phonemizer, err := phonemizer.NewPhonemizer("")
	// if err != nil {
	// 	panic(err)
	// }

	// preprocessor := preprocessor.NewPreprocessor()
	// kitten := kitten.NewKitten(nil)

	// defer kitten.Deinit()

	// var waveform_data []float32
	// var phonemized string
	// var p_sentence string
	// for _, sentence := range sentencizer.Sentencize(input) {

	// 	_, _ = os.Create("./.tts-processing.lock")

	// 	p_sentence = preprocessor.ProcessSentence(sentence)
	// 	phonemized, err = phonemizer.Phonemize(p_sentence)

	// 	fmt.Println("Original: ", sentence)
	// 	fmt.Println("Processed: ", p_sentence)
	// 	fmt.Println("Phonemized: ", phonemized)

	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	waveform_data, err = kitten.RunInference(phonemized)
	// 	if err != nil {
	// 		log.Fatalf("%+v", err)
	// 	}

	// 	fname := "out.bin"
	// 	file, _ := os.Create(fname)

	// 	for _, sample := range waveform_data {
	// 		var buf [8]byte
	// 		binary.LittleEndian.PutUint32(buf[:], math.Float32bits(float32(sample)))
	// 		_, err := file.Write(buf[:])
	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 	}

	// 	_ = os.Remove("./.tts-processing.lock")
	// 	time.Sleep(1 * time.Second)

	// 	for {
	// 		if f, _ := os.Stat("./.tts-playback.lock"); f != nil {
	// 			time.Sleep(1 * time.Second)
	// 		} else {
	// 			break
	// 		}
	// 	}
	// }
}

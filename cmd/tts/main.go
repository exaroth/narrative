package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/exaroth/narrative/internal/common"
	"github.com/exaroth/narrative/pkg/kitten"
	"github.com/exaroth/narrative/pkg/phonemizer"
	"github.com/exaroth/narrative/pkg/player"
	"github.com/exaroth/narrative/pkg/preprocessor"
)

func main() {
	var err error
	if len(os.Args) < 2 {
		panic(errors.New("No input provided"))
	}

	input := strings.Join(os.Args[1:], " ")

	n_paths := common.InitDirectoryStructure()

	data_cfg, err := common.LoadDataConfig(n_paths.DataConfigPath)
	if err != nil {
		panic(fmt.Errorf("Could not load data cfg: %w", err))
	}

	lib_path, err := data_cfg.GetLibPath(n_paths)

	if err != nil {
		panic(fmt.Errorf("Could not retrieve library path: %w", err))
	}

	repo := phonemizer.NewPhonemizerRepository("")

	if err := repo.LoadLanguage(); err != nil {
		panic(err)
	}

	phonemizer, err := phonemizer.NewPhonemizer("", false)
	if err != nil {
		panic(err)
	}
	preprocessor := preprocessor.NewPreprocessor()

	kitten := kitten.InitKittenWithParams(
		lib_path,
		data_cfg.GetModelPath(n_paths),
		data_cfg.GetVoicesPath(n_paths),
		kitten.DEFAULT_VOICE,
		1.0,
	)
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

	waveform_data, err = kitten.RunInference(phonemized)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	speaker := player.InitPlayer()
	speaker.AddSample(waveform_data)
	done := make(chan bool)
	onExit := func() {
		done <- true
	}
	speaker.Play(onExit)
	<-done

}

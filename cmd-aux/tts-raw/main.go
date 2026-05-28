package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/exaroth/narrative/internal/common"
	"github.com/exaroth/narrative/pkg/kitten"
	"github.com/exaroth/narrative/pkg/player"
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

	kitten := kitten.InitKittenWithParams(
		lib_path,
		data_cfg.GetModelPath(n_paths),
		data_cfg.GetVoicesPath(n_paths),
		kitten.DEFAULT_VOICE,
		1.0,
	)
	defer kitten.Deinit()

	var waveform_data []float32

	waveform_data, err = kitten.RunInference(input)
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

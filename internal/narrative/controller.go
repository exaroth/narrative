package narrative

import (
	"fmt"
	"os"

	"github.com/exaroth/narrative/internal/config"
	"github.com/exaroth/narrative/pkg/kitten"
	"github.com/exaroth/narrative/pkg/phonemizer"
	"github.com/exaroth/narrative/pkg/player"
	"github.com/exaroth/narrative/pkg/preprocessor"
)

type NarrativeCtrl struct {
	phonemizer   *phonemizer.Phonemizer
	preprocessor *preprocessor.Preprocessor
	ttsClient    *kitten.Kitten
	player       *player.Player
	cfg          *config.Config
	paths        *NarrativePaths
}

func initDirectoryStructure() *NarrativePaths {
	paths := InitPaths()
	MakePath(paths.ConfigDir)
	MakePath(paths.DataDir)
	MakePath(paths.ModelPath)
	MakePath(paths.LibPath)
	return paths
}

func NewCtrl() (*NarrativeCtrl, error) {

	phonemizer, err := phonemizer.NewPhonemizer("")
	if err != nil {
		return nil, fmt.Errorf("init err; phonemizer init: %w", err)
	}

	preprocessor := preprocessor.NewPreprocessor()
	kitten := kitten.NewKitten(kitten.DefaultConfig())
	paths := initDirectoryStructure()

	var cfg *config.Config
	if _, err := os.Stat(paths.ConfigPath); err != nil {
		cfg = config.DefaultConfig()
		if err := cfg.Save(paths.ConfigPath); err != nil {
			return nil, fmt.Errorf("Error saving config @ %s; %w",
				paths.ConfigPath, err,
			)
		}
	} else {
		cfg, err = config.LoadConfig(paths.ConfigPath)
		if err != nil {
			return nil, fmt.Errorf("Error loading config @ %s, %w",
				paths.ConfigPath, err,
			)
		}
	}

	return &NarrativeCtrl{
		ttsClient:    kitten,
		phonemizer:   phonemizer,
		preprocessor: preprocessor,
		player:       player.InitPlayer(),
		cfg:          cfg,
		paths:        paths,
	}, nil
}

func Run() {

}

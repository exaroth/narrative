package narrative

import (
	"fmt"
	"os"

	"github.com/exaroth/narrative/internal/config"
	"github.com/exaroth/narrative/pkg/kitten"
	"github.com/exaroth/narrative/pkg/phonemizer"
	"github.com/exaroth/narrative/pkg/player"
	"github.com/exaroth/narrative/pkg/preprocessor"
	"github.com/exaroth/narrative/pkg/reader"

	tea "charm.land/bubbletea/v2"
)

type NarrativeCtrl struct {
	phonemizer    *phonemizer.Phonemizer
	preprocessor  *preprocessor.Preprocessor
	ttsClient     *kitten.Kitten
	player        *player.Player
	cfg           *config.Config
	dataCfg       *DataConfig
	paths         *NarrativePaths
	model         *narrativeModel
	args          *NarrativeArgs
	currentSource *Source
	program       *tea.Program
}

// Create basic directory structure for narrative.
func initDirectoryStructure() *NarrativePaths {
	paths := InitPaths()
	MakePath(paths.ConfigDir)
	MakePath(paths.DataDir)
	MakePath(paths.ModelPath)
	MakePath(paths.LibPath)
	MakePath(paths.SourcesPath)
	return paths
}

// Initialize new narrative controller.
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
	data_cfg, err := LoadDataConfig(paths.DataConfigPath)
	if err != nil {
		data_cfg = NewDataConfig()
		if err := data_cfg.Save(paths.DataConfigPath); err != nil {
			return nil, fmt.Errorf("Error creating data cfg: %w", err)
		}
	}

	ctrl := &NarrativeCtrl{
		ttsClient:    kitten,
		phonemizer:   phonemizer,
		preprocessor: preprocessor,
		player:       player.InitPlayer(),
		cfg:          cfg,
		dataCfg:      data_cfg,
		paths:        paths,
		args:         ParseArgs(),
	}
	model := InitNarrativeModel(ctrl)
	ctrl.model = model
	return ctrl, nil
}

func (c NarrativeCtrl) Source() *Source {
	return c.currentSource
}

// Add new text source based on the argument provided.
func (c *NarrativeCtrl) addNewSource() error {
	_, t := GetSourceType(c.args.Source)
	var r reader.SourceReader
	var err error
	switch t {
	case SourceTypeText:
		r, err = reader.TextReader{}.Read(c.args.Source)
	default:
		return fmt.Errorf("Unable to find reader for file type: %s", t)
	}
	if err != nil {
		return fmt.Errorf("Error reading text data: %w", err)
	}
	path, err := SaveTextSource(c.paths.SourcesPath, r.Id(), r.Data())
	if err != nil {
		return fmt.Errorf("Error creating source file: %w", err)
	}
	c.dataCfg.AddSource(t, r.Title(), r.Author(), r.Id(), path)
	return c.dataCfg.Save(c.paths.DataConfigPath)
}

// Initialize text source for playback and set it as current
// source in the controller.
func (c *NarrativeCtrl) LoadSource(id string) error {
	if s, ok := c.dataCfg.Sources[id]; !ok {
		return fmt.Errorf("Text source with id %s not found", id)
	} else {
		var sentence_n int
		if sn, ok := c.dataCfg.LastSentence[id]; ok {
			sentence_n = sn
		}
		ss, err := InitSource(
			s.Path, sentence_n,
			c.cfg.BufferSize, c.cfg.MaxBufferSize,
			c.preprocessor, c.phonemizer, c.ttsClient,
		)
		if err != nil {
			return fmt.Errorf("Error initializing source @ %s: %w", s.Path, err)
		}
		c.currentSource = ss
	}
	return nil
}

// Process command line arguments.
func (c *NarrativeCtrl) handleArguments() (bool, error) {
	if len(c.args.Source) > 0 {
		return false, c.addNewSource()
	}
	return false, nil
}

// Switch text source in the model
func (c *NarrativeCtrl) selectSource(id string) error {
	err := c.LoadSource(id)
	if err != nil {
		return err
	}
	c.initSourcePlayback()
	return nil
}

// Initialize source data for playback, this will buffer
// audio data before playing anything.
func (c *NarrativeCtrl) initSourcePlayback() {
	if c.currentSource == nil {
		return
	}

	PSM.SetStatus(playbackBuffering)
	go func() {
		c.program.Send(SetSourceCmd{source: c.currentSource})
		go c.currentSource.updateCacheBufferCb(func() {
			go c.program.Send(StartPlaybackCmd{})
		})
	}()
}

// Resume paused playback
func (c *NarrativeCtrl) resume() {
	c.player.Resume()
}

// Play current sentence of the source loaded in the controller.
func (c *NarrativeCtrl) play(callback func()) {
	if c.currentSource == nil {
		return
	}
	data, err := c.currentSource.getCurrentSentenceWaveformData()
	if err != nil {
		// TODO
		panic(err)
	}
	c.player.AddSample(data)
	c.player.Play(callback)
}

// Pause playback.
func (c *NarrativeCtrl) pause() {
	c.player.Pause()
}

func (c *NarrativeCtrl) stop() {
	PSM.SetStatus(playbackIdle)
	c.player.Stop()
}

// Run the model.
func (c *NarrativeCtrl) Run() error {
	exit, err := c.handleArguments()
	if exit || err != nil {
		return err
	}
	c.program = tea.NewProgram(c.model)

	if len(c.dataCfg.Sources) > 0 {
		var s_id string
		if len(c.dataCfg.LastSource) > 0 {
			s_id = c.dataCfg.LastSource
		} else if len(c.dataCfg.Sources) > 0 {
			s_id = c.dataCfg.Sources.DateOrdered()[0].Id
		}
		if err := c.LoadSource(s_id); err != nil {
			return err
		}
		c.model.initSource(c.currentSource)
	}

	if _, err := c.program.Run(); err != nil {
		return err
	}

	return nil
}

// Cleanly close the app.
func (c *NarrativeCtrl) Deinit() {
	defer c.cfg.Save(c.paths.ConfigPath)
	defer c.dataCfg.Save(c.paths.DataConfigPath)
	defer c.ttsClient.Deinit()
	defer c.program.Quit()
}

package narrative

import (
	"fmt"
	"os"
	"slices"

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
	remote        *Remote

	// Source to be automatically played on startup.
	autoplaySource string
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

	go startSpinner("Preparing Narrative...")
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
	SpinnerMessageCh <- "Initializing text source.."
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
	c.autoplaySource = r.Id()
	return c.dataCfg.Save(c.paths.DataConfigPath)
}

// Process command line arguments.
func (c *NarrativeCtrl) handleArguments() (bool, error) {
	if len(c.args.Source) > 0 {
		return false, c.addNewSource()
	}
	return false, nil
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
		if ss.id != dummySourceId {
			c.dataCfg.LastSource = ss.id
		}
		c.currentSource = ss
		c.remote = NewRemote(c.currentSource, c.player)
	}
	return nil
}

// Switch text source in the model
func (c *NarrativeCtrl) selectSource(id string) error {
	if !PSM.AllowsSourceSwitching() {
		return nil
	}
	if c.currentSource != nil {
		c.remote.Stop()
		c.dataCfg.LastSentence[c.currentSource.id] = c.currentSource.SNum()
	}
	err := c.LoadSource(id)
	if err != nil {
		return err
	}
	// TODO: save last source sentence
	c.initSourcePlayback()
	return nil
}

// Delete source with given id.
func (c *NarrativeCtrl) deleteSource(id string) error {

	if _, ok := c.dataCfg.Sources[id]; !ok {
		return fmt.Errorf("Text source with id %s not found", id)
	}
	s := c.dataCfg.Sources[id]
	if c.currentSource.id == id {
		c.remote.Stop()
		next := c.dataCfg.Sources.Next(id)
		if next != nil {
			c.LoadSource(next.Id)
		} else {
			c.loadDummySource()
		}
	}
	err := os.Remove(s.Path)
	c.dataCfg.DeleteSource(id)
	c.dataCfg.Save(c.paths.DataConfigPath)
	go c.program.Send(UpdateSourceCmd{source: c.currentSource})
	go c.program.Send(UpdateSourceListCmd{items: c.dataCfg.Sources.ListItems()})
	return err
}

// Initialize source data for playback, this will buffer
// audio data before playing anything.
func (c *NarrativeCtrl) initSourcePlayback() {
	if c.currentSource == nil {
		return
	}
	go func() {
		PSM.SetStatus(playbackBuffering)
		MessageCh <- "Buffering, please wait..."
		c.program.Send(SetSourceCmd{source: c.currentSource})
		go c.currentSource.UpdateCacheBuffer(func() {
			c.remote.StartPlayback()
			go c.program.Send(RemoveModalCmd{})
		})
	}()
}

// Run the model.
func (c *NarrativeCtrl) Run() error {
	exit, err := c.handleArguments()
	if exit || err != nil {
		return err
	}

	go func() {
		SpinnerCloseCh <- struct{}{}
	}()
	// Wait for spinner to be closed before continuing.
	SpinnerWaitCh <- struct{}{}

	c.program = tea.NewProgram(c.model)

	if len(c.dataCfg.Sources) > 0 {
		var s_id string
		if len(c.autoplaySource) > 0 {
			s_id = c.autoplaySource
		} else if len(c.dataCfg.LastSource) > 0 {
			s_id = c.dataCfg.LastSource
		} else if len(c.dataCfg.Sources) > 0 {
			s_id = c.dataCfg.Sources.DateOrdered()[0].Id
		}
		if err := c.LoadSource(s_id); err != nil {
			return err
		}
	} else {
		c.loadDummySource()
		go func() {
			MessageCh <- "Pro Tip: Add some text sources first..."
		}()
	}
	c.model.initSource(c.currentSource)
	// In case any sources have been added remove via args.
	go c.program.Send(UpdateSourceListCmd{items: c.dataCfg.Sources.ListItems()})
	if len(c.dataCfg.LastSource) > 0 || len(c.autoplaySource) > 0 {
		var s_id string
		var autoplay bool
		if len(c.autoplaySource) > 0 {
			s_id = c.autoplaySource
			autoplay = true
		} else {
			s_id = c.dataCfg.LastSource
		}
		idx := slices.Index(c.dataCfg.Sources.Ids(), s_id)
		go c.program.Send(SelectSourceCmd{
			id:       s_id,
			autoplay: autoplay,
			index:    idx,
		})
	}
	if _, err := c.program.Run(); err != nil {
		return err
	}

	return nil
}

// Load fake source, in cases where there is
// no actual source to display.
func (c *NarrativeCtrl) loadDummySource() {
	c.currentSource = GetDummySource(
		c.preprocessor,
		c.phonemizer,
		c.ttsClient,
	)
	c.remote = NewRemote(c.currentSource, c.player)
}

// Cleanly close the app.
func (c *NarrativeCtrl) Deinit() {
	defer c.cfg.Save(c.paths.ConfigPath)
	defer c.dataCfg.Save(c.paths.DataConfigPath)
	defer c.ttsClient.Deinit()
	if c.program != nil {
		defer c.program.Quit()
	}
}

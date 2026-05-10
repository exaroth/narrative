package narrative

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/exaroth/narrative/internal/common"
	"github.com/exaroth/narrative/internal/debugger"
	"github.com/exaroth/narrative/pkg/kitten"
	"github.com/exaroth/narrative/pkg/phonemizer"
	"github.com/exaroth/narrative/pkg/player"
	"github.com/exaroth/narrative/pkg/preprocessor"

	tea "charm.land/bubbletea/v2"
)

type NarrativeCtrl struct {
	phonemizer    *phonemizer.Phonemizer
	preprocessor  *preprocessor.Preprocessor
	ttsClient     *kitten.Kitten
	player        *player.Player
	cfg           *common.Config
	dataCfg       *common.DataConfig
	paths         *common.NarrativePaths
	model         *narrativeModel
	args          *NarrativeArgs
	currentSource *Source
	program       *tea.Program
	remote        *Remote

	// Source to be automatically played on startup.
	autoplaySource string
	// Whether to run debugger after app deinit.
	runDebugger bool
}

// Get currently loaded source.
func (c NarrativeCtrl) Source() *Source {
	return c.currentSource
}

// Get currently used voice name.
func (c NarrativeCtrl) Voice() string {
	return c.cfg.Voice
}

// Initialize new narrative controller.
func NewCtrl(args *NarrativeArgs) (ctrl *NarrativeCtrl, err error) {
	var model_n, lib_n string
	var ph *phonemizer.Phonemizer
	var preproc *preprocessor.Preprocessor

	paths := common.InitDirectoryStructure()

	startSpinner("Preparing Narrative...")

	// if any errors occured close spinner gracefully.
	defer func() {
		if err != nil {
			CloseSpinner()
		}
	}()

	if paths.RequiresInit() || args.Init {
		CloseSpinner()
		model_n, lib_n, err = ShowWelcomeScreen(paths, args.Init)
		if err != nil {
			return nil, err
		}
	}

	ph, err = phonemizer.NewPhonemizer("")
	if err != nil {
		return nil, fmt.Errorf("init err; phonemizer init: %w", err)
	}

	preproc = preprocessor.NewPreprocessor()
	var cfg *common.Config
	if _, err = os.Stat(paths.ConfigPath); err != nil {
		cfg = common.DefaultConfig()
		if err = cfg.Save(paths.ConfigPath); err != nil {
			return nil, fmt.Errorf("Error saving config @ %s; %w",
				paths.ConfigPath, err,
			)
		}
	} else {
		cfg, err = common.LoadConfig(paths.ConfigPath)
		if err != nil {
			return nil, fmt.Errorf("Error loading config @ %s, %w",
				paths.ConfigPath, err,
			)
		}
	}
	var data_cfg *common.DataConfig
	data_cfg, err = common.LoadDataConfig(paths.DataConfigPath)
	if err != nil {
		if errors.Is(err, common.MissingLibErr) || errors.Is(err, common.MissingModelErr) {
			return nil, err
		}
		data_cfg = common.NewDataConfig(model_n, lib_n)
		if err = data_cfg.Save(paths.DataConfigPath); err != nil {
			return nil, fmt.Errorf("Error creating data cfg: %w", err)
		}
	}

	ctrl = &NarrativeCtrl{
		phonemizer:   ph,
		preprocessor: preproc,
		player:       player.InitPlayer(),
		cfg:          cfg,
		dataCfg:      data_cfg,
		paths:        paths,
		args:         args,
	}
	model := InitNarrativeModel(ctrl)
	ctrl.model = model
	return ctrl, nil
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
			c.preprocessor, c.phonemizer, c.ttsClient, s,
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

// Retrieve bookmarks for current source
func (c *NarrativeCtrl) GetBookmarks() []int {
	if c.currentSource.IsDummy() {
		return []int{}
	}
	return c.dataCfg.GetBookmarks(c.currentSource.id)
}

// Add bookmark for current source and sentence
func (c *NarrativeCtrl) AddBookmark() bool {
	if c.currentSource.IsDummy() {
		return false
	}
	res := c.dataCfg.AddBookmark(
		c.currentSource.id,
		c.currentSource.SNum(),
	)
	MessageCh <- "Bookmark added"
	return res
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
func (c *NarrativeCtrl) Run() (string, error) {
	exit, msg, err := c.handleArguments()
	if exit || err != nil {
		CloseSpinner()
		return msg, err
	}

	lib_path, err := c.dataCfg.GetLibPath(c.paths)
	if err != nil {
		CloseSpinner()
		return "", err
	}

	c.ttsClient = kitten.InitKittenWithParams(lib_path,
		c.dataCfg.GetModelPath(c.paths),
		c.dataCfg.GetVoicesPath(c.paths),
		c.cfg.Voice,
		c.cfg.Speed,
	)
	CloseSpinner()

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
			return "", err
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
		return "", err
	}

	return "", nil
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
	fmt.Print("\033[s")
	if c.runDebugger {
		fmt.Println("Starting debugger")
	} else {
		fmt.Println("Closing Narrative...")
	}
	if c.currentSource.id != dummySourceId {
		c.dataCfg.LastSentence[c.currentSource.id] = c.currentSource.SNum()
	}
	defer c.cfg.Save(c.paths.ConfigPath)
	defer c.dataCfg.Save(c.paths.DataConfigPath)
	// Wait for all inference to finish before quitting
	// to avoid panics.
	for {
		if len(CacheLock.Items()) == 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if c.runDebugger {
		defer c.startDebugger()
	}
	defer fmt.Print("\033[u\033[K")
	if c.program != nil {
		defer c.program.Quit()
	}
	defer c.ttsClient.Deinit()
}

// Start debugger passing current source.
func (c *NarrativeCtrl) startDebugger() {
	if c.Source() == nil || c.Source().IsDummy() {
		return
	}

	debugger, err := debugger.InitWithData(c.Source().data, c.Source().SNum())
	if err != nil {
		fmt.Printf("Error initalizing debugger:\n%s\n", err.Error())
		return
	}
	debugger.Run()
}

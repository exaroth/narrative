package narrative

import (
	"fmt"
	"maps"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/exaroth/narrative/internal/common"
	"github.com/exaroth/narrative/pkg/kitten"
	"github.com/exaroth/narrative/pkg/reader"
	cp "github.com/otiai10/copy"
)

// Holds arguments used by Narrative.
type NarrativeArgs struct {
	Source      string  `arg:"positional"`
	ListVoices  bool    `arg:"--list-voices"`
	Voice       string  `arg:"-v,--voice"`
	SelectModel string  `arg:"-m,--select-model"`
	AddModel    string  `arg:"--add-model"`
	ListModels  bool    `arg:"--list-models"`
	AddLib      string  `arg:"--add-lib"`
	SelectLib   string  `arg:"-l,--select-lib"`
	ListLibs    bool    `arg:"--list-libs"`
	Speed       float32 `arg:"-s,--speed"`
	Init        bool
	Help        bool
	Version     bool
}

// Parse command line arguments.
func ParseArgs() (*NarrativeArgs, error) {
	var args NarrativeArgs
	err := arg.Parse(&args)
	return &args, err
}

// Process command line arguments.
func (c *NarrativeCtrl) handleArguments() (bool, string, error) {
	if len(c.args.Source) > 0 {
		return false, "", c.addNewSource()
	}
	if len(c.args.Voice) > 0 {
		return c.updateVoice()
	}
	if c.args.ListVoices {
		return true, c.getVoiceList(), nil
	}
	if c.args.ListModels {
		return true, c.getModelList(), nil
	}
	if len(c.args.AddModel) > 0 {
		msg, err := c.addModel(c.args.AddModel)
		return true, msg, err
	}
	if len(c.args.SelectModel) > 0 {
		return false, "", c.selectModel(c.args.SelectModel)
	}
	if c.args.ListLibs {
		return true, c.getLibList(), nil
	}
	if len(c.args.AddLib) > 0 {
		msg, err := c.addLib(c.args.AddLib)
		return true, msg, err
	}
	if len(c.args.SelectLib) > 0 {
		return false, "", c.selectLib(c.args.SelectLib)
	}
	if c.args.Speed > 0 {
		c.cfg.Speed = c.args.Speed
		return false, "", nil
	}
	return false, "", nil
}

// Add new text source based on the argument provided.
func (c *NarrativeCtrl) addNewSource() error {
	SpinnerMessageCh <- "Initializing text source.."
	_, t := common.GetSourceType(c.args.Source)
	var r reader.SourceReader
	var err error
	switch t {
	case common.SourceTypeText:
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

// Update voice to be used in playback.
func (c *NarrativeCtrl) updateVoice() (bool, string, error) {
	all_voices := slices.Collect(maps.Keys(kitten.VOICE_MAP))
	var voice string
	switch c.args.Voice {
	case "random":
		voice = all_voices[rand.Intn(len(all_voices)-1)]
	case "male":
		voice = kitten.MALE_VOICES[rand.Intn(len(kitten.MALE_VOICES)-1)]
	case "female":
		voice = kitten.FEMALE_VOICES[rand.Intn(len(kitten.FEMALE_VOICES)-1)]
	default:
		if slices.Index(all_voices, c.args.Voice) == -1 {
			return false, "", fmt.Errorf(
				"Invalid voice passed: %s, available voices: %s",
				c.args.Voice,
				strings.Join(all_voices, ", "),
			)
		}
		voice = c.args.Voice

	}
	c.cfg.Voice = voice
	return false, "", nil
}

// Get list of all the voices.
func (c *NarrativeCtrl) getVoiceList() string {
	return fmt.Sprintf(
		"Male voices: %s\nFemale voices: %s",
		strings.Join(kitten.MALE_VOICES, ", "),
		strings.Join(kitten.FEMALE_VOICES, ", "),
	)
}

// List available models.
func (c *NarrativeCtrl) getModelList() string {
	var builder strings.Builder
	builder.WriteString("Available models:\n")
	for _, m := range [3]common.TTSModel{
		common.TTSModelNano,
		common.TTSModelMini,
		common.TTSModelMicro,
	} {
		builder.WriteString("Name: " + m.Name + "\n")
		builder.WriteString("  Source: " + m.Remote + "\n")
		builder.WriteString("  Desc: " + m.Desc + "\n")
	}
	return builder.String()
}

// Download KittenTTS model passed in the argument.
func (c *NarrativeCtrl) addModel(n string) (string, error) {
	CloseSpinner()
	kt, err := common.KittenTypeFromString(n)
	if err != nil {
		return "", err
	}
	model := common.GetKittenModel(kt)
	model_p := filepath.Join(c.paths.ModelPath, model.Name)

	if _, err := os.Stat(model_p); err == nil {
		return "", fmt.Errorf("Model %s is already installed.", n)
	}

	defer os.RemoveAll(DEFAULT_DOWNLOAD_DIR)

	if err := DownloadAndQuit(
		0,
		model.Remote+"/"+model.Fname+"?download=true",
		fmt.Sprintf("Downloading model %s...", model.Name),
		model.Fname,
		30,
	); err != nil {
		return "", fmt.Errorf("Error downlaoding model: %w", err)
	}
	if err := DownloadAndQuit(
		1,
		model.Remote+"/"+common.VOICES_FNAME+"?download=true",
		"Downloading voice data...",
		common.VOICES_FNAME,
		30,
	); err != nil {
		return "", fmt.Errorf("Error downloading voices %w", err)
	}

	if err := os.MkdirAll(model_p, os.ModePerm); err != nil {
		return "", err
	}

	if err := cp.Copy(
		filepath.Join(DEFAULT_DOWNLOAD_DIR, model.Fname),
		filepath.Join(model_p, model.Fname),
	); err != nil {
		return "", fmt.Errorf("Error copying model: %w", err)
	}
	if err := cp.Copy(
		filepath.Join(DEFAULT_DOWNLOAD_DIR, common.VOICES_FNAME),
		filepath.Join(model_p, common.VOICES_FNAME),
	); err != nil {
		return "", fmt.Errorf("Error copying voices file: %w", err)
	}

	c.dataCfg.SelectedModel = model.Name
	if err := c.dataCfg.Save(c.paths.DataConfigPath); err != nil {
		return "", fmt.Errorf("Error saving data config, %w", err)
	}
	return fmt.Sprintf("Model %s added and selected as default.", n), nil
}

// Run narrative with selected model.
func (c *NarrativeCtrl) selectModel(n string) error {
	kt, err := common.KittenTypeFromString(n)
	if err != nil {
		return err
	}
	model := common.GetKittenModel(kt)
	model_p := filepath.Join(c.paths.ModelPath, model.Name)

	if _, err := os.Stat(model_p); err != nil {
		return fmt.Errorf("Model %s is not installed.", n)
	}
	c.dataCfg.SelectedModel = model.Name
	return nil
}

// List available libraries for current os/arch.
func (c *NarrativeCtrl) getLibList() string {
	l := common.GetOnnxLib(runtime.GOOS, runtime.GOARCH, "")
	if l == nil {
		return fmt.Sprintf(
			"Narrative is not available for %s/%s systems",
			runtime.GOOS, runtime.GOARCH,
		)
	}
	matches := common.OnnxLibMap[runtime.GOOS][runtime.GOARCH]
	var builder strings.Builder
	builder.WriteString("Available libraries:\n")
	for _, m := range matches {
		builder.WriteString("Name: " + m.Name + "\n")
		builder.WriteString("   Url: " + m.Remote + "\n")
		builder.WriteString("   OS: " + m.Os + "\n")
		builder.WriteString("   Arch: " + m.Arch + "\n")
		builder.WriteString("\n")
	}
	return builder.String()
}

// Add and select library with given name.
func (c *NarrativeCtrl) addLib(n string) (string, error) {
	l := common.GetOnnxLib(runtime.GOOS, runtime.GOARCH, n)
	if l == nil {
		return "", fmt.Errorf("ONNX library %s not found.", n)
	}

	lib_p := filepath.Join(c.paths.LibPath, l.Name)
	if _, err := os.Stat(lib_p); err == nil {
		return "", fmt.Errorf("Library %s is already installed", l.Name)
	}

	defer os.RemoveAll(DEFAULT_DOWNLOAD_DIR)

	if err := DownloadAndQuit(
		0,
		l.Remote,
		"Downloading ONNX library...",
		"lib.tgz",
		30,
	); err != nil {
		return "", fmt.Errorf("Error downloading library: %w", err)
	}

	fmt.Println("Extracting archive...")
	if err := ExtractTarGz(
		filepath.Join(DEFAULT_DOWNLOAD_DIR, "lib.tgz"),
		filepath.Join(DEFAULT_DOWNLOAD_DIR, "lib"),
	); err != nil {
		return "", fmt.Errorf("Error extracting archive: %w", err)
	}
	if err := cp.Copy(
		filepath.Join(DEFAULT_DOWNLOAD_DIR, "lib", l.TarPath),
		lib_p,
	); err != nil {
		return "", fmt.Errorf("Error copying library data: %w", err)
	}

	c.dataCfg.SelectedLib = l.Name
	if err := c.dataCfg.Save(c.paths.DataConfigPath); err != nil {
		return "", fmt.Errorf("Error saving data config, %w", err)
	}
	return fmt.Sprintf("Library %s added and selected.", l.Name), nil
}

// Run narrative with installed library.
func (c *NarrativeCtrl) selectLib(n string) error {
	l := common.GetOnnxLib(runtime.GOOS, runtime.GOARCH, n)
	if l == nil {
		return fmt.Errorf("ONNX library %s not found.", n)
	}

	lib_p := filepath.Join(c.paths.LibPath, l.Name)
	if _, err := os.Stat(lib_p); err != nil {
		return fmt.Errorf("Library %s is not installed", l.Name)
	}
	c.dataCfg.SelectedLib = l.Name
	return c.dataCfg.Save(c.paths.DataConfigPath)
}

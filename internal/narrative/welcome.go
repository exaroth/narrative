package narrative

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	cp "github.com/otiai10/copy"
)

const welcomeMessage = `Welcome to Narrative.
Narrative is terminal based TTS converter and player for ebooks, websites and other text sources.
Before continuing please select TTS model that best suits your needs and hardware.
[j/k - Select, Enter - Confirm, Ctrl+C - Exit]
`

type welcomeDownloadType int

const (
	welcomeDownloadTypeModel welcomeDownloadType = iota
	welcomeDownloadTypeVoice
	welcomeDownloadTypeLib
)

func getWelcomeScreenTTSModels() []list.Item {
	var v []list.Item
	tts_models := [3]*TTSModel{
		GetKittenModel(KittenModelNano),
		GetKittenModel(KittenModelMicro),
		GetKittenModel(KittenModelMini),
	}
	for _, i := range tts_models {
		v = append(v, i)
	}
	return v
}

// Model used for rendering welcome screen
// and initializing Narrative.
type welcomeModel struct {
	modelList list.Model
	downloads map[welcomeDownloadType]*Downloader
	width     int
	ctrl      *Welcome
}

func (m welcomeModel) Init() tea.Cmd {
	return nil
}

func (m welcomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.modelList.SetSize(msg.Width, msg.Height)

	case tea.KeyPressMsg:
		k := msg.String()
		if k == "ctrl+c" {
			m.ctrl.exit = true
			return m, tea.Quit
		}
		if k == "enter" {
			it := m.modelList.SelectedItem().(*TTSModel)
			m.ctrl.selection = it.T
			return m, tea.Quit
		}
	}
	m.modelList, cmd = m.modelList.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}
func (m *welcomeModel) centered(c string) string {
	return lipgloss.Place(
		m.width,
		lipgloss.Height(c),
		lipgloss.Center,
		lipgloss.Center,
		c,
	)
}

func (m *welcomeModel) selectionView() tea.View {
	var v tea.View
	var builder = strings.Builder{}
	builder.WriteString(welcomeScreenLogoStyle.Render(m.centered(LOGO)) + "\n")
	builder.WriteString(welcomeScreenMessageStyle.Render(m.centered(welcomeMessage)) + "\n")
	builder.WriteString(m.modelList.View())

	v.SetContent(builder.String())
	return v

}

func (m welcomeModel) View() tea.View {
	var v tea.View
	v = m.selectionView()
	v.AltScreen = true
	return v
}
func (m *welcomeModel) finalPause() tea.Cmd {
	return tea.Tick(time.Millisecond*750, func(_ time.Time) tea.Msg {
		return nil
	})
}

// Controller for handling welcome screen operations.
type Welcome struct {
	paths     *NarrativePaths
	model     *welcomeModel
	program   *tea.Program
	selection KittenModelType
	err       error
	exit      bool
}

func (w *Welcome) Run() error {
	program := tea.NewProgram(w.model)
	if _, err := program.Run(); err != nil {
		return err
	}
	return nil
}

type welcomeModelListDelegate struct{}

func (d welcomeModelListDelegate) Height() int                             { return 1 }
func (d welcomeModelListDelegate) Spacing() int                            { return 1 }
func (d welcomeModelListDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d welcomeModelListDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(*TTSModel)
	if !ok {
		return
	}

	var builder = strings.Builder{}
	builder.WriteString(
		lipgloss.Place(m.Width(), 1, lipgloss.Center, lipgloss.Center,
			welcomeScreenModelDescriptionStyle.Render(i.Desc),
		),
	)
	builder.WriteString("\n")

	if index == m.Index() {
		builder.WriteString(
			welcomeScreenModelNameSelectedStyle.Render(
				lipgloss.Place(m.Width(), 1, lipgloss.Center, lipgloss.Center,
					"> "+strings.ToUpper(i.Name)+" <"),
			),
		)
	} else {
		builder.WriteString(
			welcomeScreenModelNameStyle.Render(
				lipgloss.Place(m.Width(), 1, lipgloss.Center, lipgloss.Center,
					strings.ToUpper(i.Name)),
			),
		)
	}

	fmt.Fprint(w, builder.String())
}

func initWelcomeScreen(paths *NarrativePaths) (KittenModelType, error) {
	w := &Welcome{
		paths:     paths,
		selection: -1,
	}

	l := list.New(getWelcomeScreenTTSModels(), welcomeModelListDelegate{}, 0, 0)
	l.KeyMap = ModelListKeymap()
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.SetShowPagination(false)
	l.SetShowStatusBar(false)
	l.SetShowFilter(false)

	model := welcomeModel{
		modelList: l,
		ctrl:      w,
	}
	w.model = &model

	if err := w.Run(); err != nil {
		panic(err)
	}

	if w.err != nil {
		return -1, w.err
	}

	if w.exit {
		return -1, fmt.Errorf("Program terminated by user.")
	}

	return w.selection, nil
}

func ShowWelcomeScreen(paths *NarrativePaths) (model_n string, lib_n string, err error) {
	lib := GetOnnxLib(runtime.GOOS, runtime.GOARCH, "")
	if lib == nil {
		err = fmt.Errorf("Narrative is does not support %s/%s systems.",
			runtime.GOOS,
			runtime.GOARCH,
		)
		return
	}
	var m_data *TTSModel
	if selection, e := initWelcomeScreen(paths); e != nil {
		err = e
		return
	} else {
		m_data = GetKittenModel(selection)
	}

	defer os.RemoveAll(DEFAULT_DOWNLOAD_DIR)

	// Download files
	if e := DownloadAndQuit(
		int(welcomeDownloadTypeModel),
		m_data.Remote+"/"+m_data.Fname+"?download=true",
		fmt.Sprintf("Downloading model %s...", m_data.Name),
		m_data.Fname,
		30,
	); e != nil {
		err = fmt.Errorf("Error downlaoding model: %w", e)
		return
	}
	if e := DownloadAndQuit(
		int(welcomeDownloadTypeVoice),
		m_data.Remote+"/"+VOICES_FNAME+"?download=true",
		"Downloading voice data...",
		VOICES_FNAME,
		30,
	); e != nil {
		err = fmt.Errorf("Error downloading voices %w", e)
		return
	}
	if e := DownloadAndQuit(
		int(welcomeDownloadTypeLib),
		lib.remote,
		"Downloading ONNX library...",
		"lib.tgz",
		30,
	); e != nil {
		err = fmt.Errorf("Error downloading library: %w", e)
		return
	}

	fmt.Println("Extracting archive...")
	if e := ExtractTarGz(
		filepath.Join(DEFAULT_DOWNLOAD_DIR, "lib.tgz"),
		filepath.Join(DEFAULT_DOWNLOAD_DIR, "lib"),
	); e != nil {
		err = fmt.Errorf("Error extracting archive: %w", e)
		return
	}

	// Copy library files
	lib_p := filepath.Join(paths.LibPath, lib.name)
	os.RemoveAll(lib_p)
	if e := cp.Copy(
		filepath.Join(DEFAULT_DOWNLOAD_DIR, "lib", lib.tar_path),
		lib_p,
	); e != nil {
		err = fmt.Errorf("Error copying library data: %w", e)
		return
	}

	// Copy model files
	m_path := filepath.Join(paths.ModelPath, m_data.Name)
	os.RemoveAll(m_path)
	if e := os.MkdirAll(m_path, os.ModePerm); e != nil {
		err = fmt.Errorf("Error creating model dir %w", e)
		return
	}
	if e := cp.Copy(
		filepath.Join(DEFAULT_DOWNLOAD_DIR, m_data.Fname),
		filepath.Join(m_path, m_data.Fname),
	); e != nil {
		err = fmt.Errorf("Error copying model: %w", e)
		return
	}
	if e := cp.Copy(
		filepath.Join(DEFAULT_DOWNLOAD_DIR, VOICES_FNAME),
		filepath.Join(m_path, VOICES_FNAME),
	); e != nil {
		err = fmt.Errorf("Error copying voices file: %w", e)
		return
	}
	model_n = m_data.Name
	lib_n = lib.name
	os.Remove(paths.DataConfigPath)
	return

}

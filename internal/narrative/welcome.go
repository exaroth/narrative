package narrative

import (
	"fmt"
	"io"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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

type welcomeScreenMode int

const (
	welcomeScreenModeSelect welcomeScreenMode = iota
	welcomeScreenModeDownload
	welcomeScreenModeError
	welcomeScreenModeDone
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
	mode      welcomeScreenMode
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
	paths   *NarrativePaths
	model   *welcomeModel
	program *tea.Program
	err     error
	exit    bool
}

func (w *Welcome) Run() error {
	program := tea.NewProgram(w.model)
	if _, err := program.Run(); err != nil {
		return err
	}
	return nil
}

func ShowWelcomeScreen(paths *NarrativePaths) bool {
	w := &Welcome{
		paths: paths,
	}

	l := list.New(getWelcomeScreenTTSModels(), welcomeModelListDelegate{}, 0, 0)
	l.KeyMap = ModelListKeymap()
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.SetShowPagination(false)
	l.SetShowStatusBar(false)
	l.SetShowFilter(false)

	model := welcomeModel{
		mode:      welcomeScreenModeSelect,
		modelList: l,
		ctrl:      w,
	}
	w.model = &model

	if err := w.Run(); err != nil {
		panic(err)
	}

	return w.exit
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
			welcomeScreenModelDescriptionStyle.Render(i.desc),
		),
	)
	builder.WriteString("\n")

	if index == m.Index() {
		builder.WriteString(
			welcomeScreenModelNameSelectedStyle.Render(
				lipgloss.Place(m.Width(), 1, lipgloss.Center, lipgloss.Center,
					"> "+strings.ToUpper(i.name)+" <"),
			),
		)
	} else {
		builder.WriteString(
			welcomeScreenModelNameStyle.Render(
				lipgloss.Place(m.Width(), 1, lipgloss.Center, lipgloss.Center,
					strings.ToUpper(i.name)),
			),
		)
	}

	fmt.Fprint(w, builder.String())
}

var welcomeScreenLogoStyle = lipgloss.NewStyle().Foreground(col.Color1)
var welcomeScreenMessageStyle = lipgloss.NewStyle().Margin(1, 0)
var welcomeScreenModelDescriptionStyle = lipgloss.NewStyle().Foreground(col.Color3)
var welcomeScreenModelNameStyle = lipgloss.NewStyle().Foreground(col.Color2)
var welcomeScreenModelNameSelectedStyle = lipgloss.NewStyle().
	Inherit(welcomeScreenModelNameStyle).
	Foreground(col.Color1)

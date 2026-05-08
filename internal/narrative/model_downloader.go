package narrative

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
)

const DEFAULT_DOWNLOAD_DIR = "/tmp/narrative"

type downloadProgressWriter struct {
	id         int
	total      int
	downloaded int
	file       *os.File
	reader     io.Reader
	onProgress func(float64)
	onError    func(error)
}

func (pw *downloadProgressWriter) Start() {
	_, err := io.Copy(pw.file, io.TeeReader(pw.reader, pw))
	if err != nil {
		pw.onError(err)
	}
}

func (pw *downloadProgressWriter) Write(p []byte) (int, error) {
	pw.downloaded += len(p)
	if pw.total > 0 && pw.onProgress != nil {
		pw.onProgress(float64(pw.downloaded) / float64(pw.total))
	}
	return len(p), nil
}

// Module used for downloading files such as model data
// and libraries.
type Downloader struct {
	updateCh       chan float64
	errCh          chan error
	id             int
	text           string
	url            string
	filename       string
	downloadDir    string
	err            error
	completed      float64
	pw             *downloadProgressWriter
	progress       progress.Model
	width          int
	exitOnComplete bool
}

// Initialize new Downloader.
func NewDownloader(id int, url, text, filename string, width int, exitOnComplete bool) (*Downloader, error) {
	if url == "" {
		return nil, fmt.Errorf("Url is required")
	}
	if filename == "" {
		return nil, fmt.Errorf("Filename is required")
	}

	return &Downloader{
		updateCh:       make(chan float64),
		errCh:          make(chan error),
		id:             id,
		url:            url,
		text:           text,
		filename:       filename,
		downloadDir:    DEFAULT_DOWNLOAD_DIR,
		width:          width,
		exitOnComplete: exitOnComplete,
	}, nil
}

// Initialize new downloader and download the file returning optional error if any occured.

func DownloadAndQuit(id int, url, text, filename string, width int) error {
	downloader, err := NewDownloader(id, url, text, filename, width, true)
	if err != nil {
		return fmt.Errorf("Error creating download %w", err)
	}

	if err := downloader.InitDownload(); err != nil {
		return fmt.Errorf("Error starting download %w", err)
	}

	program := tea.NewProgram(downloader)
	if d, err := program.Run(); err != nil {
		return err
	} else {
		return d.(Downloader).err
	}
}

func (d *Downloader) getResponse(url string) (*http.Response, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Got status code %d for url: %s", resp.StatusCode, url)
	}
	return resp, nil
}

func (d *Downloader) InitDownload() error {
	resp, err := d.getResponse(d.url)
	if err != nil {
		return err
	}
	if resp.ContentLength <= 0 {
		return fmt.Errorf("Invalid content length")
	}

	if err := os.MkdirAll(d.downloadDir, os.ModePerm); err != nil {
		return err
	}

	full_p := filepath.Join(d.downloadDir, d.filename)
	file, err := os.Create(full_p)
	if err != nil {
		return err
	}

	d.pw = &downloadProgressWriter{
		id:     d.id,
		total:  int(resp.ContentLength),
		file:   file,
		reader: resp.Body,
		onProgress: func(val float64) {
			if val == 100 {
				resp.Body.Close()
			}
			d.updateCh <- val
		},
		onError: func(err error) {
			resp.Body.Close()
			d.errCh <- err
		},
	}
	go d.pw.Start()

	return nil
}

func (d Downloader) Init() tea.Cmd {
	return tea.Batch(
		WaitForDownloadProgressMsg(d.updateCh),
		WaitForDownloadProgressErr(d.errCh),
	)
}

func (d Downloader) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case DownloadProgressMsg:
		d.completed = float64(msg)
		cmds = append(
			cmds,
			d.progress.SetPercent(float64(msg)),
			WaitForDownloadProgressMsg(d.updateCh),
		)
		if msg == 1 && d.exitOnComplete {
			return d, tea.Quit
		}
	case DownloadProgressErr:
		d.err = msg
		return d, tea.Quit
	case tea.WindowSizeMsg:
		var w = d.width
		if msg.Width < w {
			w = msg.Width
		}
		d.progress = progress.New(progress.WithColors(col.Color2), progress.WithWidth(w))
	}

	d.progress, cmd = d.progress.Update(msg)
	cmds = append(cmds, cmd)
	return d, tea.Batch(cmds...)
}

func (d Downloader) View() tea.View {
	var v tea.View
	var builder strings.Builder

	if d.err != nil {
		builder.WriteString(errStyle.Render("Error: ") + d.err.Error() + "\n")
	} else {
		if len(d.text) > 0 {
			builder.WriteString(d.text + "\n")
		} else {
			builder.WriteString("\n")
		}
	}
	builder.WriteString(d.progress.View())
	v.SetContent(builder.String())
	return v
}

func (d Downloader) Error() error {
	return d.err
}

func (d Downloader) Completed() bool {
	return d.completed == 100
}

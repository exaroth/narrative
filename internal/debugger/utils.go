package debugger

import (
	"encoding/csv"
	"io"
	"maps"
	"os"
	"slices"
	"sync"

	"charm.land/lipgloss/v2"
)

type CsvDict struct {
	path string
	data map[string]string
	mut  *sync.Mutex
}

func (d *CsvDict) Update(word, phoneme string) {
	d.mut.Lock()
	defer d.mut.Unlock()

	d.data[word] = phoneme
}

func (d *CsvDict) Get(word string) *string {
	d.mut.Lock()
	defer d.mut.Unlock()

	if v, ok := d.data[word]; ok {
		return &v
	}
	return nil
}

// Save new dict in sorted order.
func (d *CsvDict) Save() error {
	var err error

	d.mut.Lock()
	defer d.mut.Unlock()

	f_reader, err := os.OpenFile(d.path, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer f_reader.Close()
	os.Remove(d.path + ".bkp")

	if bkp, err := os.Create(d.path + ".bkp"); err == nil {
		defer bkp.Close()
		if _, err := io.Copy(bkp, f_reader); err != nil {
			return err
		}
	} else {
		return err
	}

	if err := f_reader.Truncate(0); err == nil {
		if _, err := f_reader.Seek(0, 0); err != nil {
			return err
		}
	} else {
		return err
	}

	var writer = csv.NewWriter(f_reader)
	writer.Comma = ' '

	keys := slices.Sorted((maps.Keys(d.data)))
	rows := [][]string{}
	for _, k := range keys {
		row := []string{k, d.data[k]}
		rows = append(rows, row)
	}
	err = writer.WriteAll(rows)
	if err != nil {
		return err
	}
	return nil
}

func LoadCsvDict(path string) (*CsvDict, error) {

	f_reader, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f_reader.Close()

	var reader = csv.NewReader(f_reader)
	reader.Comma = ' '

	recs, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	data := make(map[string]string)

	for _, rec := range recs {
		data[rec[0]] = rec[1]
	}
	return &CsvDict{
		path: path,
		data: data,
		mut:  &sync.Mutex{},
	}, nil
}

func getStatusBar(contents string, width int) string {

	w := lipgloss.Width
	title := statusStyle.Render("Debugger")
	crumbs := statusBarCrumbsStyle.Render("v0.1")
	help := statusBarStatusText.
		Width(width - w(title) - w(crumbs)).
		Render(contents)

	bar := lipgloss.JoinHorizontal(lipgloss.Top,
		title,
		help,
		crumbs,
	)

	return lipgloss.Sprint(statusBarStyle.Width(width).Render(bar))
}

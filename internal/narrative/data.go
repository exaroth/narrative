package narrative

import (
	"encoding/json"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
)

type SourceType int

const (
	SourceTypeText SourceType = iota
	SourceTypeEpub
	SourceTypeMobi
	SourceTypeAzw3
	SourceTypeUrl
	SourceTypeMarkdown
)

var SourceTypeName = map[SourceType]string{
	SourceTypeText:     "text",
	SourceTypeEpub:     "epub",
	SourceTypeMobi:     "mobi",
	SourceTypeAzw3:     "azw3",
	SourceTypeUrl:      "url",
	SourceTypeMarkdown: "markdown",
}

func (t SourceType) String() string { return SourceTypeName[t] }

// Retrieve source type and indicator whether data is remote
// and should be downloaded before processing.
// TODO: currently remote files are not supported.
func GetSourceType(input string) (remote bool, t SourceType) {
	ext := filepath.Ext(strings.ToLower(input))
	switch ext {
	case "txt":
		return false, SourceTypeText
	case "epub":
		return false, SourceTypeEpub
	case "mobi":
		return false, SourceTypeMobi
	case "md", "mkd", "markdown":
		return false, SourceTypeMarkdown
	case "azw3":
		return false, SourceTypeAzw3
	default:
		return false, SourceTypeText
	}
}

// Contains information about given text source.
type TextSource struct {
	Id, Title, Author, Path string
	SourceType              SourceType
	Added                   int64
}

func (t TextSource) FilterValue() string { return t.Title }

type Sources map[string]*TextSource

// Return source list ordered by date.
func (s Sources) DateOrdered() []*TextSource {
	return slices.SortedFunc(maps.Values(s), func(s1, s2 *TextSource) int {
		return int(s2.Added - s1.Added)
	})
}

// Return list of source ids, ordered by date.
func (s Sources) Ids() []string {
	var result []string
	for _, v := range s.DateOrdered() {
		result = append(result, v.Id)
	}
	return result
}

// Return text sources as bubbletea compatible list.
func (s Sources) ListItems() []list.Item {
	var result []list.Item
	for _, i := range s.DateOrdered() {
		result = append(result, i)
	}
	return result
}

// Retrieve source next to source with id provided,
// in order they are displayed.
// If there's only one source or there are no sources
// will return nil.
// If source is last in order will return first source.
func (s Sources) Next(id string) *TextSource {
	if len(s) == 0 || len(s) == 1 {
		return nil
	}
	ord := s.DateOrdered()
	for i, s := range ord {
		if s.Id == id {
			if i == len(ord)-1 {
				return ord[0]
			}
			return ord[i+1]
		}
	}
	// not found
	return nil
}

// DataConfig stores information about data
// managed by narrative, such us text sources and models.
type DataConfig struct {
	Models        map[string]TTSModel
	Libs          []string
	SelectedModel string
	SelectedLib   string
	Bookmarks     map[string][]string
	LastSource    string
	LastSentence  map[string]int
	Sources       Sources
}

// Add new source of text.
func (c *DataConfig) AddSource(source_type SourceType, title, author, id, path string) {
	t := &TextSource{
		Id:         id,
		Title:      title,
		Author:     author,
		SourceType: source_type,
		Path:       path,
		Added:      time.Now().Unix(),
	}
	c.Sources[id] = t
}

// Delete source with given id.
func (c *DataConfig) DeleteSource(id string) {
	delete(c.Sources, id)
	delete(c.LastSentence, id)
	delete(c.Bookmarks, id)
	if c.LastSource == id {
		c.LastSource = ""
	}
}

// Save data config as json file.
func (c *DataConfig) Save(path string) error {
	m, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(path, m, 0644)
}

// Initialize new empty data config.
func NewDataConfig() *DataConfig {
	return &DataConfig{
		Models: map[string]TTSModel{
			TTSModelNano.name:  TTSModelNano,
			TTSModelMicro.name: TTSModelMicro,
			TTSModelMini.name:  TTSModelMini,
		},
		Libs:          []string{},
		SelectedModel: "",
		SelectedLib:   "",
		LastSource:    "",
		LastSentence:  make(map[string]int),
		Bookmarks:     make(map[string][]string),
		Sources:       make(map[string]*TextSource),
	}
}

// Load data config from json file.
func LoadDataConfig(path string) (*DataConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	bytes, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var cfg DataConfig
	err = json.Unmarshal(bytes, &cfg)
	return &cfg, err
}

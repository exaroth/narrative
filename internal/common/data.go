package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	"github.com/go-softwarelab/common/pkg/seq"
)

type SourceType int

const (
	SourceTypeText SourceType = iota
	SourceTypeEpub
	SourceTypeMobi
	SourceTypeAzw3
	SourceTypeHTML
	SourceTypeMarkdown
)

var SourceTypeName = map[SourceType]string{
	SourceTypeText:     "text",
	SourceTypeEpub:     "epub",
	SourceTypeMobi:     "mobi",
	SourceTypeAzw3:     "azw3",
	SourceTypeHTML:     "html",
	SourceTypeMarkdown: "markdown",
}

var (
	MissingLibErr = errors.New(`Library is missing, fix it by running:
narrative --init
	or
narrative --add-lib <library-name>`)
	MissingModelErr = errors.New(`KittenTTS model is missing, fix it by running:
narrative --init
	or
narrative --add-model <model-name>`)
)

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
	case "html":
		return false, SourceTypeHTML
	default:
		return false, SourceTypeText
	}
}

// Contains information about given text source.
type TextSource struct {
	Id, Title, Author, Path string
	SourceType              SourceType
	Added                   int64
	Playing                 bool
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
	SelectedModel string
	SelectedLib   string
	Bookmarks     map[string][]int
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

// Add new bookmark for given source.
func (c *DataConfig) AddBookmark(id string, sn int) []int {
	if s, ok := c.Bookmarks[id]; ok {
		if slices.Index(s, sn) == -1 {
			cb := make([]int, len(c.Bookmarks[id]))
			copy(cb, c.Bookmarks[id])
			cb = append(cb, sn)
			c.Bookmarks[id] = slices.Sorted(seq.Of(cb...))
		}
		return c.Bookmarks[id]
	}
	c.Bookmarks[id] = []int{sn}
	return c.Bookmarks[id]
}

// Delete bookmark, passing sentence number,
// bookmark with minimum distance to sentence n
// will be deleted.
func (c *DataConfig) DeleteBookmark(id string, sn int) []int {
	if _, ok := c.Bookmarks[id]; !ok {
		return []int{}
	}
	bookmarks := c.Bookmarks[id]
	if len(bookmarks) == 0 {
		return []int{}
	}
	if len(bookmarks) == 1 {
		c.Bookmarks[id] = []int{}
		return c.Bookmarks[id]
	}
	if slices.Index(bookmarks, sn) > -1 {
		updated := slices.DeleteFunc(bookmarks, func(s int) bool {
			return s == sn
		})
		c.Bookmarks[id] = updated
		return updated
	}
	to_del := c.getClosestBookmark(id, sn)
	// sanity check
	if to_del == -1 {
		return c.Bookmarks[id]
	}

	updated := slices.DeleteFunc(bookmarks, func(s int) bool {
		return s == to_del
	})
	c.Bookmarks[id] = updated
	return c.Bookmarks[id]
}

// Retrieve all bookmarks associated with given source.
func (c *DataConfig) GetBookmarks(id string) []int {
	if s, ok := c.Bookmarks[id]; ok {
		return s
	}
	return []int{}
}

// Get bookmark closes to that of the passed cursor.
func (c *DataConfig) getClosestBookmark(id string, sn int) int {

	if _, ok := c.Bookmarks[id]; !ok {
		return -1
	}

	if len(c.Bookmarks[id]) == 0 {
		return -1
	}
	if len(c.Bookmarks[id]) == 1 {
		return c.Bookmarks[id][0]
	}

	cb := make([]int, len(c.Bookmarks[id]))
	copy(cb, c.Bookmarks[id])
	cb = append(cb, sn)
	cb = slices.Sorted(seq.Of(cb...))

	sel := -1
	s_i := slices.Index(cb, sn)
	switch s_i {
	case 0:
		sel = cb[1]
	case len(cb) - 1:
		sel = cb[len(cb)-2]
	default:
		prev, next := cb[s_i-1], cb[s_i+1]
		if sn-prev < next-sn {
			sel = prev
		} else {
			sel = next
		}
	}
	return sel
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

// Retrieve full path to the onnx library file.
func (c *DataConfig) GetLibPath(paths *NarrativePaths) (string, error) {
	lib := GetOnnxLib(runtime.GOOS, runtime.GOARCH, c.SelectedLib)
	if lib == nil {
		return "", fmt.Errorf("Could not find library %s", c.SelectedLib)
	}
	return filepath.Join(paths.LibPath, lib.Name, lib.Filename), nil
}

// Get currently selected model config.
func (c *DataConfig) getModel() TTSModel {
	model, ok := c.Models[c.SelectedModel]
	if !ok {
		// Should never happen
		panic(fmt.Sprintf("Model '%s' could not be selected", c.SelectedModel))
	}
	return model
}

// Retrieve full path to the voices file.
func (c *DataConfig) GetVoicesPath(paths *NarrativePaths) string {
	model := c.getModel()
	return filepath.Join(paths.ModelPath, model.Name, VOICES_FNAME)
}

// Retrieve full path to the model library file.
func (c *DataConfig) GetModelPath(paths *NarrativePaths) string {
	model := c.getModel()
	return filepath.Join(paths.ModelPath, model.Name, model.Fname)
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
func NewDataConfig(model_n, lib_n string) *DataConfig {
	return &DataConfig{
		Models: map[string]TTSModel{
			TTSModelNano.Name:  TTSModelNano,
			TTSModelMicro.Name: TTSModelMicro,
			TTSModelMini.Name:  TTSModelMini,
		},
		SelectedModel: model_n,
		SelectedLib:   lib_n,
		LastSource:    "",
		LastSentence:  make(map[string]int),
		Bookmarks:     make(map[string][]int),
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
	if err != nil {
		return nil, err
	}
	if cfg.SelectedLib == "" {
		return nil, MissingLibErr
	}
	if cfg.SelectedModel == "" {
		return nil, MissingModelErr
	}
	return &cfg, err
}

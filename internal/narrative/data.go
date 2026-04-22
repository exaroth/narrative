package narrative

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
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

func (t SourceType) String() string {
	return SourceTypeName[t]
}

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
}

// DataConfig stores information about data
// managed by narrative, such us text sources and models.
type DataConfig struct {
	Models        []string
	Libs          []string
	SelectedModel string
	SelectedLib   string
	Sources       map[string]*TextSource
}

// Add new source of text.
func (c *DataConfig) AddSource(source_type SourceType, title, author, id, path string) {
	t := &TextSource{
		Id:         id,
		Title:      title,
		Author:     author,
		SourceType: source_type,
		Path:       path,
	}
	c.Sources[id] = t
}

// Delete source with given id.
func (c *DataConfig) DelSource(id string) {
	delete(c.Sources, id)
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
		Models:        []string{},
		Libs:          []string{},
		SelectedModel: "",
		SelectedLib:   "",
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

package reader

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/exaroth/narrative/pkg/reader/kindleunpack"
	"github.com/google/uuid"
)

const (
	PYTHON_C             = "python3"
	MOBI_UNPACK_DIR_NAME = "kindle-unpack"
	MOBI_OUTPUT_NAME     = "book"
	MOBI_TEMP_DIR_P      = "/tmp/narrative_mobi"
	MOBI_AUTHOR_PREFIX   = "Author:"
	MOBI_TITLE_PREFIX    = "Title:"
)

// Mobi reader implements functionality for reading
// pdb based books (mobi, azw, etc).
type MobiReader struct {
	data     []string
	t        SourceType
	id       string
	title    string
	author   string
	metadata string
	fpath    string
	updateCh chan<- string
}

// Initialize new mobi reader.
func NewMobiReader(t SourceType, update_ch chan<- string) *MobiReader {
	return &MobiReader{t: t, updateCh: update_ch}
}

// Read data and uncompress mobi/azw files before processing output
// files into text source.
func (r MobiReader) Read(source string) (SourceReader, error) {
	var err error
	data, err := os.ReadFile(source)
	if err != nil {
		return nil, fmt.Errorf("Unable to read source file, %w", err)
	}
	// Write data to dst
	if _, err := exec.LookPath(PYTHON_C); err != nil {
		return nil, fmt.Errorf("Python 3 must be installed for mobi/azw file conversions.")
	}

	defer os.RemoveAll(MOBI_TEMP_DIR_P)

	if err := kindleunpack.Unzip(MOBI_TEMP_DIR_P); err != nil {
		return nil, fmt.Errorf("Error decompressing kindle unpack lib.")
	}

	temp_f_p := filepath.Join(MOBI_TEMP_DIR_P, MOBI_OUTPUT_NAME+filepath.Ext(source))
	if err := os.WriteFile(temp_f_p, data, 0644); err != nil {
		return nil, fmt.Errorf("Error copying source file to %s, %w", temp_f_p, err)
	}

	output_p := filepath.Join(MOBI_TEMP_DIR_P, MOBI_OUTPUT_NAME)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command(
		PYTHON_C,
		filepath.Join(MOBI_TEMP_DIR_P, MOBI_UNPACK_DIR_NAME),
		temp_f_p,
		output_p,
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf(
			"Error unpacking ebook: %s, err: %s",
			stderr.String(),
			err,
		)
	}

	r.metadata = stdout.String()

	// Retrieve title/author from stdout.
	for _, l := range strings.Split(r.metadata, "\n") {
		if strings.HasPrefix(l, MOBI_AUTHOR_PREFIX) {
			r.author = strings.Trim(strings.TrimPrefix(l, MOBI_AUTHOR_PREFIX), " ")
		}
		if strings.HasPrefix(l, MOBI_TITLE_PREFIX) {
			r.title = strings.Trim(strings.TrimPrefix(l, MOBI_TITLE_PREFIX), " ")
		}
	}

	if r.title == "" {
		// should technically never happen for successfully converted files.
		r.title = source
	}

	r.id = uuid.New().String()

	// r.data = Sentencize([]byte(result))
	return &r, nil
}

// For html we return site name as author atm.
func (r MobiReader) Author() string {
	return r.author
}

func (r MobiReader) Title() string {
	return r.title
}

// Return raw text data.
func (r MobiReader) Data() []string {
	return r.data
}

// Return random uuid.
func (r MobiReader) Id() string {
	return r.id
}

func (r MobiReader) Chapters() []int {
	return []int{}
}

func (r MobiReader) Metadata() string {
	return r.metadata
}

func (r MobiReader) Type() SourceType {
	return SourceTypeHTML
}

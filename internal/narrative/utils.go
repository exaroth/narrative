package narrative

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
)

// Save text source as gob file.
func SaveTextSource(path, id string, data []byte) (string, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(data)
	if err != nil {
		return "", err
	}
	p := filepath.Join(path, fmt.Sprintf("%s.gob", id))
	return p, os.WriteFile(p, buf.Bytes(), 0644)
}

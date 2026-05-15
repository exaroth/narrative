package kindleunpack

import (
	"archive/zip"
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

//go:embed kindle-unpack.zip
var kindleUnpack []byte

// Read zip archive, caching the contents in the process.
var dataOnce = sync.OnceValue(func() *zip.Reader {
	r, err := zip.NewReader(bytes.NewReader(kindleUnpack), int64(len(kindleUnpack)))
	if err != nil {
		panic(fmt.Sprintf("Cannot read embedded archive: %s", err))
	}
	return r
})

// Return zip file as traversable fs.
func Data() fs.FS {
	return dataOnce()
}

// Traverse contents fo zip file returning file paths.
func getAllFilenames(efs fs.FS) (files []string, err error) {
	if err := fs.WalkDir(efs, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		files = append(files, path)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("Error reading zip archive: %w", err)
	}

	return files, nil
}

// Unzip contents of the zip file into target directory.
func Unzip(dest string) error {
	zipfile := Data()
	fnames, err := getAllFilenames(zipfile)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("Cant write to target directory %s: %w", dest, err)
	}

	extractAndWriteFile := func(zip_path string) error {
		f, err := zipfile.Open(zip_path)
		if err != nil {
			return fmt.Errorf("Error reading zip contents %s, %w", zip_path, err)
		}
		defer f.Close()

		os.MkdirAll(filepath.Join(dest, filepath.Dir(zip_path)), os.ModePerm)
		path := filepath.Join(dest, zip_path)

		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("Illegal file path: %s", path)
		}

		tf, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return fmt.Errorf("Error creating target for zip file %s, %w", zip_path, err)
		}
		defer tf.Close()

		_, err = io.Copy(tf, bufio.NewReader(f))
		if err != nil {
			return fmt.Errorf("Error copying file %s, %w", zip_path, err)
		}
		return nil
	}

	for _, f := range fnames {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

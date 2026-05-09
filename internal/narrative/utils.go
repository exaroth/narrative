package narrative

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Extract tar.gz archive into destination dir, unly supports
// relative symlinks.
func ExtractTarGz(archive, dest_dir string) error {
	if err := os.MkdirAll(dest_dir, os.ModePerm); err != nil {
		return fmt.Errorf("Error creating target dir %s, %w", dest_dir, err)
	}
	reader, err := os.Open(archive)
	if err != nil {
		return fmt.Errorf("Error opening archive %s, %w", archive, err)
	}
	uncompressedStream, err := gzip.NewReader(reader)
	if err != nil {
		return fmt.Errorf("Reading gzip failed: %w", err)
	}

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()

		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		path := filepath.Join(dest_dir, header.Name)

		switch header.Typeflag {

		case tar.TypeDir:
			if err := os.Mkdir(path, 0755); err != nil {
				return fmt.Errorf("Error creating directory %s: %w", path, err)
			}
		case tar.TypeReg:
			if err := func() error {
				outFile, err := os.Create(path)
				if err != nil {
					return fmt.Errorf("Error creating file: %w", err)
				}
				defer outFile.Close()
				if _, err := io.Copy(outFile, tarReader); err != nil {
					return fmt.Errorf("Error copying file: %w", err)
				}
				return nil
			}(); err != nil {
				return err
			}
		case tar.TypeSymlink:
			os.Symlink(header.Linkname, path)
			// default:
			// 	return fmt.Errorf("Unknown type found in %s: %s", header.Name, string(header.Typeflag))
		}
	}
	return nil
}

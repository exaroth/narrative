package narrative

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
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

// Attempt to infer filename from url string.
func InferFilenameFromUrl(url string) (string, error) {
	var ext, base string
	r, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return "", err
	}
	base = path.Base(r.URL.Path)
	ext = filepath.Ext(strings.ToLower(base))
	if ext != "" {
		return base, nil
	}
	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return "", err
	}
	if res.StatusCode != 200 {
		return "", fmt.Errorf("Invalid status code %d returned", res.StatusCode)
	}
	content_disp := res.Header.Get("Content-Disposition")
	if len(content_disp) > 0 {
		_, d_params, err := mime.ParseMediaType(content_disp)
		if err != nil {
			return "", fmt.Errorf("Error parsing content disposition header: %w", err)
		}
		if df, ok := d_params["filename"]; ok {
			return df, nil
		}
	}
	if res.Header.Get("Content-Type") != "" {
		c_t, _, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
		if err != nil {
			return "", fmt.Errorf("Error parsing Content-Type header: %w", err)
		}
		switch c_t {
		case "text/html":
			ext = ".html"
		case "text/plain":
			ext = ".txt"
		case "application/epub+zip", "application/epub":
			ext = ".epub"
		case "text/markdown":
			ext = ".md"
		case "application/x-mobipocket-ebook":
			ext = ".mobi"
		case "application/vnd.amazon.ebook":
			ext = ".azw3"
		}
	}

	if len(base) > 1 {
		return base + ext, nil
	}
	return r.URL.Host + ext, nil
}

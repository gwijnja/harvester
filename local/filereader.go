package local

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"

	"github.com/gwijnja/harvester"
)

// Reader reads a file from disk and presents it to the next processor in the chain.
type FileReader struct {
	ToLoad              string
	Loaded              string
	FollowSymlinks      bool
	DeleteAfterDownload bool
	Regex               string
	next                harvester.FileWriter
}

func (r *FileReader) SetNext(next harvester.FileWriter) {
	r.next = next
}

func (d *FileReader) List() ([]string, error) {
	slog.Info("Listing files", slog.String("toload", d.ToLoad))
	files, err := os.ReadDir(d.ToLoad)
	if err != nil {
		return nil, fmt.Errorf("failed to list files in %s: %s", d.ToLoad, err)
	}

	if d.Regex != "" {
		slog.Debug("Filtering files with regex", slog.String("regex", d.Regex))
	}

	filenames := make([]string, 0, len(files))
	for _, file := range files {

		// Skip directories
		if file.IsDir() {
			continue
		}

		// Skip symlinks if FollowSymlinks is false
		if file.Type()&fs.ModeSymlink != 0 && !d.FollowSymlinks {
			continue
		}

		// Skip files that do not match the regex
		if d.Regex != "" {
			matched, err := regexp.MatchString(d.Regex, file.Name())
			if err != nil {
				return nil, fmt.Errorf("invalid regex, error matching %s with %s: %s", d.Regex, file.Name(), err)
			}
			if !matched {
				slog.Warn("Skipping non-matching file", slog.String("filename", file.Name()))
				continue
			}
		}

		slog.Info("File found", slog.String("filename", file.Name()))

		filenames = append(filenames, file.Name())
	}
	return filenames, nil
}

func (r *FileReader) Process(filename string) error {
	from := filepath.Join(r.ToLoad, filename)
	slog.Debug("Opening", slog.String("path", from))
	f, err := os.Open(from)
	if err != nil {
		return fmt.Errorf("unable to open file %s: %s", from, err)
	}
	defer f.Close()

	// Call the next processor in the chain
	if err := r.next.Process(filename, f); err != nil {
		return err
	}

	// After the transfer has completed succesfully, either delete the file or move it
	if r.DeleteAfterDownload {
		slog.Info("Deleting", slog.String("path", from))
		if err := os.Remove(from); err != nil {
			return fmt.Errorf("unable to remove file %s: %s", from, err)
		}
		return nil
	}

	// Move the file from ToLoad to Loaded
	to := filepath.Join(r.Loaded, filename)
	slog.Info("Moving file", slog.String("from", from), slog.String("to", to))
	err = os.Rename(from, to)
	if err != nil {
		return fmt.Errorf("Reader.MoveFile(): unable to move %s to %s: %s", from, to, err)
	}
	return nil
}

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
	DeleteAfterDownload bool
	FollowSymlinks      bool
	Regex               string
	MaxFiles            int
	next                harvester.FileWriter
}

func (r *FileReader) SetNext(next harvester.FileWriter) {
	r.next = next
}

func (d *FileReader) List() ([]string, error) {

	// List files in the ToLoad directory
	files, err := os.ReadDir(d.ToLoad)
	if err != nil {
		return nil, fmt.Errorf("local: Failed to list files in %s: %s", d.ToLoad, err)
	}
	slog.Debug("local: Listed files in ToLoad", slog.String("path", d.ToLoad))

	if d.Regex != "" {
		slog.Debug("local: Filtering files with regex", slog.String("regex", d.Regex))
	}

	// Create a list of filenames
	filenames := make([]string, 0, len(files))
	for _, file := range files {

		// Skip directories
		if file.IsDir() {
			slog.Debug("local: Skipping directory", slog.String("filename", file.Name()))
			continue
		}

		// Skip symlinks if FollowSymlinks is false
		if file.Type()&fs.ModeSymlink != 0 && !d.FollowSymlinks {
			slog.Debug("local: Skipping symlink", slog.String("filename", file.Name()))
			continue
		}

		// Skip files that do not match the regex
		if d.Regex != "" {
			matched, err := regexp.MatchString(d.Regex, file.Name())
			if err != nil {
				return nil, fmt.Errorf("local: Failed to compile regex %s: %s", d.Regex, err)
			}
			if !matched {
				slog.Warn("local: Skipping non-matching file", slog.String("filename", file.Name()))
				continue
			}
		}

		// Add the filename to the list
		filenames = append(filenames, file.Name())
		slog.Info("local: Found file", slog.String("filename", file.Name()))

		// Stop if the maximum number of files has been reached
		if d.MaxFiles > 0 && len(filenames) >= d.MaxFiles {
			slog.Info("local: Reached maximum number of files", slog.Int("max_files", d.MaxFiles))
			break
		}
	}
	return filenames, nil
}

// Process reads a file from disk and presents it to the next processor in the chain.
func (r *FileReader) Process(filename string) error {

	// Open the file
	from := filepath.Join(r.ToLoad, filename)
	f, err := os.Open(from)
	if err != nil {
		return fmt.Errorf("local: Failed to open file %s: %s", from, err)
	}
	slog.Debug("local: Opened file", slog.String("path", from))

	defer func() {
		f.Close()
		slog.Info("local: Closed file", slog.String("path", from))
	}()

	// Call the next processor in the chain
	if err := r.next.Process(filename, f); err != nil {
		return err
	}

	// After the transfer has completed succesfully, either delete the file or move it
	if r.DeleteAfterDownload {
		if err := os.Remove(from); err != nil {
			return fmt.Errorf("local: Failed to remove file %s: %s", from, err)
		}
		slog.Info("local: Deleted file", slog.String("path", from))
		return nil
	}

	// Move the file from ToLoad to Loaded
	to := filepath.Join(r.Loaded, filename)
	err = os.Rename(from, to)
	if err != nil {
		return fmt.Errorf("local: Failed to move file %s to %s: %s", from, to, err)
	}
	slog.Info("local: Moved file", slog.String("from", from), slog.String("to", to))
	return nil
}

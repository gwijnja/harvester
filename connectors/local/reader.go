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
type Reader struct {
	harvester.BaseProcessor
	ToLoad              string
	Loaded              string
	FollowSymlinks      bool
	DeleteAfterDownload bool
	Regex               string
}

// Process reads a file from disk and presents it to the next processor in the chain.
func (d *Reader) Process(ctx *harvester.FileContext) error {

	path := filepath.Join(d.ToLoad, ctx.Filename)
	slog.Debug("Opening", slog.String("path", path))
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("Reader.Process(): unable to open file %s: %s", path, err)
	}
	defer f.Close()
	slog.Debug("Opened", slog.String("path", path))

	ctx.Reader = f
	// Remember the filename, because we need it for moving or deleting the file, and ctx.Filename may change during processing.
	origFilename := ctx.Filename

	slog.Debug("Calling the next processor")
	err = d.BaseProcessor.Process(ctx)
	if err != nil {
		return fmt.Errorf("Reader.Process(): error processing %s: %s", path, err)
	}

	// After the transfer has completed succesfully, either delete the file or move it
	if d.DeleteAfterDownload {
		slog.Info("Deleting", slog.String("path", path))
		err = os.Remove(path)
		if err != nil {
			return fmt.Errorf("Reader.Process(): the transfer was successful, but the source file (%s) could not be deleted: %s", path, err)
		}
		slog.Info("Deleted", slog.String("path", path))
		return nil
	}

	// Move the file from ToLoad to Loaded
	slog.Info("Moving", slog.String("path", path))
	err = d.MoveFile(origFilename)
	if err != nil {
		return fmt.Errorf("Reader.Process(): error moving %s: %s", ctx.Filename, err)
	}
	return nil
}

func (d *Reader) List() ([]string, error) {
	slog.Info("Listing files", slog.String("toload", d.ToLoad))
	files, err := os.ReadDir(d.ToLoad)
	if err != nil {
		return nil, fmt.Errorf("Reader.List(): unable to list files in %s: %s", d.ToLoad, err)
	}
	slog.Info("Listed files", slog.Int("count", len(files)))

	if d.Regex != "" {
		slog.Info("Filtering files with regex", slog.String("regex", d.Regex))
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
				return nil, fmt.Errorf("Reader.List(): the regex seems invalid, error matching %s with %s: %s", d.Regex, file.Name(), err)
			}
			if !matched {
				slog.Warn("Reader.List(): skipping non-matching file", slog.String("filename", file.Name()))
				continue
			}
		}

		slog.Info("Found file", slog.String("filename", file.Name()))

		filenames = append(filenames, file.Name())
	}
	return filenames, nil
}

func (d *Reader) MoveFile(filename string) error {
	from := filepath.Join(d.ToLoad, filename)
	to := filepath.Join(d.Loaded, filename)
	slog.Info("Moving", slog.String("from", from), slog.String("to", to))
	err := os.Rename(from, to)
	if err != nil {
		return fmt.Errorf("Reader.MoveFile(): unable to move %s to %s: %s", from, to, err)
	}
	slog.Info("Moved", slog.String("from", from), slog.String("to", to))
	return nil
}

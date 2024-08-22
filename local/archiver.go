package local

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gwijnja/harvester"
)

// Archiver is a connector that stores the file a directory with the
// date from the filename, if the filename matches a regex.
type Archiver struct {
	harvester.NextProcessor
	Transmit string
	Archive  string
	Regex    string // Example: "(\\d{4})-(\\d{2})-(\\d{2})"
	Format   string // Example: "$1/$2/$3"
}

// Process receives a file and writes it to the local filesystem.
func (a *Archiver) Process(filename string, r io.Reader) error {

	// Prepare the archive directory
	archiveDir, err := a.PrepArchiveDir(filename)
	if err != nil {
		return fmt.Errorf("unable to prepare archive directory: %s", err)
	}

	transmitPath := filepath.Join(a.Transmit, filename)
	slog.Info("Creating transmit file", slog.String("path", transmitPath))
	f, err := os.Create(transmitPath)
	if err != nil {
		return fmt.Errorf("unable to open transmit file %s: %s", transmitPath, err)
	}
	defer f.Close()
	slog.Info("File created", slog.String("path", transmitPath))

	tee := io.TeeReader(r, f)

	// Forward the reader to the next processor
	slog.Info("Calling NextProcessor.Process")
	err = a.NextProcessor.Process(filename, tee)
	if err != nil {
		slog.Error("tee process returned error, deleting transming file from archive.")
		os.Remove(transmitPath)
		return fmt.Errorf("error processing %s: %s", filename, err)
	}

	// Close the file
	f.Close()

	// Move it to the final archive directory
	archivePath := filepath.Join(archiveDir, filename)
	slog.Info("Moving", slog.String("from", transmitPath), slog.String("to", archivePath))
	err = os.Rename(transmitPath, archivePath)
	if err != nil {
		slog.Error("Error while moving", slog.String("from", transmitPath), slog.String("to", archivePath), slog.Any("error", err))
		return fmt.Errorf("error moving %s: %s", transmitPath, err)
	}

	return nil
}

// PrepArchiveDir creates the archive directory if it does not exist.
func (r *Archiver) PrepArchiveDir(filename string) (string, error) {
	subPath, err := r.MatchPath(filename)
	if err != nil {
		slog.Warn("Unable to match archive path to filename, file will be placed in archive root", slog.String("filename", filename))
		subPath = ""
	}

	fullPath := filepath.Join(r.Archive, subPath)
	slog.Info("Creating archive directory", slog.String("path", fullPath))
	err = os.MkdirAll(fullPath, 0755)
	if err != nil {
		return "", fmt.Errorf("unable to create archive directory %s: %s", fullPath, err)
	}

	return fullPath, nil
}

// MatchPath matches the filename with the regex and formats the archive path.
func (r *Archiver) MatchPath(filename string) (string, error) {

	slog.Debug("Compiling regex", slog.String("regex", r.Regex))
	re, err := regexp.Compile(r.Regex)
	if err != nil {
		return "", fmt.Errorf("unable to compile regex: %s", err)
	}

	slog.Debug("Matching regex", slog.String("filename", filename))
	matches := re.FindStringSubmatch(filename)
	if len(matches) == 0 {
		return "", fmt.Errorf("no matches found for regex: %s", r.Regex)
	}

	slog.Debug("Formatting archive path")
	path := r.Format
	for i, match := range matches {
		path = strings.Replace(path, fmt.Sprintf("$%d", i), match, -1)
	}

	slog.Info("Subpath formatted", slog.String("subpath", path))

	return path, nil
}

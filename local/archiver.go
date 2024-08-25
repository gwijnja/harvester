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

	// Prepare archive directory
	archiveDir, err := a.PrepArchiveDir(filename)
	if err != nil {
		return err
	}

	// Create the transmit file
	transmitPath := filepath.Join(a.Transmit, filename)
	f, err := os.Create(transmitPath)
	if err != nil {
		return fmt.Errorf("local: Failed to create transmit file %s: %s", transmitPath, err)
	}
	slog.Info("local: Created transmit file", slog.String("path", transmitPath))

	defer func() {
		f.Close()
		slog.Info("local: Closed transmit file", slog.String("path", transmitPath))
	}()

	tee := io.TeeReader(r, f)

	// Forward the reader to the next processor
	slog.Debug("Calling NextProcessor.Process")
	err = a.NextProcessor.Process(filename, tee)
	if err != nil {
		f.Close()
		slog.Info("local: Closed transmit file", slog.String("path", transmitPath))

		os.Remove(transmitPath)
		slog.Info("local: Removed transmit file", slog.String("path", transmitPath))
		return err
	}

	// Close the file
	f.Close()
	slog.Info("local: Closed transmit file", slog.String("path", transmitPath))

	// Move to archive directory
	archivePath := filepath.Join(archiveDir, filename)
	err = os.Rename(transmitPath, archivePath)
	if err != nil {
		return fmt.Errorf("local: Failed to move %s to %s: %s", transmitPath, archivePath, err)
	}
	slog.Info("Moved", slog.String("from", transmitPath), slog.String("to", archivePath))

	return nil
}

// PrepArchiveDir creates the archive directory if it does not exist.
func (r *Archiver) PrepArchiveDir(filename string) (string, error) {

	// Match the filename with the regex
	subPath, err := r.MatchPath(filename)
	if err != nil {
		slog.Warn("local: Failed to match filename, file will be placed in archive root", slog.String("filename", filename))
		subPath = ""
	}

	fullPath := filepath.Join(r.Archive, subPath)

	// If path already exists, return
	if _, err := os.Stat(fullPath); err == nil {
		return fullPath, nil
	}

	// Create the directory
	err = os.MkdirAll(fullPath, 0755)
	if err != nil {
		return "", fmt.Errorf("local: Failed to create archive directory %s: %s", fullPath, err)
	}
	slog.Info("local: Created archive directory", slog.String("path", fullPath))

	return fullPath, nil
}

// MatchPath matches the filename with the regex and formats the archive path.
func (r *Archiver) MatchPath(filename string) (string, error) {

	// Compile the regex
	re, err := regexp.Compile(r.Regex)
	if err != nil {
		return "", fmt.Errorf("local: Failed to compile regex: %s", err)
	}
	slog.Debug("local: Compiled regex", slog.String("regex", r.Regex))

	// Match the filename
	matches := re.FindStringSubmatch(filename)
	if len(matches) == 0 {
		return "", fmt.Errorf("local: Failed to match filename %s against regex %s", filename, r.Regex)
	}
	slog.Debug("local: Matched regex", slog.String("filename", filename), slog.Int("matches", len(matches)))

	// Format the archive path
	path := r.Format
	for i, match := range matches {
		path = strings.Replace(path, fmt.Sprintf("$%d", i), match, -1)
	}
	slog.Info("local: Formatted subpath", slog.String("subpath", path))

	return path, nil
}

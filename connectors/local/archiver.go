package local

import (
	"fmt"
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
func (r *Archiver) Process(ctx *harvester.FileContext) error {

	// Prepare the archive directory
	archiveDir, err := r.PrepArchiveDir(ctx.Filename)
	if err != nil {
		return fmt.Errorf("[local] Archiver.Process(): unable to prepare archive directory: %s", err)
	}

	transmitPath := r.Transmit + "/" + ctx.Filename
	slog.Info("Creating file", slog.String("path", transmitPath))

	f, err := os.Create(transmitPath)
	if err != nil {
		return fmt.Errorf("[local] Archiver.Process(): unable to open file %s: %s", transmitPath, err)
	}
	defer f.Close()
	slog.Info("File created", slog.String("path", transmitPath))

	// Copy the ctx.Reader to the file
	slog.Info("Calling AuditCopy")
	written, err := harvester.AuditCopy(f, ctx.Reader)
	if err != nil {
		// If the copy fails, close the file and delete it if something was created
		slog.Error("Error while copying", slog.String("path", transmitPath), slog.Any("error", err), slog.Int64("written", written))
		slog.Info("Closing and removing", slog.String("path", transmitPath))
		f.Close()
		os.Remove(transmitPath)
		return fmt.Errorf("[local] Archiver.Process(): error copying %s after %d bytes: %s", transmitPath, written, err)
	}
	slog.Info("Copy complete", slog.String("path", transmitPath), slog.Int64("written", written))

	// Move the file from Transmit to ToLoad
	archivePath := filepath.Join(archiveDir, ctx.Filename)
	slog.Info("Moving", slog.String("from", transmitPath), slog.String("to", archivePath))
	err = os.Rename(transmitPath, archivePath)
	if err != nil {
		// TODO: if the ToLoad directory does not exist, the error is wait too long and confusing.
		slog.Error("Error while moving", slog.String("from", transmitPath), slog.String("to", archivePath), slog.Any("error", err))
		slog.Info("Closing and removing", slog.String("path", transmitPath))
		f.Close()
		os.Remove(transmitPath)
		// TODO: dubbele errors, moet eigenlijk niet he...
		return fmt.Errorf("error moving %s: %s", transmitPath, err)
	}

	return nil
}

// PrepArchiveDir creates the archive directory if it does not exist.
func (r *Archiver) PrepArchiveDir(filename string) (string, error) {
	subPath, err := r.MatchPath(filename)
	if err != nil {
		return "", fmt.Errorf("unable to match path: %s", err)
	}

	fullPath := filepath.Join(r.Archive, subPath)
	slog.Info("Creating archive directory", slog.String("path", fullPath))
	err = os.MkdirAll(fullPath, 0755)
	if err != nil {
		return "", fmt.Errorf("unable to create archive directory %s: %s", fullPath, err)
	}

	return fullPath, nil
}

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

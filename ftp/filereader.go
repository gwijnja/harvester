package ftp

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"regexp"

	"github.com/gwijnja/harvester"
	"github.com/jlaffaye/ftp"
)

type FileReader struct {
	Connector
	ToLoad              string
	Loaded              string
	Regex               string
	MaxFiles            int // set to 0 for no limit
	DeleteAfterDownload bool
	next                harvester.FileWriter
}

func (c *FileReader) SetNext(next harvester.FileWriter) {
	c.next = next
}

func (c *FileReader) List() ([]string, error) {

	// Connect
	conn, err := c.connect()
	if err != nil {
		return nil, err
	}
	defer func() {
		slog.Debug("ftp: closing connection")
		conn.Quit()
	}()

	// List files
	slog.Info("Reader: listing files", slog.String("path", c.ToLoad))
	entries, err := conn.List(c.ToLoad)
	if err != nil {
		return nil, fmt.Errorf("reader: error listing files in %s: %s", c.ToLoad, err)
	}

	// Filter files
	return c.filterEntries(entries)
}

func (c *FileReader) Process(filename string) error {
	conn, err := c.connect()
	if err != nil {
		return err
	}
	defer func() {
		slog.Debug("ftp: closing connection")
		conn.Quit()
	}()

	// Set the transfer type to binary
	slog.Debug("Reader: setting transfer type to binary")
	err = conn.Type(ftp.TransferTypeBinary)
	if err != nil {
		return fmt.Errorf("reader: error setting transfer type to binary: %s", err)
	}

	toLoadPath := filepath.Join(c.ToLoad, filename)
	r, err := conn.Retr(toLoadPath)
	if err != nil {
		return fmt.Errorf("reader: error retrieving file %s: %s", toLoadPath, err)
	}
	defer r.Close() // just in case we exit early for some reason

	err = c.next.Process(filename, r)
	if err != nil {
		return fmt.Errorf("reader: error processing file %s: %s", filename, err)
	}
	r.Close() // close implicitly, because we're going to delete or move the file

	// Move the file from ToLoad to Loaded, or delete it
	if c.DeleteAfterDownload {
		slog.Info("Reader: deleting file", slog.String("path", toLoadPath))
		err = conn.Delete(toLoadPath)
		if err != nil {
			return fmt.Errorf("reader: error deleting file %s: %s", toLoadPath, err)
		}
		return nil
	}

	// Move the file from toLoad to loaded
	loadedPath := filepath.Join(c.Loaded, filename)
	slog.Info("Reader: renaming file", slog.String("from", toLoadPath), slog.String("to", loadedPath))
	err = conn.Rename(toLoadPath, loadedPath)
	if err != nil {
		return fmt.Errorf("reader: error renaming file %s to %s: %s", toLoadPath, loadedPath, err)
	}

	return nil
}

func (c *FileReader) filterEntries(entries []*ftp.Entry) ([]string, error) {

	filenames := []string{}

	for _, entry := range entries {
		slog.Debug("Entry", slog.String("name", entry.Name), slog.String("type", entry.Type.String()))
		if entry.Type != ftp.EntryTypeFile {
			continue
		}
		if c.Regex != "" {
			matched, err := regexp.MatchString(c.Regex, entry.Name)
			if err != nil {
				return nil, fmt.Errorf("invalid regex, error matching %s with %s: %s", c.Regex, entry.Name, err)
			}
			if !matched {
				slog.Warn("Skipping non-matching file", slog.String("filename", entry.Name))
				continue
			}
		}
		slog.Info("File found", slog.String("filename", entry.Name))
		filenames = append(filenames, entry.Name)
		if c.MaxFiles > 0 && len(filenames) >= c.MaxFiles {
			break
		}
	}

	return filenames, nil
}

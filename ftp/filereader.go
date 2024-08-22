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
	FtpConnector
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
	err := c.connect()
	if err != nil {
		return nil, err
	}
	defer func() {
		slog.Debug("Reader: quitting connection")
		c.connection.Quit()
	}()

	// List files
	slog.Info("Reader: listing files", slog.String("path", c.ToLoad))
	entries, err := c.connection.List(c.ToLoad)
	if err != nil {
		return nil, fmt.Errorf("reader: error listing files in %s: %s", c.ToLoad, err)
	}

	// Filter files
	return c.filterEntries(entries)
}

func (c *FileReader) Process(filename string) error {
	err := c.connect()
	if err != nil {
		return err
	}
	defer func() {
		slog.Debug("Reader: quitting connection")
		c.connection.Quit()
	}()

	// Set the transfer type to binary
	slog.Debug("Reader: setting transfer type to binary")
	err = c.connection.Type(ftp.TransferTypeBinary)
	if err != nil {
		return fmt.Errorf("reader: error setting transfer type to binary: %s", err)
	}

	path := filepath.Join(c.ToLoad, filename)
	r, err := c.connection.Retr(path)
	if err != nil {
		return fmt.Errorf("reader: error retrieving file %s: %s", path, err)
	}

	err = c.next.Process(filename, r)
	if err != nil {
		return fmt.Errorf("reader: error processing file %s: %s", filename, err)
	}
	r.Close()

	// Move the file from ToLoad to Loaded, or delete it
	if c.DeleteAfterDownload {
		slog.Info("Reader: deleting file", slog.String("path", path))
		err = c.connection.Delete(path)
		if err != nil {
			return fmt.Errorf("reader: error deleting file %s: %s", path, err)
		}
		return nil
	}

	newPath := filepath.Join(c.Loaded, filename)
	slog.Info("Reader: renaming file", slog.String("from", path), slog.String("to", newPath))
	err = c.connection.Rename(path, newPath)
	if err != nil {
		return fmt.Errorf("reader: error renaming file %s to %s: %s", path, newPath, err)
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

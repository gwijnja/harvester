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
		slog.Debug("Quitting connection")
		c.connection.Quit()
	}()

	// List files
	slog.Info("Listing files", slog.String("path", c.ToLoad))
	entries, err := c.connection.List(c.ToLoad)
	if err != nil {
		return nil, fmt.Errorf("error listing files in %s: %s", c.ToLoad, err)
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
		slog.Debug("Quitting connection")
		c.connection.Quit()
	}()

	// Set the transfer type to binary
	slog.Debug("Setting transfer type to binary")
	err = c.connection.Type(ftp.TransferTypeBinary)
	if err != nil {
		return fmt.Errorf("error setting transfer type to binary: %s", err)
	}

	path := filepath.Join(c.ToLoad, filename)
	r, err := c.connection.Retr(path)
	if err != nil {
		return fmt.Errorf("error retrieving file %s: %s", path, err)
	}

	err = c.next.Process(filename, r)
	if err != nil {
		return fmt.Errorf("error processing file %s: %s", filename, err)
	}
	r.Close()

	// Move the file from ToLoad to Loaded, or delete it
	if c.DeleteAfterDownload {
		slog.Info("Deleting file", slog.String("path", path))
		err = c.connection.Delete(path)
		if err != nil {
			return fmt.Errorf("error deleting file %s: %s", path, err)
		}
		slog.Debug("File deleted", slog.String("path", path))
		return nil
	}

	newPath := filepath.Join(c.Loaded, filename)
	slog.Info("Renaming file", slog.String("from", path), slog.String("to", newPath))
	err = c.connection.Rename(path, newPath)
	if err != nil {
		return fmt.Errorf("error renaming file %s to %s: %s", path, newPath, err)
	}
	slog.Debug("File renamed", slog.String("from", path), slog.String("to", newPath))

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

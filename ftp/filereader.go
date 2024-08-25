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
	DeleteAfterDownload bool
	Regex               string
	MaxFiles            int // set to 0 for no limit
	next                harvester.FileWriter
}

// SetNext sets the next FileWriter in the chain
func (c *FileReader) SetNext(next harvester.FileWriter) {
	c.next = next
}

// List lists files in the ToLoad directory
func (c *FileReader) List() ([]string, error) {

	// Connect
	conn, err := c.connect()
	if err != nil {
		return nil, err
	}
	defer func() {
		conn.Quit()
		slog.Info("ftp: Closed connection")
	}()

	// List files
	entries, err := conn.List(c.ToLoad)
	if err != nil {
		return nil, fmt.Errorf("ftp: Failed to list files in %s: %s", c.ToLoad, err)
	}
	slog.Debug("ftp: Listed files", slog.Int("count", len(entries)))

	// Filter files
	filtered, err := c.filterEntries(entries)
	if err != nil {
		return nil, err
	}

	return harvester.SortAndLimit(filtered, c.MaxFiles), nil
}

// Process downloads a file from the FTP server and processes it
func (c *FileReader) Process(filename string) error {

	// Connect
	conn, err := c.connect()
	if err != nil {
		return err
	}
	defer func() {
		conn.Quit()
		slog.Info("ftp: Closed connection")
	}()

	// Set the transfer type to binary
	err = conn.Type(ftp.TransferTypeBinary)
	if err != nil {
		return fmt.Errorf("ftp: Failed to set transfer type to binary: %s", err)
	}
	slog.Debug("ftp: Set transfer type to binary")

	// Retrieve the file
	toLoadPath := filepath.Join(c.ToLoad, filename)
	r, err := conn.Retr(toLoadPath)
	if err != nil {
		return fmt.Errorf("ftp: Failed to retrieve file %s: %s", toLoadPath, err)
	}
	defer func() {
		r.Close() // just in case we exit early for some reason
		slog.Info("ftp: Closed data connection")
	}()
	slog.Info("ftp: Retrieved file", slog.String("path", toLoadPath))

	// Call the next processor
	err = c.next.Process(filename, r)
	if err != nil {
		return err
	}
	r.Close() // close implicitly, because we're going to delete or move the file
	slog.Info("ftp: Closed data connection")

	// Move the file from ToLoad to Loaded, or delete it
	if c.DeleteAfterDownload {
		err = conn.Delete(toLoadPath)
		if err != nil {
			return fmt.Errorf("ftp: Failed to delete file %s: %s", toLoadPath, err)
		}
		slog.Info("ftp: Deleted file", slog.String("path", toLoadPath))
		return nil
	}

	// Move the file from toLoad to loaded
	loadedPath := filepath.Join(c.Loaded, filename)
	err = conn.Rename(toLoadPath, loadedPath)
	if err != nil {
		return fmt.Errorf("ftp: Failed to rename file %s to %s: %s", toLoadPath, loadedPath, err)
	}
	slog.Info("ftp: Renamed file", slog.String("from", toLoadPath), slog.String("to", loadedPath))

	return nil
}

// filterEntries filters the entries based on the regex and maxFiles
func (c *FileReader) filterEntries(entries []*ftp.Entry) ([]string, error) {

	filenames := []string{}

	// Loop over the entries and filter them
	for _, entry := range entries {

		// Skip non-file entries
		if entry.Type != ftp.EntryTypeFile {
			slog.Warn("ftp: Skipping non-file entry", slog.String("filename", entry.Name))
			continue
		}

		// Skip files that don't match the regex
		if c.Regex != "" {
			matched, err := regexp.MatchString(c.Regex, entry.Name)
			if err != nil {
				return nil, fmt.Errorf("ftp: Failed to compile regex %s: %s", c.Regex, err)
			}
			if !matched {
				slog.Warn("ftp: Skipping non-matching file", slog.String("filename", entry.Name))
				continue
			}
		}

		// Add the file to the list
		slog.Info("ftp: Found file", slog.String("filename", entry.Name))
		filenames = append(filenames, entry.Name)
	}

	return filenames, nil
}

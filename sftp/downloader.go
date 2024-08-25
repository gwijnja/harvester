package sftp

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"

	"github.com/gwijnja/harvester"
)

type Downloader struct {
	Connector
	ToLoad              string
	Loaded              string
	Regex               string
	MaxFiles            int
	DeleteAfterDownload bool
	harvester.NextProcessor
}

// List returns a list of files in the ToLoad directory that match the regex.
func (d *Downloader) List() ([]string, error) {

	// Connect to the SFTP server
	conn, err := d.Connector.connect()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// List the files in the ToLoad directory, relative to the current root.
	ff, err := conn.sftpClient.ReadDir(d.ToLoad)
	if err != nil {
		return nil, fmt.Errorf("sftp: Failed to read directory %s: %s", d.ToLoad, err)
	}
	slog.Info("sftp: Read directory", slog.Int("entries", len(ff)))

	// Exclude directories, we are only interested in files
	ff = excludeDirectories(ff)

	// Filter the files
	re, err := regexp.Compile(d.Regex)
	if err != nil {
		return nil, fmt.Errorf("sftp: Failed to compile regex %s: %s", d.Regex, err)
	}
	files := []string{}
	for _, f := range ff {
		if !re.MatchString(f.Name()) {
			slog.Warn("sftp: Skipping non-matching file", slog.String("filename", f.Name()))
			continue
		}
		files = append(files, f.Name())
		slog.Info("sftp: Found file", slog.String("filename", f.Name()))
	}

	return harvester.SortAndLimit(files, d.MaxFiles), nil
}

// Process downloads the file from the SFTP server and calls the next processor.
func (d *Downloader) Process(filename string) error {

	// Connect to the SFTP server
	conn, err := d.Connector.connect()
	if err != nil {
		return err
	}
	defer func() {
		conn.Close()
	}()

	// Open the file
	toloadPath := filepath.Join(d.ToLoad, filename)
	remoteFile, err := conn.sftpClient.Open(toloadPath)
	if err != nil {
		return fmt.Errorf("sftp: Failed to open remote file %s: %s", toloadPath, err)
	}
	slog.Info("sftp: Opened remote file", slog.String("filename", filename))
	defer func() {
		remoteFile.Close()
		slog.Info("sftp: Closed remote file", slog.String("filename", filename))
	}()

	// Call the next processor
	err = d.NextProcessor.Process(filename, remoteFile)
	if err != nil {
		return err
	}

	// Check if the file should be deleted
	if d.DeleteAfterDownload {
		err = conn.sftpClient.Remove(toloadPath)
		if err != nil {
			return fmt.Errorf("sftp: Failed to delete remote file %s: %s", toloadPath, err)
		}
		slog.Info("sftp: Deleted remote file", slog.String("filename", filename))
		return nil
	}

	// Move the file to the Loaded directory
	loadedPath := filepath.Join(d.Loaded, filename)
	err = conn.sftpClient.Rename(toloadPath, loadedPath)
	if err != nil {
		return fmt.Errorf("sftp: Failed to move remote file %s to %s: %s", toloadPath, loadedPath, err)
	}
	slog.Info("sftp: Moved remote file", slog.String("from", toloadPath), slog.String("to", loadedPath))

	return nil
}

// excludeDirectories returns a slice of FileInfo objects that are not directories.
func excludeDirectories(ff []os.FileInfo) []os.FileInfo {
	filenames := make([]os.FileInfo, 0, len(ff))
	for _, f := range ff {
		if !f.IsDir() {
			filenames = append(filenames, f)
		}
	}
	return filenames
}

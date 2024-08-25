package sftp

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"

	"github.com/gwijnja/harvester"
)

type FileReader struct {
	Connector
	ToLoad              string
	Loaded              string
	DeleteAfterDownload bool
	Regex               string
	MaxFiles            int
	harvester.NextProcessor
}

// List returns a list of files in the ToLoad directory that match the regex.
func (r *FileReader) List() ([]string, error) {

	// Connect to the SFTP server
	conn, err := r.Connector.connect()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// List the files in the ToLoad directory, relative to the current root.
	ff, err := conn.sftpClient.ReadDir(r.ToLoad)
	if err != nil {
		return nil, fmt.Errorf("sftp: Failed to read directory %s: %s", r.ToLoad, err)
	}
	slog.Info("sftp: Read directory", slog.Int("entries", len(ff)))

	// Exclude directories, we are only interested in files
	ff = excludeDirectories(ff)

	// Only match files that match the regex
	ff, err = filterFiles(ff, r.Regex)
	if err != nil {
		return nil, err
	}
	slog.Info("sftp: Filtered files", slog.Int("remaining", len(ff)))

	// Convert the FileInfo objects to a slice of strings
	files := make([]string, 0, len(ff))
	for _, f := range ff {
		files = append(files, f.Name())
		slog.Info("sftp: Found file", slog.String("filename", f.Name()))
		if r.MaxFiles > 0 && len(files) >= r.MaxFiles {
			slog.Info("sftp: Reached maximum number of files", slog.Int("max", r.MaxFiles))
			break
		}
	}

	return files, nil
}

// Process downloads the file from the SFTP server and calls the next processor.
func (r *FileReader) Process(filename string) error {

	// Connect to the SFTP server
	conn, err := r.Connector.connect()
	if err != nil {
		return err
	}
	defer func() {
		conn.Close()
	}()

	// Open the file
	toloadPath := filepath.Join(r.ToLoad, filename)
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
	err = r.NextProcessor.Process(filename, remoteFile)
	if err != nil {
		return err
	}

	// Check if the file should be deleted
	if r.DeleteAfterDownload {
		err = conn.sftpClient.Remove(toloadPath)
		if err != nil {
			return fmt.Errorf("sftp: Failed to delete remote file %s: %s", toloadPath, err)
		}
		slog.Info("sftp: Deleted remote file", slog.String("filename", filename))
		return nil
	}

	// Move the file to the Loaded directory
	loadedPath := filepath.Join(r.Loaded, filename)
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

// filterFiles returns a slice of FileInfo objects that match the regex.
func filterFiles(ff []os.FileInfo, regex string) ([]os.FileInfo, error) {
	filtered := []fs.FileInfo{}
	for _, f := range ff {
		if regex != "" {

			// Match the regex
			matched, err := regexp.MatchString(regex, f.Name())
			if err != nil {
				return nil, fmt.Errorf("sftp: Failed to compile regex %s: %s", regex, err)
			}

			// Skip non-matching files
			if !matched {
				slog.Warn("sftp: Skipping non-matching file", slog.String("filename", f.Name()))
				continue
			}
		}
		filtered = append(filtered, f)
		slog.Info("sftp: Found file", slog.String("filename", f.Name()))
	}
	return filtered, nil
}

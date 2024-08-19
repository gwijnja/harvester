package sftp

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
)

// List returns a list of files in the ToLoad directory that match the regex.
func (c *Connector) List() ([]string, error) {
	if c.client == nil {
		slog.Info("List() called, connecting", slog.String("host", c.Host), slog.Int("port", c.Port))
		err := c.connect()
		if err != nil {
			return nil, fmt.Errorf("failed to connect before listing files: %s", err)
		}
	}

	slog.Info("Connected to SFTP server", slog.String("host", c.Host), slog.Int("port", c.Port))

	// List the files in the ToLoad directory, relative to the current root.
	ff, err := c.client.ReadDir(c.ToLoad)
	if err != nil {
		return nil, fmt.Errorf("failed to list files in %s: %s", c.ToLoad, err)
	}

	// Exclude directories, we are only interested in files
	ff = excludeDirectories(ff)

	// Only match files that match the regex
	ff = filterFiles(ff, c.Regex)

	// Convert the FileInfo objects to a slice of strings
	files := make([]string, 0, len(ff))
	for _, f := range ff {
		files = append(files, f.Name())
	}

	return files, nil
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
func filterFiles(ff []os.FileInfo, regex *regexp.Regexp) []os.FileInfo {
	filenames := make([]os.FileInfo, 0, len(ff))
	for _, f := range ff {
		if regex.MatchString(f.Name()) {
			filenames = append(filenames, f)
		}
	}
	return filenames
}

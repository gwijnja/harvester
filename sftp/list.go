package sftp

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"regexp"
)

// List returns a list of files in the ToLoad directory that match the regex.
func (r *FileReader) List() ([]string, error) {
	r.Connector.reconnectIfNeeded()

	// List the files in the ToLoad directory, relative to the current root.
	slog.Info("sftp: Listing files", slog.String("directory", r.ToLoad))
	ff, err := r.Connector.client.ReadDir(r.ToLoad)
	if err != nil {
		return nil, fmt.Errorf("failed to list files in %s: %s", r.ToLoad, err)
	}

	// Exclude directories, we are only interested in files
	ff = excludeDirectories(ff)

	// Only match files that match the regex
	slog.Info("sftp: Filtering files", slog.String("regex", r.Regex))
	ff, err = filterFiles(ff, r.Regex)
	if err != nil {
		return nil, err
	}

	// Convert the FileInfo objects to a slice of strings
	files := make([]string, 0, len(ff))
	for _, f := range ff {
		files = append(files, f.Name())
		if r.MaxFiles > 0 && len(files) >= r.MaxFiles {
			break
		}
	}

	slog.Info("sftp: Found files", slog.Int("count", len(files)))

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
func filterFiles(ff []os.FileInfo, regex string) ([]os.FileInfo, error) {
	filtered := []fs.FileInfo{}
	for _, f := range ff {
		if regex != "" {
			matched, err := regexp.MatchString(regex, f.Name())
			if err != nil {
				return nil, fmt.Errorf("invalid regex, error matching %s with %s: %s", regex, f.Name(), err)
			}
			if !matched {
				slog.Warn("Skipping non-matching file", slog.String("filename", f.Name()))
				continue
			}
		}
		filtered = append(filtered, f)
	}
	return filtered, nil
}

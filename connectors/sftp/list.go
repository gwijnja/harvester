package sftp

import (
	"fmt"
	"log"
	"os"
	"regexp"
)

func (c *Connector) List() ([]string, error) {
	if c.client == nil {
		log.Println("[sftp] List() called, connecting first.")
		err := c.connect()
		if err != nil {
			return nil, fmt.Errorf("failed to connect before listing files: %s", err)
		}
	}

	log.Println("[sftp] Connected, listing files.")

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

func excludeDirectories(ff []os.FileInfo) []os.FileInfo {
	filenames := make([]os.FileInfo, 0, len(ff))
	for _, f := range ff {
		if !f.IsDir() {
			filenames = append(filenames, f)
		}
	}
	return filenames
}

func filterFiles(ff []os.FileInfo, regex *regexp.Regexp) []os.FileInfo {
	filenames := make([]os.FileInfo, 0, len(ff))
	for _, f := range ff {
		if regex.MatchString(f.Name()) {
			filenames = append(filenames, f)
		}
	}
	return filenames
}

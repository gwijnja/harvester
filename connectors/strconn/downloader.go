package strconn

import (
	"strings"

	"github.com/gwijnja/harvester"
)

// Downloader is a structure that holds the configuration for a string-based downloader.
type Downloader struct {
	harvester.BaseProcessor
	Files map[string]string
}

// Process reads a file and writes the contents to the next processor
func (d *Downloader) Process(ctx *harvester.FileContext) error {
	ctx.Reader = strings.NewReader(d.Files[ctx.Filename])
	return d.BaseProcessor.Process(ctx)
}

// List returns a list of filenames that are available for download
func (d *Downloader) List() ([]string, error) {
	keys := make([]string, 0, len(d.Files))
	for k := range d.Files {
		keys = append(keys, k)
	}
	return keys, nil
}

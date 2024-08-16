package strconn

import (
	"gwijnja/mft"
	"strings"
)

type Downloader struct {
	mft.BaseProcessor
	Files map[string]string
}

func (d *Downloader) Process(ctx *mft.FileContext) error {
	ctx.Reader = strings.NewReader(d.Files[ctx.Filename])
	return d.BaseProcessor.Process(ctx)
}

func (d *Downloader) List() ([]string, error) {
	keys := make([]string, 0, len(d.Files))
	for k := range d.Files {
		keys = append(keys, k)
	}
	return keys, nil
}

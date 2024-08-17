package local

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/gwijnja/harvester"
)

type Reader struct {
	harvester.BaseProcessor
	ToLoad              string
	Loaded              string
	FollowSymlinks      bool
	DeleteAfterDownload bool
	Regex               string
}

// Process reads a file from disk and presents it to the next processor in the chain.
func (d *Reader) Process(ctx *harvester.FileContext) error {
	log.Println("[local] Reader.Process(): Called for", ctx.Filename)

	path := filepath.Join(d.ToLoad, ctx.Filename)
	log.Println("[local] Reader.Process(): Opening", path)
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("[local] Reader.Process(): unable to open file %s: %s", path, err)
	}
	defer f.Close()

	ctx.Reader = f
	// Remember the filename, because we need it for moving or deleting the file, and ctx.Filename may change during processing.
	origFilename := ctx.Filename

	log.Println("[local] Reader.Process(): Calling the next processor")
	err = d.BaseProcessor.Process(ctx)
	if err != nil {
		return fmt.Errorf("[local] Reader.Process(): error processing %s: %s", path, err)
	}

	// After the transfer has completed succesfully, either delete the file or move it
	if d.DeleteAfterDownload {
		log.Println("[local] Reader.Process(): deleting", path)
		err = os.Remove(path)
		if err != nil {
			return fmt.Errorf("[local] Reader.Process(): the transfer was successful, but the source file (%s) could not be deleted: %s", path, err)
		}
		return nil
	}

	// Move the file from ToLoad to Loaded
	err = d.MoveFile(origFilename)
	if err != nil {
		return fmt.Errorf("[local] Reader.Process(): error moving %s: %s", ctx.Filename, err)
	}
	return nil
}

func (d *Reader) List() ([]string, error) {
	log.Println("[local] Reader.List(): Listing files in", d.ToLoad)

	files, err := os.ReadDir(d.ToLoad)
	if err != nil {
		return nil, fmt.Errorf("[local] Reader.List(): unable to list files in %s: %s", d.ToLoad, err)
	}
	log.Println("[local] Reader.List(): Found", len(files), "files")

	if d.Regex != "" {
		log.Println("[local] Reader.List(): Filtering files with regex:", d.Regex)
	}

	filenames := make([]string, 0, len(files))
	for _, file := range files {

		// Skip directories
		if file.IsDir() {
			continue
		}

		// Skip symlinks if FollowSymlinks is false
		if file.Type()&fs.ModeSymlink != 0 && !d.FollowSymlinks {
			continue
		}

		// Skip files that do not match the regex
		if d.Regex != "" {
			matched, err := regexp.MatchString(d.Regex, file.Name())
			if err != nil {
				return nil, fmt.Errorf("[local] Reader.List(): the regex seems invalid, error matching %s with %s: %s", d.Regex, file.Name(), err)
			}
			if !matched {
				log.Println("[local] Reader.List(): skipping", file.Name(), "because it does not match the regex")
				continue
			}
		}

		filenames = append(filenames, file.Name())
	}
	return filenames, nil
}

func (d *Reader) MoveFile(filename string) error {
	from := filepath.Join(d.ToLoad, filename)
	to := filepath.Join(d.Loaded, filename)
	log.Println("[local] Reader.MoveFile(): moving", from, "to", to)
	err := os.Rename(from, to)
	if err != nil {
		return fmt.Errorf("[local] Reader.MoveFile(): unable to move %s to %s: %s", from, to, err)
	}
	return nil
}

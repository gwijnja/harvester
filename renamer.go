package harvester

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"
)

type Renamer struct {
	NextProcessor
	Regex  string // Example: "(\\d{4})-(\\d{2})-(\\d{2})"
	Format string // Example: "$1$2$3.txt"
}

func (r *Renamer) Process(ctx *FileContext) error {

	slog.Debug("Compiling regex", slog.String("regex", r.Regex))
	re, err := regexp.Compile(r.Regex)
	if err != nil {
		return fmt.Errorf("unable to compile regex: %s", err)
	}

	slog.Debug("Matching regex", slog.String("regex", r.Regex), slog.String("filename", ctx.Filename))
	matches := re.FindStringSubmatch(ctx.Filename)
	if len(matches) == 0 {
		return fmt.Errorf("no matches found for regex: %s", r.Regex)
	}
	slog.Debug("Matches found", slog.Int("num_matches", len(matches)))

	slog.Debug("Renaming context filename", slog.String("filename", ctx.Filename))
	newFilename := r.Format
	for i, match := range matches {
		newFilename = strings.Replace(newFilename, fmt.Sprintf("$%d", i), match, -1)
	}

	slog.Debug("Renaming context filename", slog.String("old", ctx.Filename), slog.String("new", newFilename))
	ctx.Filename = newFilename

	// Call next processor
	return r.NextProcessor.Process(ctx)
}

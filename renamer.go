package harvester

import (
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"strings"
)

type Renamer struct {
	Regex  string // Example: "(\\d{4})-(\\d{2})-(\\d{2})"
	Format string // Example: "$1$2$3.txt"
	NextProcessor
}

func (r *Renamer) Process(oldFilename string, reader io.Reader) error {

	// Compile the regex
	re, err := regexp.Compile(r.Regex)
	if err != nil {
		return fmt.Errorf("harvester: Failed to compile regex: %s", err)
	}
	slog.Debug("harvester: Compiled regex", slog.String("regex", r.Regex))

	// Match the regex
	matches := re.FindStringSubmatch(oldFilename)
	if len(matches) == 0 {
		return fmt.Errorf("harvester: Failed to match %s", oldFilename)
	}
	slog.Debug("harvester: Matched regex", slog.Int("num_matches", len(matches)))

	// Replace the matches in the format string
	newFilename := r.Format
	for i, match := range matches {
		newFilename = strings.Replace(newFilename, fmt.Sprintf("$%d", i), match, -1)
	}
	slog.Info("harvester: Renamed file", slog.String("old", oldFilename), slog.String("new", newFilename))

	// Call next processor
	return r.NextProcessor.Process(newFilename, reader)
}

package harvester

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log/slog"
	"time"
)

// AuditCopy copies data from src to dst, while calculating a SHA-1 hash of the data and logging statistics.
func AuditCopy(dst io.Writer, src io.Reader) (written int64, err error) {

	// Prepare copy
	hasher := sha1.New()
	writer := io.MultiWriter(dst, hasher)
	start := time.Now()

	// Copy data
	slog.Debug("Copying data")
	written, err = io.Copy(writer, src)
	if err != nil {
		return written, fmt.Errorf("[mft] AuditCopy(): error copying data after %d bytes: %s", written, err)
	}

	// Gather statistics
	elapsed := time.Since(start)
	megabytesPerSecond := float64(written) / (1048576 * elapsed.Seconds())
	sha1hash := hasher.Sum(nil)

	// Log results
	slog.Info(
		"Copy complete",
		slog.Int64("written", written),
		slog.Duration("elapsed", elapsed),
		slog.Float64("megabytespersecond", megabytesPerSecond),
		slog.String("sha1hash", fmt.Sprintf("%x", sha1hash)),
	)

	return
}

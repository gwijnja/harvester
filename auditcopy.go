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
	written, err = io.Copy(writer, src)
	if err != nil {
		return written, fmt.Errorf("harvester: Failed copying data after %d bytes: %s", written, err)
	}

	// Gather statistics
	elapsed := time.Since(start)
	sha1hash := hasher.Sum(nil)
	mebibytes := float64(written) / 1024 / 1024

	// Log results
	slog.Info(
		"harvester: Copy complete",
		slog.Int64("bytes", written),
		slog.Int64("msec", elapsed.Milliseconds()),
		slog.Float64("MB/s", mebibytes/elapsed.Seconds()),
		slog.String("sha1", fmt.Sprintf("%x", sha1hash)),
	)

	return
}

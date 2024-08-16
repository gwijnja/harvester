package mft

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"time"
)

func AuditCopy(dst io.Writer, src io.Reader) (written int64, err error) {

	// Prepare copy
	log.Println("[mft] AuditCopy(): Setting up the MultiWriter")
	hasher := sha1.New()
	writer := io.MultiWriter(dst, hasher)
	start := time.Now()

	// Copy data
	log.Println("[mft] AuditCopy(): Copying the data")
	written, err = io.Copy(writer, src)
	if err != nil {
		return written, fmt.Errorf("[mft] AuditCopy(): error copying data after %d bytes: %s", written, err)
	}

	// Gather statistics
	elapsed := time.Since(start)
	megabytesPerSecond := float64(written) / (1048576 * elapsed.Seconds())
	sha1hash := hasher.Sum(nil)

	// Log results
	log.Printf(
		"[mft] AuditCopy(): %d bytes in %s (%.1f MiB/s), SHA-1: %x\n",
		written, elapsed.String(), megabytesPerSecond, sha1hash,
	)

	return
}

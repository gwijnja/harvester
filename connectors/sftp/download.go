package sftp

import (
	"fmt"
	"io"
	"log/slog"
)

// DownloadNew downloads a file from the SFTP server, moves it to the ToLoad directory, and returns a LocalFile.
func (c *Connector) DownloadNew(filename string, localStorage harvester.LocalStorage) (*harvester.LocalFile, error) {
	// connect if needed
	c.ReconnectIfNeeded()

	// copy file
	// if copy fails, return

	// delete or move remote
	// if delete or move fails, delete file in transmit and return.

	// move file to ToLoad
	// if move fails, log an error

	// return local file
}

// Download downloads a file from the SFTP server, moves it to the ToLoad directory, and returns a LocalFile.
func (c *Connector) Download(filename string, localStorage harvester.LocalStorage) (*harvester.LocalFile, error) {
	if c.client == nil {
		return nil, fmt.Errorf("SFTP client is not connected")
	}

	// Open the remote file
	remoteFile, err := c.client.Open(c.ToLoad + "/" + filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open remote file %s: %s", filename, err)
	}
	defer remoteFile.Close()

	// Open the local file
	f, err := localStorage.OpenFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open local file: %s", err)
	}

	// Copy the remote file to the local file
	written, err := io.Copy(f, remoteFile)
	f.Close() // Do not use defer, because the file needs to be closed before I call NewLocalFile below.
	if err != nil {
		return nil, fmt.Errorf("failed to copy remote file to local file: %s", err)
	}
	slog.Info("Copied", slog.Int64("bytes", written))

	// Move local file from Transmit to ToLoad
	localFile, err := localStorage.MoveToToLoad(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to move local file to ToLoad: %s", err)
	}
	return localFile, nil
}

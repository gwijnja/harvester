package sftp

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/gwijnja/harvester"
)

type FileReader struct {
	Connector
	ToLoad              string
	Loaded              string
	Regex               string
	DeleteAfterDownload bool
	harvester.NextProcessor
}

// Process downloads the file from the SFTP server and calls the next processor.
func (r *FileReader) Process(filename string) error {
	err := r.ReconnectIfNeeded()
	if err != nil {
		return fmt.Errorf("failed to reconnect: %s", err)
	}

	// Open the file
	toloadPath := filepath.Join(r.ToLoad, filename)
	remoteFile, err := r.client.Open(toloadPath)
	if err != nil {
		return fmt.Errorf("failed to open remote file %s: %s", toloadPath, err)
	}
	defer remoteFile.Close()

	// Call the next processor
	slog.Info("sftp: Calling the next processor", slog.String("filename", filename))
	err = r.NextProcessor.Process(filename, remoteFile)
	if err != nil {
		return fmt.Errorf("failed to process file %s: %s", filename, err)
	}

	if r.DeleteAfterDownload {
		slog.Info("sftp: Deleting remote file", slog.String("filename", filename))
		err = r.client.Remove(toloadPath)
		if err != nil {
			return fmt.Errorf("failed to delete remote file %s: %s", toloadPath, err)
		}
		return nil
	}

	// Move the file to the Loaded directory
	loadedPath := filepath.Join(r.Loaded, filename)
	slog.Info("sftp: Moving remote file", slog.String("from", toloadPath), slog.String("to", loadedPath))
	err = r.client.Rename(toloadPath, loadedPath)
	if err != nil {
		return fmt.Errorf("failed to move remote file %s to %s: %s", toloadPath, loadedPath, err)
	}

	return nil

}

func (r *FileReader) ReconnectIfNeeded() error {
	if r.Connector.client == nil {
		slog.Info("SFTP client is not connected, connecting")
		return r.Connector.connect()
	}

	if !r.Connector.isAlive() {
		slog.Info("SFTP client is dead, reconnecting")
		return r.Connector.connect()
	}

	return nil
}

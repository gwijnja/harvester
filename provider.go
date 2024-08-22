package harvester

/*
type Provider interface {
	NextProcessor
	List() ([]string, error)
	ProcessFiles() error
}

func (p *Provider) ProcessFiles() error {

	slog.Info("Listing files")
	filenames, err := p.List()
	if err != nil {
		slog.Error("Unable to list filenames", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Debug("Files found", slog.Int("count", len(filenames)))

	for _, filename := range filenames {
		slog.Debug("Handling file", slog.String("filename", filename))
		ctx := FileContext{Filename: filename}
		err := p.Process(&ctx)
		if err != nil {
			slog.Error("Error while processing", slog.Any("error", err))
			continue
		}
		slog.Info("Completed file copy", slog.String("filename", filename))
	}

	return nil
}
*/

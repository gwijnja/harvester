package harvester

import (
	"log/slog"
	"time"
)

type job struct {
	Reader     FileReader
	Processors []FileWriter
	Writer     FileWriter
	Interval   time.Duration
}

func NewJob(r FileReader, w FileWriter) *job {
	return &job{
		Reader:     r,
		Processors: []FileWriter{},
		Writer:     w,
	}
}

func (j *job) Add(w FileWriter) {
	j.Processors = append(j.Processors, w)
}

func (j *job) RunOnce() error {
	j.createChain()
	return j.processFiles()
}

func (j *job) Run(interval time.Duration) error {
	j.createChain()

	for {
		err := j.processFiles()
		if err != nil {
			slog.Error("Error while processing files", slog.Any("error", err))
		}

		slog.Info("Sleeping", slog.Duration("duration", interval))
		time.Sleep(interval)
	}
}

func (j *job) processFiles() error {
	slog.Info("Getting a list of files from the reader")
	filenames, err := j.Reader.List()
	slog.Debug("here")
	if err != nil {
		return err
	}
	slog.Debug("Got a list of files from the reader", slog.Any("filenames", filenames))

	for _, filename := range filenames {
		slog.Info("Processing file", slog.String("filename", filename))
		err := j.Reader.Process(filename)
		if err != nil {
			slog.Error("Error while processing file", slog.String("filename", filename), slog.Any("error", err))
			continue
		}
	}

	return nil
}

func (j *job) createChain() {
	// Link the processors
	for i, p := range j.Processors {
		if i == 0 {
			j.Reader.SetNext(p)
		} else {
			j.Processors[i-1].SetNext(p)
		}
	}

	// Add the writer to the end
	if len(j.Processors) > 0 {
		j.Processors[len(j.Processors)-1].SetNext(j.Writer)
	} else {
		j.Reader.SetNext(j.Writer)
	}
}

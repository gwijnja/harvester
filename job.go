package harvester

import (
	"log/slog"
)

type job struct {
	Reader     FileReader
	Processors []FileWriter
	Writer     FileWriter
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

func (j *job) Run() error {
	j.createChain()

	slog.Info("Getting a list of files from the reader")
	filenames, err := j.Reader.List()
	if err != nil {
		return err
	}

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

package harvester

import (
	"log/slog"
	"runtime"
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

func (j *job) Insert(w FileWriter) {
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
			slog.Error("harvester: Failed to process files", slog.Any("error", err))
		}

		j.logMemoryUsage()
		slog.Info("harvester: Sleeping", slog.Duration("duration", interval))
		time.Sleep(interval)
		slog.Info("harvester: Waking up")
	}
}

func (j *job) processFiles() error {
	filenames, err := j.Reader.List()
	if err != nil {
		return err
	}

	for _, filename := range filenames {
		slog.Info("job: Processing file", slog.String("filename", filename))
		err := j.Reader.Process(filename)
		if err != nil {
			slog.Error("job: Failed to process file", slog.String("filename", filename), slog.Any("error", err))
			continue
		}
	}
	slog.Info("job: Done processing files")

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

func (j *job) logMemoryUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	allocated := int64(float64(m.Alloc) / 1024)
	slog.Info("job: Current allocated heap memory", slog.Int64("kB", allocated))

}

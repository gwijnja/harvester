package harvester

import "io"

type FileWriter interface {
	SetNext(next FileWriter)
	Process(filename string, r io.Reader) error
}

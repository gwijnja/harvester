package harvester

import "io"

type FileWriter interface {
	Process(filename string, r io.Reader) error
	SetNext(next FileWriter)
}

package harvester

import "io"

// Processor is an interface that defines a Process method that takes a FileContext and returns an error.
type Processor interface {
	Process(ctx *FileContext) error
	SetNext(next Processor)
}

// Receiver is an interface that extends Processor with a List method that returns a list of filenames.
type Receiver interface {
	Processor
	List() ([]string, error)
}

// BaseProcessor is a struct that holds the next processor in the chain.
type BaseProcessor struct {
	next Processor
}

// SetNext sets the next processor in the chain.
func (b *BaseProcessor) SetNext(next Processor) {
	b.next = next
}

// Process calls the next processor in the chain, if it exists.
func (b *BaseProcessor) Process(ctx *FileContext) error {
	if b.next != nil {
		return b.next.Process(ctx)
	}
	return nil
}

// FileContext is a struct that contains a filename and a reader.
type FileContext struct {
	Filename string
	Reader   io.Reader
}

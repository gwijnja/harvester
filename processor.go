package mft

import "io"

type Processor interface {
	Process(ctx *FileContext) error
	SetNext(next Processor)
}

type Receiver interface {
	Processor
	List() ([]string, error)
}

type BaseProcessor struct {
	next Processor
}

func (b *BaseProcessor) SetNext(next Processor) {
	b.next = next
}

func (b *BaseProcessor) Process(ctx *FileContext) error {
	if b.next != nil {
		return b.next.Process(ctx)
	}
	return nil
}

type FileContext struct {
	Filename string
	Reader   io.Reader
}

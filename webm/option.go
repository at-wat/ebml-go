package webm

import (
	"github.com/at-wat/ebml-go"
)

var (
	// DefaultBlockInterceptor is the default BlockInterceptor used by BlockWriter.
	DefaultBlockInterceptor = NewMultiTrackBlockSorter(10)
)

// BlockWriterOption configures a BlockWriterOptions.
type BlockWriterOption func(*BlockWriterOptions) error

// BlockWriterOptions stores options for BlockWriter.
type BlockWriterOptions struct {
	ebmlHeader  interface{}
	segmentInfo interface{}
	seekHead    interface{}
	marshalOpts []ebml.MarshalOption
	onError     func(error)
	onFatal     func(error)
	interceptor BlockInterceptor
}

// WithEBMLHeader sets EBML header of WebM.
func WithEBMLHeader(h interface{}) BlockWriterOption {
	return func(o *BlockWriterOptions) error {
		o.ebmlHeader = h
		return nil
	}
}

// WithSegmentInfo sets Segment.Info of WebM.
func WithSegmentInfo(i interface{}) BlockWriterOption {
	return func(o *BlockWriterOptions) error {
		o.segmentInfo = i
		return nil
	}
}

// WithSeekHead sets Segment.SeekHead of WebM.
func WithSeekHead(s interface{}) BlockWriterOption {
	return func(o *BlockWriterOptions) error {
		o.seekHead = s
		return nil
	}
}

// WithMarshalOptions passes ebml.MarshalOption to ebml.Marshal.
func WithMarshalOptions(opts ...ebml.MarshalOption) BlockWriterOption {
	return func(o *BlockWriterOptions) error {
		o.marshalOpts = opts
		return nil
	}
}

// WithOnErrorHandler registers marshal error handler
func WithOnErrorHandler(handler func(error)) BlockWriterOption {
	return func(o *BlockWriterOptions) error {
		o.onError = handler
		return nil
	}
}

// WithOnFatalHandler registers marshal error handler
func WithOnFatalHandler(handler func(error)) BlockWriterOption {
	return func(o *BlockWriterOptions) error {
		o.onFatal = handler
		return nil
	}
}

// WithBlockInterceptor registers BlockInterceptor
func WithBlockInterceptor(interceptor BlockInterceptor) BlockWriterOption {
	return func(o *BlockWriterOptions) error {
		o.interceptor = interceptor
		return nil
	}
}

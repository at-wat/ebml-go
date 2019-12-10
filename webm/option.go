package webm

import (
	"github.com/at-wat/ebml-go"
)

// SimpleWriterOption configures a SimpleWriterOptions.
type SimpleWriterOption func(*SimpleWriterOptions) error

// SimpleWriterOptions stores options for NewSimpleWriter.
type SimpleWriterOptions struct {
	ebmlHeader  interface{}
	segmentInfo interface{}
	seekHead    interface{}
	marshalOpts []ebml.MarshalOption
	onError     func(error)
	onFatal     func(error)
}

// WithEBMLHeader sets EBML header of WebM.
func WithEBMLHeader(h interface{}) SimpleWriterOption {
	return func(o *SimpleWriterOptions) error {
		o.ebmlHeader = h
		return nil
	}
}

// WithSegmentInfo sets Segment.Info of WebM.
func WithSegmentInfo(i interface{}) SimpleWriterOption {
	return func(o *SimpleWriterOptions) error {
		o.segmentInfo = i
		return nil
	}
}

// WithSeekHead sets Segment.SeekHead of WebM.
func WithSeekHead(s interface{}) SimpleWriterOption {
	return func(o *SimpleWriterOptions) error {
		o.seekHead = s
		return nil
	}
}

// WithMarshalOptions passes ebml.MarshalOption to ebml.Marshal.
func WithMarshalOptions(opts ...ebml.MarshalOption) SimpleWriterOption {
	return func(o *SimpleWriterOptions) error {
		o.marshalOpts = opts
		return nil
	}
}

// WithOnErrorHandler registers marshal error handler
func WithOnErrorHandler(handler func(error)) SimpleWriterOption {
	return func(o *SimpleWriterOptions) error {
		o.onError = handler
		return nil
	}
}

// WithOnFatalHandler registers marshal error handler
func WithOnFatalHandler(handler func(error)) SimpleWriterOption {
	return func(o *SimpleWriterOptions) error {
		o.onFatal = handler
		return nil
	}
}

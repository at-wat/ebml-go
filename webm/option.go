package webm

import (
	"github.com/at-wat/ebml-go"
)

// FrameWriterOption configures a FrameWriterOptions.
type FrameWriterOption func(*FrameWriterOptions) error

// FrameWriterOptions stores options for NewFrameWriter.
type FrameWriterOptions struct {
	ebmlHeader  interface{}
	segmentInfo interface{}
	seekHead    interface{}
	marshalOpts []ebml.MarshalOption
}

// WithEBMLHeader sets EBML header of WebM.
func WithEBMLHeader(h interface{}) FrameWriterOption {
	return func(o *FrameWriterOptions) error {
		o.ebmlHeader = h
		return nil
	}
}

// WithSegmentInfo sets Segment.Info of WebM.
func WithSegmentInfo(i interface{}) FrameWriterOption {
	return func(o *FrameWriterOptions) error {
		o.segmentInfo = i
		return nil
	}
}

// WithSeekHead sets Segment.SeekHead of WebM.
func WithSeekHead(s interface{}) FrameWriterOption {
	return func(o *FrameWriterOptions) error {
		o.seekHead = s
		return nil
	}
}

// WithMarshalOptions passes ebml.MarshalOption to ebml.Marshal.
func WithMarshalOptions(opts ...ebml.MarshalOption) FrameWriterOption {
	return func(o *FrameWriterOptions) error {
		o.marshalOpts = opts
		return nil
	}
}

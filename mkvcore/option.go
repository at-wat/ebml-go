// Copyright 2019 The ebml-go authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mkvcore

import (
	"errors"

	"github.com/at-wat/ebml-go"
)

// ErrInvalidTrackNumber means that a track number is invalid. The track number must be larger than 0.
var ErrInvalidTrackNumber = errors.New("invalid track number")

// ErrCuesRequiresSeekHead means WithCues was used without WithSeekHead.
var ErrCuesRequiresSeekHead = errors.New("WithCues requires WithSeekHead")

// ErrCuesRequiresSeeker means WithCues was used with a writer that does not implement io.WriteSeeker.
var ErrCuesRequiresSeeker = errors.New("WithCues requires an io.WriteSeeker")

// ErrCuesReservedTooSmall means WithCues was called with a reservedSize smaller than 9 bytes.
var ErrCuesReservedTooSmall = errors.New("WithCues reservedSize must be at least 9")

// durationSettable is implemented by segment info types that support
// having their Duration field set automatically (e.g. webm.Info).
type durationSettable interface {
	SetDuration(float64)
}

// BlockWriterOption configures a BlockWriterOptions.
type BlockWriterOption interface {
	ApplyToBlockWriterOptions(opts *BlockWriterOptions) error
}

// BlockReaderOption configures a BlockReaderOptions.
type BlockReaderOption interface {
	ApplyToBlockReaderOptions(opts *BlockReaderOptions) error
}

// BlockReadWriterOptionFn configures a BlockReadWriterOptions.
type BlockReadWriterOptionFn func(*BlockReadWriterOptions) error

// ApplyToBlockWriterOptions implements BlockWriterOption.
func (o BlockReadWriterOptionFn) ApplyToBlockWriterOptions(opts *BlockWriterOptions) error {
	return o(&opts.BlockReadWriterOptions)
}

// ApplyToBlockReaderOptions implements BlockReaderOption.
func (o BlockReadWriterOptionFn) ApplyToBlockReaderOptions(opts *BlockReaderOptions) error {
	return o(&opts.BlockReadWriterOptions)
}

// BlockReadWriterOptions stores options for BlockWriter and BlockReader.
type BlockReadWriterOptions struct {
	onError func(error)
	onFatal func(error)
}

// WithOnErrorHandler registers marshal error handler.
func WithOnErrorHandler(handler func(error)) BlockReadWriterOptionFn {
	return func(o *BlockReadWriterOptions) error {
		o.onError = handler
		return nil
	}
}

// WithOnFatalHandler registers marshal error handler.
func WithOnFatalHandler(handler func(error)) BlockReadWriterOptionFn {
	return func(o *BlockReadWriterOptions) error {
		o.onFatal = handler
		return nil
	}
}

// BlockWriterOptionFn configures a BlockWriterOptions.
type BlockWriterOptionFn func(*BlockWriterOptions) error

// ApplyToBlockWriterOptions implements BlockWriterOption.
func (o BlockWriterOptionFn) ApplyToBlockWriterOptions(opts *BlockWriterOptions) error {
	return o(opts)
}

// BlockWriterOptions stores options for BlockWriter.
type BlockWriterOptions struct {
	BlockReadWriterOptions
	ebmlHeader          interface{}
	segmentInfo         interface{}
	seekHead            bool
	marshalOpts         []ebml.MarshalOption
	interceptor         BlockInterceptor
	mainTrackNumber     uint64
	maxKeyframeInterval int64
	cuesReservedSize    int
}

// WithEBMLHeader sets EBML header.
func WithEBMLHeader(h interface{}) BlockWriterOptionFn {
	return func(o *BlockWriterOptions) error {
		o.ebmlHeader = h
		return nil
	}
}

// WithSegmentInfo sets Segment.Info.
func WithSegmentInfo(i interface{}) BlockWriterOptionFn {
	return func(o *BlockWriterOptions) error {
		o.segmentInfo = i
		return nil
	}
}

// WithSeekHead enables SeekHead calculation
func WithSeekHead(enable bool) BlockWriterOptionFn {
	return func(o *BlockWriterOptions) error {
		o.seekHead = enable
		return nil
	}
}

// WithMarshalOptions passes ebml.MarshalOption to ebml.Marshal.
func WithMarshalOptions(opts ...ebml.MarshalOption) BlockWriterOptionFn {
	return func(o *BlockWriterOptions) error {
		o.marshalOpts = opts
		return nil
	}
}

// WithBlockInterceptor registers BlockInterceptor.
func WithBlockInterceptor(interceptor BlockInterceptor) BlockWriterOptionFn {
	return func(o *BlockWriterOptions) error {
		o.interceptor = interceptor
		return nil
	}
}

// WithCues enables writing a Cues (seek index) element and calculates Duration on finish.
// This effectively allows writing seekable streams without additional remuxing, at the cost
// pre-allocated and likely inefficient index space.
// reservedSize is the number of bytes (>=9) to reserve at the front of the file for Cues.
// Requires WithSeekHead(true) and an io.WriteSeeker, not just io.WriteCloser.
// If the index space is underallocated, Cues data is not written at all.
func WithCues(reservedSize int) BlockWriterOptionFn {
	return func(o *BlockWriterOptions) error {
		if reservedSize < 9 {
			return ErrCuesReservedTooSmall
		}
		o.cuesReservedSize = reservedSize
		return nil
	}
}

// WithMaxKeyframeInterval sets maximum keyframe interval of the main (video) track.
// Using this option starts the cluster with a key frame if possible.
// interval must be given in the scale of timecode.
func WithMaxKeyframeInterval(mainTrackNumber uint64, interval int64) BlockWriterOptionFn {
	return func(o *BlockWriterOptions) error {
		if mainTrackNumber == 0 {
			return ErrInvalidTrackNumber
		}
		o.mainTrackNumber = mainTrackNumber
		o.maxKeyframeInterval = interval
		return nil
	}
}

// BlockReaderOptionFn configures a BlockReaderOptions.
type BlockReaderOptionFn func(*BlockReaderOptions) error

// ApplyToBlockReaderOptions implements BlockReaderOption.
func (o BlockReaderOptionFn) ApplyToBlockReaderOptions(opts *BlockReaderOptions) error {
	return o(opts)
}

// BlockReaderOptions stores options for BlockReader.
type BlockReaderOptions struct {
	BlockReadWriterOptions
	unmarshalOpts []ebml.UnmarshalOption
}

// WithUnmarshalOptions passes ebml.UnmarshalOption to ebml.Unmarshal.
func WithUnmarshalOptions(opts ...ebml.UnmarshalOption) BlockReaderOptionFn {
	return func(o *BlockReaderOptions) error {
		o.unmarshalOpts = opts
		return nil
	}
}

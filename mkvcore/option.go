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

// BlockWriterOption configures a BlockWriterOptions.
type BlockWriterOption func(*BlockWriterOptions) error

// BlockWriterOptions stores options for BlockWriter.
type BlockWriterOptions struct {
	ebmlHeader          interface{}
	segmentInfo         interface{}
	seekHead            bool
	marshalOpts         []ebml.MarshalOption
	onError             func(error)
	onFatal             func(error)
	interceptor         BlockInterceptor
	mainTrackNumber     uint64
	maxKeyframeInterval int64
}

// BlockReaderOptions stores options for BlockReader.
type BlockReaderOptions struct {
	unmarshalOpts []ebml.UnmarshalOption
	onError       func(error)
	onFatal       func(error)
}

// WithEBMLHeader sets EBML header.
func WithEBMLHeader(h interface{}) BlockWriterOption {
	return func(o *BlockWriterOptions) error {
		o.ebmlHeader = h
		return nil
	}
}

// WithSegmentInfo sets Segment.Info.
func WithSegmentInfo(i interface{}) BlockWriterOption {
	return func(o *BlockWriterOptions) error {
		o.segmentInfo = i
		return nil
	}
}

// WithSeekHead enables SeekHead calculation
func WithSeekHead(enable bool) BlockWriterOption {
	return func(o *BlockWriterOptions) error {
		o.seekHead = enable
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

// WithOnErrorHandler registers marshal error handler.
func WithOnErrorHandler(handler func(error)) BlockWriterOption {
	return func(o *BlockWriterOptions) error {
		o.onError = handler
		return nil
	}
}

// WithOnFatalHandler registers marshal error handler.
func WithOnFatalHandler(handler func(error)) BlockWriterOption {
	return func(o *BlockWriterOptions) error {
		o.onFatal = handler
		return nil
	}
}

// WithBlockInterceptor registers BlockInterceptor.
func WithBlockInterceptor(interceptor BlockInterceptor) BlockWriterOption {
	return func(o *BlockWriterOptions) error {
		o.interceptor = interceptor
		return nil
	}
}

// WithMaxKeyframeInterval sets maximum keyframe interval of the main (video) track.
// Using this option starts the cluster with a key frame if possible.
// interval must be given in the scale of timecode.
func WithMaxKeyframeInterval(mainTrackNumber uint64, interval int64) BlockWriterOption {
	return func(o *BlockWriterOptions) error {
		if mainTrackNumber == 0 {
			return ErrInvalidTrackNumber
		}
		o.mainTrackNumber = mainTrackNumber
		o.maxKeyframeInterval = interval
		return nil
	}
}

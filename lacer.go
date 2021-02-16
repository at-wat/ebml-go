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

package ebml

import (
	"errors"
	"io"
)

// ErrUnevenFixedLace means that an uneven number of frames are stored in a fixed lacing block.
var ErrUnevenFixedLace = errors.New("uneven size of frames in fixed lace")

// ErrTooManyFrames means that a number of frames exceeds the limit.
var ErrTooManyFrames = errors.New("too many frames")

// Lacer is the interface to read laced frames in Block.
type Lacer interface {
	Write([][]byte) error
}

type noLacer struct {
	w io.Writer
}
type xiphLacer struct {
	w io.Writer
}
type fixedLacer struct {
	w io.Writer
}
type ebmlLacer struct {
	w io.Writer
}

func (l *noLacer) Write(b [][]byte) error {
	nFrames := len(b)
	switch {
	case nFrames == 0:
		return nil
	case nFrames != 1:
		return wrapErrorf(ErrTooManyFrames, "lacing %d frames by no-lacer", nFrames)
	}
	_, err := l.w.Write(b[0])
	return err
}

func (l *xiphLacer) Write(b [][]byte) error {
	nFrames := len(b)
	switch {
	case nFrames == 0:
		return nil
	case nFrames > 0xFF:
		return wrapErrorf(ErrTooManyFrames, "lacing %d frames", nFrames)
	}
	size := []byte{byte(nFrames - 1)}
	for i := 0; i < nFrames-1; i++ {
		n := len(b[i])
		for ; n > 0xFF; n -= 0xFF {
			size = append(size, 0xFF)
		}
		size = append(size, byte(n))
	}
	if _, err := l.w.Write(size); err != nil {
		return err
	}
	for i := 0; i < nFrames; i++ {
		if _, err := l.w.Write(b[i]); err != nil {
			return err
		}
	}
	return nil
}

func (l *fixedLacer) Write(b [][]byte) error {
	nFrames := len(b)
	switch {
	case nFrames == 0:
		return nil
	case nFrames > 0xFF:
		return wrapErrorf(ErrTooManyFrames, "lacing %d frames", nFrames)
	}
	for i := 1; i < nFrames; i++ {
		if len(b[i]) != len(b[0]) {
			return wrapErrorf(
				ErrUnevenFixedLace, "lacing %d bytes on %d bytes frame", len(b[i]), len(b[0]),
			)
		}
	}
	if _, err := l.w.Write([]byte{byte(nFrames - 1)}); err != nil {
		return err
	}
	for i := 0; i < nFrames; i++ {
		if _, err := l.w.Write(b[i]); err != nil {
			return err
		}
	}
	return nil
}

func (l *ebmlLacer) Write(b [][]byte) error {
	nFrames := len(b)
	switch {
	case nFrames == 0:
		return nil
	case nFrames > 0xFF:
		return wrapErrorf(ErrTooManyFrames, "lacing %d frames", nFrames)
	}
	size := []byte{byte(nFrames - 1)}

	n, err := encodeElementID(uint64(len(b[0])))
	if err != nil {
		return err
	}
	size = append(size, n...)

	frameSizePrev := int64(len(b[0]))
	for i := 1; i < nFrames-1; i++ {
		frameSize := int64(len(b[i]))
		n, err := encodeVInt(frameSize - frameSizePrev)
		frameSizePrev = frameSize
		if err != nil {
			return err
		}
		size = append(size, n...)
	}
	if _, err := l.w.Write(size); err != nil {
		return err
	}
	for i := 0; i < nFrames; i++ {
		if _, err := l.w.Write(b[i]); err != nil {
			return err
		}
	}
	return nil
}

// NewNoLacer creates pass-through Lacer for not laced data.
func NewNoLacer(w io.Writer) Lacer {
	return &noLacer{w}
}

// NewXiphLacer creates Lacer for Xiph laced data.
func NewXiphLacer(w io.Writer) Lacer {
	return &xiphLacer{w}
}

// NewFixedLacer creates Lacer for Fixed laced data.
func NewFixedLacer(w io.Writer) Lacer {
	return &fixedLacer{w}
}

// NewEBMLLacer creates Lacer for EBML laced data.
func NewEBMLLacer(w io.Writer) Lacer {
	return &ebmlLacer{w}
}

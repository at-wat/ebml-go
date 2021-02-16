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
	"bytes"
	"io"
	"testing"

	"github.com/at-wat/ebml-go/internal/errs"
)

func TestLacer(t *testing.T) {
	cases := map[string]struct {
		newLacer func(io.Writer) Lacer
		frames   [][]byte
		b        []byte
		err      error
	}{
		"NoLaceEmpty": {
			newLacer: NewNoLacer,
			frames:   [][]byte{},
			b:        []byte{},
			err:      nil,
		},
		"NoLaceTooMany": {
			newLacer: NewNoLacer,
			frames:   make([][]byte, 2),
			b:        []byte{},
			err:      ErrTooManyFrames,
		},
		"Xiph": {
			newLacer: NewXiphLacer,
			frames: [][]byte{
				bytes.Repeat([]byte{0xAA}, 256),
				bytes.Repeat([]byte{0xCC}, 16),
				bytes.Repeat([]byte{0x55}, 8),
			},
			b: append(
				[]byte{
					0x02,
					0xFF, 0x01, // 256 bytes
					0x10, // 16 bytes
				},
				bytes.Join(
					[][]byte{
						bytes.Repeat([]byte{0xAA}, 256),
						bytes.Repeat([]byte{0xCC}, 16),
						bytes.Repeat([]byte{0x55}, 8)}, []byte{})...),
			err: nil,
		},
		"XiphEmpty": {
			newLacer: NewXiphLacer,
			frames:   [][]byte{},
			b:        []byte{},
			err:      nil,
		},
		"XiphTooLong": {
			newLacer: NewXiphLacer,
			frames:   make([][]byte, 256),
			b:        nil,
			err:      ErrTooManyFrames,
		},
		"Fixed": {
			newLacer: NewFixedLacer,
			frames: [][]byte{
				bytes.Repeat([]byte{0xAA}, 16),
				bytes.Repeat([]byte{0xCC}, 16),
				bytes.Repeat([]byte{0x55}, 16),
			},
			b: append(
				[]byte{
					0x02,
				},
				bytes.Join(
					[][]byte{
						bytes.Repeat([]byte{0xAA}, 16),
						bytes.Repeat([]byte{0xCC}, 16),
						bytes.Repeat([]byte{0x55}, 16),
					}, []byte{})...),
			err: nil,
		},
		"FixedEmpty": {
			newLacer: NewFixedLacer,
			frames:   [][]byte{},
			b:        []byte{},
			err:      nil,
		},
		"FixedUneven": {
			newLacer: NewFixedLacer,
			frames: [][]byte{
				bytes.Repeat([]byte{0xAA}, 16),
				bytes.Repeat([]byte{0xCC}, 16),
				bytes.Repeat([]byte{0x55}, 15),
			},
			b:   nil,
			err: ErrUnevenFixedLace,
		},
		"FixedTooLong": {
			newLacer: NewFixedLacer,
			frames:   make([][]byte, 256),
			b:        nil,
			err:      ErrTooManyFrames,
		},
		"EBML": {
			newLacer: NewEBMLLacer,
			frames: [][]byte{
				bytes.Repeat([]byte{0xAA}, 800),
				bytes.Repeat([]byte{0xCC}, 500),
				bytes.Repeat([]byte{0x55}, 100),
			},
			b: append(
				[]byte{
					0x02,
					0x43, 0x20, // 800 bytes
					0x5E, 0xD3, // 500 bytes
				},
				bytes.Join(
					[][]byte{
						bytes.Repeat([]byte{0xAA}, 800),
						bytes.Repeat([]byte{0xCC}, 500),
						bytes.Repeat([]byte{0x55}, 100),
					}, []byte{})...),
			err: nil,
		},
		"EBMLEmpty": {
			newLacer: NewEBMLLacer,
			frames:   [][]byte{},
			b:        []byte{},
			err:      nil,
		},
		"EBMLTooLong": {
			newLacer: NewEBMLLacer,
			frames:   make([][]byte, 256),
			b:        nil,
			err:      ErrTooManyFrames,
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			l := c.newLacer(&buf)
			err := l.Write(c.frames)
			if !errs.Is(err, c.err) {
				t.Fatalf("Expected error: '%v', got: '%v'", c.err, err)
			}
			if !bytes.Equal(c.b, buf.Bytes()) {
				t.Errorf("Expected data: %v, \n         got: %v", c.b, buf.Bytes())
			}
		})
	}
}

func TestLacer_WriterError(t *testing.T) {
	lacers := map[string]struct {
		frames   [][]byte
		newLacer func(io.Writer) Lacer
		n        int
	}{
		"NoLacer":    {[][]byte{{0x01, 0x02}}, NewNoLacer, 3},
		"XiphLacer":  {[][]byte{{0x01}, {0x02}}, NewXiphLacer, 4},
		"FixedLacer": {[][]byte{{0x01}, {0x02}}, NewFixedLacer, 3},
		"EBMLLacer":  {[][]byte{{0x01}, {0x02}}, NewEBMLLacer, 4},
	}
	for name, c := range lacers {
		t.Run(name, func(t *testing.T) {
			for l := 0; l < c.n-1; l++ {
				lacer := c.newLacer(&limitedDummyWriter{limit: l})
				if err := lacer.Write(c.frames); !errs.Is(err, bytes.ErrTooLarge) {
					t.Errorf("Expected error against too large data (Writer size limit: %d): '%v', got '%v'", l, bytes.ErrTooLarge, err)
				}
			}
		})
	}
}

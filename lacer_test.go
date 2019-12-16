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
)

func TestLacer(t *testing.T) {
	cases := map[string]struct {
		newLacer func(io.Writer) Lacer
		frames   [][]byte
		b        []byte
		err      error
	}{
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
			err:      errTooManyFrames,
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
			err: errUnevenFixedLace,
		},
		"EBML": {
			newLacer: NewEBMLLacer,
			frames: [][]byte{
				bytes.Repeat([]byte{0xAA}, 256),
				bytes.Repeat([]byte{0xCC}, 16),
				bytes.Repeat([]byte{0x55}, 8),
			},
			b: append(
				[]byte{
					0x02,
					0x41, 0x00, // 256 bytes
					0x90, // 16 bytes
				},
				bytes.Join(
					[][]byte{
						bytes.Repeat([]byte{0xAA}, 256),
						bytes.Repeat([]byte{0xCC}, 16),
						bytes.Repeat([]byte{0x55}, 8),
					}, []byte{})...),
			err: nil,
		},
		"EBMLEmpty": {
			newLacer: NewEBMLLacer,
			frames:   [][]byte{},
			b:        []byte{},
			err:      nil,
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			l := c.newLacer(&buf)
			err := l.Write(c.frames)
			if !isErr(err, c.err) {
				t.Fatalf("Unexpected error, expected: %v, got: %v", c.err, err)
			}
			if !bytes.Equal(c.b, buf.Bytes()) {
				t.Errorf("Unexpected data, \nexpected: %v, \n     got: %v", c.b, buf.Bytes())
			}
		})
	}
}
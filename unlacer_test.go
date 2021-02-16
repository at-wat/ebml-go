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

func TestUnlacer(t *testing.T) {
	cases := map[string]struct {
		newUnlacer func(io.Reader, int64) (Unlacer, error)
		header     []byte
		frames     [][]byte
		err        error
	}{
		"Xiph": {
			newUnlacer: NewXiphUnlacer,
			header: []byte{
				0x02,
				0xFF, 0x01, // 256 bytes
				0x10, // 16 bytes
			},
			frames: [][]byte{
				bytes.Repeat([]byte{0xAA}, 256),
				bytes.Repeat([]byte{0xCC}, 16),
				bytes.Repeat([]byte{0x55}, 8),
			},
			err: nil,
		},
		"XiphEmpty": {
			newUnlacer: NewXiphUnlacer,
			header:     []byte{},
			frames:     [][]byte{},
			err:        io.ErrUnexpectedEOF,
		},
		"XiphShort": {
			newUnlacer: NewXiphUnlacer,
			header: []byte{
				0x02, 0xFF,
			},
			frames: [][]byte{},
			err:    io.ErrUnexpectedEOF,
		},
		"XiphMissingFrame": {
			newUnlacer: NewXiphUnlacer,
			header: []byte{
				0x02,
				0x02,
				0x01,
			},
			frames: [][]byte{{0x00, 0x01}},
			err:    io.ErrUnexpectedEOF,
		},
		"XiphMissingLastFrame": {
			newUnlacer: NewXiphUnlacer,
			header: []byte{
				0x02,
				0x02,
				0x01,
			},
			frames: [][]byte{{0x00, 0x01}, {0x02}},
			err:    io.ErrUnexpectedEOF,
		},
		"Fixed": {
			newUnlacer: NewFixedUnlacer,
			header: []byte{
				0x02,
			},
			frames: [][]byte{
				bytes.Repeat([]byte{0xAA}, 16),
				bytes.Repeat([]byte{0xCC}, 16),
				bytes.Repeat([]byte{0x55}, 16),
			},
			err: nil,
		},
		"FixedEmpty": {
			newUnlacer: NewFixedUnlacer,
			header:     []byte{},
			frames:     [][]byte{},
			err:        io.ErrUnexpectedEOF,
		},
		"FixedUndivisible": {
			newUnlacer: NewFixedUnlacer,
			header: []byte{
				0x02,
			},
			frames: [][]byte{
				bytes.Repeat([]byte{0xAA}, 16),
				bytes.Repeat([]byte{0xCC}, 16),
				bytes.Repeat([]byte{0x55}, 15),
			},
			err: ErrFixedLaceUndivisible,
		},
		"EBML": {
			newUnlacer: NewEBMLUnlacer,
			header: []byte{
				0x02,
				0x43, 0x20, // 800 bytes
				0x5E, 0xD3, // 500 bytes
			},
			frames: [][]byte{
				bytes.Repeat([]byte{0xAA}, 800),
				bytes.Repeat([]byte{0xCC}, 500),
				bytes.Repeat([]byte{0x55}, 100),
			},
			err: nil,
		},
		"EBMLEmpty": {
			newUnlacer: NewEBMLUnlacer,
			header:     []byte{},
			frames:     [][]byte{},
			err:        io.ErrUnexpectedEOF,
		},
		"EBMLInvalidSize": {
			newUnlacer: NewEBMLUnlacer,
			header: []byte{
				0x02,
				0x41,
			},
			frames: [][]byte{},
			err:    io.ErrUnexpectedEOF,
		},
		"EBMLInvalidSize2": {
			newUnlacer: NewEBMLUnlacer,
			header: []byte{
				0x03,
				0x81,
			},
			frames: [][]byte{},
			err:    io.ErrUnexpectedEOF,
		},
		"EBMLNegativeSize": {
			newUnlacer: NewEBMLUnlacer,
			header: []byte{
				0x03,
				0x81, // 1 byte
				0x80, // -62 bytes
			},
			frames: [][]byte{},
			err:    io.ErrUnexpectedEOF,
		},
		"EBMLMissingFrame": {
			newUnlacer: NewEBMLUnlacer,
			header: []byte{
				0x02,
				0x82, // 2 bytes
				0x9F, // 2 bytes
			},
			frames: [][]byte{{0x00, 0x01}},
			err:    io.ErrUnexpectedEOF,
		},
		"EBMLMissingLastFrame": {
			newUnlacer: NewEBMLUnlacer,
			header: []byte{
				0x02,
				0x82, // 2 bytes
				0x9E, // 1 byte
			},
			frames: [][]byte{{0x00, 0x01}, {0x02}},
			err:    io.ErrUnexpectedEOF,
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			b := append([]byte{}, c.header...)
			for _, f := range c.frames {
				b = append(b, f...)
			}

			ul, err := c.newUnlacer(bytes.NewReader(b), int64(len(b)))
			if !errs.Is(err, c.err) {
				t.Fatalf("Expected error: '%v', got: '%v'", c.err, err)
			}
			if err != nil {
				return
			}

			for _, f := range c.frames {
				b, err := ul.Read()
				if err != nil {
					t.Fatalf("Unexpected error: '%v'", err)
				}
				if !bytes.Equal(f, b) {
					t.Errorf("Unexpected data, \nexpected: %v, \n     got: %v", f, b)
				}
			}
			if _, err := ul.Read(); err != io.EOF {
				t.Fatalf("Unexpected error: '%v'", err)
			}
		})
	}
}

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
	"reflect"
	"testing"
)

func TestUnmarshalBlock(t *testing.T) {
	testCases := map[string]struct {
		input    []byte
		expected Block
	}{
		"Track1BKeyframeInvisible": {
			[]byte{0x82, 0x01, 0x23, 0x88, 0xAA, 0xCC},
			Block{0x02, 0x0123, true, true, LacingNo, false, nil, [][]byte{{0xAA, 0xCC}}},
		},
		"Track2BDiscardable": {
			[]byte{0x42, 0x13, 0x01, 0x23, 0x01, 0x11, 0x22, 0x33},
			Block{0x0213, 0x0123, false, false, LacingNo, true, nil, [][]byte{{0x11, 0x22, 0x33}}},
		},
		"Track3BNoData": {
			[]byte{0x21, 0x23, 0x45, 0x00, 0x02, 0x00},
			Block{0x012345, 0x0002, false, false, LacingNo, false, nil, [][]byte{{}}},
		},
	}
	for n, c := range testCases {
		t.Run(n, func(t *testing.T) {
			block, err := UnmarshalBlock(bytes.NewBuffer(c.input))
			if err != nil {
				t.Fatalf("Failed to unmarshal block: %v", err)
			}
			if !reflect.DeepEqual(c.expected, *block) {
				t.Errorf("Unexpected unmarshal result, expected: %v, got: %v", c.expected, *block)
			}
		})
	}
}

func TestUnmarshalBlock_Error(t *testing.T) {
	input := []byte{0x21, 0x23, 0x45, 0x00, 0x02, 0x00}

	t.Run("EOF",
		func(t *testing.T) {
			for l := 0; l < len(input); l++ {
				if _, err := UnmarshalBlock(bytes.NewBuffer(input[:l])); err != io.ErrUnexpectedEOF {
					t.Errorf("UnmarshalBlock should return %v against short data (%d bytes), but got %v",
						io.ErrUnexpectedEOF, l, err)
				}
			}
		},
	)
}

func TestMarshalBlock(t *testing.T) {
	testCases := map[string]struct {
		input    Block
		expected []byte
	}{
		"Track1BKeyframeInvisible": {
			Block{0x02, 0x0123, true, true, LacingNo, false, nil, [][]byte{{0xAA, 0xCC}}},
			[]byte{0x82, 0x01, 0x23, 0x88, 0xAA, 0xCC},
		},
		"Track2BDiscardable": {
			Block{0x0213, 0x0123, false, false, LacingNo, true, nil, [][]byte{{0x11, 0x22, 0x33}}},
			[]byte{0x42, 0x13, 0x01, 0x23, 0x01, 0x11, 0x22, 0x33},
		},
		"Track3BNoData": {
			Block{0x012345, 0x0002, false, false, LacingNo, false, nil, [][]byte{{}}},
			[]byte{0x21, 0x23, 0x45, 0x00, 0x02, 0x00},
		},
	}
	for n, c := range testCases {
		t.Run(n, func(t *testing.T) {
			var b bytes.Buffer
			err := MarshalBlock(&c.input, &b)
			if err != nil {
				t.Fatalf("Failed to marshal block: %v", err)
			}
			if !reflect.DeepEqual(c.expected, b.Bytes()) {
				t.Errorf("Unexpected marshal result, expected: %v, got: %v", c.expected, b.Bytes())
			}
		})
	}
}

func TestMarshalBlock_Error(t *testing.T) {
	input := &Block{0x012345, 0x0002, false, false, LacingNo, false, nil, [][]byte{{0x00}}} // 7 bytes

	t.Run("EOF",
		func(t *testing.T) {
			for l := 0; l < 7; l++ {
				err := MarshalBlock(input, &limitedWriter{limit: l})
				if err != bytes.ErrTooLarge {
					t.Errorf("UnmarshalBlock should return bytes.ErrTooLarge against too large data, but got %v", err)
				}
			}
		},
	)
}

type limitedWriter struct {
	n     int
	limit int
}

func (s *limitedWriter) Write(b []byte) (int, error) {
	s.n += len(b)
	if s.n > s.limit {
		return len(b) - (s.n - s.limit), bytes.ErrTooLarge
	}
	return len(b), nil
}

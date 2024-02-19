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
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/at-wat/ebml-go/internal/errs"
)

func TestMarshal(t *testing.T) {
	type TestOmitempty struct {
		DocType        string `ebml:"EBMLDocType,omitempty"`
		DocTypeVersion uint64 `ebml:"EBMLDocTypeVersion,omitempty"`
		SeekID         []byte `ebml:"SeekID,omitempty"`
	}
	type TestNoOmitempty struct {
		DocType        string `ebml:"EBMLDocType"`
		DocTypeVersion uint64 `ebml:"EBMLDocTypeVersion"`
		SeekID         []byte `ebml:"SeekID"`
	}
	type TestSliceOmitempty struct {
		DocTypeVersion []uint64 `ebml:"EBMLDocTypeVersion,omitempty"`
	}
	type TestSliceNoOmitempty struct {
		DocTypeVersion []uint64 `ebml:"EBMLDocTypeVersion"`
	}
	type TestSized struct {
		DocType        string  `ebml:"EBMLDocType,size=3"`
		DocTypeVersion uint64  `ebml:"EBMLDocTypeVersion,size=2"`
		Duration0      float32 `ebml:"Duration,size=8"`
		Duration1      float64 `ebml:"Duration,size=4"`
		SeekID         []byte  `ebml:"SeekID,size=2"`
	}
	type TestPtr struct {
		DocType        *string `ebml:"EBMLDocType"`
		DocTypeVersion *uint64 `ebml:"EBMLDocTypeVersion"`
	}
	type TestPtrOmitempty struct {
		DocType        *string `ebml:"EBMLDocType,omitempty"`
		DocTypeVersion *uint64 `ebml:"EBMLDocTypeVersion,omitempty"`
	}
	type TestInterface struct {
		DocType        interface{} `ebml:"EBMLDocType"`
		DocTypeVersion interface{} `ebml:"EBMLDocTypeVersion"`
	}
	type TestBlocks struct {
		Block Block `ebml:"SimpleBlock"`
	}

	var str string
	var uinteger uint64

	testCases := map[string]struct {
		input    interface{}
		expected [][]byte // one of
	}{
		"Omitempty": {
			&struct{ EBML TestOmitempty }{},
			[][]byte{{0x1a, 0x45, 0xDF, 0xA3, 0x80}},
		},
		"NoOmitempty": {
			&struct{ EBML TestNoOmitempty }{},
			[][]byte{
				{
					0x1A, 0x45, 0xDF, 0xA3, 0x8A,
					0x42, 0x82, 0x80,
					0x42, 0x87, 0x81, 0x00,
					0x53, 0xAB, 0x80,
				},
			},
		},
		"SliceOmitempty": {
			&struct {
				EBML TestSliceOmitempty
			}{TestSliceOmitempty{make([]uint64, 0)}},
			[][]byte{{0x1a, 0x45, 0xDF, 0xA3, 0x80}},
		},
		"SliceOmitemptyNested": {
			&struct {
				EBML []TestSliceOmitempty `ebml:"EBML,omitempty"`
			}{make([]TestSliceOmitempty, 3)},
			[][]byte{{}},
		},
		"SliceNoOmitempty": {
			&struct {
				EBML TestSliceNoOmitempty
			}{TestSliceNoOmitempty{make([]uint64, 2)}},
			[][]byte{
				{
					0x1a, 0x45, 0xDF, 0xA3, 0x88,
					0x42, 0x87, 0x81, 0x00,
					0x42, 0x87, 0x81, 0x00,
				},
			},
		},
		"Sized": {
			&struct{ EBML TestSized }{TestSized{"a", 1, 0.0, 0.0, []byte{0x01}}},
			[][]byte{
				{
					0x1A, 0x45, 0xDF, 0xA3, 0xA2,
					0x42, 0x82, 0x83, 0x61, 0x00, 0x00,
					0x42, 0x87, 0x82, 0x00, 0x01,
					0x44, 0x89, 0x88, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x44, 0x89, 0x84, 0x00, 0x00, 0x00, 0x00,
					0x53, 0xAB, 0x82, 0x01, 0x00,
				},
			},
		},
		"SizedAndOverflow": {
			&struct{ EBML TestSized }{TestSized{"abc", 0x012345, 0.0, 0.0, []byte{0x01, 0x02, 0x03}}},
			[][]byte{
				{
					0x1A, 0x45, 0xDF, 0xA3, 0xA4,
					0x42, 0x82, 0x83, 0x61, 0x62, 0x63,
					0x42, 0x87, 0x83, 0x01, 0x23, 0x45,
					0x44, 0x89, 0x88, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x44, 0x89, 0x84, 0x00, 0x00, 0x00, 0x00,
					0x53, 0xAB, 0x83, 0x01, 0x02, 0x03,
				},
			},
		},
		"Ptr": {
			&struct{ EBML TestPtr }{TestPtr{&str, &uinteger}},
			[][]byte{
				{
					0x1A, 0x45, 0xDF, 0xA3, 0x87,
					0x42, 0x82, 0x80,
					0x42, 0x87, 0x81, 0x00,
				},
			},
		},
		"PtrOmitempty": {
			&struct{ EBML TestPtrOmitempty }{TestPtrOmitempty{&str, &uinteger}},
			[][]byte{{0x1A, 0x45, 0xDF, 0xA3, 0x80}},
		},
		"Interface": {
			&struct{ EBML TestInterface }{TestInterface{str, uinteger}},
			[][]byte{
				{
					0x1A, 0x45, 0xDF, 0xA3, 0x87,
					0x42, 0x82, 0x80,
					0x42, 0x87, 0x81, 0x00,
				},
			},
		},
		"InterfacePtr": {
			&struct{ EBML TestInterface }{TestInterface{&str, &uinteger}},
			[][]byte{
				{
					0x1A, 0x45, 0xDF, 0xA3, 0x87,
					0x42, 0x82, 0x80,
					0x42, 0x87, 0x81, 0x00,
				},
			},
		},
		"Map": {
			&map[string]interface{}{
				"Info": map[string]interface{}{
					"MuxingApp":  "test",
					"WritingApp": "abcd",
				},
			},
			[][]byte{
				{
					0x15, 0x49, 0xA9, 0x66, 0x8E,
					0x4D, 0x80, 0x84, 0x74, 0x65, 0x73, 0x74,
					0x57, 0x41, 0x84, 0x61, 0x62, 0x63, 0x64,
				},
				{ // Go map element order is unstable
					0x15, 0x49, 0xA9, 0x66, 0x8e,
					0x57, 0x41, 0x84, 0x61, 0x62, 0x63, 0x64,
					0x4D, 0x80, 0x84, 0x74, 0x65, 0x73, 0x74,
				},
			},
		},
		"Block": {
			&TestBlocks{
				Block: Block{
					TrackNumber: 0x01, Timecode: 0x0123, Lacing: LacingNo, Data: [][]byte{{0x01}},
				},
			},
			[][]byte{{0xA3, 0x85, 0x81, 0x01, 0x23, 0x00, 0x01}},
		},
		"BlockXiph": {
			&TestBlocks{
				Block: Block{
					TrackNumber: 0x01, Timecode: 0x0123, Lacing: LacingXiph, Data: [][]byte{{0x01}, {0x02}},
				},
			},
			[][]byte{{0xA3, 0x88, 0x81, 0x01, 0x23, 0x02, 0x01, 0x01, 0x01, 0x02}},
		},
		"BlockFixed": {
			&TestBlocks{
				Block: Block{
					TrackNumber: 0x01, Timecode: 0x0123, Lacing: LacingFixed, Data: [][]byte{{0x01}, {0x02}},
				},
			},
			[][]byte{{0xA3, 0x87, 0x81, 0x01, 0x23, 0x04, 0x01, 0x01, 0x02}},
		},
		"BlockEBML": {
			&TestBlocks{
				Block: Block{
					TrackNumber: 0x01, Timecode: 0x0123, Lacing: LacingEBML, Data: [][]byte{{0x01}, {0x02}},
				},
			},
			[][]byte{{0xA3, 0x88, 0x81, 0x01, 0x23, 0x06, 0x01, 0x81, 0x01, 0x02}},
		},
	}

	for n, c := range testCases {
		t.Run(n, func(t *testing.T) {
			var b bytes.Buffer
			if err := Marshal(c.input, &b); err != nil {
				t.Fatalf("Unexpected error: '%v'", err)
			}
			for _, expected := range c.expected {
				if bytes.Equal(expected, b.Bytes()) {
					return
				}
			}
			t.Errorf("Marshaled binary doesn't match:\n expected one of:\n%v,\ngot:\n%v", c.expected, b.Bytes())
		})
	}
}

func TestMarshal_Error(t *testing.T) {
	testCases := map[string]struct {
		input interface{}
		err   error
	}{
		"InvalidInput": {
			struct{}{},
			ErrInvalidType,
		},
		"InvalidElementName": {
			&struct {
				Invalid uint64 `ebml:"Invalid"`
			}{},
			ErrUnknownElementName,
		},
		"InvalidMapKey": {
			&map[int]interface{}{1: "test"},
			ErrNonStringMapKey,
		},
		"InvalidType": {
			&[]int{},
			ErrIncompatibleType,
		},
	}
	for n, c := range testCases {
		t.Run(n, func(t *testing.T) {
			var b bytes.Buffer
			if err := Marshal(c.input, &b); !errs.Is(err, c.err) {
				t.Fatalf("Expected error: '%v', got: '%v'", c.err, err)
			}
		})
	}
}

func TestMarshal_OptionError(t *testing.T) {
	errExpected := errors.New("an error")
	if err := Marshal(&struct{}{}, &bytes.Buffer{},
		func(*MarshalOptions) error {
			return errExpected
		},
	); err != errExpected {
		t.Errorf("Expected error against failing MarshalOption: '%v', got: '%v'", errExpected, err)
	}
}

func TestMarshal_WriterError(t *testing.T) {
	type EBMLHeader struct {
		DocTypeVersion  uint64 `ebml:"EBMLDocTypeVersion"`              // 2 + 1 + 1 bytes
		DocTypeVersion2 uint64 `ebml:"EBMLDocTypeVersion,size=unknown"` // 2 + 8 + 8 bytes
	} // 22 bytes
	s := struct {
		Header  EBMLHeader `ebml:"EBML"`              // 4 + 1 + 22 bytes
		Header2 EBMLHeader `ebml:"EBML,size=unknown"` // 4 + 8 + 22 bytes
	}{} // 61 bytes

	for l := 0; l < 61; l++ {
		if err := Marshal(&s, &limitedDummyWriter{limit: l}); !errs.Is(err, bytes.ErrTooLarge) {
			t.Errorf("Expected error against too large data (Writer size limit: %d): '%v', got '%v'",
				l, bytes.ErrTooLarge, err,
			)
		}
	}
}

func TestMarshal_EncodeError(t *testing.T) {
	s := struct {
		SimpleBlock Block
	}{
		SimpleBlock: Block{
			Lacing: LacingFixed,
			Data:   [][]byte{{0x01}, {0x01, 0x02}},
		},
	}
	if err := Marshal(&s, &bytes.Buffer{}); !errs.Is(err, ErrUnevenFixedLace) {
		t.Errorf("Expected error on encoding uneven fixed lace Block: '%v', got: '%v'",
			ErrUnevenFixedLace, err)
	}
}

func TestMarshal_WithWriteHooks(t *testing.T) {
	type DummyCluster struct {
		Timecode uint64 `ebml:"Timecode"` // 2 + 1 + 1 bytes
	}
	s := struct {
		Header struct {
			DocTypeVersion uint64 `ebml:"EBMLDocTypeVersion"` // 2 + 1 + 1 bytes
		} `ebml:"EBML"` // 4 + 1 + 4 bytes
		Segment struct {
			Cluster []DummyCluster `ebml:"Cluster,size=unknown"` // 4 + 8 + 4 bytes
		} `ebml:"Segment,size=unknown"` // 4 + 8 + (16 * n) bytes
	}{}
	s.Segment.Cluster = make([]DummyCluster, 2)

	m := make(map[string][]*Element)
	hook := withElementMap(m)
	if err := Marshal(&s, &bytes.Buffer{}, WithElementWriteHooks(hook)); err != nil {
		t.Errorf("Unexpected error: '%v'", err)
	}

	expected := map[string][]uint64{
		"EBML":                     {0},
		"EBML.EBMLDocTypeVersion":  {4},
		"Segment":                  {9},
		"Segment.Cluster":          {21, 36},
		"Segment.Cluster.Timecode": {33, 48},
	}
	posMap := elementPositionMap(m)
	if !reflect.DeepEqual(expected, posMap) {
		t.Errorf("Unexpected write hook positions, \nexpected: %v, \n     got: %v", expected, posMap)
	}
	checkTarget := "Segment.Cluster.Timecode"
	switch {
	case len(m[checkTarget]) != 2:
		t.Fatalf("%s write hook should be called twice, but called %d times",
			checkTarget, len(m[checkTarget]))
	case m[checkTarget][0].Type != ElementTimecode:
		t.Fatalf("ElementType of %s should be %s, got %s",
			checkTarget, ElementTimecode, m[checkTarget][0].Type)
	}
	switch v, ok := m[checkTarget][0].Value.(uint64); {
	case !ok:
		t.Errorf("Invalid type of data: %T", v)
	case v != 0:
		t.Errorf("The value should be 0, got %d", v)
	}
}

func ExampleMarshal() {
	type EBMLHeader struct {
		DocType            string `ebml:"EBMLDocType"`
		DocTypeVersion     uint64 `ebml:"EBMLDocTypeVersion"`
		DocTypeReadVersion uint64 `ebml:"EBMLDocTypeReadVersion"`
	}
	type TestEBML struct {
		Header EBMLHeader `ebml:"EBML"`
	}
	s := TestEBML{
		Header: EBMLHeader{
			DocType:            "webm",
			DocTypeVersion:     2,
			DocTypeReadVersion: 2,
		},
	}

	var b bytes.Buffer
	if err := Marshal(&s, &b); err != nil {
		panic(err)
	}
	for _, b := range b.Bytes() {
		fmt.Printf("0x%02x, ", int(b))
	}
	// Output:
	// 0x1a, 0x45, 0xdf, 0xa3, 0x8f, 0x42, 0x82, 0x84, 0x77, 0x65, 0x62, 0x6d, 0x42, 0x87, 0x81, 0x02, 0x42, 0x85, 0x81, 0x02,
}

func ExampleWithDataSizeLen() {
	type EBMLHeader struct {
		DocType            string `ebml:"EBMLDocType"`
		DocTypeVersion     uint64 `ebml:"EBMLDocTypeVersion"`
		DocTypeReadVersion uint64 `ebml:"EBMLDocTypeReadVersion"`
	}
	type TestEBML struct {
		Header EBMLHeader `ebml:"EBML"`
	}
	s := TestEBML{
		Header: EBMLHeader{
			DocType:            "webm",
			DocTypeVersion:     2,
			DocTypeReadVersion: 2,
		},
	}

	var b bytes.Buffer
	if err := Marshal(&s, &b, WithDataSizeLen(2)); err != nil {
		panic(err)
	}
	for _, b := range b.Bytes() {
		fmt.Printf("0x%02x, ", int(b))
	}
	// Output:
	// 0x1a, 0x45, 0xdf, 0xa3, 0x40, 0x12, 0x42, 0x82, 0x40, 0x04, 0x77, 0x65, 0x62, 0x6d, 0x42, 0x87, 0x40, 0x01, 0x02, 0x42, 0x85, 0x40, 0x01, 0x02,
}

func TestMarshal_Tag(t *testing.T) {
	tagged := struct {
		DocCustomNamedType string `ebml:"EBMLDocType"`
	}{
		DocCustomNamedType: "hoge",
	}
	untagged := struct {
		EBMLDocType string
	}{
		EBMLDocType: "hoge",
	}

	var bTagged, bUntagged bytes.Buffer
	if err := Marshal(&tagged, &bTagged); err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}
	if err := Marshal(&untagged, &bUntagged); err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}

	if !bytes.Equal(bTagged.Bytes(), bUntagged.Bytes()) {
		t.Errorf("Tagged struct and untagged struct must be marshal-ed to same binary, tagged: %v, untagged: %v", bTagged.Bytes(), bUntagged.Bytes())
	}
}

func TestMarshal_InvalidTag(t *testing.T) {
	input := struct {
		DocCustomNamedType string `ebml:"EBMLDocType,invalidtag"`
	}{
		DocCustomNamedType: "hoge",
	}

	var buf bytes.Buffer
	if err := Marshal(&input, &buf); !errs.Is(err, ErrInvalidTag) {
		t.Errorf("Expected error against invalid tag: '%v', got: '%v'", ErrInvalidTag, err)
	}
}

func TestMarshal_Chan(t *testing.T) {
	expected := []byte{
		0x18, 0x53, 0x80, 0x67, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0x1F, 0x43, 0xB6, 0x75, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xE7, 0x81, 0x01,
		0x1F, 0x43, 0xB6, 0x75, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xE7, 0x81, 0x02,
	}
	type Cluster struct {
		Timecode uint64 `ebml:"Timecode"`
	}

	t.Run("ChanStruct", func(t *testing.T) {
		ch := make(chan Cluster, 100)
		input := &struct {
			Segment struct {
				Cluster chan Cluster `ebml:"Cluster,size=unknown"`
			} `ebml:"Segment,size=unknown"`
		}{}
		input.Segment.Cluster = ch
		ch <- Cluster{Timecode: 0x01}
		ch <- Cluster{Timecode: 0x02}
		close(ch)

		var b bytes.Buffer
		if err := Marshal(input, &b); err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}
		if !bytes.Equal(expected, b.Bytes()) {
			t.Errorf("Marshaled binary doesn't match:\n expected: %v,\n      got: %v", expected, b.Bytes())
		}
	})
	t.Run("ChanStructPtr", func(t *testing.T) {
		input := &struct {
			Segment struct {
				Cluster chan *Cluster `ebml:"Cluster,size=unknown"`
			} `ebml:"Segment,size=unknown"`
		}{}

		t.Run("Valid", func(t *testing.T) {
			ch := make(chan *Cluster, 100)
			input.Segment.Cluster = ch
			ch <- &Cluster{Timecode: 0x01}
			ch <- &Cluster{Timecode: 0x02}
			close(ch)

			var b bytes.Buffer
			if err := Marshal(input, &b); err != nil {
				t.Fatalf("Unexpected error: '%v'", err)
			}
			if !bytes.Equal(expected, b.Bytes()) {
				t.Errorf("Marshaled binary doesn't match:\n expected: %v,\n      got: %v", expected, b.Bytes())
			}
		})
		t.Run("Nil", func(t *testing.T) {
			ch := make(chan *Cluster, 100)
			input.Segment.Cluster = ch
			ch <- nil
			close(ch)

			if err := Marshal(input, &bytes.Buffer{}); !errs.Is(err, ErrIncompatibleType) {
				t.Fatalf("Expected error: '%v', got: '%v'", ErrIncompatibleType, err)
			}
		})
	})
	t.Run("ChanStructSlice", func(t *testing.T) {
		input := &struct {
			Segment struct {
				Cluster chan []Cluster `ebml:"Cluster,size=unknown"`
			} `ebml:"Segment,size=unknown"`
		}{}
		ch := make(chan []Cluster, 100)
		input.Segment.Cluster = ch
		ch <- make([]Cluster, 2)
		close(ch)

		if err := Marshal(input, &bytes.Buffer{}); !errs.Is(err, ErrIncompatibleType) {
			t.Fatalf("Expected error: '%v', got: '%v'", ErrIncompatibleType, err)
		}
	})
}

func TestMarshal_Func(t *testing.T) {
	expected := []byte{
		0x18, 0x53, 0x80, 0x67, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0x1F, 0x43, 0xB6, 0x75, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xE7, 0x81, 0x01,
	}
	type Cluster struct {
		Timecode uint64 `ebml:"Timecode"`
	}

	t.Run("FuncStruct", func(t *testing.T) {
		input := &struct {
			Segment struct {
				Cluster func() Cluster `ebml:"Cluster,size=unknown"`
			} `ebml:"Segment,size=unknown"`
		}{}
		input.Segment.Cluster = func() Cluster {
			return Cluster{Timecode: 0x01}
		}

		var b bytes.Buffer
		if err := Marshal(input, &b); err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}
		if !bytes.Equal(expected, b.Bytes()) {
			t.Errorf("Marshaled binary doesn't match:\n expected: %v,\n      got: %v", expected, b.Bytes())
		}
	})
	t.Run("FuncStructPtr", func(t *testing.T) {
		input := &struct {
			Segment struct {
				Cluster func() *Cluster `ebml:"Cluster,size=unknown"`
			} `ebml:"Segment,size=unknown"`
		}{}

		t.Run("Valid", func(t *testing.T) {
			input.Segment.Cluster = func() *Cluster {
				return &Cluster{Timecode: 0x01}
			}

			var b bytes.Buffer
			if err := Marshal(input, &b); err != nil {
				t.Fatalf("Unexpected error: '%v'", err)
			}
			if !bytes.Equal(expected, b.Bytes()) {
				t.Errorf("Marshaled binary doesn't match:\n expected: %v,\n      got: %v", expected, b.Bytes())
			}
		})
		t.Run("Nil", func(t *testing.T) {
			input.Segment.Cluster = func() *Cluster {
				return nil
			}

			if err := Marshal(input, &bytes.Buffer{}); !errs.Is(err, ErrIncompatibleType) {
				t.Fatalf("Expected error: '%v', got: '%v'", ErrIncompatibleType, err)
			}
		})
	})
	t.Run("FuncStructSlice", func(t *testing.T) {
		input := &struct {
			Segment struct {
				Cluster func() []Cluster `ebml:"Cluster,size=unknown"`
			} `ebml:"Segment,size=unknown"`
		}{}
		input.Segment.Cluster = func() []Cluster {
			return []Cluster{
				{Timecode: 0x01},
				{Timecode: 0x02},
			}
		}

		var b bytes.Buffer
		if err := Marshal(input, &b); err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}
		expected := []byte{
			0x18, 0x53, 0x80, 0x67, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0x1F, 0x43, 0xB6, 0x75, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xE7, 0x81, 0x01,
			0x1F, 0x43, 0xB6, 0x75, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xE7, 0x81, 0x02,
		}
		if !bytes.Equal(expected, b.Bytes()) {
			t.Errorf("Marshaled binary doesn't match:\n expected: %v,\n      got: %v", expected, b.Bytes())
		}
	})
	t.Run("FuncWithError", func(t *testing.T) {
		input := &struct {
			Segment struct {
				Cluster func() (Cluster, error) `ebml:"Cluster,size=unknown"`
			} `ebml:"Segment,size=unknown"`
		}{}

		t.Run("Valid", func(t *testing.T) {
			input.Segment.Cluster = func() (Cluster, error) {
				return Cluster{Timecode: 0x01}, nil
			}

			var b bytes.Buffer
			if err := Marshal(input, &b); err != nil {
				t.Fatalf("Unexpected error: '%v'", err)
			}
			if !bytes.Equal(expected, b.Bytes()) {
				t.Errorf("Marshaled binary doesn't match:\n expected: %v,\n      got: %v", expected, b.Bytes())
			}
		})
		t.Run("Error", func(t *testing.T) {
			expectedErr := errors.New("an error")
			input.Segment.Cluster = func() (Cluster, error) {
				return Cluster{Timecode: 0x01}, expectedErr
			}

			if err := Marshal(input, &bytes.Buffer{}); !errs.Is(err, expectedErr) {
				t.Fatalf("Expected error: '%v', got: '%v'", expectedErr, err)
			}
		})
		t.Run("NonErrorType", func(t *testing.T) {
			input := &struct {
				Segment struct {
					Cluster func() (*Cluster, int) `ebml:"Cluster,size=unknown"`
				} `ebml:"Segment,size=unknown"`
			}{}
			input.Segment.Cluster = func() (*Cluster, int) {
				return nil, 1
			}

			if err := Marshal(input, &bytes.Buffer{}); !errs.Is(err, ErrIncompatibleType) {
				t.Fatalf("Expected error: '%v', got: '%v'", ErrIncompatibleType, err)
			}
		})
	})
	t.Run("InvalidFunc", func(t *testing.T) {
		input := &struct {
			Segment struct {
				Cluster func() `ebml:"Cluster,size=unknown"`
			} `ebml:"Segment,size=unknown"`
		}{}
		input.Segment.Cluster = func() {}

		if err := Marshal(input, &bytes.Buffer{}); !errs.Is(err, ErrIncompatibleType) {
			t.Fatalf("Expected error: '%v', got: '%v'", ErrIncompatibleType, err)
		}
	})
}

func BenchmarkMarshal(b *testing.B) {
	type EBMLHeader struct {
		DocType            string `ebml:"EBMLDocType"`
		DocTypeVersion     uint64 `ebml:"EBMLDocTypeVersion"`
		DocTypeReadVersion uint64 `ebml:"EBMLDocTypeReadVersion"`
	}
	type TestEBML struct {
		Header EBMLHeader `ebml:"EBML"`
	}
	s := TestEBML{
		Header: EBMLHeader{
			DocType:            "webm",
			DocTypeVersion:     2,
			DocTypeReadVersion: 2,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		if err := Marshal(&s, &buf); err != nil {
			b.Fatalf("Unexpected error: '%v'", err)
		}
	}
}

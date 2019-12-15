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
	"io"
	"reflect"
	"testing"
)

func ExampleUnmarshal() {
	TestBinary := []byte{
		0x1a, 0x45, 0xdf, 0xa3, // EBML
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, // 0x10
		0x42, 0x82, 0x85, 0x77, 0x65, 0x62, 0x6d, 0x00,
		0x42, 0x87, 0x81, 0x02, 0x42, 0x85, 0x81, 0x02,
	}
	type TestEBML struct {
		Header struct {
			DocType            string `ebml:"EBMLDocType"`
			DocTypeVersion     uint64 `ebml:"EBMLDocTypeVersion"`
			DocTypeReadVersion uint64 `ebml:"EBMLDocTypeReadVersion"`
		} `ebml:"EBML"`
	}

	r := bytes.NewReader(TestBinary)

	var ret TestEBML
	if err := Unmarshal(r, &ret); err != nil {
		fmt.Printf("error: %+v\n", err)
	}
	fmt.Println(ret)

	// Output: {{webm 2 2}}
}

func TestUnmarshal_Convert(t *testing.T) {
	cases := map[string]struct {
		b        []byte
		expected interface{}
	}{
		"UInt64ToUInt64": {
			[]byte{0x42, 0x87, 0x81, 0x02},
			struct {
				DocTypeVersion uint64 `ebml:"EBMLDocTypeVersion"`
			}{2},
		},
		"UInt64ToUInt32": {
			[]byte{0x42, 0x87, 0x81, 0x02},
			struct {
				DocTypeVersion uint32 `ebml:"EBMLDocTypeVersion"`
			}{2},
		},
		"UInt64ToUInt16": {
			[]byte{0x42, 0x87, 0x81, 0x02},
			struct {
				DocTypeVersion uint16 `ebml:"EBMLDocTypeVersion"`
			}{2},
		},
		"UInt64ToUInt8": {
			[]byte{0x42, 0x87, 0x81, 0x02},
			struct {
				DocTypeVersion uint8 `ebml:"EBMLDocTypeVersion"`
			}{2},
		},
		"UInt64ToUInt": {
			[]byte{0x42, 0x87, 0x81, 0x02},
			struct {
				DocTypeVersion uint `ebml:"EBMLDocTypeVersion"`
			}{2},
		},
		"Int64ToInt64": {
			[]byte{0xFB, 0x81, 0xFF},
			struct {
				ReferenceBlock int64 `ebml:"ReferenceBlock"`
			}{-1},
		},
		"Int64ToInt32": {
			[]byte{0xFB, 0x81, 0xFF},
			struct {
				ReferenceBlock int32 `ebml:"ReferenceBlock"`
			}{-1},
		},
		"Int64ToInt16": {
			[]byte{0xFB, 0x81, 0xFF},
			struct {
				ReferenceBlock int16 `ebml:"ReferenceBlock"`
			}{-1},
		},
		"Int64ToInt8": {
			[]byte{0xFB, 0x81, 0xFF},
			struct {
				ReferenceBlock int8 `ebml:"ReferenceBlock"`
			}{-1},
		},
		"Int64ToInt": {
			[]byte{0xFB, 0x81, 0xFF},
			struct {
				ReferenceBlock int `ebml:"ReferenceBlock"`
			}{-1},
		},
		"Float64ToFloat64": {
			[]byte{0x44, 0x89, 0x84, 0x00, 0x00, 0x00, 0x00},
			struct {
				Duration float64 `ebml:"Duration"`
			}{0.0},
		},
		"Float64ToFloat32": {
			[]byte{0x44, 0x89, 0x84, 0x00, 0x00, 0x00, 0x00},
			struct {
				Duration float32 `ebml:"Duration"`
			}{0.0},
		},
		"UInt64ToUInt32Slice": {
			[]byte{0x42, 0x87, 0x81, 0x02},
			struct {
				DocTypeVersion []uint32 `ebml:"EBMLDocTypeVersion"`
			}{[]uint32{2}},
		},
		"Int64ToInt32Slice": {
			[]byte{0xFB, 0x81, 0xFF},
			struct {
				ReferenceBlock []int32 `ebml:"ReferenceBlock"`
			}{[]int32{-1}},
		},
		"Float64ToFloat32Slice": {
			[]byte{0x44, 0x89, 0x84, 0x00, 0x00, 0x00, 0x00},
			struct {
				Duration []float32 `ebml:"Duration"`
			}{[]float32{0.0}},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			ret := reflect.New(reflect.ValueOf(c.expected).Type())
			if err := Unmarshal(bytes.NewReader(c.b), ret.Interface()); err != nil {
				t.Fatalf("Unexpected error: %v\n", err)
			}

			if !reflect.DeepEqual(c.expected, ret.Elem().Interface()) {
				t.Errorf("Unexpected convert result, expected: %v, got %v",
					c.expected, ret.Elem().Interface())
			}
		})
	}
}

func TestUnmarshal_OptionError(t *testing.T) {
	errExpected := errors.New("an error")
	err := Unmarshal(&bytes.Buffer{}, &struct{}{},
		func(*UnmarshalOptions) error {
			return errExpected
		},
	)
	if err != errExpected {
		t.Errorf("Unexpected error for failing UnmarshalOption, expected: %v, got: %v", errExpected, err)
	}
}

func TestUnmarshal_WithElementReadHooks(t *testing.T) {
	TestBinary := []byte{
		0x18, 0x53, 0x80, 0x67, 0xa6, // Segment
		0x1c, 0x53, 0xbb, 0x6b, 0x80, // Cues (empty)
		0x16, 0x54, 0xae, 0x6b, 0x9c, // Tracks
		0xae, 0x8c, // TrackEntry[0]
		0x53, 0x6e, 0x86, 0x56, 0x69, 0x64, 0x65, 0x6f, 0x00, // Name=Video
		0xd7, 0x81, 0x01, // TrackNumber=1
		0xae, 0x8c, // TrackEntry[1]
		0x53, 0x6e, 0x86, 0x41, 0x75, 0x64, 0x69, 0x6f, 0x00, // Name=Audio
		0xd7, 0x81, 0x02, // TrackNumber=2
	}

	type TestEBML struct {
		Segment struct {
			Tracks struct {
				TrackEntry []struct {
					Name        string `ebml:"Name,omitempty"`
					TrackNumber uint64 `ebml:"TrackNumber"`
				} `ebml:"TrackEntry"`
			} `ebml:"Tracks"`
		} `ebml:"Segment"`
	}

	r := bytes.NewReader(TestBinary)

	var ret TestEBML
	m := make(map[string][]*Element)
	hook := withElementMap(m)
	if err := Unmarshal(r, &ret, WithElementReadHooks(hook)); err != nil {
		t.Errorf("Unexpected error: %+v", err)
	}

	// Verify positions of elements
	expected := map[string][]uint64{
		"Segment":                               {0},
		"Segment.Tracks":                        {10},
		"Segment.Tracks.TrackEntry":             {15, 29},
		"Segment.Tracks.TrackEntry.Name":        {17, 31},
		"Segment.Tracks.TrackEntry.TrackNumber": {26, 40},
	}
	posMap := elementPositionMap(m)
	if !reflect.DeepEqual(expected, posMap) {
		t.Errorf("Unexpected read hook positions, \nexpected: %v, \n     got: %v", expected, posMap)
	}
}

func TestUnmarshal_Tag(t *testing.T) {
	var tagged struct {
		DocCustomNamedType string `ebml:"EBMLDocType"`
	}
	var untagged struct {
		EBMLDocType string
	}

	b := []byte{0x42, 0x82, 0x85, 0x68, 0x6F, 0x67, 0x65, 0x00}

	if err := Unmarshal(bytes.NewBuffer(b), &tagged); err != nil {
		t.Fatalf("Unexpected error: %+v", err)
	}
	if err := Unmarshal(bytes.NewBuffer(b), &untagged); err != nil {
		t.Fatalf("Unexpected error: %+v", err)
	}

	if tagged.DocCustomNamedType != untagged.EBMLDocType {
		t.Errorf("Unmarshal result to tagged and and untagged struct must be same, tagged: %v, untagged: %v", tagged, untagged)
	}
}

func TestUnmarshal_Error(t *testing.T) {
	type TestEBML struct {
		Header struct {
		} `ebml:"EBML"`
	}
	t.Run("NilValue", func(t *testing.T) {
		if err := Unmarshal(bytes.NewBuffer([]byte{}), nil); err != errIndefiniteType {
			t.Errorf("Unexpected error, %v, got %v\n", errIndefiniteType, err)
		}
	})
	t.Run("Short", func(t *testing.T) {
		TestBinaries := map[string][]byte{
			"ElementID": {0x1a, 0x45, 0xdf},
			"DataSize":  {0x42, 0x86, 0x40},
			"UInt(0)":   {0x42, 0x86, 0x84},
			"UInt":      {0x42, 0x86, 0x84, 0x00},
			"Float(0)":  {0x44, 0x89, 0x84},
			"Float":     {0x44, 0x89, 0x84, 0x00},
			"String(0)": {0x42, 0x82, 0x84},
			"String":    {0x42, 0x82, 0x84, 0x00},
		}
		for name, b := range TestBinaries {
			t.Run(name, func(t *testing.T) {
				var val TestEBML
				if err := Unmarshal(bytes.NewBuffer(b), &val); err != io.ErrUnexpectedEOF {
					t.Errorf("Unexpected error, expected: %v, got: %v\n", io.ErrUnexpectedEOF, err)
				}
			})
		}
	})
	t.Run("Incompatible", func(t *testing.T) {
		cases := map[string]struct {
			b   []byte
			ret interface{}
			err error
		}{
			"UInt64ToInt64": {
				b: []byte{0x42, 0x87, 0x81, 0x02},
				ret: &struct {
					DocTypeVersion int64 `ebml:"EBMLDocTypeVersion"`
				}{},
				err: errIncompatibleType,
			},
			"Int64ToUInt64": {
				b: []byte{0xFB, 0x81, 0xFF},
				ret: &struct {
					ReferenceBlock uint64 `ebml:"ReferenceBlock"`
				}{},
				err: errIncompatibleType,
			},
			"Float64ToInt64": {
				b: []byte{0x44, 0x89, 0x84, 0x00, 0x00, 0x00, 0x00},
				ret: &struct {
					Duration int64 `ebml:"Duration"`
				}{},
				err: errIncompatibleType,
			},
			"StringToInt64": {
				b: []byte{0x42, 0x82, 0x85, 0x77, 0x65, 0x62, 0x6d, 0x00},
				ret: &struct {
					EBMLDocType int64 `ebml:"EBMLDocType"`
				}{},
				err: errIncompatibleType,
			},
			"UInt64ToInt64Slice": {
				b: []byte{0x42, 0x87, 0x81, 0x02},
				ret: &struct {
					DocTypeVersion []int64 `ebml:"EBMLDocTypeVersion"`
				}{},
				err: errIncompatibleType,
			},
			"Int64ToUInt64Slice": {
				b: []byte{0xFB, 0x81, 0xFF},
				ret: &struct {
					ReferenceBlock []uint64 `ebml:"ReferenceBlock"`
				}{},
				err: errIncompatibleType,
			},
			"Float64ToInt64Slice": {
				b: []byte{0x44, 0x89, 0x84, 0x00, 0x00, 0x00, 0x00},
				ret: &struct {
					Duration []int64 `ebml:"Duration"`
				}{},
				err: errIncompatibleType,
			},
			"StringToInt64Slice": {
				b: []byte{0x42, 0x82, 0x85, 0x77, 0x65, 0x62, 0x6d, 0x00},
				ret: &struct {
					EBMLDocType []int64 `ebml:"EBMLDocType"`
				}{},
				err: errIncompatibleType,
			},
		}
		for name, c := range cases {
			t.Run(name, func(t *testing.T) {
				if err := Unmarshal(bytes.NewBuffer(c.b), c.ret); err != c.err {
					t.Errorf("Unexpected error, expected: %v, got: %v\n", c.err, err)
				}
			})
		}
	})

}

func BenchmarkUnmarshal(b *testing.B) {
	TestBinary := []byte{
		0x1a, 0x45, 0xdf, 0xa3, // EBML
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, // 0x10
		0x42, 0x82, 0x85, 0x77, 0x65, 0x62, 0x6d, 0x00,
		0x42, 0x87, 0x81, 0x02, 0x42, 0x85, 0x81, 0x02,
	}
	type TestEBML struct {
		Header struct {
			DocType            string `ebml:"EBMLDocType"`
			DocTypeVersion     uint64 `ebml:"EBMLDocTypeVersion"`
			DocTypeReadVersion uint64 `ebml:"EBMLDocTypeReadVersion"`
		} `ebml:"EBML"`
	}

	var ret TestEBML

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := Unmarshal(bytes.NewReader(TestBinary), &ret); err != nil {
			b.Fatalf("Unexpected error: %+v", err)
		}
	}
}

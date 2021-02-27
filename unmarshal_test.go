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
	"time"

	"github.com/at-wat/ebml-go/internal/errs"
)

func ExampleUnmarshal() {
	TestBinary := []byte{
		0x1a, 0x45, 0xdf, 0xa3, // EBML
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, // 0x10
		0x42, 0x82, 0x85, 0x77, 0x65, 0x62, 0x6d, 0x00, // EBMLDocType = webm
		0x42, 0x87, 0x81, 0x02, // DocTypeVersion = 2
		0x42, 0x85, 0x81, 0x02, // DocTypeReadVersion = 2
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
		fmt.Printf("error: %v\n", err)
	}
	fmt.Println(ret)

	// Output: {{webm 2 2}}
}

func TestUnmarshal_MultipleUnknownSize(t *testing.T) {
	b := []byte{
		0x18, 0x53, 0x80, 0x67, // Segment
		0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0x1F, 0x43, 0xB6, 0x75, // Cluster
		0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0x42, 0x87, 0x81, 0x01,
		0x1F, 0x43, 0xB6, 0x75, // Cluster
		0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0x42, 0x87, 0x81, 0x02,
	}
	type Cluster struct {
		DocTypeVersion uint64 `ebml:"EBMLDocTypeVersion"`
	}
	type Segment struct {
		Cluster []Cluster `ebml:"Cluster"`
	}
	type TestEBML struct {
		Segment Segment `ebml:"Segment"`
	}
	expected := TestEBML{
		Segment: Segment{
			Cluster: []Cluster{{0x01}, {0x02}},
		},
	}

	var ret TestEBML
	if err := Unmarshal(bytes.NewReader(b), &ret); err != nil {
		t.Fatalf("Unexpected error: '%v'\n", err)
	}
	if !reflect.DeepEqual(expected, ret) {
		t.Errorf("Expected result: %v, got: %v", expected, ret)
	}
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
		"UInt64ToUInt64Slice": {
			[]byte{0x42, 0x87, 0x81, 0x02},
			struct {
				DocTypeVersion []uint64 `ebml:"EBMLDocTypeVersion"`
			}{[]uint64{2}},
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
				t.Fatalf("Unexpected error: '%v'\n", err)
			}

			if !reflect.DeepEqual(c.expected, ret.Elem().Interface()) {
				t.Errorf("Expected convert result: %v, got %v",
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
		t.Errorf("Expected error against failing UnmarshalOption: '%v', got: '%v'", errExpected, err)
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
		t.Errorf("Unexpected error: '%v'", err)
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
	checkTarget := "Segment.Tracks.TrackEntry.Name"
	switch {
	case len(m[checkTarget]) != 2:
		t.Fatalf("%s read hook should be called twice, but called %d times",
			checkTarget, len(m[checkTarget]))
	case m[checkTarget][0].Type != ElementName:
		t.Fatalf("ElementType of %s should be %s, got %s",
			checkTarget, ElementName, m[checkTarget][0].Type)
	}
	switch v, ok := m[checkTarget][0].Value.(string); {
	case !ok:
		t.Errorf("Invalid type of data: %T", v)
	case v != "Video":
		t.Errorf("The value should be Video, got %s", v)
	}
}

func TestUnmarshal_Chan(t *testing.T) {
	TestBinary := []byte{
		0x18, 0x53, 0x80, 0x67, 0x8f, // Segment
		0x16, 0x54, 0xae, 0x6b, 0x8a, // Tracks
		0xae, 0x83, // TrackEntry[0]
		0xd7, 0x81, 0x01, // TrackNumber=1
		0xae, 0x83, // TrackEntry[0]
		0xd7, 0x81, 0x02, // TrackNumber=2
	}
	type TestEBML struct {
		Segment struct {
			Tracks struct {
				TrackEntry struct {
					TrackNumber chan uint64 `ebml:"TrackNumber"`
				} `ebml:"TrackEntry"`
			} `ebml:"Tracks"`
		} `ebml:"Segment"`
	}

	var ret TestEBML
	ch := make(chan uint64, 100)
	ret.Segment.Tracks.TrackEntry.TrackNumber = ch

	done := make(chan struct{})
	go func() {
		select {
		case <-time.After(5 * time.Second):
			panic("test timeout")
		case <-done:
		}
	}()
	if err := Unmarshal(bytes.NewReader(TestBinary), &ret); err != nil {
		t.Errorf("Unexpected error: '%v'", err)
	}
	close(done)
	if len(ch) != 2 {
		t.Fatalf("Element chan should be sent twice, but sent %d times", len(ch))
	}
	if v := <-ch; v != 1 {
		t.Errorf("First value should be 1, got %d", v)
	}
	if v := <-ch; v != 2 {
		t.Errorf("Second value should be 2, got %d", v)
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
		t.Fatalf("Unexpected error: '%v'", err)
	}
	if err := Unmarshal(bytes.NewBuffer(b), &untagged); err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}

	if tagged.DocCustomNamedType != untagged.EBMLDocType {
		t.Errorf("Unmarshal result to tagged and and untagged struct must be same, tagged: %v, untagged: %v", tagged, untagged)
	}
}

func TestUnmarshal_Map(t *testing.T) {
	b := []byte{
		0x1A, 0x45, 0xDF, 0xA3, 0x8C,
		0x42, 0x82, 0x85, 0x68, 0x6F, 0x67, 0x65, 0x00,
		0xEC, 0x82, 0x00, 0x00,
		0x18, 0x53, 0x80, 0x67, 0xFF,
		0x1F, 0x43, 0xB6, 0x75, 0x80,
		0x1F, 0x43, 0xB6, 0x75, 0x80,
		0x1F, 0x43, 0xB6, 0x75, 0x80,
	}
	expected := map[string]interface{}{
		"EBML": map[string]interface{}{
			"EBMLDocType": "hoge",
			"Void":        []uint8{0, 0},
		},
		"Segment": map[string]interface{}{
			"Cluster": []interface{}{
				map[string]interface{}{},
				map[string]interface{}{},
				map[string]interface{}{},
			},
		},
	}

	t.Run("AllocatedMap", func(t *testing.T) {
		ret := make(map[string]interface{})
		if err := Unmarshal(bytes.NewBuffer(b), &ret); err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}

		if !reflect.DeepEqual(expected, ret) {
			t.Errorf("Unmarshal to map differs from expected:\n%#+v\ngot:\n%#+v", expected, ret)
		}
	})

	t.Run("NilMap", func(t *testing.T) {
		var ret map[string]interface{}
		if err := Unmarshal(bytes.NewBuffer(b), &ret); err != nil {
			t.Fatalf("Unexpected error: '%v'", err)
		}

		if !reflect.DeepEqual(expected, ret) {
			t.Errorf("Unmarshal to map differs from expected:\n%#+v\ngot:\n%#+v", expected, ret)
		}
	})
}

func TestUnmarshal_IgnoreUnknown(t *testing.T) {
	b := []byte{
		0x1A, 0x45, 0xDF, 0xA3, 0x8A,
		0x42, 0x82, 0x85, 0x68, 0x6F, 0x67, 0x65, 0x00,
		0x81, 0x81, // 0x81 is not defined in Matroska v4
		0x18, 0x53, 0x80, 0x67, 0xFF,
		0x1F, 0x43, 0xB6, 0x75, 0x80,
		0x1F, 0x43, 0xB6, 0x75, 0x80,
	}
	expected := map[string]interface{}{
		"EBML": map[string]interface{}{
			"EBMLDocType": "hoge",
		},
		"Segment": map[string]interface{}{
			"Cluster": []interface{}{
				map[string]interface{}{},
				map[string]interface{}{},
			},
		},
	}

	ret := make(map[string]interface{})
	if err := Unmarshal(bytes.NewBuffer(b), &ret, WithIgnoreUnknown(true)); err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}

	if !reflect.DeepEqual(expected, ret) {
		t.Errorf("Unmarshal with IgnoreUnknown differs from expected:\n%#+v\ngot:\n%#+v", expected, ret)
	}
}

func TestUnmarshal_Error(t *testing.T) {
	type TestEBML struct {
		Header struct {
		} `ebml:"EBML"`
	}
	t.Run("NilValue", func(t *testing.T) {
		if err := Unmarshal(bytes.NewBuffer([]byte{}), nil); !errs.Is(err, ErrIndefiniteType) {
			t.Errorf("Expected error: '%v', got: '%v'", ErrIndefiniteType, err)
		}
	})
	t.Run("NonPtr", func(t *testing.T) {
		if err := Unmarshal(bytes.NewBuffer([]byte{}), struct{}{}); !errs.Is(err, ErrIncompatibleType) {
			t.Errorf("Expected error: '%v', got: '%v'", ErrIncompatibleType, err)
		}
	})
	t.Run("UnknownElementName", func(t *testing.T) {
		input := &struct {
			Header struct {
			} `ebml:"Unknown"`
		}{}
		if err := Unmarshal(bytes.NewBuffer([]byte{}), input); !errs.Is(err, ErrUnknownElementName) {
			t.Errorf("Expected error: '%v', got: '%v'", ErrUnknownElementName, err)
		}
	})
	t.Run("InvalidTag", func(t *testing.T) {
		input := &struct {
			Header struct {
			} `ebml:"EBML,ivalid"`
		}{}
		if err := Unmarshal(bytes.NewBuffer([]byte{}), input); !errs.Is(err, ErrInvalidTag) {
			t.Errorf("Expected error: '%v', got: '%v'", ErrInvalidTag, err)
		}
	})
	t.Run("UnknownElement", func(t *testing.T) {
		input := &TestEBML{}
		b := []byte{0x81}
		if err := Unmarshal(bytes.NewBuffer(b), input); !errs.Is(err, ErrUnknownElement) {
			t.Errorf("Expected error: '%v', got: '%v'", ErrUnknownElement, err)
		}
	})
	t.Run("NonStaticUnknownElementWithIgnoreUnknown", func(t *testing.T) {
		input := &TestEBML{}
		b := []byte{0x81, 0xFF}
		if err := Unmarshal(
			bytes.NewBuffer(b), input, WithIgnoreUnknown(true),
		); err != nil {
			t.Errorf("Unexpected error: '%v'", err)
		}
	})
	t.Run("ShortUnknownElementWithIgnoreUnknown", func(t *testing.T) {
		input := &TestEBML{}
		b := []byte{0x81, 0x85, 0x00}
		if err := Unmarshal(
			bytes.NewBuffer(b), input, WithIgnoreUnknown(true),
		); err != nil {
			t.Errorf("Unexpected error: '%v'", err)
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
				if err := Unmarshal(bytes.NewBuffer(b), &val); !errs.Is(err, io.ErrUnexpectedEOF) {
					t.Errorf("Expected error: '%v', got: '%v'", io.ErrUnexpectedEOF, err)
				}
			})
		}
	})
	t.Run("ErrorPropagation", func(t *testing.T) {
		TestBinaries := map[string][]byte{
			"UInt":       {0x42, 0x86, 0x84, 0x00, 0x00, 0x00, 0x00},
			"Float":      {0x44, 0x89, 0x84, 0x00, 0x00, 0x00, 0x00},
			"String":     {0x42, 0x82, 0x84, 0x00, 0x00, 0x00, 0x00},
			"Block":      {0xA3, 0x85, 0x81, 0x00, 0x00, 0x00, 0x00},
			"BlockXiph":  {0xA3, 0x88, 0x81, 0x00, 0x00, 0x02, 0x01, 0x01, 0x00, 0x00},
			"BlockFixed": {0xA3, 0x87, 0x81, 0x00, 0x00, 0x04, 0x01, 0x00, 0x00},
			"BlockEBML":  {0xA3, 0x88, 0x98, 0x00, 0x00, 0x06, 0x01, 0x81, 0x00, 0x00},
		}
		for name, b := range TestBinaries {
			t.Run(name, func(t *testing.T) {
				for i := 1; i < len(b)-1; i++ {
					var val TestEBML
					r := &delayedBrokenReader{b: b, limit: i}
					if err := Unmarshal(r, &val); !errs.Is(err, io.ErrClosedPipe) {
						t.Errorf("Error is not propagated from Reader, limit: %d, expected: '%v', got: '%v'", i, io.ErrClosedPipe, err)
					}
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
				err: ErrIncompatibleType,
			},
			"Int64ToUInt64": {
				b: []byte{0xFB, 0x81, 0xFF},
				ret: &struct {
					ReferenceBlock uint64 `ebml:"ReferenceBlock"`
				}{},
				err: ErrIncompatibleType,
			},
			"Float64ToInt64": {
				b: []byte{0x44, 0x89, 0x84, 0x00, 0x00, 0x00, 0x00},
				ret: &struct {
					Duration int64 `ebml:"Duration"`
				}{},
				err: ErrIncompatibleType,
			},
			"StringToInt64": {
				b: []byte{0x42, 0x82, 0x85, 0x77, 0x65, 0x62, 0x6d, 0x00},
				ret: &struct {
					EBMLDocType int64 `ebml:"EBMLDocType"`
				}{},
				err: ErrIncompatibleType,
			},
			"UInt64ToInt64Slice": {
				b: []byte{0x42, 0x87, 0x81, 0x02},
				ret: &struct {
					DocTypeVersion []int64 `ebml:"EBMLDocTypeVersion"`
				}{},
				err: ErrIncompatibleType,
			},
			"Int64ToUInt64Slice": {
				b: []byte{0xFB, 0x81, 0xFF},
				ret: &struct {
					ReferenceBlock []uint64 `ebml:"ReferenceBlock"`
				}{},
				err: ErrIncompatibleType,
			},
			"Float64ToInt64Slice": {
				b: []byte{0x44, 0x89, 0x84, 0x00, 0x00, 0x00, 0x00},
				ret: &struct {
					Duration []int64 `ebml:"Duration"`
				}{},
				err: ErrIncompatibleType,
			},
			"StringToInt64Slice": {
				b: []byte{0x42, 0x82, 0x85, 0x77, 0x65, 0x62, 0x6d, 0x00},
				ret: &struct {
					EBMLDocType []int64 `ebml:"EBMLDocType"`
				}{},
				err: ErrIncompatibleType,
			},
		}
		for name, c := range cases {
			t.Run(name, func(t *testing.T) {
				if err := Unmarshal(bytes.NewBuffer(c.b), c.ret); !errs.Is(err, c.err) {
					t.Errorf("Expected error: '%v', got: '%v'", c.err, err)
				}
			})
		}
	})
}

func ExampleUnmarshal_partial() {
	TestBinary := []byte{
		0x1a, 0x45, 0xdf, 0xa3, 0x84, // EBML
		0x42, 0x87, 0x81, 0x02, // DocTypeVersion = 2
		0x18, 0x53, 0x80, 0x67, 0xFF, // Segment
		0x16, 0x54, 0xae, 0x6b, 0x85, // Tracks
		0xae, 0x83, // TrackEntry[0]
		0xd7, 0x81, 0x01, // TrackNumber=1
		0x1F, 0x43, 0xB6, 0x75, 0xFF, // Cluster
		0xE7, 0x81, 0x00, // Timecode
		0xA3, 0x86, 0x81, 0x00, 0x00, 0x88, 0xAA, 0xCC, // SimpleBlock
	}

	type TestHeader struct {
		Header  map[string]interface{} `ebml:"EBML"`
		Segment struct {
			Tracks map[string]interface{} `ebml:"Tracks,stop"` // Stop unmarshalling after reading this element
		}
	}
	type TestClusters struct {
		Cluster []struct {
			Timecode    uint64
			SimpleBlock []Block
		}
	}

	r := bytes.NewReader(TestBinary)

	var header TestHeader
	if err := Unmarshal(r, &header); !errs.Is(err, ErrReadStopped) {
		panic("Unmarshal failed")
	}
	fmt.Printf("First unmarshal: %v\n", header)

	var clusters TestClusters
	if err := Unmarshal(r, &clusters); err != nil {
		panic("Unmarshal failed")
	}
	fmt.Printf("Second unmarshal: %v\n", clusters)

	// Output:
	// First unmarshal: {map[EBMLDocTypeVersion:2] {map[TrackEntry:map[TrackNumber:1]]}}
	// Second unmarshal: {[{0 [{1 0 true true 0 false [[170 204]]}]}]}
}

func BenchmarkUnmarshal(b *testing.B) {
	TestBinary := []byte{
		0x1a, 0x45, 0xdf, 0xa3, // EBML
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, // 0x10
		0x42, 0x82, 0x85, 0x77, 0x65, 0x62, 0x6d, 0x00, // DocType = webm
		0x42, 0x87, 0x81, 0x02, // DocTypeVersion = 2
		0x42, 0x85, 0x81, 0x02, // DocTypeReadVersion = 2
		0x18, 0x53, 0x80, 0x67, 0xFF, // Segment
		0x1F, 0x43, 0xB6, 0x75, 0xFF, // Cluster
		0xE7, 0x81, 0x00, // Timecode
		0xA3, 0x86, 0x81, 0x00, 0x00, 0x88, 0xAA, 0xCC, // SimpleBlock
		0xA3, 0x86, 0x81, 0x00, 0x10, 0x88, 0xAA, 0xCC, // SimpleBlock
		0xA3, 0x86, 0x81, 0x00, 0x20, 0x88, 0xAA, 0xCC, // SimpleBlock
		0x1F, 0x43, 0xB6, 0x75, 0xFF, // Cluster
		0xE7, 0x81, 0x10, // Timecode
		0xA3, 0x86, 0x81, 0x00, 0x00, 0x88, 0xAA, 0xCC, // SimpleBlock
		0xA3, 0x86, 0x81, 0x00, 0x10, 0x88, 0xAA, 0xCC, // SimpleBlock
		0xA3, 0x86, 0x81, 0x00, 0x20, 0x88, 0xAA, 0xCC, // SimpleBlock
	}
	type TestEBML struct {
		Header struct {
			DocType            string `ebml:"EBMLDocType"`
			DocTypeVersion     uint64 `ebml:"EBMLDocTypeVersion"`
			DocTypeReadVersion uint64 `ebml:"EBMLDocTypeReadVersion"`
		} `ebml:"EBML"`
		Segment []struct {
			Cluster struct {
				Timecode    uint64
				SimpleBlock []Block
			}
		}
	}

	var ret TestEBML

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := Unmarshal(bytes.NewReader(TestBinary), &ret); err != nil {
			b.Fatalf("Unexpected error: '%v'", err)
		}
	}
}

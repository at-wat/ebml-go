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
	"testing"
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

	var str string
	var uinteger uint64

	testCases := map[string]struct {
		input    interface{}
		expected []byte
	}{
		"Omitempty": {
			&struct{ EBML TestOmitempty }{},
			[]byte{
				0x1a, 0x45, 0xDF, 0xA3, 0x80,
			},
		},
		"NoOmitempty": {
			&struct{ EBML TestNoOmitempty }{},
			[]byte{
				0x1A, 0x45, 0xDF, 0xA3, 0x8B,
				0x42, 0x82, 0x81, 0x00,
				0x42, 0x87, 0x81, 0x00,
				0x53, 0xAB, 0x80,
			},
		},
		"Sized": {
			&struct{ EBML TestSized }{TestSized{"a", 1, 0.0, 0.0, []byte{0x01}}},
			[]byte{
				0x1A, 0x45, 0xDF, 0xA3, 0xA2,
				0x42, 0x82, 0x83, 0x61, 0x00, 0x00,
				0x42, 0x87, 0x82, 0x00, 0x01,
				0x44, 0x89, 0x88, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x44, 0x89, 0x84, 0x00, 0x00, 0x00, 0x00,
				0x53, 0xAB, 0x82, 0x01, 0x00,
			},
		},
		"SizedAndOverflow": {
			&struct{ EBML TestSized }{TestSized{"abc", 0x012345, 0.0, 0.0, []byte{0x01, 0x02, 0x03}}},
			[]byte{
				0x1A, 0x45, 0xDF, 0xA3, 0xA5,
				0x42, 0x82, 0x84, 0x61, 0x62, 0x63, 0x00,
				0x42, 0x87, 0x83, 0x01, 0x23, 0x45,
				0x44, 0x89, 0x88, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x44, 0x89, 0x84, 0x00, 0x00, 0x00, 0x00,
				0x53, 0xAB, 0x83, 0x01, 0x02, 0x03,
			},
		},
		"Ptr": {
			&struct{ EBML TestPtr }{TestPtr{&str, &uinteger}},
			[]byte{
				0x1A, 0x45, 0xDF, 0xA3, 0x88,
				0x42, 0x82, 0x81, 0x00,
				0x42, 0x87, 0x81, 0x00,
			},
		},
		"PtrOmitempty": {
			&struct{ EBML TestPtrOmitempty }{TestPtrOmitempty{&str, &uinteger}},
			[]byte{
				0x1A, 0x45, 0xDF, 0xA3, 0x80,
			},
		},
		"Interface": {
			&struct{ EBML TestInterface }{TestInterface{str, uinteger}},
			[]byte{
				0x1A, 0x45, 0xDF, 0xA3, 0x88,
				0x42, 0x82, 0x81, 0x00,
				0x42, 0x87, 0x81, 0x00,
			},
		},
		"InterfacePtr": {
			&struct{ EBML TestInterface }{TestInterface{&str, &uinteger}},
			[]byte{
				0x1A, 0x45, 0xDF, 0xA3, 0x88,
				0x42, 0x82, 0x81, 0x00,
				0x42, 0x87, 0x81, 0x00,
			},
		},
	}

	for n, c := range testCases {
		t.Run(n, func(t *testing.T) {
			var b bytes.Buffer
			if err := Marshal(c.input, &b); err != nil {
				t.Fatalf("error: %+v\n", err)
			}
			if !bytes.Equal(c.expected, b.Bytes()) {
				t.Errorf("Marshaled binary doesn't match:\n expected: %v,\n      got: %v", c.expected, b.Bytes())
			}
		})
	}
}

func TestMarshal_OptionError(t *testing.T) {
	errExpected := errors.New("an error")
	err := Marshal(&struct{}{}, &bytes.Buffer{},
		func(*MarshalOptions) error {
			return errExpected
		},
	)
	if err != errExpected {
		t.Errorf("Unexpected error for failing MarshalOption, expected: %v, got: %v", errExpected, err)
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
	// 0x1a, 0x45, 0xdf, 0xa3, 0x90, 0x42, 0x82, 0x85, 0x77, 0x65, 0x62, 0x6d, 0x00, 0x42, 0x87, 0x81, 0x02, 0x42, 0x85, 0x81, 0x02,
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
	// 0x1a, 0x45, 0xdf, 0xa3, 0x40, 0x13, 0x42, 0x82, 0x40, 0x05, 0x77, 0x65, 0x62, 0x6d, 0x00, 0x42, 0x87, 0x40, 0x01, 0x02, 0x42, 0x85, 0x40, 0x01, 0x02,
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
		t.Fatalf("error: %+v\n", err)
	}
	if err := Marshal(&untagged, &bUntagged); err != nil {
		t.Fatalf("error: %+v\n", err)
	}

	if !bytes.Equal(bTagged.Bytes(), bUntagged.Bytes()) {
		t.Errorf("Tagged struct and untagged struct must be marshal-ed to same binary, tagged: %v, untagged: %v", bTagged.Bytes(), bUntagged.Bytes())
	}
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
			b.Fatalf("error: %+v\n", err)
		}
	}
}

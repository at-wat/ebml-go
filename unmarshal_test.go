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
	"fmt"
	"testing"
)

func ExampleUnmarshal() {
	TestBinary := []byte{
		0x1a, 0x45, 0xdf, 0xa3, // EBML
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, // 0x10
		0x42, 0x82, 0x85, 0x77, 0x65, 0x62, 0x6d, 0x00, // EBMLDocType
		0x42, 0x87, 0x81, 0x02, // EBMLDocTypeVersion
		0x42, 0x85, 0x81, 0x02, // EBMLDocTypeReadVersion
		0x18, 0x53, 0x80, 0x67, 0x85, // Segment
		0x2a, 0xd7, 0xb1, 0x81, 0x03, // TimecodeScale
	}
	type TestEBML struct {
		Header struct {
			DocType            string `ebml:"EBMLDocType"`
			DocTypeVersion     uint64 `ebml:"EBMLDocTypeVersion"`
			DocTypeReadVersion uint64 `ebml:"EBMLDocTypeReadVersion"`
		} `ebml:"EBML"`
		Segment struct {
			TimecodeScale uint64 `ebml:"TimecodeScale"`
		} `ebml:"Segment"`
	}
	type TestEBMLWithMetadata struct {
		Header struct {
			Metadata           Metadata `ebml:"-"`
			DocType            string   `ebml:"EBMLDocType"`
			DocTypeVersion     uint64   `ebml:"EBMLDocTypeVersion"`
			DocTypeReadVersion uint64   `ebml:"EBMLDocTypeReadVersion"`
		} `ebml:"EBML"`
		Segment struct {
			Metadata      Metadata `ebml:"-"`
			TimecodeScale uint64   `ebml:"TimecodeScale"`
		} `ebml:"Segment"`
	}

	r := bytes.NewReader(TestBinary)

	var ret TestEBML
	if err := Unmarshal(r, &ret); err != nil {
		fmt.Printf("error: %+v\n", err)
	}
	fmt.Println(ret)

	r = bytes.NewReader(TestBinary)

	var ret2 TestEBMLWithMetadata
	if err := Unmarshal(r, &ret2); err != nil {
		fmt.Printf("error: %+v\n", err)
	}
	fmt.Println(ret2)

	// Output:
	// {{webm 2 2} {3}}
	// {{{0} webm 2 2} {{28} 3}}
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
		t.Fatalf("error: %+v\n", err)
	}
	if err := Unmarshal(bytes.NewBuffer(b), &untagged); err != nil {
		t.Fatalf("error: %+v\n", err)
	}

	if tagged.DocCustomNamedType != untagged.EBMLDocType {
		t.Errorf("Unmarshal result to tagged and and untagged struct must be same, tagged: %v, untagged: %v", tagged, untagged)
	}
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
			b.Fatalf("error: %+v\n", err)
		}
	}
}

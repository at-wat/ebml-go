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

func TestUnmarshal_WithElementHooks(t *testing.T) {
	TestBinary := []byte{
		0x18, 0x53, 0x80, 0x67, 0xa1, // Segment
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
	if err := Unmarshal(r, &ret, WithElementHooks(hook)); err != nil {
		t.Errorf("error: %+v\n", err)
	}

	// Verify positions of elements
	expected := map[string][]uint64{
		"Segment.Tracks":            {5},
		"Segment.Tracks.TrackEntry": {10, 24},
	}
	for key, positions := range expected {
		elem, ok := m[key]
		if !ok {
			t.Errorf("Key '%s' doesn't exist\n", key)
		}
		if len(elem) != len(positions) {
			t.Errorf("Unexpected element size of '%s', expected: %d, got: %d\n", key, len(positions), len(elem))
		}
		for i, pos := range positions {
			if elem[i].Position != pos {
				t.Errorf("Unexpected element positon of '%s[%d]', expected: %d, got: %d\n", key, i, pos, elem[i].Position)
			}
		}
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

func withElementMap(m map[string][]*Element) func(*Element) {
	return func(elem *Element) {
		key := elem.Name
		e := elem
		for {
			if e.Parent == nil {
				break
			}
			e = e.Parent
			key = fmt.Sprintf("%s.%s", e.Name, key)
		}
		elements, ok := m[key]
		if !ok {
			elements = make([]*Element, 0)
		}
		elements = append(elements, elem)
		m[key] = elements
	}
}

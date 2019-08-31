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

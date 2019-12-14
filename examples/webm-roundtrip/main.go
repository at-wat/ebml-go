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

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/at-wat/ebml-go"
	"github.com/at-wat/ebml-go/webm"
)

func main() {
	r, err := os.Open("sample.webm")
	if err != nil {
		panic(err)
	}
	defer r.Close()

	var ret struct {
		Header  webm.EBMLHeader `ebml:"EBML"`
		Segment webm.Segment    `ebml:"Segment"`
	}
	if err := ebml.Unmarshal(r, &ret); err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	j, err := json.MarshalIndent(ret, "", "  ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Printf("%s\n", string(j))

	w, err := os.OpenFile("copy.webm", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer w.Close()
	if err := ebml.Marshal(&ret, w); err != nil {
		panic(err)
	}
}

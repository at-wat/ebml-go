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

package mkvcore

import (
	"io/ioutil"
	"strings"
	"testing"
)

func TestReaderWithSizeCount(t *testing.T) {
	closer := ioutil.NopCloser(strings.NewReader("hello!"))
	r := &readerWithSizeCount{r: closer}

	buf := make([]byte, 1024)
	if n, err := r.Read(buf); err != nil {
		t.Fatalf("Failed to Write: '%v'", err)
	} else if n != 6 {
		t.Errorf("Expected return value of writerWithSizeCount.Write: 6, got: %d", n)
	}
	if n := r.Size(); n != 6 {
		t.Errorf("Expected return value of writerWithSizeCount.Size(): 6, got: %d", n)
	}

	r.Clear()

	if n := r.Size(); n != 0 {
		t.Errorf("Expected return value of writerWithSizeCount.Size(): 0, got: %d", n)
	}
}

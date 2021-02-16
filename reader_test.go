// Copyright 2021 The ebml-go authors.
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
	"testing"
)

func TestRollbackReader(t *testing.T) {
	r := &rollbackReaderImpl{
		Reader: bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7}),
	}

	b := make([]byte, 3)
	n, err := io.ReadFull(r, b)
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("Expected to read 3 bytes, got %d bytes", n)
	}
	if !bytes.Equal([]byte{0, 1, 2}, b) {
		t.Fatalf("Unexpected read result: %v", b)
	}

	r.Reset()

	n, err = io.ReadFull(r, b)
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("Expected to read 3 bytes, got %d bytes", n)
	}
	if !bytes.Equal([]byte{3, 4, 5}, b) {
		t.Fatalf("Unexpected read result: %v", b)
	}

	r.RollbackTo(1)

	n, err = io.ReadFull(r, b)
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("Expected to read 3 bytes, got %d bytes", n)
	}
	if !bytes.Equal([]byte{4, 5, 6}, b) {
		t.Fatalf("Unexpected read result: %v", b)
	}
}

func TestRollbackReaderNop(t *testing.T) {
	r := &rollbackReaderNop{
		Reader: bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7}),
	}

	b := make([]byte, 3)
	n, err := r.Read(b)
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("Expected to read 3 bytes, got %d bytes", n)
	}
	if !bytes.Equal([]byte{0, 1, 2}, b) {
		t.Fatalf("Unexpected read result: %v", b)
	}

	r.Reset()

	n, err = r.Read(b)
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("Expected to read 3 bytes, got %d bytes", n)
	}
	if !bytes.Equal([]byte{3, 4, 5}, b) {
		t.Fatalf("Unexpected read result: %v", b)
	}

	defer func() {
		if err := recover(); err == nil {
			t.Error("Expected panic")
		}
	}()
	r.RollbackTo(1)
}

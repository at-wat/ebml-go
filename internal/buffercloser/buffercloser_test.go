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

package buffercloser

import (
	"bytes"
	"testing"
	"time"
)

func TestBufferCloser(t *testing.T) {
	data := []byte{0x01, 0x02}
	buf := New()

	switch n, err := buf.Write(data); {
	case err != nil:
		t.Errorf("Unexpected error on Write(): %v", err)
	case n != len(data):
		t.Errorf("Number of wrote bytes should be %d, got %d", len(data), n)
	}

	select {
	case <-buf.Closed():
		t.Errorf("Closed() should not be sent to unclosed buffer")
	case <-time.After(5 * time.Millisecond):
	}

	if !bytes.Equal(data, buf.Bytes()) {
		t.Errorf("Expected bytes in the buffer: %d, got: %d", data, buf.Bytes())
	}

	if err := buf.Close(); err != nil {
		t.Errorf("Unexpected error on Close(): %v", err)
	}
	if _, err := buf.Write(data); err != ErrClosedBuffer {
		t.Errorf("Write() should return %d on closed buffer, got %v", ErrClosedBuffer, err)
	}

	select {
	case <-buf.Closed():
	case <-time.After(5 * time.Millisecond):
		t.Errorf("Closed() should be sent to closed buffer")
	}
}

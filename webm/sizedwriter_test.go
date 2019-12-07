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

package webm

import (
	"bytes"
	"testing"
)

type bufferCloser struct {
	bytes.Buffer
	closed chan struct{}
}

func (b *bufferCloser) Close() error {
	close(b.closed)
	return nil
}

func TestWriterWithSizeCount(t *testing.T) {
	buf := &bufferCloser{closed: make(chan struct{})}
	w := &writerWithSizeCount{w: buf}

	if n, err := w.Write([]byte{0x01, 0x02}); err != nil {
		t.Fatalf("Failed to Write: %v", err)
	} else if n != 2 {
		t.Errorf("Unexpected return value of writerWithSizeCount.Write, expected: 2, got: %d", n)
	}
	if n := w.Size(); n != 2 {
		t.Errorf("Unexpected return value of writerWithSizeCount.Size(), expected: 2, got: %d", n)
	}

	w.Clear()

	if n := w.Size(); n != 0 {
		t.Errorf("Unexpected return value of writerWithSizeCount.Size(), expected: 0, got: %d", n)
	}

	if err := w.Close(); err != nil {
		t.Errorf("writerWithSizeCount.Close() doesn't propagate base io.WriteCloser.Close() return value")
	}
	select {
	case <-buf.closed:
	default:
		t.Errorf("Base io.WriteCloser is not closed by writerWithSizeCount.Close()")
	}
}

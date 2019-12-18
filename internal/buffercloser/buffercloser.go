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
	"errors"
	"io"
	"sync"
)

// ErrClosedBuffer is the error returned by Write on closed buffer.
var ErrClosedBuffer = errors.New("write on closed buffer")

// BufferCloser is bytes.Buffer with io.Closer interface.
type BufferCloser interface {
	io.Writer
	io.Closer
	Bytes() []byte
	Closed() <-chan struct{}
}

type bufferCloser struct {
	buf       bytes.Buffer
	closed    chan struct{}
	closeOnce sync.Once
}

// New creates and initialize a new BufferCloser.
func New() BufferCloser {
	return &bufferCloser{
		closed: make(chan struct{}),
	}
}

func (b *bufferCloser) Write(p []byte) (int, error) {
	select {
	case <-b.closed:
		return 0, ErrClosedBuffer
	default:
		return b.buf.Write(p)
	}
}

func (b *bufferCloser) Close() error {
	b.closeOnce.Do(func() { close(b.closed) })
	return nil
}

func (b *bufferCloser) Bytes() []byte {
	return b.buf.Bytes()
}

func (b *bufferCloser) Closed() <-chan struct{} {
	return b.closed
}

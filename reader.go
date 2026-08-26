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
)

type rollbackReader interface {
	Set(io.Reader)
	Get() io.Reader
	Read([]byte) (int, error)
	Reset()
	JumpTo(uint64) error
}

type rollbackReaderImpl struct {
	io.Reader
	buf []byte
}

func (r *rollbackReaderImpl) Set(v io.Reader) {
	r.Reader = v
}

func (r *rollbackReaderImpl) Get() io.Reader {
	return r.Reader
}

func (r *rollbackReaderImpl) Read(b []byte) (int, error) {
	n, err := r.Reader.Read(b)
	r.buf = append(r.buf, b[:n]...)
	if n != 0 && err == io.EOF {
		err = nil
	}
	return n, err
}

func (r *rollbackReaderImpl) Reset() {
	r.buf = r.buf[0:0]
}

func (r *rollbackReaderImpl) JumpTo(i uint64) error {
	n := uint64(len(r.buf))
	switch {
	case i < n:
		buf := r.buf
		r.Reader = io.MultiReader(
			bytes.NewReader(buf[i:]),
			r.Reader,
		)
		r.buf = nil
		return nil
	case i == n:
		r.buf = nil
		return nil
	default:
		err := readSkip(r.Reader, i-n)
		r.buf = nil
		return err
	}
}

type rollbackReaderNop struct {
	io.Reader
}

func (r *rollbackReaderNop) Set(v io.Reader) {
	r.Reader = v
}

func (r *rollbackReaderNop) Get() io.Reader {
	return r.Reader
}

func (r *rollbackReaderNop) Read(b []byte) (int, error) {
	n, err := r.Reader.Read(b)
	if n != 0 && err == io.EOF {
		err = nil
	}
	return n, err
}

func (*rollbackReaderNop) Reset() {
}

func (*rollbackReaderNop) JumpTo(_ uint64) error {
	panic("can't jump nop rollback reader")
}

func readSkip(r io.Reader, n uint64) error {
	if seeker, ok := r.(io.Seeker); ok {
		_, err := seeker.Seek(int64(n), io.SeekCurrent)
		return err
	}
	var buf [1024]byte
	for {
		if n < uint64(len(buf)) {
			_, err := r.Read(buf[:n])
			return err
		}
		if _, err := r.Read(buf[:]); err != nil {
			return err
		}
	}
}

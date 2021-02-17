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
	RollbackTo(int)
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
	return n, err
}

func (r *rollbackReaderImpl) Reset() {
	r.buf = r.buf[0:0]
}

func (r *rollbackReaderImpl) RollbackTo(i int) {
	buf := r.buf
	r.Reader = io.MultiReader(
		bytes.NewReader(buf[i:]),
		r.Reader,
	)
	r.buf = nil
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

func (*rollbackReaderNop) Reset() {
}

func (*rollbackReaderNop) RollbackTo(i int) {
	panic("can't rollback nop rollback reader")
}

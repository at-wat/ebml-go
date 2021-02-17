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
	"errors"
	"io"
)

// ErrFixedLaceUndivisible means that a length of a fixed lacing data is undivisible.
var ErrFixedLaceUndivisible = errors.New("undivisible fixed lace")

// Unlacer is the interface to read laced frames in Block.
type Unlacer interface {
	Read() ([]byte, error)
}

type unlacer struct {
	r    io.Reader
	i    int
	size []int
}

func (u *unlacer) Read() ([]byte, error) {
	if u.i >= len(u.size) {
		return nil, io.EOF
	}
	n := u.size[u.i]
	u.i++

	b := make([]byte, n)
	_, err := io.ReadFull(u.r, b)
	return b, err
}

// NewNoUnlacer creates pass-through Unlacer for not laced data.
func NewNoUnlacer(r io.Reader, n int64) (Unlacer, error) {
	return &unlacer{r: r, size: []int{int(n)}}, nil
}

// NewXiphUnlacer creates Unlacer for Xiph laced data.
func NewXiphUnlacer(r io.Reader, n int64) (Unlacer, error) {
	var nFrame int
	var b [1]byte
	switch _, err := r.Read(b[:]); err {
	case nil:
		nFrame = int(b[0]) + 1
	case io.EOF:
		return nil, io.ErrUnexpectedEOF
	default:
		return nil, err
	}
	n--

	ul := &unlacer{
		r:    r,
		size: make([]int, nFrame),
	}
	for i := 0; i < nFrame-1; i++ {
		for {
			switch _, err := ul.r.Read(b[:]); err {
			case nil:
			case io.EOF:
				return nil, io.ErrUnexpectedEOF
			default:
				return nil, err
			}
			n--
			ul.size[i] += int(b[0])
			if b[0] != 0xFF {
				ul.size[nFrame-1] -= ul.size[i]
				break
			}
		}
	}
	ul.size[nFrame-1] += int(n)
	if ul.size[nFrame-1] <= 0 {
		return nil, io.ErrUnexpectedEOF
	}

	return ul, nil
}

// NewFixedUnlacer creates Unlacer for Fixed laced data.
func NewFixedUnlacer(r io.Reader, n int64) (Unlacer, error) {
	var nFrame int
	var b [1]byte
	switch _, err := r.Read(b[:]); err {
	case nil:
		nFrame = int(b[0]) + 1
	case io.EOF:
		return nil, io.ErrUnexpectedEOF
	default:
		return nil, err
	}

	ul := &unlacer{
		r:    r,
		size: make([]int, nFrame),
	}
	ul.size[0] = (int(n) - 1) / nFrame
	for i := 1; i < nFrame; i++ {
		ul.size[i] = ul.size[0]
	}
	if ul.size[0]*nFrame+1 != int(n) {
		return nil, wrapErrorf(
			ErrFixedLaceUndivisible, "unlacing %d bytes of %d frames", n-1, nFrame,
		)
	}
	return ul, nil
}

// NewEBMLUnlacer creates Unlacer for EBML laced data.
func NewEBMLUnlacer(r io.Reader, n int64) (Unlacer, error) {
	var nFrame int
	var b [1]byte
	switch _, err := r.Read(b[:]); err {
	case nil:
		nFrame = int(b[0]) + 1
	case io.EOF:
		return nil, io.ErrUnexpectedEOF
	default:
		return nil, err
	}
	n--

	vd := &valueDecoder{}

	ul := &unlacer{
		r:    r,
		size: make([]int, nFrame),
	}
	un64, nRead, err := vd.readVUInt(ul.r)
	if err != nil {
		return nil, err
	}
	n64 := int64(un64)
	n -= int64(nRead)
	ul.size[0] = int(n64)
	ul.size[nFrame-1] -= int(n64)

	for i := 1; i < nFrame-1; i++ {
		n64Diff, nRead, err := vd.readVInt(ul.r)
		n64 += int64(n64Diff)
		if err != nil {
			return nil, err
		}
		if n64 <= 0 {
			return nil, io.ErrUnexpectedEOF
		}
		n -= int64(nRead)
		ul.size[i] = int(n64)
		ul.size[nFrame-1] -= int(n64)
	}
	ul.size[nFrame-1] += int(n)
	if ul.size[nFrame-1] <= 0 {
		return nil, io.ErrUnexpectedEOF
	}

	return ul, nil
}

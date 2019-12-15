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
	"errors"
	"io"
)

var (
	errFixedLaceUndivisible = errors.New("undivisible fixed lace")
)

// Unlacer is the interface to read laced frames in Block.
type Unlacer interface {
	Read() ([]byte, error)
}

type unlacer struct {
	b    []byte
	p    int
	i    int
	size []int
}

func (u *unlacer) Read() ([]byte, error) {
	if u.i >= len(u.size) {
		return nil, io.EOF
	}
	n := u.size[u.i]
	ret := u.b[u.p : u.p+n]
	u.i++
	u.p += n

	if u.i >= len(u.size) {
		return ret, io.EOF
	}
	return ret, nil
}

// NewXiphUnlacer creates Unlacer for Xiph laced data.
func NewXiphUnlacer(b []byte) (Unlacer, error) {
	if len(b) < 1 {
		return nil, io.ErrUnexpectedEOF
	}
	n := int(b[0]) + 1

	ul := &unlacer{
		b:    b,
		p:    1,
		size: make([]int, n),
	}
	ul.size[n-1] = len(b)
	for i := 0; i < n-1; i++ {
		for {
			pos := ul.p
			if len(b) <= pos {
				return nil, io.ErrUnexpectedEOF
			}
			ul.size[i] += int(b[pos])
			ul.p++
			if b[pos] != 0xFF {
				ul.size[n-1] -= ul.size[i]
				break
			}
		}
	}
	ul.size[n-1] -= ul.p

	return ul, nil
}

// NewFixedUnlacer creates Unlacer for Fixed laced data.
func NewFixedUnlacer(b []byte) (Unlacer, error) {
	if len(b) < 1 {
		return nil, io.ErrUnexpectedEOF
	}
	n := int(b[0]) + 1

	ul := &unlacer{
		b:    b,
		p:    1,
		size: make([]int, n),
	}
	ul.size[0] = (len(b) - 1) / n
	for i := 1; i < n; i++ {
		ul.size[i] = ul.size[0]
	}
	if ul.size[0]*n+1 != len(b) {
		return nil, errFixedLaceUndivisible
	}
	return ul, nil
}

// NewEBMLUnlacer creates Unlacer for EBML laced data.
func NewEBMLUnlacer(b []byte) (Unlacer, error) {
	if len(b) < 1 {
		return nil, io.ErrUnexpectedEOF
	}
	n := int(b[0]) + 1

	ul := &unlacer{
		b:    b,
		size: make([]int, n),
	}
	ul.size[n-1] = len(b)
	r := bytes.NewReader(b[1:])
	for i := 0; i < n-1; i++ {
		n64, _, err := readVInt(r)
		if err != nil {
			return nil, err
		}
		ul.size[i] = int(n64)
		ul.size[n-1] -= ul.size[i]
	}
	ul.p = len(b) - r.Len() + 1
	ul.size[n-1] -= ul.p

	return ul, nil
}

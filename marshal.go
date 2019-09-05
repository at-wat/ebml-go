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
	"reflect"
	"strings"
)

var (
	errUnsupportedElement = errors.New("Unsupported element")
)

// Marshal struct to EBML bytes
func Marshal(val interface{}, w io.Writer) error {
	vo := reflect.ValueOf(val).Elem()

	return marshalImpl(vo, w)
}

func marshalImpl(vo reflect.Value, w io.Writer) error {
	for i := 0; i < vo.NumField(); i++ {
		vn := vo.Field(i)
		tn := vo.Type().Field(i)

		var nn []string
		if n, ok := tn.Tag.Lookup("ebml"); ok {
			nn = strings.Split(n, ",")
		}
		var name string
		if len(nn) > 0 && len(nn[0]) > 0 {
			name = nn[0]
		} else {
			name = tn.Name
		}
		if t, err := ElementTypeFromString(name); err == nil {
			e, ok := table[t]
			if !ok {
				return errUnsupportedElement
			}

			var inf, omitempty bool
			for _, n := range nn {
				if n == "inf" {
					inf = true
				} else if n == "omitempty" {
					omitempty = true
				}
			}

			var lst []reflect.Value
			if vn.Kind() == reflect.Ptr {
				if !vn.IsNil() {
					lst = []reflect.Value{vn.Elem()}
				} else {
					continue
				}
			} else if vn.Kind() == reflect.Slice && e.t != TypeBinary {
				l := vn.Len()
				for i := 0; i < l; i++ {
					lst = append(lst, vn.Index(i))
				}
			} else {
				if omitempty && reflect.DeepEqual(reflect.Zero(vn.Type()).Interface(), vn.Interface()) {
					continue
				}
				lst = []reflect.Value{vn}
			}

			for _, vn := range lst {
				// Write element ID
				if _, err := w.Write(e.b); err != nil {
					return err
				}
				var bw io.Writer
				if inf {
					// Directly write length unspecified element
					bsz := encodeVInt(uint64(sizeInf))
					if _, err := w.Write(bsz); err != nil {
						return err
					}
					bw = w
				} else {
					bw = &bytes.Buffer{}
				}

				if e.t == TypeMaster {
					if err := marshalImpl(vn, bw); err != nil {
						return err
					}
				} else {
					bc, err := perTypeEncoder[e.t](vn.Interface())
					if err != nil {
						return err
					}
					if _, err := bw.Write(bc); err != nil {
						return err
					}
				}

				// Write element with length
				if !inf {
					bsz := encodeVInt(uint64(bw.(*bytes.Buffer).Len()))
					if _, err := w.Write(bsz); err != nil {
						return err
					}
					if _, err := w.Write(bw.(*bytes.Buffer).Bytes()); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

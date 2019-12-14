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
	errUnknownElement = errors.New("unknown element")
	errIndefiniteType = errors.New("unmarshal to indefinite type")
)

// Unmarshal EBML stream.
func Unmarshal(r io.Reader, val interface{}, opts ...UnmarshalOption) error {
	options := &UnmarshalOptions{}
	for _, o := range opts {
		if err := o(options); err != nil {
			return err
		}
	}

	vo := reflect.ValueOf(val)
	if !vo.IsValid() {
		return errIndefiniteType
	}
	voe := vo.Elem()
	for {
		if _, err := readElement(r, sizeUnknown, voe, 0, nil, options); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

func readElement(r0 io.Reader, n int64, vo reflect.Value, pos uint64, parent *Element, options *UnmarshalOptions) (io.Reader, error) {
	var r io.Reader
	if n != sizeUnknown {
		r = io.LimitReader(r0, n)
	} else {
		r = r0
	}

	type field struct {
		v reflect.Value
		t reflect.Type
	}
	fieldMap := make(map[string]field)
	if vo.IsValid() {
		for i := 0; i < vo.NumField(); i++ {
			var nn []string
			if n, ok := vo.Type().Field(i).Tag.Lookup("ebml"); ok {
				nn = strings.Split(n, ",")
			}
			var name string
			if len(nn) > 0 && len(nn[0]) > 0 {
				name = nn[0]
			} else {
				name = vo.Type().Field(i).Name
			}
			fieldMap[name] = field{vo.Field(i), vo.Type().Field(i).Type}
		}
	}

	for {
		var headerSize uint64
		e, nb, err := readVInt(r)
		headerSize += uint64(nb)
		if err != nil {
			if nb == 0 && err == io.ErrUnexpectedEOF {
				return nil, io.EOF
			}
			return nil, err
		}
		v, ok := revTable[uint32(e)]
		if !ok {
			return nil, errUnknownElement
		}

		size, nb, err := readVInt(r)
		headerSize += uint64(nb)
		if err != nil {
			return nil, err
		}
		var vnext reflect.Value
		if fm, ok := fieldMap[v.e.String()]; ok {
			vnext = fm.v
		}

		var elem *Element
		if len(options.hooks) > 0 {
			elem = &Element{
				Value:    vnext.Interface(),
				Name:     v.e.String(),
				Position: pos,
				Size:     size,
				Parent:   parent,
			}
		}

		switch v.t {
		case TypeMaster:
			if v.top && !vnext.IsValid() {
				b := bytes.Join([][]byte{table[v.e].b, encodeDataSize(size, uint64(nb))}, []byte{})
				return bytes.NewBuffer(b), io.EOF
			}
			var vn reflect.Value
			if vnext.IsValid() && vnext.CanSet() {
				switch vnext.Kind() {
				case reflect.Ptr:
					vnext.Set(reflect.New(vnext.Type().Elem()))
					vn = vnext.Elem()
				case reflect.Slice:
					vnext.Set(reflect.Append(vnext, reflect.New(vnext.Type().Elem()).Elem()))
					vn = vnext.Index(vnext.Len() - 1)
				default:
					vn = vnext
				}
			}
			r0, err := readElement(r, int64(size), vn, pos+headerSize, elem, options)

			if err != nil && err != io.EOF {
				return r0, err
			}
			if r0 != nil {
				r = io.MultiReader(r0, r)
			}
		default:
			val, err := perTypeReader[v.t](r, size)
			if err != nil {
				return nil, err
			}
			vr := reflect.ValueOf(val)
			if vnext.IsValid() && vnext.CanSet() {
				if vr.Type() == vnext.Type() {
					vnext.Set(reflect.ValueOf(val))
				} else if vnext.Kind() == reflect.Slice && vr.Type() == vnext.Type().Elem() {
					vnext.Set(reflect.Append(vnext, reflect.ValueOf(val)))
				}
			}
		}
		for _, hook := range options.hooks {
			hook(elem)
		}

		pos += headerSize + size
	}
}

// UnmarshalOption configures a UnmarshalOptions struct.
type UnmarshalOption func(*UnmarshalOptions) error

// UnmarshalOptions stores options for unmarshalling.
type UnmarshalOptions struct {
	hooks []func(elem *Element)
}

// WithElementReadHooks returns an UnmarshalOption which registers element hooks.
func WithElementReadHooks(hooks ...func(*Element)) UnmarshalOption {
	return func(opts *UnmarshalOptions) error {
		opts.hooks = hooks
		return nil
	}
}

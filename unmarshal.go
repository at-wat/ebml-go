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

// ErrUnknownElement means that a decoded element is not known.
var ErrUnknownElement = errors.New("unknown element")

// ErrIndefiniteType means that a marshal/unmarshal destination type is not valid.
var ErrIndefiniteType = errors.New("marshal/unmarshal to indefinite type")

// ErrIncompatibleType means that an element is not convertible to a corresponding struct field.
var ErrIncompatibleType = errors.New("marshal/unmarshal to incompatible type")

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
		return wrapErrorf(ErrIndefiniteType, "unmarshalling to %T", val)
	}
	if vo.Kind() != reflect.Ptr {
		return wrapErrorf(ErrIncompatibleType, "unmarshalling to %T", val)
	}

	voe := vo.Elem()
	for {
		if _, err := readElement(r, SizeUnknown, voe, 0, 0, nil, options); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

func readElement(r0 io.Reader, n int64, vo reflect.Value, depth int, pos uint64, parent *Element, options *UnmarshalOptions) (io.Reader, error) {
	var r io.Reader
	if n != SizeUnknown {
		r = io.LimitReader(r0, n)
	} else {
		r = r0
	}

	var mapOut bool
	fieldMap := make(map[ElementType]reflect.Value)
	switch vo.Kind() {
	case reflect.Struct:
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
			t, err := ElementTypeFromString(name)
			if err != nil {
				return nil, err
			}
			fieldMap[t] = vo.Field(i)
		}
	case reflect.Map:
		mapOut = true
	}

	for {
		var headerSize uint64
		e, nb, err := readVUInt(r)
		headerSize += uint64(nb)
		if err != nil {
			if nb == 0 && err == io.ErrUnexpectedEOF {
				return nil, io.EOF
			}
			return nil, err
		}
		v, ok := revTable[uint32(e)]
		if !ok {
			if !options.ignoreUnknown {
				return nil, wrapErrorf(ErrUnknownElement, "unmarshalling element 0x%x", e)
			}
			remain, _ := encodeElementID(e)
			r = io.MultiReader(
				bytes.NewReader(remain[1:]),
				r,
			)
			pos++
			continue
		}

		size, nb, err := readDataSize(r)
		headerSize += uint64(nb)
		if err != nil {
			return nil, err
		}

		if !ok {
			if size == SizeUnknown {
				return nil, wrapErrorf(ErrUnknownElement, "unmarshalling unknown size element 0x%x", e)
			}
			if _, err := readBinary(r, size); err != nil {
				return nil, err
			}
			pos += headerSize + size
			continue
		}

		var vnext reflect.Value
		if !mapOut {
			if vn, ok := fieldMap[v.e]; ok {
				vnext = vn
			}
		}

		var chanSend reflect.Value
		var elem *Element
		if len(options.hooks) > 0 && vnext.IsValid() {
			elem = &Element{
				Name:     v.e.String(),
				Type:     v.e,
				Position: pos,
				Size:     size,
				Parent:   parent,
			}
		}
		if vnext.Kind() == reflect.Chan {
			chanSend = vnext
			vnext = reflect.New(vnext.Type().Elem()).Elem()
		}

		switch v.t {
		case DataTypeMaster:
			if v.top && depth > 1 {
				b := bytes.Join([][]byte{table[v.e].b, encodeDataSize(size, uint64(nb))}, []byte{})
				return bytes.NewBuffer(b), io.EOF
			}
			var vn reflect.Value
			if mapOut {
				vnext = reflect.ValueOf(make(map[string]interface{}))
				vn = vnext
			} else {
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
			}
			if elem != nil {
				elem.Value = vn.Interface()
			}
			r0, err := readElement(r, int64(size), vn, depth+1, pos+headerSize, elem, options)
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
			if mapOut {
				vnext = vr
			} else {
				if vnext.IsValid() && vnext.CanSet() {
					switch {
					case vr.Type() == vnext.Type():
						vnext.Set(vr)
					case isConvertible(vr.Type(), vnext.Type()):
						vnext.Set(vr.Convert(vnext.Type()))
					case vnext.Kind() == reflect.Slice:
						t := vnext.Type().Elem()
						switch {
						case vr.Type() == t:
							vnext.Set(reflect.Append(vnext, vr))
						case isConvertible(vr.Type(), t):
							vnext.Set(reflect.Append(vnext, vr.Convert(t)))
						default:
							return nil, wrapErrorf(
								ErrIncompatibleType, "unmarshalling %s to %s", vnext.Type(), vr.Type(),
							)
						}
					default:
						return nil, wrapErrorf(
							ErrIncompatibleType, "unmarshalling %s to %s", vnext.Type(), vr.Type(),
						)
					}
				}
			}
			if elem != nil {
				elem.Value = vr.Interface()
			}
		}
		if mapOut {
			t := vo.Type()
			if vo.IsNil() && t.Kind() == reflect.Map {
				vo.Set(reflect.MakeMap(t))
			}
			key := reflect.ValueOf(v.e.String())
			if e := vo.MapIndex(key); e.IsValid() {
				switch {
				case e.Elem().Kind() == reflect.Slice && v.t != DataTypeBinary:
					vnext = reflect.Append(e.Elem(), vnext)
				default:
					vnext = reflect.ValueOf([]interface{}{
						e.Elem().Interface(),
						vnext.Interface()},
					)
				}
			}
			vo.SetMapIndex(key, vnext)
		}
		if chanSend.IsValid() {
			chanSend.Send(vnext)
		}
		if elem != nil {
			for _, hook := range options.hooks {
				hook(elem)
			}
		}

		pos += headerSize + size
	}
}

// UnmarshalOption configures a UnmarshalOptions struct.
type UnmarshalOption func(*UnmarshalOptions) error

// UnmarshalOptions stores options for unmarshalling.
type UnmarshalOptions struct {
	hooks         []func(elem *Element)
	ignoreUnknown bool
}

// WithElementReadHooks returns an UnmarshalOption which registers element hooks.
func WithElementReadHooks(hooks ...func(*Element)) UnmarshalOption {
	return func(opts *UnmarshalOptions) error {
		opts.hooks = hooks
		return nil
	}
}

// WithIgnoreUnknown returns an UnmarshalOption which makes Unmarshal ignoring unknown element with static length.
func WithIgnoreUnknown(ignore bool) UnmarshalOption {
	return func(opts *UnmarshalOptions) error {
		opts.ignoreUnknown = ignore
		return nil
	}
}

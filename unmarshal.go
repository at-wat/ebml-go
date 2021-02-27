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
)

// ErrUnknownElement means that a decoded element is not known.
var ErrUnknownElement = errors.New("unknown element")

// ErrIndefiniteType means that a marshal/unmarshal destination type is not valid.
var ErrIndefiniteType = errors.New("marshal/unmarshal to indefinite type")

// ErrIncompatibleType means that an element is not convertible to a corresponding struct field.
var ErrIncompatibleType = errors.New("marshal/unmarshal to incompatible type")

// ErrInvalidElementSize means that an element has inconsistent size. e.g. element size is larger than its parent element size.
var ErrInvalidElementSize = errors.New("invalid element size")

// ErrReadStopped is returned if unmarshaler finished to read element which has stop tag.
var ErrReadStopped = errors.New("read stopped")

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

	vd := &valueDecoder{}

	voe := vo.Elem()
	for {
		if _, err := vd.readElement(r, SizeUnknown, voe, 0, 0, nil, options); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

func (vd *valueDecoder) readElement(r0 io.Reader, n int64, vo reflect.Value, depth int, pos uint64, parent *Element, options *UnmarshalOptions) (io.Reader, error) {
	pos0 := pos
	var r rollbackReader
	if options.ignoreUnknown {
		r = &rollbackReaderImpl{}
	} else {
		r = &rollbackReaderNop{}
	}
	if n != SizeUnknown {
		r.Set(io.LimitReader(r0, n))
	} else {
		r.Set(r0)
	}

	var mapOut bool
	type fieldDef struct {
		v    reflect.Value
		stop bool
	}
	fieldMap := make(map[ElementType]fieldDef)
	switch vo.Kind() {
	case reflect.Struct:
		for i := 0; i < vo.NumField(); i++ {
			f := fieldDef{
				v: vo.Field(i),
			}
			var name string
			if n, ok := vo.Type().Field(i).Tag.Lookup("ebml"); ok {
				t, err := parseTag(n)
				if err != nil {
					return nil, err
				}
				name = t.name
				f.stop = t.stop
			}
			if name == "" {
				name = vo.Type().Field(i).Name
			}
			t, err := ElementTypeFromString(name)
			if err != nil {
				return nil, err
			}
			fieldMap[t] = f
		}
	case reflect.Map:
		mapOut = true
	}

	for {
		r.Reset()

		var headerSize uint64
		e, nb, err := vd.readVUInt(r)
		headerSize += uint64(nb)
		if err != nil {
			if nb == 0 && err == io.ErrUnexpectedEOF {
				return nil, io.EOF
			}
			if options.ignoreUnknown {
				return nil, nil
			}
			return nil, err
		}
		v, ok := revTable[uint32(e)]
		if !ok {
			if options.ignoreUnknown {
				r.RollbackTo(1)
				pos++
				continue
			}
			return nil, wrapErrorf(ErrUnknownElement, "unmarshalling element 0x%x", e)
		}

		size, nb, err := vd.readDataSize(r)
		headerSize += uint64(nb)

		if n != SizeUnknown && pos+headerSize+size > pos0+uint64(n) {
			err = ErrInvalidElementSize
		}

		if err != nil {
			if options.ignoreUnknown {
				r.RollbackTo(1)
				pos++
				continue
			}
			return nil, err
		}

		var vnext reflect.Value
		var stopHere bool
		if vn, ok := fieldMap[v.e]; ok {
			if !mapOut {
				vnext = vn.v
			}
			stopHere = vn.stop
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
			r0, err := vd.readElement(r, int64(size), vn, depth+1, pos+headerSize, elem, options)
			if err != nil && err != io.EOF {
				return r0, err
			}
			if r0 != nil {
				r.Set(io.MultiReader(r0, r.Get()))
			}
		default:
			val, err := vd.decode(v.t, r, size)
			if err != nil {
				if options.ignoreUnknown {
					r.RollbackTo(1)
					pos++
					continue
				}
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
		if stopHere {
			return nil, ErrReadStopped
		}
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

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

// ErrUnsupportedElement means that a element name is known but unsupported in this version of ebml-go.
var ErrUnsupportedElement = errors.New("unsupported element")

// ErrNonStringMapKey is returned if input is  map and key is not a string.
var ErrNonStringMapKey = errors.New("non-string map key")

// Marshal struct to EBML bytes.
//
// Examples of struct field tags:
//
//   // Field appears as element "EBMLVersion".
//   Field uint64 `ebml:EBMLVersion`
//
//   // Field appears as element "EBMLVersion" and
//   // the field is omitted from the output if the value is empty.
//   Field uint64 `ebml:TheElement,omitempty`
//
//   // Field appears as element "EBMLVersion" and
//   // the field is omitted from the output if the value is empty.
//   EBMLVersion uint64 `ebml:,omitempty`
//
//   // Field appears as master element "Segment" and
//   // the size of the element contents is left unknown for streaming data.
//   Field struct{} `ebml:Segment,size=unknown`
//
//   // Field appears as master element "Segment" and
//   // the size of the element contents is left unknown for streaming data.
//   // This style may be deprecated in the future.
//   Field struct{} `ebml:Segment,inf`
//
//   // Field appears as element "EBMLVersion" and
//   // the size of the element data is reserved by 4 bytes.
//   Field uint64 `ebml:EBMLVersion,size=4`
func Marshal(val interface{}, w io.Writer, opts ...MarshalOption) error {
	options := &MarshalOptions{}
	for _, o := range opts {
		if err := o(options); err != nil {
			return err
		}
	}
	vo := reflect.ValueOf(val)
	if vo.Kind() != reflect.Ptr {
		return wrapErrorf(ErrInvalidType, "marshalling to %T", val)
	}

	_, err := marshalImpl(vo.Elem(), w, 0, nil, options)
	return err
}

func pealElem(v reflect.Value, binary, omitEmpty bool) ([]reflect.Value, bool) {
	for {
		switch v.Kind() {
		case reflect.Interface, reflect.Ptr:
			if v.IsNil() {
				return nil, false
			}
			v = v.Elem()
		case reflect.Slice:
			if binary {
				if omitEmpty && v.Len() == 0 {
					return nil, false
				}
				return []reflect.Value{v}, true
			}
			var lst []reflect.Value
			l := v.Len()
			for i := 0; i < l; i++ {
				vv, ok := pealElem(v.Index(i), false, omitEmpty)
				if !ok {
					continue
				}
				lst = append(lst, vv...)
			}
			return lst, true
		case reflect.Chan:
			return []reflect.Value{v}, true
		default:
			if omitEmpty && deepIsZero(v) {
				return nil, false
			}
			return []reflect.Value{v}, true
		}
	}
}

func deepIsZero(v reflect.Value) bool {
	return reflect.DeepEqual(reflect.Zero(v.Type()).Interface(), v.Interface())
}

func marshalImpl(vo reflect.Value, w io.Writer, pos uint64, parent *Element, options *MarshalOptions) (uint64, error) {
	var l int
	var tagFieldFunc func(int) (*structTag, reflect.Value, error)

	switch vo.Kind() {
	case reflect.Struct:
		l = vo.NumField()
		tagFieldFunc = func(i int) (*structTag, reflect.Value, error) {
			tag := &structTag{}
			tn := vo.Type().Field(i)
			if n, ok := tn.Tag.Lookup("ebml"); ok {
				var err error
				if tag, err = parseTag(n); err != nil {
					return nil, reflect.Value{}, err
				}
			}
			if tag.name == "" {
				tag.name = tn.Name
			}
			return tag, vo.Field(i), nil
		}
	case reflect.Map:
		l = vo.Len()
		keys := vo.MapKeys()
		tagFieldFunc = func(i int) (*structTag, reflect.Value, error) {
			name := keys[i]
			if name.Kind() != reflect.String {
				return nil, reflect.Value{}, ErrNonStringMapKey
			}
			return &structTag{name: name.String()}, vo.MapIndex(name), nil
		}
	default:
		return pos, ErrIncompatibleType
	}

	for i := 0; i < l; i++ {
		tag, vn, err := tagFieldFunc(i)
		if err != nil {
			return pos, err
		}

		t, err := ElementTypeFromString(tag.name)
		if err != nil {
			return pos, err
		}
		e, ok := table[t]
		if !ok {
			return pos, wrapErrorf(ErrUnsupportedElement, "marshalling \"%s\"", t)
		}

		unknown := tag.size == SizeUnknown

		lst, ok := pealElem(vn, e.t == DataTypeBinary, tag.omitEmpty)
		if !ok {
			continue
		}

		writeOne := func(vn reflect.Value) (uint64, error) {
			// Write element ID
			var headerSize uint64
			n, err := w.Write(e.b)
			if err != nil {
				return pos, err
			}
			headerSize += uint64(n)

			var bw io.Writer
			if unknown {
				// Directly write length unspecified element
				bsz := encodeDataSize(uint64(SizeUnknown), 0)
				n, err := w.Write(bsz)
				if err != nil {
					return pos, err
				}
				headerSize += uint64(n)
				bw = w
			} else {
				bw = &bytes.Buffer{}
			}

			var elem *Element
			if len(options.hooks) > 0 {
				elem = &Element{
					Value:    vn.Interface(),
					Name:     tag.name,
					Type:     t,
					Position: pos,
					Size:     SizeUnknown,
					Parent:   parent,
				}
			}

			var size uint64
			if e.t == DataTypeMaster {
				p, err := marshalImpl(vn, bw, pos+headerSize, elem, options)
				if err != nil {
					return pos, err
				}
				size = p - pos - headerSize
			} else {
				bc, err := perTypeEncoder[e.t](vn.Interface(), tag.size)
				if err != nil {
					return pos, err
				}
				n, err := bw.Write(bc)
				if err != nil {
					return pos, err
				}
				size = uint64(n)
			}

			// Write element with length
			if !unknown {
				if len(options.hooks) > 0 {
					elem.Size = size
				}
				bsz := encodeDataSize(size, options.dataSizeLen)
				n, err := w.Write(bsz)
				if err != nil {
					return pos, err
				}
				headerSize += uint64(n)

				if _, err := w.Write(bw.(*bytes.Buffer).Bytes()); err != nil {
					return pos, err
				}
			}
			for _, cb := range options.hooks {
				cb(elem)
			}
			pos += headerSize + size
			return pos, nil
		}

		for _, vn := range lst {
			var err error
			switch vn.Kind() {
			case reflect.Chan:
				for {
					val, ok := vn.Recv()
					if !ok {
						break
					}
					lst, ok := pealElem(val, e.t == DataTypeBinary, tag.omitEmpty)
					if !ok {
						return pos, wrapErrorf(
							ErrIncompatibleType, "marshalling %s from channel", val.Type(),
						)
					}
					if len(lst) != 1 {
						return pos, wrapErrorf(
							ErrIncompatibleType, "marshalling %s from channel", val.Type(),
						)
					}
					pos, err = writeOne(lst[0])
				}
			case reflect.Func:
				ret := vn.Call(nil)
				lenRet := len(ret)
				if lenRet != 1 && lenRet != 2 {
					return pos, wrapErrorf(ErrIncompatibleType, "number of return value must be 1 or 2 but %d", lenRet)
				}
				val := ret[0]
				if lenRet == 2 {
					errVal := ret[1]
					if errVal.Type().String() != "error" {
						return pos, wrapErrorf(ErrIncompatibleType, "2nd return value must be error but %s", errVal.Type())
					}
					if iFace := errVal.Interface(); iFace != nil {
						return pos, iFace.(error)
					}
				}
				lst, ok := pealElem(val, e.t == DataTypeBinary, tag.omitEmpty)
				if !ok {
					return pos, wrapErrorf(
						ErrIncompatibleType, "marshalling %s from func", val.Type(),
					)
				}
				for _, l := range lst {
					pos, err = writeOne(l)
				}
			default:
				pos, err = writeOne(vn)
			}
			if err != nil {
				return pos, err
			}
		}
	}
	return pos, nil
}

// MarshalOption configures a MarshalOptions struct.
type MarshalOption func(*MarshalOptions) error

// MarshalOptions stores options for marshalling.
type MarshalOptions struct {
	dataSizeLen uint64
	hooks       []func(elem *Element)
}

// WithDataSizeLen returns an MarshalOption which sets number of reserved bytes of element data size.
func WithDataSizeLen(l int) MarshalOption {
	return func(opts *MarshalOptions) error {
		opts.dataSizeLen = uint64(l)
		return nil
	}
}

// WithElementWriteHooks returns an MarshalOption which registers element hooks.
func WithElementWriteHooks(hooks ...func(*Element)) MarshalOption {
	return func(opts *MarshalOptions) error {
		opts.hooks = hooks
		return nil
	}
}

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

var (
	errUnsupportedElement = errors.New("unsupported element")
)

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
	vo := reflect.ValueOf(val).Elem()

	_, err := marshalImpl(vo, w, 0, nil, options)
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
	l := vo.NumField()
	for i := 0; i < l; i++ {
		vn := vo.Field(i)
		tn := vo.Type().Field(i)

		tag := &structTag{}
		if n, ok := tn.Tag.Lookup("ebml"); ok {
			var err error
			if tag, err = parseTag(n); err != nil {
				return pos, err
			}
		}
		if tag.name == "" {
			tag.name = tn.Name
		}
		if t, err := ElementTypeFromString(tag.name); err == nil {
			e, ok := table[t]
			if !ok {
				return pos, errUnsupportedElement
			}

			unknown := tag.size == sizeUnknown

			lst, ok := pealElem(vn, e.t == TypeBinary, tag.omitEmpty)
			if !ok {
				continue
			}

			for _, vn := range lst {
				// Write element ID
				var headerSize uint64
				if n, err := w.Write(e.b); err != nil {
					return pos, err
				} else {
					headerSize += uint64(n)
				}
				var bw io.Writer
				if unknown {
					// Directly write length unspecified element
					bsz := encodeDataSize(uint64(sizeUnknown), 0)
					if n, err := w.Write(bsz); err != nil {
						return pos, err
					} else {
						headerSize += uint64(n)
					}
					bw = w
				} else {
					bw = &bytes.Buffer{}
				}

				elem := &Element{
					Value:    vn.Interface(),
					Name:     tag.name,
					Position: pos,
					Size:     sizeUnknown,
					Parent:   parent,
				}

				var size uint64
				if e.t == TypeMaster {
					if p, err := marshalImpl(vn, bw, pos+headerSize, elem, options); err != nil {
						return pos, err
					} else {
						size = p - pos - headerSize
					}
				} else {
					bc, err := perTypeEncoder[e.t](vn.Interface(), tag.size)
					if err != nil {
						return pos, err
					}
					if n, err := bw.Write(bc); err != nil {
						return pos, err
					} else {
						size = uint64(n)
					}
				}

				// Write element with length
				if !unknown {
					elem.Size = size
					bsz := encodeDataSize(elem.Size, options.dataSizeLen)
					if n, err := w.Write(bsz); err != nil {
						return pos, err
					} else {
						headerSize += uint64(n)
					}
					if _, err := w.Write(bw.(*bytes.Buffer).Bytes()); err != nil {
						return pos, err
					}
				}
				for _, cb := range options.hooks {
					cb(elem)
				}
				pos += headerSize + size
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

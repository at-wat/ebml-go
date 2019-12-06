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

// Marshal struct to EBML bytes
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

	return marshalImpl(vo, w, options)
}

func marshalImpl(vo reflect.Value, w io.Writer, options *MarshalOptions) error {
	for i := 0; i < vo.NumField(); i++ {
		vn := vo.Field(i)
		tn := vo.Type().Field(i)

		tag := &structTag{}
		if n, ok := tn.Tag.Lookup("ebml"); ok {
			var err error
			if tag, err = parseTag(n); err != nil {
				return err
			}
		}
		if tag.name == "" {
			tag.name = tn.Name
		}
		if t, err := ElementTypeFromString(tag.name); err == nil {
			e, ok := table[t]
			if !ok {
				return errUnsupportedElement
			}

			unknown := tag.size == sizeUnknown

			var lst []reflect.Value
			switch {
			case vn.Kind() == reflect.Ptr:
				if vn.IsNil() {
					continue
				}
				lst = []reflect.Value{vn.Elem()}
			case vn.Kind() == reflect.Slice && e.t != TypeBinary:
				l := vn.Len()
				for i := 0; i < l; i++ {
					lst = append(lst, vn.Index(i))
				}
			default:
				if tag.omitEmpty && reflect.DeepEqual(reflect.Zero(vn.Type()).Interface(), vn.Interface()) {
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
				if unknown {
					// Directly write length unspecified element
					bsz := encodeDataSize(uint64(sizeUnknown), 0)
					if _, err := w.Write(bsz); err != nil {
						return err
					}
					bw = w
				} else {
					bw = &bytes.Buffer{}
				}

				if e.t == TypeMaster {
					if err := marshalImpl(vn, bw, options); err != nil {
						return err
					}
				} else {
					bc, err := perTypeEncoder[e.t](vn.Interface(), tag.size)
					if err != nil {
						return err
					}
					if _, err := bw.Write(bc); err != nil {
						return err
					}
				}

				// Write element with length
				if !unknown {
					bsz := encodeDataSize(uint64(bw.(*bytes.Buffer).Len()), options.dataSizeLen)
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

type MarshalOption func(*MarshalOptions) error

// MarshalOptions stores options for marshalling
type MarshalOptions struct {
	dataSizeLen uint64
}

// WithDataSizeLen returns an MarshalOption which sets number of reserved bytes of element data size
func WithDataSizeLen(l int) MarshalOption {
	return func(opts *MarshalOptions) error {
		opts.dataSizeLen = uint64(l)
		return nil
	}
}

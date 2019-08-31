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

		if n, ok := tn.Tag.Lookup("ebml"); ok {
			nn := strings.Split(n, ",")
			var name string
			if len(nn) > 0 {
				name = nn[0]
			} else {
				name = tn.Name
			}
			if t, err := ElementTypeFromString(name); err == nil {
				e, ok := table[t]
				if !ok {
					return errUnsupportedElement
				}

				var inf bool
				for _, n := range nn {
					if n == "inf" {
						inf = true
						break
					}
				}

				var lst []reflect.Value
				if vn.Kind() == reflect.Slice && e.t != TypeBinary {
					l := vn.Len()
					for i := 0; i < l; i++ {
						lst = append(lst, vn.Index(i))
					}
				} else {
					lst = []reflect.Value{vn}
				}

				for _, vn := range lst {
					if _, err := w.Write(e.b); err != nil {
						return err
					}
					var b bytes.Buffer
					if e.t == TypeMaster {
						if err := marshalImpl(vn, &b); err != nil {
							return err
						}
					} else {
						bc, err := perTypeEncoder[e.t](vn.Interface())
						if err != nil {
							return err
						}
						if _, err := b.Write(bc); err != nil {
							return err
						}
					}
					l := b.Len()
					if inf {
						l = sizeInf
					}
					bsz := encodeVInt(uint64(l))
					if _, err := w.Write(bsz); err != nil {
						return err
					}
					if _, err := w.Write(b.Bytes()); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

package ebml

import (
	"bytes"
	"errors"
	"io"
	"reflect"
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
			if t, err := ElementTypeFromString(n); err == nil {
				e, ok := table[t]
				if !ok {
					return errUnsupportedElement
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

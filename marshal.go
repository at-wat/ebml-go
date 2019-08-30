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
	to := reflect.TypeOf(val).Elem()

	return marshalImpl(vo, to, w)
}

func marshalImpl(vo reflect.Value, to reflect.Type, w io.Writer) error {
	for i := 0; i < vo.NumField(); i++ {
		vn := vo.Field(i)
		tn := to.Field(i)

		if n, ok := tn.Tag.Lookup("ebml"); ok {
			if t, err := ElementTypeFromString(n); err == nil {
				e, ok := table[t]
				if !ok {
					return errUnsupportedElement
				}
				if _, err := w.Write(e.b); err != nil {
					return err
				}

				var bc []byte
				if e.t == TypeMaster {
					var b bytes.Buffer
					if err := marshalImpl(vn, tn.Type, &b); err != nil {
						return err
					}
					bc = b.Bytes()
				} else {
					var err error
					if bc, err = perTypeEncoder[e.t](vn.Interface()); err != nil {
						return err
					}
				}
				bsz := encodeVInt(uint64(len(bc)))
				if _, err := w.Write(bsz); err != nil {
					return err
				}
				if _, err := w.Write(bc); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

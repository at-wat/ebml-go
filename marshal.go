package ebml

import (
	"bytes"
	"errors"
	"reflect"
)

var (
	errUnsupportedElement = errors.New("Unsupported element")
)

// Marshal struct to EBML bytes
func Marshal(val interface{}) ([]byte, error) {
	vo := reflect.ValueOf(val).Elem()
	to := reflect.TypeOf(val).Elem()

	return marshalImpl(vo, to)
}

func marshalImpl(vo reflect.Value, to reflect.Type) ([]byte, error) {
	var b []byte
	for i := 0; i < vo.NumField(); i++ {
		vn := vo.Field(i)
		tn := to.Field(i)

		if n, ok := tn.Tag.Lookup("ebml"); ok {
			if t, err := ElementTypeFromString(n); err == nil {
				e, ok := table[t]
				if !ok {
					return []byte{}, errUnsupportedElement
				}
				b = bytes.Join([][]byte{b, e.b}, []byte{})

				var bc []byte
				if e.t == TypeMaster {
					var err error
					bc, err = marshalImpl(vn, tn.Type)
					if err != nil {
						return []byte{}, err
					}
				} else {
					bc, err = perTypeEncoder[e.t](vn.Interface())
					if err != nil {
						return []byte{}, err
					}
				}
				bsz := encodeVInt(uint64(len(bc)))
				b = bytes.Join([][]byte{b, bsz, bc}, []byte{})
			}
		}
	}
	return b, nil
}

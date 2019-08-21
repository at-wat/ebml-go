package ebml

import (
	"errors"
	"io"
	"reflect"
)

const (
	sizeInf = 0xffffffffffffff
)

var (
	errUnknownElement = errors.New("Unknown element")
	errInvalidIntSize = errors.New("Invalid int size")
)

// Unmarshal EBML stream
func Unmarshal(r io.Reader, val interface{}) error {
	vo := reflect.ValueOf(val).Elem()
	to := reflect.TypeOf(val).Elem()

	for {
		if err := readElement(r, sizeInf, vo, to); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

func readElement(r0 io.Reader, n int64, vo reflect.Value, to reflect.Type) error {
	var r io.Reader
	if n != sizeInf {
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
			if n, ok := to.Field(i).Tag.Lookup("ebml"); ok {
				fieldMap[n] = field{vo.Field(i), to.Field(i).Type}
			}
		}
	}

	tb := revTable
	for {
		var bs [1]byte
		_, err := r.Read(bs[:])
		if err != nil {
			return err
		}
		b := bs[0]

		n, ok := tb[b]
		if !ok {
			return errUnknownElement
		}

		switch v := n.(type) {
		case elementRevTable:
			tb = v
		case element:
			size, err := readVInt(r)
			if err != nil {
				return err
			}
			var vnext reflect.Value
			var tnext reflect.Type
			if fm, ok := fieldMap[v.e.String()]; ok {
				vnext = fm.v
				tnext = fm.t
			}

			switch v.t {
			case TypeMaster:
				err := readElement(r, int64(size), vnext, tnext)
				if err != nil && err != io.EOF {
					return err
				}
			default:
				val, err := perTypeReader[v.t](r, size)
				if err != nil {
					return err
				}
				vr := reflect.ValueOf(val)
				if vnext.IsValid() && vnext.CanSet() && vr.Type() == vnext.Type() {
					vnext.Set(reflect.ValueOf(val))
				}
			}
			tb = revTable
		}
	}
}

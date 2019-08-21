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

	for {
		if err := readElement(r, sizeInf, vo); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

func readElement(r0 io.Reader, n int64, vo reflect.Value) error {
	var r io.Reader
	if n != sizeInf {
		r = io.LimitReader(r0, n)
	} else {
		r = r0
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
			if vo.IsValid() {
				vnext = vo.FieldByName(v.e.String())
			}
			switch v.t {
			case TypeMaster:
				err := readElement(r, int64(size), vnext)
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

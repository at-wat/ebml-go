package ebml

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"strings"
)

const (
	sizeInf = 0xffffffffffffff
)

var (
	errUnknownElement = errors.New("Unknown element")
	errInvalidIntSize = errors.New("Invalid int size")
	errIndefiniteType = errors.New("Unmarshal to indefinite type")
)

// Unmarshal EBML stream
func Unmarshal(r io.Reader, val interface{}) error {
	vo := reflect.ValueOf(val)
	if !vo.IsValid() {
		return errIndefiniteType
	}
	voe := vo.Elem()

	for {
		if _, err := readElement(r, sizeInf, voe, ElementRoot); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

func readElement(r0 io.Reader, n int64, vo reflect.Value, parent ElementType) (io.Reader, error) {
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
			if n, ok := vo.Type().Field(i).Tag.Lookup("ebml"); ok {
				nn := strings.Split(n, ",")
				var name string
				if len(nn) > 0 {
					name = nn[0]
				} else {
					name = vo.Type().Field(i).Name
				}
				fieldMap[name] = field{vo.Field(i), vo.Type().Field(i).Type}
			}
		}
	}

	tb := revTable
	for {
		var bs [1]byte
		_, err := r.Read(bs[:])
		if err != nil {
			return nil, err
		}
		b := bs[0]

		n, ok := tb[b]
		if !ok {
			return nil, errUnknownElement
		}

		switch v := n.(type) {
		case elementRevTable:
			tb = v
		case element:
			size, err := readVInt(r)
			if err != nil {
				return nil, err
			}
			var vnext reflect.Value
			if fm, ok := fieldMap[v.e.String()]; ok {
				vnext = fm.v
			}

			switch v.t {
			case TypeMaster:
				if v.top && v.e == parent {
					b := bytes.Join([][]byte{table[v.e].b, encodeVInt(sizeInf)}, []byte{})
					return bytes.NewBuffer(b), io.EOF
				}
				var vn reflect.Value
				if vnext.IsValid() && vnext.CanSet() {
					if vnext.Kind() == reflect.Ptr {
						vnext.Set(reflect.New(vnext.Type().Elem()))
						vn = vnext.Elem()
					} else if vnext.Kind() == reflect.Slice {
						vnext.Set(reflect.Append(vnext, reflect.New(vnext.Type().Elem()).Elem()))
						vn = vnext.Index(vnext.Len() - 1)
					} else {
						vn = vnext
					}
				}
				r0, err := readElement(r, int64(size), vn, v.e)
				if err != nil && err != io.EOF {
					return r0, err
				}
				if r0 != nil {
					r0 = io.MultiReader(r0, r)
				}
			default:
				val, err := perTypeReader[v.t](r, size)
				if err != nil {
					return nil, err
				}
				vr := reflect.ValueOf(val)
				if vnext.IsValid() && vnext.CanSet() {
					if vr.Type() == vnext.Type() {
						vnext.Set(reflect.ValueOf(val))
					} else if vnext.Kind() == reflect.Slice && vr.Type() == vnext.Type().Elem() {
						vnext.Set(reflect.Append(vnext, reflect.ValueOf(val)))
					}
				}
			}
			tb = revTable
		}
	}
}

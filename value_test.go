package ebml

import (
	"bytes"
	"reflect"
	"testing"
	"time"
)

func TestVInt(t *testing.T) {
	testCases := map[string]struct {
		b []byte
		i uint64
	}{
		"1 byte":     {[]byte{0x81}, 0x01},
		"2 bytes":    {[]byte{0x41, 0x23}, 0x0123},
		"3 bytes":    {[]byte{0x21, 0x23, 0x45}, 0x012345},
		"4 bytes":    {[]byte{0x11, 0x23, 0x45, 0x67}, 0x01234567},
		"5 bytes":    {[]byte{0x09, 0x23, 0x45, 0x67, 0x89}, 0x0123456789},
		"6 bytes":    {[]byte{0x05, 0x23, 0x45, 0x67, 0x89, 0xab}, 0x0123456789ab},
		"7 bytes":    {[]byte{0x03, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd}, 0x0123456789abcd},
		"8 bytes":    {[]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}, 0x23456789abcdef},
		"Indefinite": {[]byte{0x01, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, sizeInf},
	}

	for n, c := range testCases {
		t.Run("Decode "+n, func(t *testing.T) {
			r, err := readVInt(bytes.NewBuffer(c.b))
			if err != nil {
				t.Fatalf("Failed to readVInt: %v", err)
			}
			if r != c.i {
				t.Errorf("Unexpected readVInt result, expected: %d, got: %d", c.i, r)
			}
		})
	}
	for n, c := range testCases {
		t.Run("Encode "+n, func(t *testing.T) {
			b := encodeVInt(c.i)
			if bytes.Compare(b, c.b) != 0 {
				t.Errorf("Unexpected encodeVInt result, expected: %d, got: %d", c.b, b)
			}
		})
	}
}

func TestValue(t *testing.T) {
	testCases := map[string]struct {
		b    []byte
		t    Type
		v    interface{}
		vEnc interface{}
	}{
		"Binary":  {[]byte{0x01, 0x02, 0x03}, TypeBinary, []byte{0x01, 0x02, 0x03}, nil},
		"String":  {[]byte{0x31, 0x32, 0x00}, TypeString, "12", nil},
		"Int(3B)": {[]byte{0x01, 0x02, 0x03}, TypeInt, int64(0x010203), nil},
		"Int(4B)": {[]byte{0x01, 0x02, 0x03, 0x04}, TypeInt, int64(0x01020304), nil},
		"Int(5B)": {[]byte{0x01, 0x02, 0x03, 0x04, 0x05}, TypeInt, int64(0x0102030405), nil},
		"Int(6B)": {[]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, TypeInt, int64(0x010203040506), nil},
		"Int(7B)": {[]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}, TypeInt, int64(0x01020304050607), nil},
		"Int(8B)": {[]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}, TypeInt, int64(0x0102030405060708), nil},
		"UInt":    {[]byte{0x01, 0x02, 0x03}, TypeUInt, uint64(0x010203), nil},
		"Date":    {[]byte{0x01, 0x02, 0x03}, TypeDate, time.Unix(dateEpochInUnixtime, 0x010203), nil},
		"Float32": {[]byte{0x40, 0x10, 0x00, 0x00}, TypeFloat, float64(2.25), float32(2.25)},
		"Float64": {[]byte{0x40, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, TypeFloat, float64(2.25), nil},
		"Block": {[]byte{0x85, 0x12, 0x34, 0x80, 0x34, 0x56}, TypeBlock,
			Block{uint64(5), int16(0x1234), true, false, LacingNo, false, nil, [][]byte{[]byte{0x34, 0x56}}}, nil,
		},
	}
	for n, c := range testCases {
		t.Run("Read "+n, func(t *testing.T) {
			v, err := perTypeReader[c.t](bytes.NewBuffer(c.b), uint64(len(c.b)))
			if err != nil {
				t.Fatalf("Failed to read%s: %v", n, err)
			}
			if !reflect.DeepEqual(v, c.v) {
				t.Errorf("Unexpected read%s result, expected: %v, got: %v", n, c.v, v)
			}
		})
		t.Run("Encode "+n, func(t *testing.T) {
			var v interface{}
			if c.vEnc != nil {
				v = c.vEnc
			} else {
				v = c.v
			}
			b, err := perTypeEncoder[c.t](v)
			if err != nil {
				t.Fatalf("Failed to encode%s: %v", n, err)
			}
			if bytes.Compare(b, c.b) != 0 {
				t.Errorf("Unexpected encode%s result, expected: %v, got: %v", n, c.b, b)
			}
		})
	}
}

func TestEncodeValue_WrongInputType(t *testing.T) {
	testCases := map[string]struct {
		t   Type
		v   []interface{}
		err error
	}{
		"Binary": {
			TypeBinary,
			[]interface{}{"aaa", int64(1), uint64(1), time.Unix(1, 0), float32(1.0), float64(1.0), Block{}},
			errInvalidType,
		},
		"String": {
			TypeString,
			[]interface{}{[]byte{0x01}, int64(1), uint64(1), time.Unix(1, 0), float32(1.0), float64(1.0), Block{}},
			errInvalidType,
		},
		"Int": {
			TypeInt,
			[]interface{}{"aaa", []byte{0x01}, uint64(1), time.Unix(1, 0), float32(1.0), float64(1.0), Block{}},
			errInvalidType,
		},
		"UInt": {
			TypeUInt,
			[]interface{}{"aaa", []byte{0x01}, int64(1), time.Unix(1, 0), float32(1.0), float64(1.0), Block{}},
			errInvalidType,
		},
		"Date": {
			TypeDate,
			[]interface{}{"aaa", []byte{0x01}, int64(1), uint64(1), float32(1.0), float64(1.0), Block{}},
			errInvalidType,
		},
		"Float": {
			TypeFloat,
			[]interface{}{"aaa", []byte{0x01}, int64(1), uint64(1), time.Unix(1, 0), Block{}},
			errInvalidType,
		},
		"Block": {
			TypeBlock,
			[]interface{}{"aaa", []byte{0x01}, int64(1), uint64(1), time.Unix(1, 0), float32(1.0), float64(1.0)},
			errInvalidType,
		},
	}
	for n, c := range testCases {
		t.Run("Encode "+n, func(t *testing.T) {
			for _, v := range c.v {
				_, err := perTypeEncoder[c.t](v)
				if err != c.err {
					t.Fatalf("encode%s returned unexpected error to wrong input type: %v", n, err)
				}
			}
		})
	}
}

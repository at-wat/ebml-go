package ebml

import (
	"bytes"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/at-wat/ebml-go/internal/errs"
)

func TestDataSize(t *testing.T) {
	testCases := map[string]struct {
		b []byte
		i uint64
	}{
		"1 byte (upper bound)":  {[]byte{0xFE}, 0x80 - 2},
		"2 bytes (lower bound)": {[]byte{0x40, 0x7F}, 0x80 - 1},
		"2 bytes (upper bound)": {[]byte{0x7F, 0xFE}, 0x4000 - 2},
		"3 bytes (lower bound)": {[]byte{0x20, 0x3F, 0xFF}, 0x4000 - 1},
		"3 bytes (upper bound)": {[]byte{0x3F, 0xFF, 0xFE}, 0x200000 - 2},
		"4 bytes (lower bound)": {[]byte{0x10, 0x1F, 0xFF, 0xFF}, 0x200000 - 1},
		"4 bytes (upper bound)": {[]byte{0x1F, 0xFF, 0xFF, 0xFE}, 0x10000000 - 2},
		"5 bytes (lower bound)": {[]byte{0x08, 0x0F, 0xFF, 0xFF, 0xFF}, 0x10000000 - 1},
		"5 bytes (upper bound)": {[]byte{0x0F, 0xFF, 0xFF, 0xFF, 0xFE}, 0x800000000 - 2},
		"6 bytes (lower bound)": {[]byte{0x04, 0x07, 0xFF, 0xFF, 0xFF, 0xFF}, 0x800000000 - 1},
		"6 bytes (upper bound)": {[]byte{0x07, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE}, 0x40000000000 - 2},
		"7 bytes (lower bound)": {[]byte{0x02, 0x03, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, 0x40000000000 - 1},
		"7 bytes (upper bound)": {[]byte{0x03, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE}, 0x2000000000000 - 2},
		"8 bytes (lower bound)": {[]byte{0x01, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, 0x2000000000000 - 1},
		"8 bytes (upper bound)": {[]byte{0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE}, 0xffffffffffffff - 1},
		"Indefinite":            {[]byte{0x01, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, SizeUnknown},
	}

	vd := &valueDecoder{}

	for n, c := range testCases {
		t.Run("DecodeVInt "+n, func(t *testing.T) {
			r, _, err := vd.readVUInt(bytes.NewBuffer(c.b))
			if err != nil {
				t.Fatalf("Failed to readVUInt: '%v'", err)
			}
			if r != c.i {
				t.Errorf("Expected readVUInt result: %d, got: %d", c.i, r)
			}
		})
	}
	for n, c := range testCases {
		t.Run("DecodeDataSize "+n, func(t *testing.T) {
			r, _, err := vd.readDataSize(bytes.NewBuffer(c.b))
			if err != nil {
				t.Fatalf("Failed to readDataSize: '%v'", err)
			}
			if r != c.i {
				t.Errorf("Expected readVUInt result: %d, got: %d", c.i, r)
			}
		})
	}
	for n, c := range testCases {
		t.Run("Encode "+n, func(t *testing.T) {
			b := encodeDataSize(c.i, 0)
			if !bytes.Equal(b, c.b) {
				t.Errorf("Expected encodeDataSize result: %d, got: %d", c.b, b)
			}
		})
	}
}

func TestDataSize_Unknown(t *testing.T) {
	testCases := map[string][]byte{
		"1 byte":  {0xFF},
		"2 bytes": {0x7F, 0xFF},
		"3 bytes": {0x3F, 0xFF, 0xFF},
		"4 bytes": {0x1F, 0xFF, 0xFF, 0xFF},
		"5 bytes": {0x0F, 0xFF, 0xFF, 0xFF, 0xFF},
		"6 bytes": {0x07, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		"7 bytes": {0x03, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		"8 bytes": {0x01, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
	}

	vd := &valueDecoder{}

	for n, b := range testCases {
		t.Run("DecodeDataSize "+n, func(t *testing.T) {
			r, _, err := vd.readDataSize(bytes.NewBuffer(b))
			if err != nil {
				t.Fatalf("Failed to readDataSize: '%v'", err)
			}
			if r != SizeUnknown {
				t.Errorf("Expected readDataSize result: %d, got: %d", SizeUnknown, r)
			}
		})
	}
}

func TestElementID(t *testing.T) {
	testCases := map[string]struct {
		b []byte
		i uint64
	}{
		"1 byte (upper bound)":  {[]byte{0xFF}, 0x80 - 1},
		"2 bytes (lower bound)": {[]byte{0x40, 0x80}, 0x80},
		"2 bytes (upper bound)": {[]byte{0x7F, 0xFF}, 0x4000 - 1},
		"3 bytes (lower bound)": {[]byte{0x20, 0x40, 0x00}, 0x4000},
		"3 bytes (upper bound)": {[]byte{0x3F, 0xFF, 0xFF}, 0x200000 - 1},
		"4 bytes (lower bound)": {[]byte{0x10, 0x20, 0x00, 0x00}, 0x200000},
		"4 bytes (upper bound)": {[]byte{0x1F, 0xFF, 0xFF, 0xFF}, 0x10000000 - 1},
		"5 bytes (lower bound)": {[]byte{0x08, 0x10, 0x00, 0x00, 0x00}, 0x10000000},
		"5 bytes (upper bound)": {[]byte{0x0F, 0xFF, 0xFF, 0xFF, 0xFF}, 0x800000000 - 1},
		"6 bytes (lower bound)": {[]byte{0x04, 0x08, 0x00, 0x00, 0x00, 0x00}, 0x800000000},
		"6 bytes (upper bound)": {[]byte{0x07, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, 0x40000000000 - 1},
		"7 bytes (lower bound)": {[]byte{0x02, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00}, 0x40000000000},
		"7 bytes (upper bound)": {[]byte{0x03, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, 0x2000000000000 - 1},
	}

	vd := &valueDecoder{}

	for n, c := range testCases {
		t.Run("Decode "+n, func(t *testing.T) {
			r, _, err := vd.readVUInt(bytes.NewBuffer(c.b))
			if err != nil {
				t.Fatalf("Failed to readVUInt: '%v'", err)
			}
			if r != c.i {
				t.Errorf("Expected readVUInt result: %d, got: %d", c.i, r)
			}
		})
	}
	for n, c := range testCases {
		t.Run("Encode "+n, func(t *testing.T) {
			b, err := encodeElementID(c.i)
			if err != nil {
				t.Fatalf("Failed to encodeElementID: '%v'", err)
			}
			if !bytes.Equal(b, c.b) {
				t.Errorf("Expected encodeDataSize result: %d, got: %d", c.b, b)
			}
		})
	}

	_, err := encodeElementID(0x2000000000000)
	if err != ErrUnsupportedElementID {
		t.Errorf("Expected error type result: %s, got: %s", ErrUnsupportedElementID, err)
	}
}

func TestVInt(t *testing.T) {
	testCases := map[string]struct {
		b []byte
		i int64
	}{
		"1 byte (lower bound)":  {[]byte{0x80}, -0x3F},
		"1 byte (upper bound)":  {[]byte{0xFE}, 0x3F},
		"2 bytes (lower bound)": {[]byte{0x40, 0x00}, -0x1FFF},
		"2 bytes (upper bound)": {[]byte{0x7F, 0xFE}, 0x1FFF},
		"3 bytes (lower bound)": {[]byte{0x20, 0x00, 0x00}, -0xFFFFF},
		"3 bytes (upper bound)": {[]byte{0x3F, 0xFF, 0xFE}, 0xFFFFF},
		"4 bytes (lower bound)": {[]byte{0x10, 0x00, 0x00, 0x00}, -0x7FFFFFF},
		"4 bytes (upper bound)": {[]byte{0x1F, 0xFF, 0xFF, 0xFE}, 0x7FFFFFF},
		"5 bytes (lower bound)": {[]byte{0x08, 0x00, 0x00, 0x00, 0x00}, -0x3FFFFFFFF},
		"5 bytes (upper bound)": {[]byte{0x0F, 0xFF, 0xFF, 0xFF, 0xFE}, 0x3FFFFFFFF},
		"6 bytes (lower bound)": {[]byte{0x04, 0x00, 0x00, 0x00, 0x00, 0x00}, -0x1FFFFFFFFFF},
		"6 bytes (upper bound)": {[]byte{0x07, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE}, 0x1FFFFFFFFFF},
		"7 bytes (lower bound)": {[]byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, -0xFFFFFFFFFFFF},
		"7 bytes (upper bound)": {[]byte{0x03, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE}, 0xFFFFFFFFFFFF},
		"8 bytes (lower bound)": {[]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, -0x7FFFFFFFFFFFFF},
		"8 bytes (upper bound)": {[]byte{0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE}, 0x7FFFFFFFFFFFFF},
	}

	vd := &valueDecoder{}

	for n, c := range testCases {
		t.Run("Decode "+n, func(t *testing.T) {
			r, _, err := vd.readVInt(bytes.NewBuffer(c.b))
			if err != nil {
				t.Fatalf("Failed to readVUInt: '%v'", err)
			}
			if r != c.i {
				t.Errorf("Expected readVUInt result: %d, got: %d", c.i, r)
			}
		})
	}
	for n, c := range testCases {
		t.Run("Encode "+n, func(t *testing.T) {
			b, err := encodeVInt(c.i)
			if err != nil {
				t.Fatalf("Failed to encodeVInt: '%v'", err)
			}
			if !bytes.Equal(b, c.b) {
				t.Errorf("Expected encodeVInt result: %d, got: %d", c.b, b)
			}
		})
	}
}

func TestValue(t *testing.T) {
	testCases := map[string]struct {
		b    []byte
		t    DataType
		v    interface{}
		n    uint64
		vEnc interface{}
	}{
		"Binary":      {[]byte{0x01, 0x02, 0x03}, DataTypeBinary, []byte{0x01, 0x02, 0x03}, 0, nil},
		"Binary(4B)":  {[]byte{0x01, 0x02, 0x03, 0x00}, DataTypeBinary, []byte{0x01, 0x02, 0x03, 0x00}, 4, []byte{0x01, 0x02, 0x03}},
		"String":      {[]byte{0x31, 0x32}, DataTypeString, "12", 0, nil},
		"String(3B)":  {[]byte{0x31, 0x32, 0x00}, DataTypeString, "12", 3, nil},
		"String(4B)":  {[]byte{0x31, 0x32, 0x00, 0x00}, DataTypeString, "12", 4, nil},
		"Int8":        {[]byte{0x01}, DataTypeInt, int64(0x01), 0, nil},
		"Int16":       {[]byte{0x01, 0x02}, DataTypeInt, int64(0x0102), 0, nil},
		"Int24":       {[]byte{0x01, 0x02, 0x03}, DataTypeInt, int64(0x010203), 0, nil},
		"Int32":       {[]byte{0x01, 0x02, 0x03, 0x04}, DataTypeInt, int64(0x01020304), 0, nil},
		"Int40":       {[]byte{0x01, 0x02, 0x03, 0x04, 0x05}, DataTypeInt, int64(0x0102030405), 0, nil},
		"Int48":       {[]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, DataTypeInt, int64(0x010203040506), 0, nil},
		"Int56":       {[]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}, DataTypeInt, int64(0x01020304050607), 0, nil},
		"Int64":       {[]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}, DataTypeInt, int64(0x0102030405060708), 0, nil},
		"Int8(1B)":    {[]byte{0x01}, DataTypeInt, int64(0x01), 1, nil},
		"Int16(2B)":   {[]byte{0x01, 0x02}, DataTypeInt, int64(0x0102), 2, nil},
		"Int24(3B)":   {[]byte{0x01, 0x02, 0x03}, DataTypeInt, int64(0x010203), 3, nil},
		"Int32(4B)":   {[]byte{0x01, 0x02, 0x03, 0x04}, DataTypeInt, int64(0x01020304), 4, nil},
		"Int40(5B)":   {[]byte{0x01, 0x02, 0x03, 0x04, 0x05}, DataTypeInt, int64(0x0102030405), 5, nil},
		"Int48(6B)":   {[]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, DataTypeInt, int64(0x010203040506), 6, nil},
		"Int56(7B)":   {[]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}, DataTypeInt, int64(0x01020304050607), 7, nil},
		"Int64(8B)":   {[]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}, DataTypeInt, int64(0x0102030405060708), 8, nil},
		"Int8(2B)":    {[]byte{0x00, 0x01}, DataTypeInt, int64(0x01), 2, nil},
		"Int16(3B)":   {[]byte{0x00, 0x01, 0x02}, DataTypeInt, int64(0x0102), 3, nil},
		"Int24(4B)":   {[]byte{0x00, 0x01, 0x02, 0x03}, DataTypeInt, int64(0x010203), 4, nil},
		"Int32(5B)":   {[]byte{0x00, 0x01, 0x02, 0x03, 0x04}, DataTypeInt, int64(0x01020304), 5, nil},
		"Int40(6B)":   {[]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, DataTypeInt, int64(0x0102030405), 6, nil},
		"Int48(7B)":   {[]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, DataTypeInt, int64(0x010203040506), 7, nil},
		"Int56(8B)":   {[]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 07}, DataTypeInt, int64(0x01020304050607), 8, nil},
		"UInt":        {[]byte{0x01, 0x02, 0x03}, DataTypeUInt, uint64(0x010203), 0, nil},
		"Date":        {[]byte{0x01, 0x02, 0x03}, DataTypeDate, time.Unix(DateEpochInUnixtime, 0x010203), 0, nil},
		"Float32":     {[]byte{0x40, 0x10, 0x00, 0x00}, DataTypeFloat, float64(2.25), 0, float32(2.25)},
		"Float64":     {[]byte{0x40, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, DataTypeFloat, float64(2.25), 0, nil},
		"Float32(8B)": {[]byte{0x40, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, DataTypeFloat, float64(2.25), 8, float32(2.25)},
		"Float64(4B)": {[]byte{0x40, 0x10, 0x00, 0x00}, DataTypeFloat, float64(2.25), 4, float64(2.25)},
		"Float32(4B)": {[]byte{0x40, 0x10, 0x00, 0x00}, DataTypeFloat, float64(2.25), 4, float32(2.25)},
		"Float64(8B)": {[]byte{0x40, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, DataTypeFloat, float64(2.25), 8, nil},
		"Block": {[]byte{0x85, 0x12, 0x34, 0x80, 0x34, 0x56}, DataTypeBlock,
			Block{uint64(5), int16(0x1234), true, false, LacingNo, false, [][]byte{{0x34, 0x56}}}, 0, nil,
		},
		"ConvertInt8":   {[]byte{0x01}, DataTypeInt, int64(0x01), 0, int8(0x01)},
		"ConvertInt16":  {[]byte{0x01, 0x02}, DataTypeInt, int64(0x0102), 0, int16(0x0102)},
		"ConvertInt32":  {[]byte{0x01, 0x02, 0x03, 0x04}, DataTypeInt, int64(0x01020304), 0, int32(0x01020304)},
		"ConvertInt":    {[]byte{0x01, 0x02, 0x03, 0x04}, DataTypeInt, int64(0x01020304), 0, int(0x01020304)},
		"ConvertUInt8":  {[]byte{0x01}, DataTypeUInt, uint64(0x01), 0, uint8(0x01)},
		"ConvertUInt16": {[]byte{0x01, 0x02}, DataTypeUInt, uint64(0x0102), 0, uint16(0x0102)},
		"ConvertUInt32": {[]byte{0x01, 0x02, 0x03, 0x04}, DataTypeUInt, uint64(0x01020304), 0, uint32(0x01020304)},
		"ConvertUInt":   {[]byte{0x01, 0x02, 0x03, 0x04}, DataTypeUInt, uint64(0x01020304), 0, uint(0x01020304)},
	}

	vd := &valueDecoder{}

	for n, c := range testCases {
		t.Run("Read "+n, func(t *testing.T) {
			v, err := vd.decode(c.t, bytes.NewBuffer(c.b), uint64(len(c.b)))
			if err != nil {
				t.Fatalf("Failed to read%s: '%v'", n, err)
			}
			if !reflect.DeepEqual(v, c.v) {
				t.Errorf("Expected read%s result: %v, got: %v", n, c.v, v)
			}
		})
		t.Run("Encode "+n, func(t *testing.T) {
			var v interface{}
			if c.vEnc != nil {
				v = c.vEnc
			} else {
				v = c.v
			}
			b, err := perTypeEncoder[c.t](v, c.n)
			if err != nil {
				t.Fatalf("Failed to encode%s: '%v'", n, err)
			}
			if !bytes.Equal(b, c.b) {
				t.Errorf("Expected encode%s result: %v, got: %v", n, c.b, b)
			}
		})
	}
}

func TestEncodeValue_WrongInputType(t *testing.T) {
	testCases := []struct {
		t   DataType
		v   []interface{}
		err error
	}{
		{
			DataTypeBinary,
			[]interface{}{"aaa", int64(1), uint64(1), time.Unix(1, 0), float32(1.0), float64(1.0), Block{}},
			ErrInvalidType,
		},
		{
			DataTypeString,
			[]interface{}{[]byte{0x01}, int64(1), uint64(1), time.Unix(1, 0), float32(1.0), float64(1.0), Block{}},
			ErrInvalidType,
		},
		{
			DataTypeInt,
			[]interface{}{"aaa", []byte{0x01}, uint64(1), time.Unix(1, 0), float32(1.0), float64(1.0), Block{}},
			ErrInvalidType,
		},
		{
			DataTypeUInt,
			[]interface{}{"aaa", []byte{0x01}, int64(1), time.Unix(1, 0), float32(1.0), float64(1.0), Block{}},
			ErrInvalidType,
		},
		{
			DataTypeDate,
			[]interface{}{"aaa", []byte{0x01}, int64(1), uint64(1), float32(1.0), float64(1.0), Block{}},
			ErrInvalidType,
		},
		{
			DataTypeFloat,
			[]interface{}{"aaa", []byte{0x01}, int64(1), uint64(1), time.Unix(1, 0), Block{}},
			ErrInvalidType,
		},
		{
			DataTypeBlock,
			[]interface{}{"aaa", []byte{0x01}, int64(1), uint64(1), time.Unix(1, 0), float32(1.0), float64(1.0)},
			ErrInvalidType,
		},
	}
	for _, c := range testCases {
		t.Run("Encode "+c.t.String(), func(t *testing.T) {
			for _, v := range c.v {
				_, err := perTypeEncoder[c.t](v, 0)
				if !errs.Is(err, c.err) {
					t.Fatalf("Expected error against wrong input type %s: '%v, got: '%v'", c.t.String(), c.err, err)
				}
			}
		})
	}
}

func TestEncodeValue_WrongSize(t *testing.T) {
	testCases := map[string]struct {
		t   DataType
		v   interface{}
		n   uint64
		err error
	}{
		"Float32(3B)": {
			DataTypeFloat,
			float32(1.0),
			3,
			ErrInvalidFloatSize,
		},
		"Float64(9B)": {
			DataTypeFloat,
			float64(1.0),
			9,
			ErrInvalidFloatSize,
		},
	}
	for n, c := range testCases {
		t.Run("Encode "+n, func(t *testing.T) {
			_, err := perTypeEncoder[c.t](c.v, c.n)
			if !errs.Is(err, c.err) {
				t.Fatalf("Expected error against wrong input type %s: '%v', got: '%v'", n, c.err, err)
			}
		})
	}
}

func TestEncodeValue_OutOfRange(t *testing.T) {
	_, err := encodeVInt(1<<63 - 1)
	if err != ErrOutOfRange {
		t.Fatalf("Expected error: '%v', got: '%v'", ErrOutOfRange, err)
	}
}

func TestReadValue_WrongSize(t *testing.T) {
	testCases := map[string]struct {
		t   DataType
		b   []byte
		n   uint64
		err error
	}{
		"Float32(3B)": {
			DataTypeFloat,
			[]byte{0, 0, 0},
			3,
			ErrInvalidFloatSize,
		},
	}

	vd := &valueDecoder{}

	for n, c := range testCases {
		t.Run("Read "+n, func(t *testing.T) {
			_, err := vd.decode(c.t, bytes.NewReader(c.b), c.n)
			if !errs.Is(err, c.err) {
				t.Fatalf("Expected error against wrong data size of %s: %v, got: %v", n, c.err, err)
			}
		})
	}
}

func TestReadValue_ReadUnexpectedEOF(t *testing.T) {
	testCases := []struct {
		t DataType
		b []byte
	}{
		{DataTypeBinary, []byte{0x00, 0x00}},
		{DataTypeString, []byte{0x00, 0x00}},
		{DataTypeInt, []byte{0x00, 0x00}},
		{DataTypeUInt, []byte{0x00, 0x00}},
		{DataTypeDate, []byte{0x00, 0x00}},
		{DataTypeFloat, []byte{0x00, 0x00, 0x00, 0x00}},
	}

	vd := &valueDecoder{}

	for _, c := range testCases {
		t.Run("Read "+c.t.String(), func(t *testing.T) {
			for l := 0; l < len(c.b)-1; l++ {
				r := bytes.NewReader(c.b[:l])
				_, err := vd.decode(c.t, r, uint64(len(c.b)))
				if !errs.Is(err, io.ErrUnexpectedEOF) {
					t.Errorf("Expected error against short (%d bytes) %s: %v, got: %v",
						l, c.t.String(), io.ErrUnexpectedEOF, err)
				}
			}
		})
	}
}

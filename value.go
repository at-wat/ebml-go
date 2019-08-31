package ebml

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"time"
)

const (
	dateEpochInUnixtime = 978307200
)

var (
	errInvalidFloatSize = errors.New("Invalid float size")
	errInvalidType      = errors.New("Invalid type")
)

var perTypeReader = map[Type]func(io.Reader, uint64) (interface{}, error){
	TypeInt:    readInt,
	TypeUInt:   readUInt,
	TypeDate:   readDate,
	TypeFloat:  readFloat,
	TypeBinary: readBinary,
	TypeString: readString,
	TypeBlock:  readBlock,
}

func readVInt(r io.Reader) (uint64, error) {
	var bs [1]byte
	_, err := r.Read(bs[:])
	if err != nil {
		return 0, err
	}

	var vc int
	var value uint64

	b := bs[0]
	if b&0x80 == 0x80 {
		vc = 0
		value = uint64(b & 0x7F)
	} else if b&0xC0 == 0x40 {
		vc = 1
		value = uint64(b & 0x3F)
	} else if b&0xE0 == 0x20 {
		vc = 2
		value = uint64(b & 0x1F)
	} else if b&0xF0 == 0x10 {
		vc = 3
		value = uint64(b & 0x0F)
	} else if b&0xF8 == 0x08 {
		vc = 4
		value = uint64(b & 0x07)
	} else if b&0xFC == 0x04 {
		vc = 5
		value = uint64(b & 0x03)
	} else if b&0xFE == 0x02 {
		vc = 6
		value = uint64(b & 0x01)
	} else if b == 0x01 {
		vc = 7
		value = 0
	}

	for {
		if vc == 0 {
			return value, nil
		}

		var bs [1]byte
		_, err := r.Read(bs[:])
		if err != nil {
			return 0, err
		}
		value = value<<8 | uint64(bs[0])
		vc--
	}
}
func readBinary(r io.Reader, n uint64) (interface{}, error) {
	bs := make([]byte, n)
	_, err := r.Read(bs)
	if err != nil {
		return []byte{}, err
	}
	return bs, nil
}
func readString(r io.Reader, n uint64) (interface{}, error) {
	bs, err := readBinary(r, n)
	if err != nil {
		return "", err
	}
	s := string(bs.([]byte))
	// null terminated
	if s[len(s)-1] == '\x00' {
		s = s[:len(s)-1]
	}
	return s, nil
}
func readInt(r io.Reader, n uint64) (interface{}, error) {
	bs := make([]byte, n)
	_, err := r.Read(bs[:])
	if err != nil {
		return 0, err
	}
	var v int64
	for _, b := range bs {
		v = v<<8 | int64(b)
	}
	return v, nil
}
func readUInt(r io.Reader, n uint64) (interface{}, error) {
	bs := make([]byte, n)
	_, err := r.Read(bs[:])
	if err != nil {
		return 0, err
	}
	var v uint64
	for _, b := range bs {
		v = v<<8 | uint64(b)
	}
	return v, nil
}
func readDate(r io.Reader, n uint64) (interface{}, error) {
	i, err := readInt(r, n)
	if err != nil {
		return time.Unix(0, 0), err
	}
	return time.Unix(dateEpochInUnixtime, i.(int64)), nil
}
func readFloat(r io.Reader, n uint64) (interface{}, error) {
	if n != 4 && n != 8 {
		return 0.0, errInvalidFloatSize
	}
	bs := make([]byte, n)
	_, err := r.Read(bs[:])
	if err != nil {
		return 0, err
	}
	switch n {
	case 4:
		return float64(math.Float32frombits(binary.BigEndian.Uint32(bs))), nil
	case 8:
		return math.Float64frombits(binary.BigEndian.Uint64(bs)), nil
	default:
		panic("Invalid float size validation")
	}
}
func readBlock(r io.Reader, n uint64) (interface{}, error) {
	b, err := UnmarshalBlock(io.LimitReader(r, int64(n)))
	if err != nil {
		return nil, err
	}
	return *b, nil
}

var perTypeEncoder = map[Type]func(interface{}) ([]byte, error){
	TypeInt:    encodeInt,
	TypeUInt:   encodeUInt,
	TypeDate:   encodeDate,
	TypeFloat:  encodeFloat,
	TypeBinary: encodeBinary,
	TypeString: encodeString,
}

func encodeVInt(v uint64) []byte {
	if v < 0x80 {
		return []byte{byte(v) | 0x80}
	} else if v < 0x4000 {
		return []byte{byte(v>>8) | 0x40, byte(v)}
	} else if v < 0x200000 {
		return []byte{byte(v>>16) | 0x20, byte(v >> 8), byte(v)}
	} else if v < 0x10000000 {
		return []byte{byte(v>>24) | 0x10, byte(v >> 16), byte(v >> 8), byte(v)}
	} else if v < 0x80000000 {
		return []byte{byte(v>>32) | 0x8, byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}
	} else if v < 0x4000000000 {
		return []byte{byte(v>>40) | 0x4, byte(v >> 32), byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}
	} else if v < 0x2000000000 {
		return []byte{byte(v>>48) | 0x2, byte(v >> 40), byte(v >> 32), byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}
	} else if v < 0x1000000000 {
		return []byte{byte(v>>56) | 0x1, byte(v >> 48), byte(v >> 40), byte(v >> 32), byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}
	} else {
		return []byte{0x01, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	}
}
func encodeBinary(i interface{}) ([]byte, error) {
	v, ok := i.([]byte)
	if !ok {
		return []byte{}, errInvalidType
	}
	return v, nil
}
func encodeString(i interface{}) ([]byte, error) {
	v, ok := i.(string)
	if !ok {
		return []byte{}, errInvalidType
	}
	return append([]byte(v), 0x00), nil
}
func encodeInt(i interface{}) ([]byte, error) {
	v, ok := i.(int64)
	if !ok {
		return []byte{}, errInvalidType
	}
	return encodeUInt(uint64(v))
}
func encodeUInt(i interface{}) ([]byte, error) {
	v, ok := i.(uint64)
	if !ok {
		return []byte{}, errInvalidType
	}
	if v < 0x100 {
		return []byte{byte(v)}, nil
	} else if v < 0x10000 {
		return []byte{byte(v >> 8), byte(v)}, nil
	} else if v < 0x1000000 {
		return []byte{byte(v >> 16), byte(v >> 8), byte(v)}, nil
	} else if v < 0x100000000 {
		return []byte{byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}, nil
	} else if v < 0x10000000000 {
		return []byte{byte(v >> 32), byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}, nil
	} else if v < 0x1000000000000 {
		return []byte{byte(v >> 40), byte(v >> 32), byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}, nil
	} else if v < 0x100000000000000 {
		return []byte{byte(v >> 48), byte(v >> 40), byte(v >> 32), byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}, nil
	} else {
		return []byte{byte(v >> 56), byte(v >> 48), byte(v >> 40), byte(v >> 32), byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}, nil
	}
}
func encodeDate(i interface{}) ([]byte, error) {
	v, ok := i.(time.Time)
	if !ok {
		return []byte{}, errInvalidType
	}
	dtns := v.Sub(time.Unix(dateEpochInUnixtime, 0)).Nanoseconds()
	return encodeInt(int64(dtns))
}
func encodeFloat(i interface{}) ([]byte, error) {
	var b []byte
	switch v := i.(type) {
	case float64:
		binary.BigEndian.PutUint64(b, math.Float64bits(v))
		return b, nil
	case float32:
		binary.BigEndian.PutUint32(b, math.Float32bits(v))
		return b, nil
	default:
		return []byte{}, errInvalidType
	}
}

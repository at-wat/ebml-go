package ebml

import (
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
)

var perTypeReader = map[Type]func(io.Reader, uint64) (interface{}, error){
	TypeInt:    readInt,
	TypeUInt:   readUInt,
	TypeDate:   readDate,
	TypeFloat:  readFloat,
	TypeBinary: readBinary,
	TypeString: readString,
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

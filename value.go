// Copyright 2019 The ebml-go authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ebml

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"strings"
	"time"
)

const (
	// DateEpochInUnixtime is the Unixtime of EBML date epoch.
	DateEpochInUnixtime = 978307200
	// SizeUnknown is the longest unknown size value.
	SizeUnknown = 0xffffffffffffff
)

// ErrInvalidFloatSize means that a element size is invalid for float type. Float must be 4 or 8 bytes.
var ErrInvalidFloatSize = errors.New("invalid float size")

// ErrInvalidType means that a value is not convertible to the element data.
var ErrInvalidType = errors.New("invalid type")

// ErrUnsupportedElementID means that a value is out of range of EBML encoding.
var ErrUnsupportedElementID = errors.New("unsupported Element ID")

var perTypeReader = map[DataType]func(io.Reader, uint64) (interface{}, error){
	DataTypeInt:    readInt,
	DataTypeUInt:   readUInt,
	DataTypeDate:   readDate,
	DataTypeFloat:  readFloat,
	DataTypeBinary: readBinary,
	DataTypeString: readString,
	DataTypeBlock:  readBlock,
}

func readDataSize(r io.Reader) (uint64, int, error) {
	v, n, err := readVInt(r)
	if v == (uint64(0xFFFFFFFFFFFFFFFF) >> uint(64-n*7)) {
		return SizeUnknown, n, err
	}
	return v, n, err
}
func readVInt(r io.Reader) (uint64, int, error) {
	var bs [1]byte
	bytesRead, err := io.ReadFull(r, bs[:])
	switch err {
	case nil:
	case io.EOF:
		return 0, bytesRead, io.ErrUnexpectedEOF
	default:
		return 0, bytesRead, err
	}

	var vc int
	var value uint64

	b := bs[0]
	switch {
	case b&0x80 == 0x80:
		vc = 0
		value = uint64(b & 0x7F)
	case b&0xC0 == 0x40:
		vc = 1
		value = uint64(b & 0x3F)
	case b&0xE0 == 0x20:
		vc = 2
		value = uint64(b & 0x1F)
	case b&0xF0 == 0x10:
		vc = 3
		value = uint64(b & 0x0F)
	case b&0xF8 == 0x08:
		vc = 4
		value = uint64(b & 0x07)
	case b&0xFC == 0x04:
		vc = 5
		value = uint64(b & 0x03)
	case b&0xFE == 0x02:
		vc = 6
		value = uint64(b & 0x01)
	case b == 0x01:
		vc = 7
		value = 0
	}

	for {
		if vc == 0 {
			return value, bytesRead, nil
		}

		var bs [1]byte
		n, err := io.ReadFull(r, bs[:])
		switch err {
		case nil:
		case io.EOF:
			return 0, bytesRead, io.ErrUnexpectedEOF
		default:
			return 0, bytesRead, err
		}
		bytesRead += n
		value = value<<8 | uint64(bs[0])
		vc--
	}
}
func readBinary(r io.Reader, n uint64) (interface{}, error) {
	bs := make([]byte, n)

	switch _, err := io.ReadFull(r, bs); err {
	case nil:
		return bs, nil
	case io.EOF:
		return bs, io.ErrUnexpectedEOF
	default:
		return []byte{}, err
	}
}
func readString(r io.Reader, n uint64) (interface{}, error) {
	bs, err := readBinary(r, n)
	if err != nil {
		return "", err
	}
	s := string(bs.([]byte))
	// Remove trailing null characters
	ss := strings.Split(s, "\x00")
	return ss[0], nil
}
func readInt(r io.Reader, n uint64) (interface{}, error) {
	v, err := readUInt(r, n)
	if err != nil {
		return 0, err
	}
	v64 := v.(uint64)
	if n != 8 && (v64&(1<<(n*8-1))) != 0 {
		// negative value
		for i := n; i < 8; i++ {
			v64 |= 0xFF << (i * 8)
		}
	}
	return int64(v64), nil
}
func readUInt(r io.Reader, n uint64) (interface{}, error) {
	bs := make([]byte, n)

	switch _, err := io.ReadFull(r, bs); err {
	case nil:
	case io.EOF:
		return 0, io.ErrUnexpectedEOF
	default:
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
	return time.Unix(DateEpochInUnixtime, i.(int64)), nil
}
func readFloat(r io.Reader, n uint64) (interface{}, error) {
	bs := make([]byte, n)

	switch _, err := io.ReadFull(r, bs); err {
	case nil:
	case io.EOF:
		return bs, io.ErrUnexpectedEOF
	default:
		return []byte{}, err
	}

	switch n {
	case 4:
		return float64(math.Float32frombits(binary.BigEndian.Uint32(bs))), nil
	case 8:
		return math.Float64frombits(binary.BigEndian.Uint64(bs)), nil
	default:
		return 0.0, wrapErrorf(ErrInvalidFloatSize, "reading %d bytes float", n)
	}
}
func readBlock(r io.Reader, n uint64) (interface{}, error) {
	b, err := UnmarshalBlock(r, int64(n))
	if err != nil {
		return nil, err
	}
	return *b, nil
}

var perTypeEncoder = map[DataType]func(interface{}, uint64) ([]byte, error){
	DataTypeInt:    encodeInt,
	DataTypeUInt:   encodeUInt,
	DataTypeDate:   encodeDate,
	DataTypeFloat:  encodeFloat,
	DataTypeBinary: encodeBinary,
	DataTypeString: encodeString,
	DataTypeBlock:  encodeBlock,
}

func encodeDataSize(v, n uint64) []byte {
	switch {
	case v < 0x80-1 && n < 2:
		return []byte{byte(v) | 0x80}
	case v < 0x4000-1 && n < 3:
		return []byte{byte(v>>8) | 0x40, byte(v)}
	case v < 0x200000-1 && n < 4:
		return []byte{byte(v>>16) | 0x20, byte(v >> 8), byte(v)}
	case v < 0x10000000-1 && n < 5:
		return []byte{byte(v>>24) | 0x10, byte(v >> 16), byte(v >> 8), byte(v)}
	case v < 0x800000000-1 && n < 6:
		return []byte{byte(v>>32) | 0x8, byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}
	case v < 0x40000000000-1 && n < 7:
		return []byte{byte(v>>40) | 0x4, byte(v >> 32), byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}
	case v < 0x2000000000000-1 && n < 8:
		return []byte{byte(v>>48) | 0x2, byte(v >> 40), byte(v >> 32), byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}
	case v < SizeUnknown:
		return []byte{0x1, byte(v >> 48), byte(v >> 40), byte(v >> 32), byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}
	default:
		return []byte{0x01, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	}
}
func encodeElementID(v uint64) ([]byte, error) {
	switch {
	case v < 0x80:
		return []byte{byte(v) | 0x80}, nil
	case v < 0x4000:
		return []byte{byte(v>>8) | 0x40, byte(v)}, nil
	case v < 0x200000:
		return []byte{byte(v>>16) | 0x20, byte(v >> 8), byte(v)}, nil
	case v < 0x10000000:
		return []byte{byte(v>>24) | 0x10, byte(v >> 16), byte(v >> 8), byte(v)}, nil
	case v < 0x800000000:
		return []byte{byte(v>>32) | 0x8, byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}, nil
	case v < 0x40000000000:
		return []byte{byte(v>>40) | 0x4, byte(v >> 32), byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}, nil
	case v < 0x2000000000000:
		return []byte{byte(v>>48) | 0x2, byte(v >> 40), byte(v >> 32), byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}, nil
	}
	return nil, ErrUnsupportedElementID
}
func encodeBinary(i interface{}, n uint64) ([]byte, error) {
	v, ok := i.([]byte)
	if !ok {
		return []byte{}, ErrInvalidType
	}
	if uint64(len(v)) >= n {
		return v, nil
	}
	return append(v, bytes.Repeat([]byte{0x00}, int(n)-len(v))...), nil
}
func encodeString(i interface{}, n uint64) ([]byte, error) {
	v, ok := i.(string)
	if !ok {
		return []byte{}, wrapErrorf(ErrInvalidType, "writing %T as string", i)
	}
	if uint64(len(v)+1) >= n {
		return append([]byte(v), 0x00), nil
	}
	return append([]byte(v), bytes.Repeat([]byte{0x00}, int(n)-len(v))...), nil
}
func encodeInt(i interface{}, n uint64) ([]byte, error) {
	var v int64
	switch v2 := i.(type) {
	case int:
		v = int64(v2)
	case int8:
		v = int64(v2)
	case int16:
		v = int64(v2)
	case int32:
		v = int64(v2)
	case int64:
		v = v2
	default:
		return []byte{}, wrapErrorf(ErrInvalidType, "writing %T as int", i)
	}
	return encodeUInt(uint64(v), n)
}
func encodeUInt(i interface{}, n uint64) ([]byte, error) {
	var v uint64
	switch v2 := i.(type) {
	case uint:
		v = uint64(v2)
	case uint8:
		v = uint64(v2)
	case uint16:
		v = uint64(v2)
	case uint32:
		v = uint64(v2)
	case uint64:
		v = v2
	default:
		return []byte{}, wrapErrorf(ErrInvalidType, "writing %T as uint", i)
	}
	switch {
	case v < 0x100 && n < 2:
		return []byte{byte(v)}, nil
	case v < 0x10000 && n < 3:
		return []byte{byte(v >> 8), byte(v)}, nil
	case v < 0x1000000 && n < 4:
		return []byte{byte(v >> 16), byte(v >> 8), byte(v)}, nil
	case v < 0x100000000 && n < 5:
		return []byte{byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}, nil
	case v < 0x10000000000 && n < 6:
		return []byte{byte(v >> 32), byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}, nil
	case v < 0x1000000000000 && n < 7:
		return []byte{byte(v >> 40), byte(v >> 32), byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}, nil
	case v < 0x100000000000000 && n < 8:
		return []byte{byte(v >> 48), byte(v >> 40), byte(v >> 32), byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}, nil
	default:
		return []byte{byte(v >> 56), byte(v >> 48), byte(v >> 40), byte(v >> 32), byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}, nil
	}
}
func encodeDate(i interface{}, n uint64) ([]byte, error) {
	v, ok := i.(time.Time)
	if !ok {
		return []byte{}, wrapErrorf(ErrInvalidType, "writing %T as date", i)
	}
	dtns := v.Sub(time.Unix(DateEpochInUnixtime, 0)).Nanoseconds()
	return encodeInt(int64(dtns), n)
}
func encodeFloat32(i float32) ([]byte, error) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, math.Float32bits(i))
	return b, nil
}
func encodeFloat64(i float64) ([]byte, error) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, math.Float64bits(i))
	return b, nil
}
func encodeFloat(i interface{}, n uint64) ([]byte, error) {
	switch v := i.(type) {
	case float64:
		switch n {
		case 0:
			return encodeFloat64(v)
		case 4:
			return encodeFloat32(float32(v))
		case 8:
			return encodeFloat64(v)
		default:
			return []byte{}, wrapErrorf(ErrInvalidFloatSize, "writing %d bytes float", n)
		}
	case float32:
		switch n {
		case 0:
			return encodeFloat32(v)
		case 4:
			return encodeFloat32(v)
		case 8:
			return encodeFloat64(float64(v))
		default:
			return []byte{}, wrapErrorf(ErrInvalidFloatSize, "writing %d bytes float", n)
		}
	default:
		return []byte{}, wrapErrorf(ErrInvalidType, "writing %T as float", i)
	}
}
func encodeBlock(i interface{}, n uint64) ([]byte, error) {
	v, ok := i.(Block)
	if !ok {
		return []byte{}, wrapErrorf(ErrInvalidType, "writing %T as block", i)
	}
	var b bytes.Buffer
	if err := MarshalBlock(&v, &b); err != nil {
		return []byte{}, err
	}
	return b.Bytes(), nil
}

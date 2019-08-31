package ebml

import (
	"errors"
	"io"
	"io/ioutil"
)

var (
	errLaceUnimplemented = errors.New("Lacing in unimplementd")
)

// LacingMode is type of laced data
type LacingMode uint8

// Type of laced data
const (
	LacingNo    LacingMode = 0
	LacingXiph  LacingMode = 1
	LacingFixed LacingMode = 2
	LacingEBML  LacingMode = 3
)

const (
	blockFlagMaskKeyframe    = 0x01
	blockFlagMaskInvisible   = 0x10
	blockFlagMaskLacing      = 0x60
	blockFlagMaskDiscardable = 0x80
)

// Block represents EBML Block/SimpleBlock element
type Block struct {
	TrackNumber uint64
	Timecode    int16
	Keyframe    bool
	Invisible   bool
	Lacing      LacingMode
	Discardable bool
	Lace        *Lace
	Data        [][]byte
}

// Lace represents Lace header of EBML Block/SimpleBlock element
type Lace struct {
	NumFrames uint8
	Len       []uint8
}

// UnmarshalBlock unmarshals EBML Block structure
func UnmarshalBlock(r io.Reader) (*Block, error) {
	var b Block
	var err error
	if b.TrackNumber, err = readVInt(r); err != nil {
		return nil, err
	}
	if v, err := readInt(r, 2); err != nil {
		return nil, err
	} else {
		b.Timecode = int16(v.(int64))
	}

	var bs [1]byte
	if _, err := r.Read(bs[:]); err != nil {
		return nil, err
	}
	if bs[0]&blockFlagMaskKeyframe != 0 {
		b.Keyframe = true
	}
	if bs[0]&blockFlagMaskInvisible != 0 {
		b.Invisible = true
	}
	if bs[0]&blockFlagMaskDiscardable != 0 {
		b.Discardable = true
	}
	b.Lacing = LacingMode((bs[0] & blockFlagMaskLacing) >> 5)

	if b.Lacing != LacingNo {
		return nil, errLaceUnimplemented
	}

	b.Data = [][]byte{[]byte{}}
	b.Data[0], err = ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// MarshalBlock marshals EBML Block structure
func MarshalBlock(b *Block, w io.Writer) error {
	return nil
}

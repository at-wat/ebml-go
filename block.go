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
	"errors"
	"io"
	"io/ioutil"
)

var (
	errLaceUnimplemented = errors.New("lacing in unimplementd")
)

// LacingMode is type of laced data.
type LacingMode uint8

// Type of laced data.
const (
	LacingNo    LacingMode = 0
	LacingFixed LacingMode = 1
	LacingXiph  LacingMode = 2
	LacingEBML  LacingMode = 3
)

const (
	blockFlagMaskKeyframe    = 0x80
	blockFlagMaskInvisible   = 0x08
	blockFlagMaskLacing      = 0x06
	blockFlagMaskDiscardable = 0x01
)

// Block represents EBML Block/SimpleBlock element.
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

func (b *Block) packFlags() byte {
	var f byte
	if b.Keyframe {
		f |= blockFlagMaskKeyframe
	}
	if b.Invisible {
		f |= blockFlagMaskInvisible
	}
	if b.Discardable {
		f |= blockFlagMaskDiscardable
	}
	f |= byte(b.Lacing) << 1
	return f
}

// Lace represents Lace header of EBML Block/SimpleBlock element.
type Lace struct {
	NumFrames uint8
	Len       []uint8
}

// UnmarshalBlock unmarshals EBML Block structure.
func UnmarshalBlock(r io.Reader) (*Block, error) {
	var b Block
	var err error
	if b.TrackNumber, _, err = readVInt(r); err != nil {
		return nil, err
	}
	if v, err := readInt(r, 2); err == nil {
		b.Timecode = int16(v.(int64))
	} else {
		return nil, err
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
	b.Lacing = LacingMode((bs[0] & blockFlagMaskLacing) >> 1)

	if b.Lacing != LacingNo {
		return nil, errLaceUnimplemented
	}

	b.Data = [][]byte{{}}
	b.Data[0], err = ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// MarshalBlock marshals EBML Block structure.
func MarshalBlock(b *Block, w io.Writer) error {
	n, err := encodeElementID(b.TrackNumber)
	if err != nil {
		return err
	}
	if _, err := w.Write(n); err != nil {
		return err
	}
	if _, err := w.Write([]byte{byte(b.Timecode >> 8), byte(b.Timecode)}); err != nil {
		return err
	}
	if _, err := w.Write([]byte{b.packFlags()}); err != nil {
		return err
	}

	if b.Lacing != LacingNo {
		return errLaceUnimplemented
	}

	if _, err := w.Write(b.Data[0]); err != nil {
		return err
	}

	return nil
}

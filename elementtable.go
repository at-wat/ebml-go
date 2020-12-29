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
)

type elementDef struct {
	b   []byte
	t   DataType
	top bool
}
type elementTable map[ElementType]elementDef

var table = elementTable{
	ElementSeekHead:               elementDef{[]byte{0x11, 0x4D, 0x9B, 0x74}, DataTypeMaster, true},
	ElementTags:                   elementDef{[]byte{0x12, 0x54, 0xC3, 0x67}, DataTypeMaster, true},
	ElementInfo:                   elementDef{[]byte{0x15, 0x49, 0xA9, 0x66}, DataTypeMaster, true},
	ElementTracks:                 elementDef{[]byte{0x16, 0x54, 0xAE, 0x6B}, DataTypeMaster, true},
	ElementSegment:                elementDef{[]byte{0x18, 0x53, 0x80, 0x67}, DataTypeMaster, false},
	ElementEBML:                   elementDef{[]byte{0x1A, 0x45, 0xDF, 0xA3}, DataTypeMaster, false},
	ElementCues:                   elementDef{[]byte{0x1C, 0x53, 0xBB, 0x6B}, DataTypeMaster, true},
	ElementCluster:                elementDef{[]byte{0x1F, 0x43, 0xB6, 0x75}, DataTypeMaster, true},
	ElementLanguage:               elementDef{[]byte{0x22, 0xB5, 0x9C}, DataTypeString, false},
	ElementDefaultDuration:        elementDef{[]byte{0x23, 0xE3, 0x83}, DataTypeUInt, false},
	ElementCodecName:              elementDef{[]byte{0x25, 0x86, 0x88}, DataTypeString, false},
	ElementTimecodeScale:          elementDef{[]byte{0x2A, 0xD7, 0xB1}, DataTypeUInt, false},
	ElementEBMLVersion:            elementDef{[]byte{0x42, 0x86}, DataTypeUInt, false},
	ElementEBMLMaxIDLength:        elementDef{[]byte{0x42, 0xF2}, DataTypeUInt, false},
	ElementEBMLMaxSizeLength:      elementDef{[]byte{0x42, 0xF3}, DataTypeUInt, false},
	ElementEBMLReadVersion:        elementDef{[]byte{0x42, 0xF7}, DataTypeUInt, false},
	ElementEBMLDocType:            elementDef{[]byte{0x42, 0x82}, DataTypeString, false},
	ElementEBMLDocTypeReadVersion: elementDef{[]byte{0x42, 0x85}, DataTypeUInt, false},
	ElementEBMLDocTypeVersion:     elementDef{[]byte{0x42, 0x87}, DataTypeUInt, false},
	ElementDateUTC:                elementDef{[]byte{0x44, 0x61}, DataTypeDate, false},
	ElementTagBinary:              elementDef{[]byte{0x44, 0x85}, DataTypeBinary, false},
	ElementTagString:              elementDef{[]byte{0x44, 0x87}, DataTypeString, false},
	ElementDuration:               elementDef{[]byte{0x44, 0x89}, DataTypeFloat, false},
	ElementTagName:                elementDef{[]byte{0x45, 0xA3}, DataTypeString, false},
	ElementSeek:                   elementDef{[]byte{0x4D, 0xBB}, DataTypeMaster, false},
	ElementSeekID:                 elementDef{[]byte{0x53, 0xAB}, DataTypeBinary, false},
	ElementSeekPosition:           elementDef{[]byte{0x53, 0xAC}, DataTypeUInt, false},
	ElementMuxingApp:              elementDef{[]byte{0x4D, 0x80}, DataTypeString, false},
	ElementName:                   elementDef{[]byte{0x53, 0x6E}, DataTypeString, false},
	ElementCueBlockNumber:         elementDef{[]byte{0x53, 0x78}, DataTypeUInt, false},
	ElementDisplayWidth:           elementDef{[]byte{0x54, 0xB0}, DataTypeUInt, false},
	ElementDisplayHeight:          elementDef{[]byte{0x54, 0xBA}, DataTypeUInt, false},
	ElementFlagForced:             elementDef{[]byte{0x55, 0xAA}, DataTypeUInt, false},
	ElementColour:                 elementDef{[]byte{0x55, 0xB0}, DataTypeMaster, false},
	ElementMatrixCoefficients:     elementDef{[]byte{0x55, 0xB1}, DataTypeUInt, false},
	ElementChromaSitingHorz:       elementDef{[]byte{0x55, 0xB7}, DataTypeUInt, false},
	ElementChromaSitingVert:       elementDef{[]byte{0x55, 0xB8}, DataTypeUInt, false},
	ElementRange:                  elementDef{[]byte{0x55, 0xB9}, DataTypeUInt, false},
	ElementTransferCharacteristics:elementDef{[]byte{0x55, 0xBA}, DataTypeUInt, false},
	ElementPrimaries:              elementDef{[]byte{0x55, 0xBB}, DataTypeUInt, false},
	ElementMaxBlockAdditionID:     elementDef{[]byte{0x55, 0xEE}, DataTypeUInt, false},
	ElementCodecDelay:             elementDef{[]byte{0x56, 0xAA}, DataTypeUInt, false},
	ElementSeekPreRoll:            elementDef{[]byte{0x56, 0xBB}, DataTypeUInt, false},
	ElementWritingApp:             elementDef{[]byte{0x57, 0x41}, DataTypeString, false},
	ElementBitDepth:               elementDef{[]byte{0x62, 0x64}, DataTypeUInt, false},
	ElementCodecPrivate:           elementDef{[]byte{0x63, 0xA2}, DataTypeBinary, false},
	ElementSimpleTag:              elementDef{[]byte{0x67, 0xC8}, DataTypeMaster, false},
	ElementMinCache:               elementDef{[]byte{0x6D, 0xE7}, DataTypeUInt, false},
	ElementTag:                    elementDef{[]byte{0x73, 0x73}, DataTypeMaster, false},
	ElementSegmentFilename:        elementDef{[]byte{0x73, 0x84}, DataTypeString, false},
	ElementSegmentUID:             elementDef{[]byte{0x73, 0xA4}, DataTypeBinary, false},
	ElementTrackUID:               elementDef{[]byte{0x73, 0xC5}, DataTypeUInt, false},
	ElementTitle:                  elementDef{[]byte{0x7B, 0xA9}, DataTypeString, false},
	ElementTrackType:              elementDef{[]byte{0x83}, DataTypeUInt, false},
	ElementCodecID:                elementDef{[]byte{0x86}, DataTypeString, false},
	ElementFlagDefault:            elementDef{[]byte{0x88}, DataTypeUInt, false},
	ElementFlagInterlaced:         elementDef{[]byte{0x9A}, DataTypeUInt, false},
	ElementBlockDuration:          elementDef{[]byte{0x9B}, DataTypeUInt, false},
	ElementFlagLacing:             elementDef{[]byte{0x9C}, DataTypeUInt, false},
	ElementChannels:               elementDef{[]byte{0x9F}, DataTypeUInt, false},
	ElementBlockGroup:             elementDef{[]byte{0xA0}, DataTypeMaster, false},
	ElementBlock:                  elementDef{[]byte{0xA1}, DataTypeBlock, false},
	ElementSimpleBlock:            elementDef{[]byte{0xA3}, DataTypeBlock, false},
	ElementPosition:               elementDef{[]byte{0xA7}, DataTypeUInt, false},
	ElementCodecDecodeAll:         elementDef{[]byte{0xAA}, DataTypeUInt, false},
	ElementPrevSize:               elementDef{[]byte{0xAB}, DataTypeUInt, false},
	ElementTrackEntry:             elementDef{[]byte{0xAE}, DataTypeMaster, false},
	ElementPixelWidth:             elementDef{[]byte{0xB0}, DataTypeUInt, false},
	ElementCueDuration:            elementDef{[]byte{0xB2}, DataTypeUInt, false},
	ElementCueTime:                elementDef{[]byte{0xB3}, DataTypeUInt, false},
	ElementSamplingFrequency:      elementDef{[]byte{0xB5}, DataTypeFloat, false},
	ElementCueTrackPositions:      elementDef{[]byte{0xB7}, DataTypeMaster, false},
	ElementFlagEnabled:            elementDef{[]byte{0xB9}, DataTypeUInt, false},
	ElementPixelHeight:            elementDef{[]byte{0xBA}, DataTypeUInt, false},
	ElementCuePoint:               elementDef{[]byte{0xBB}, DataTypeMaster, false},
	ElementCRC32:                  elementDef{[]byte{0xBF}, DataTypeBinary, false},
	ElementTrackNumber:            elementDef{[]byte{0xD7}, DataTypeUInt, false},
	ElementVideo:                  elementDef{[]byte{0xE0}, DataTypeMaster, false},
	ElementAudio:                  elementDef{[]byte{0xE1}, DataTypeMaster, false},
	ElementTimecode:               elementDef{[]byte{0xE7}, DataTypeUInt, false},
	ElementVoid:                   elementDef{[]byte{0xEC}, DataTypeBinary, false},
	ElementCueRelativePosition:    elementDef{[]byte{0xF0}, DataTypeUInt, false},
	ElementCueClusterPosition:     elementDef{[]byte{0xF1}, DataTypeUInt, false},
	ElementCueTrack:               elementDef{[]byte{0xF7}, DataTypeUInt, false},
	ElementReferenceBlock:         elementDef{[]byte{0xFB}, DataTypeInt, false},
}

type elementRevTable map[uint32]element
type element struct {
	e   ElementType
	t   DataType
	top bool
}

var revTable elementRevTable

func init() {
	revTable = make(elementRevTable)
	initReverseLookupTable(revTable, table)
}

func initReverseLookupTable(revTb elementRevTable, tb elementTable) {
	for k, v := range tb {
		e, _, err := readVInt(bytes.NewBuffer(v.b))
		if err != nil {
			panic(err)
		}
		revTb[uint32(e)] = element{e: k, t: v.t, top: v.top}
	}
}

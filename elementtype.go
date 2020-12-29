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
)

// ErrUnknownElementName means that a element name is not found in the ElementType list.
var ErrUnknownElementName = errors.New("unknown element name")

// ElementType represents EBML Element type.
type ElementType int

// EBML Element types.
const (
	ElementInvalid ElementType = iota

	ElementEBML
	ElementEBMLVersion
	ElementEBMLReadVersion
	ElementEBMLMaxIDLength
	ElementEBMLMaxSizeLength
	ElementEBMLDocType
	ElementEBMLDocTypeVersion
	ElementEBMLDocTypeReadVersion

	ElementVoid
	ElementSegment

	ElementSeekHead
	ElementSeek
	ElementSeekID
	ElementSeekPosition

	ElementInfo
	ElementSegmentUID
	ElementSegmentFilename
	ElementTimestampScale
	ElementDuration
	ElementDateUTC
	ElementTitle
	ElementMuxingApp
	ElementWritingApp

	ElementCluster
	ElementTimestamp
	ElementPosition
	ElementPrevSize
	ElementSimpleBlock
	ElementBlockGroup
	ElementBlock
	ElementBlockAdditions
	ElementBlockMore
	ElementBlockAddID
	ElementBlockAdditional
	ElementBlockDuration
	ElementReferenceBlock
	ElementDiscardPadding

	ElementTracks
	ElementTrackEntry
	ElementTrackNumber
	ElementTrackUID
	ElementTrackType
	ElementFlagEnabled
	ElementFlagDefault
	ElementFlagForced
	ElementFlagLacing
	ElementMinCache
	ElementDefaultDuration
	ElementMaxBlockAdditionID
	ElementName
	ElementLanguage
	ElementCodecID
	ElementCodecPrivate
	ElementCodecName
	ElementCodecDecodeAll
	ElementCodecDelay
	ElementSeekPreRoll
	ElementVideo
	ElementFlagInterlaced
	ElementStereoMode
	ElementAlphaMode
	ElementPixelWidth
	ElementPixelHeight
	ElementPixelCropBottom
	ElementPixelCropTop
	ElementPixelCropLeft
	ElementPixelCropRight
	ElementDisplayWidth
	ElementDisplayHeight
	ElementDisplayUnit
	ElementAspectRatioType
	ElementAudio
	ElementSamplingFrequency
	ElementOutputSamplingFrequency
	ElementChannels
	ElementBitDepth
	ElementContentEncodings
	ElementContentEncoding
	ElementContentEncodingOrder
	ElementContentEncodingScope
	ElementContentEncodingType
	ElementContentEncryption
	ElementContentEncAlgo
	ElementContentEncKeyID
	ElementContentEncAESSettings
	ElementAESSettingsCipherMode

	ElementCues
	ElementCuePoint
	ElementCueTime
	ElementCueTrackPositions
	ElementCueTrack
	ElementCueClusterPosition
	ElementCueRelativePosition
	ElementCueDuration
	ElementCueBlockNumber

	ElementTags
	ElementTag
	ElementSimpleTag
	ElementTagName
	ElementTagString
	ElementTagBinary

	elementMax
)

// WebM aliases
const (
	ElementTimecodeScale = ElementTimestampScale
	ElementTimecode      = ElementTimestamp
)

var elementTypeName = map[ElementType]string{
	ElementEBML:                    "EBML",
	ElementEBMLVersion:             "EBMLVersion",
	ElementEBMLReadVersion:         "EBMLReadVersion",
	ElementEBMLMaxIDLength:         "EBMLMaxIDLength",
	ElementEBMLMaxSizeLength:       "EBMLMaxSizeLength",
	ElementEBMLDocType:             "EBMLDocType",
	ElementEBMLDocTypeVersion:      "EBMLDocTypeVersion",
	ElementEBMLDocTypeReadVersion:  "EBMLDocTypeReadVersion",
	ElementVoid:                    "Void",
	ElementSegment:                 "Segment",
	ElementSeekHead:                "SeekHead",
	ElementSeek:                    "Seek",
	ElementSeekID:                  "SeekID",
	ElementSeekPosition:            "SeekPosition",
	ElementInfo:                    "Info",
	ElementSegmentUID:              "SegmentUID",
	ElementSegmentFilename:         "SegmentFilename",
	ElementTimestampScale:          "TimestampScale",
	ElementDuration:                "Duration",
	ElementDateUTC:                 "DateUTC",
	ElementTitle:                   "Title",
	ElementMuxingApp:               "MuxingApp",
	ElementWritingApp:              "WritingApp",
	ElementCluster:                 "Cluster",
	ElementTimestamp:               "Timestamp",
	ElementPosition:                "Position",
	ElementPrevSize:                "PrevSize",
	ElementSimpleBlock:             "SimpleBlock",
	ElementBlockGroup:              "BlockGroup",
	ElementBlock:                   "Block",
	ElementBlockAdditions:          "BlockAdditions",
	ElementBlockMore:               "BlockMore",
	ElementBlockAddID:              "BlockAddID",
	ElementBlockAdditional:         "BlockAdditional",
	ElementBlockDuration:           "BlockDuration",
	ElementReferenceBlock:          "ReferenceBlock",
	ElementDiscardPadding:          "DiscardPadding",
	ElementTracks:                  "Tracks",
	ElementTrackEntry:              "TrackEntry",
	ElementTrackNumber:             "TrackNumber",
	ElementTrackUID:                "TrackUID",
	ElementTrackType:               "TrackType",
	ElementFlagEnabled:             "FlagEnabled",
	ElementFlagDefault:             "FlagDefault",
	ElementFlagForced:              "FlagForced",
	ElementFlagLacing:              "FlagLacing",
	ElementMinCache:                "MinCache",
	ElementDefaultDuration:         "DefaultDuration",
	ElementMaxBlockAdditionID:      "MaxBlockAdditionID",
	ElementName:                    "Name",
	ElementLanguage:                "Language",
	ElementCodecID:                 "CodecID",
	ElementCodecPrivate:            "CodecPrivate",
	ElementCodecName:               "CodecName",
	ElementCodecDecodeAll:          "CodecDecodeAll",
	ElementCodecDelay:              "CodecDelay",
	ElementSeekPreRoll:             "SeekPreRoll",
	ElementVideo:                   "Video",
	ElementFlagInterlaced:          "FlagInterlaced",
	ElementStereoMode:              "StereoMode",
	ElementAlphaMode:               "AlphaMode",
	ElementPixelWidth:              "PixelWidth",
	ElementPixelHeight:             "PixelHeight",
	ElementPixelCropBottom:         "PixelCropBottom",
	ElementPixelCropTop:            "PixelCropTop",
	ElementPixelCropLeft:           "PixelCropLeft",
	ElementPixelCropRight:          "PixelCropRight",
	ElementDisplayWidth:            "DisplayWidth",
	ElementDisplayHeight:           "DisplayHeight",
	ElementDisplayUnit:             "DisplayUnit",
	ElementAspectRatioType:         "AspectRatioType",
	ElementAudio:                   "Audio",
	ElementSamplingFrequency:       "SamplingFrequency",
	ElementOutputSamplingFrequency: "OutputSamplingFrequency",
	ElementChannels:                "Channels",
	ElementBitDepth:                "BitDepth",
	ElementContentEncodings:        "ContentEncodings",
	ElementContentEncoding:         "ContentEncoding",
	ElementContentEncodingOrder:    "ContentEncodingOrder",
	ElementContentEncodingScope:    "ContentEncodingScope",
	ElementContentEncodingType:     "ContentEncodingType",
	ElementContentEncryption:       "ContentEncryption",
	ElementContentEncAlgo:          "ContentEncAlgo",
	ElementContentEncKeyID:         "ContentEncKeyID",
	ElementContentEncAESSettings:   "ContentEncAESSettings",
	ElementAESSettingsCipherMode:   "AESSettingsCipherMode",
	ElementCues:                    "Cues",
	ElementCuePoint:                "CuePoint",
	ElementCueTime:                 "CueTime",
	ElementCueTrackPositions:       "CueTrackPositions",
	ElementCueTrack:                "CueTrack",
	ElementCueClusterPosition:      "CueClusterPosition",
	ElementCueRelativePosition:     "CueRelativePosition",
	ElementCueDuration:             "CueDuration",
	ElementCueBlockNumber:          "CueBlockNumber",
	ElementTags:                    "Tags",
	ElementTag:                     "Tag",
	ElementSimpleTag:               "SimpleTag",
	ElementTagName:                 "TagName",
	ElementTagString:               "TagString",
	ElementTagBinary:               "TagBinary",
}

func (i ElementType) String() string {
	if name, ok := elementTypeName[i]; ok {
		return name
	}
	return "unknown"
}

// Bytes returns []byte representation of the element ID.
func (i ElementType) Bytes() []byte {
	return table[i].b
}

// DataType returns DataType of the element.
func (i ElementType) DataType() DataType {
	return table[i].t
}

var elementNameType map[string]ElementType

// ElementTypeFromString converts string to ElementType.
func ElementTypeFromString(s string) (ElementType, error) {
	if t, ok := elementNameType[s]; ok {
		return t, nil
	}
	return 0, wrapErrorf(ErrUnknownElementName, "parsing \"%s\"", s)
}

func init() {
	elementNameType = make(map[string]ElementType)
	for t, name := range elementTypeName {
		elementNameType[name] = t
	}
	// WebM aliases
	elementNameType["TimecodeScale"] = ElementTimecodeScale
	elementNameType["Timecode"] = ElementTimecode
}

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

	ElementCRC32
	ElementVoid
	ElementSegment

	ElementSeekHead
	ElementSeek
	ElementSeekID
	ElementSeekPosition

	ElementInfo
	ElementSegmentUID
	ElementSegmentFilename
	ElementPrevUID
	ElementPrevFilename
	ElementNextUID
	ElementNextFilename
	ElementSegmentFamily
	ElementChapterTranslate
	ElementChapterTranslateEditionUID
	ElementChapterTranslateCodec
	ElementChapterTranslateID
	ElementTimestampScale
	ElementDuration
	ElementDateUTC
	ElementTitle
	ElementMuxingApp
	ElementWritingApp

	ElementCluster
	ElementTimestamp
	ElementSilentTracks
	ElementSilentTrackNumber
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
	ElementReferencePriority
	ElementReferenceBlock
	ElementCodecState
	ElementDiscardPadding
	ElementSlices
	// Deprecated: Dropped in v2
	ElementTimeSlice
	// Deprecated: Dropped in v2
	ElementLaceNumber

	ElementTracks
	ElementTrackEntry
	ElementTrackNumber
	ElementTrackUID
	ElementTrackType
	ElementFlagEnabled
	ElementFlagDefault
	ElementFlagForced
	ElementFlagHearingImpaired
	ElementFlagVisualImpaired
	ElementFlagTextDescriptions
	ElementFlagOriginal
	ElementFlagCommentary
	ElementFlagLacing
	ElementMinCache
	ElementMaxCache
	ElementDefaultDuration
	ElementDefaultDecodedFieldDuration
	// Deprecated: Dropped in v4
	ElementTrackTimestampScale
	ElementMaxBlockAdditionID
	ElementBlockAdditionMapping
	ElementBlockAddIDValue
	ElementBlockAddIDName
	ElementBlockAddIDType
	ElementBlockAddIDExtraData
	ElementName
	ElementLanguage
	ElementLanguageIETF
	ElementCodecID
	ElementCodecPrivate
	ElementCodecName
	// Deprecated: Dropped in v4
	ElementAttachmentLink
	ElementCodecDecodeAll
	ElementTrackOverlay
	ElementCodecDelay
	ElementSeekPreRoll
	ElementTrackTranslate
	ElementTrackTranslateEditionUID
	ElementTrackTranslateCodec
	ElementTrackTranslateTrackID
	ElementVideo
	ElementFlagInterlaced
	ElementFieldOrder
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
	ElementColourSpace
	ElementColour
	ElementMatrixCoefficients
	ElementBitsPerChannel
	ElementChromaSubsamplingHorz
	ElementChromaSubsamplingVert
	ElementCbSubsamplingHorz
	ElementCbSubsamplingVert
	ElementChromaSitingHorz
	ElementChromaSitingVert
	ElementRange
	ElementTransferCharacteristics
	ElementPrimaries
	ElementMaxCLL
	ElementMaxFALL
	ElementMasteringMetadata
	ElementPrimaryRChromaticityX
	ElementPrimaryRChromaticityY
	ElementPrimaryGChromaticityX
	ElementPrimaryGChromaticityY
	ElementPrimaryBChromaticityX
	ElementPrimaryBChromaticityY
	ElementWhitePointChromaticityX
	ElementWhitePointChromaticityY
	ElementLuminanceMax
	ElementLuminanceMin
	ElementProjection
	ElementProjectionType
	ElementProjectionPrivate
	ElementProjectionPoseYaw
	ElementProjectionPosePitch
	ElementProjectionPoseRoll
	ElementAudio
	ElementSamplingFrequency
	ElementOutputSamplingFrequency
	ElementChannels
	ElementBitDepth
	ElementTrackOperation
	ElementTrackCombinePlanes
	ElementTrackPlane
	ElementTrackPlaneUID
	ElementTrackPlaneType
	ElementTrackJoinBlocks
	ElementTrackJoinUID
	ElementContentEncodings
	ElementContentEncoding
	ElementContentEncodingOrder
	ElementContentEncodingScope
	ElementContentEncodingType
	ElementContentCompression
	ElementContentCompAlgo
	ElementContentCompSettings
	ElementContentEncryption
	ElementContentEncAlgo
	ElementContentEncKeyID
	ElementContentEncAESSettings
	ElementAESSettingsCipherMode
	ElementContentSignature
	ElementContentSigKeyID
	ElementContentSigAlgo
	ElementContentSigHashAlgo

	ElementCues
	ElementCuePoint
	ElementCueTime
	ElementCueTrackPositions
	ElementCueTrack
	ElementCueClusterPosition
	ElementCueRelativePosition
	ElementCueDuration
	ElementCueBlockNumber
	ElementCueCodecState
	ElementCueReference
	ElementCueRefTime

	ElementAttachments
	ElementAttachedFile
	ElementFileDescription
	ElementFileName
	ElementFileMimeType
	ElementFileData
	ElementFileUID

	ElementChapters
	ElementEditionEntry
	ElementEditionUID
	ElementEditionFlagHidden
	ElementEditionFlagDefault
	ElementEditionFlagOrdered
	ElementChapterAtom
	ElementChapterUID
	ElementChapterStringUID
	ElementChapterTimeStart
	ElementChapterTimeEnd
	ElementChapterFlagHidden
	ElementChapterFlagEnabled
	ElementChapterSegmentUID
	ElementChapterSegmentEditionUID
	ElementChapterPhysicalEquiv
	ElementChapterTrack
	ElementChapterTrackUID
	ElementChapterDisplay
	ElementChapString
	ElementChapLanguage
	ElementChapLanguageIETF
	ElementChapCountry
	ElementChapProcess
	ElementChapProcessCodecID
	ElementChapProcessPrivate
	ElementChapProcessCommand
	ElementChapProcessTime
	ElementChapProcessData

	ElementTags
	ElementTag
	ElementTargets
	ElementTargetTypeValue
	ElementTargetType
	ElementTagTrackUID
	ElementTagEditionUID
	ElementTagChapterUID
	ElementTagAttachmentUID
	ElementSimpleTag
	ElementTagName
	ElementTagLanguage
	ElementTagLanguageIETF
	ElementTagDefault
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
	ElementEBML:                        "EBML",
	ElementEBMLVersion:                 "EBMLVersion",
	ElementEBMLReadVersion:             "EBMLReadVersion",
	ElementEBMLMaxIDLength:             "EBMLMaxIDLength",
	ElementEBMLMaxSizeLength:           "EBMLMaxSizeLength",
	ElementEBMLDocType:                 "EBMLDocType",
	ElementEBMLDocTypeVersion:          "EBMLDocTypeVersion",
	ElementEBMLDocTypeReadVersion:      "EBMLDocTypeReadVersion",
	ElementCRC32:                       "CRC32",
	ElementVoid:                        "Void",
	ElementSegment:                     "Segment",
	ElementSeekHead:                    "SeekHead",
	ElementSeek:                        "Seek",
	ElementSeekID:                      "SeekID",
	ElementSeekPosition:                "SeekPosition",
	ElementInfo:                        "Info",
	ElementSegmentUID:                  "SegmentUID",
	ElementSegmentFilename:             "SegmentFilename",
	ElementPrevUID:                     "PrevUID",
	ElementPrevFilename:                "PrevFilename",
	ElementNextUID:                     "NextUID",
	ElementNextFilename:                "NextFilename",
	ElementSegmentFamily:               "SegmentFamily",
	ElementChapterTranslate:            "ChapterTranslate",
	ElementChapterTranslateEditionUID:  "ChapterTranslateEditionUID",
	ElementChapterTranslateCodec:       "ChapterTranslateCodec",
	ElementChapterTranslateID:          "ChapterTranslateID",
	ElementTimestampScale:              "TimestampScale",
	ElementDuration:                    "Duration",
	ElementDateUTC:                     "DateUTC",
	ElementTitle:                       "Title",
	ElementMuxingApp:                   "MuxingApp",
	ElementWritingApp:                  "WritingApp",
	ElementCluster:                     "Cluster",
	ElementTimestamp:                   "Timestamp",
	ElementSilentTracks:                "SilentTracks",
	ElementSilentTrackNumber:           "SilentTrackNumber",
	ElementPosition:                    "Position",
	ElementPrevSize:                    "PrevSize",
	ElementSimpleBlock:                 "SimpleBlock",
	ElementBlockGroup:                  "BlockGroup",
	ElementBlock:                       "Block",
	ElementBlockAdditions:              "BlockAdditions",
	ElementBlockMore:                   "BlockMore",
	ElementBlockAddID:                  "BlockAddID",
	ElementBlockAdditional:             "BlockAdditional",
	ElementBlockDuration:               "BlockDuration",
	ElementReferencePriority:           "ReferencePriority",
	ElementReferenceBlock:              "ReferenceBlock",
	ElementCodecState:                  "CodecState",
	ElementDiscardPadding:              "DiscardPadding",
	ElementSlices:                      "Slices",
	ElementTimeSlice:                   "TimeSlice",
	ElementLaceNumber:                  "LaceNumber",
	ElementTracks:                      "Tracks",
	ElementTrackEntry:                  "TrackEntry",
	ElementTrackNumber:                 "TrackNumber",
	ElementTrackUID:                    "TrackUID",
	ElementTrackType:                   "TrackType",
	ElementFlagEnabled:                 "FlagEnabled",
	ElementFlagDefault:                 "FlagDefault",
	ElementFlagForced:                  "FlagForced",
	ElementFlagHearingImpaired:         "FlagHearingImpaired",
	ElementFlagVisualImpaired:          "FlagVisualImpaired",
	ElementFlagTextDescriptions:        "FlagTextDescriptions",
	ElementFlagOriginal:                "FlagOriginal",
	ElementFlagCommentary:              "FlagCommentary",
	ElementFlagLacing:                  "FlagLacing",
	ElementMinCache:                    "MinCache",
	ElementMaxCache:                    "MaxCache",
	ElementDefaultDuration:             "DefaultDuration",
	ElementDefaultDecodedFieldDuration: "DefaultDecodedFieldDuration",
	ElementTrackTimestampScale:         "TrackTimestampScale",
	ElementMaxBlockAdditionID:          "MaxBlockAdditionID",
	ElementBlockAdditionMapping:        "BlockAdditionMapping",
	ElementBlockAddIDValue:             "BlockAddIDValue",
	ElementBlockAddIDName:              "BlockAddIDName",
	ElementBlockAddIDType:              "BlockAddIDType",
	ElementBlockAddIDExtraData:         "BlockAddIDExtraData",
	ElementName:                        "Name",
	ElementLanguage:                    "Language",
	ElementLanguageIETF:                "LanguageIETF",
	ElementCodecID:                     "CodecID",
	ElementCodecPrivate:                "CodecPrivate",
	ElementCodecName:                   "CodecName",
	ElementAttachmentLink:              "AttachmentLink",
	ElementCodecDecodeAll:              "CodecDecodeAll",
	ElementTrackOverlay:                "TrackOverlay",
	ElementCodecDelay:                  "CodecDelay",
	ElementSeekPreRoll:                 "SeekPreRoll",
	ElementTrackTranslate:              "TrackTranslate",
	ElementTrackTranslateEditionUID:    "TrackTranslateEditionUID",
	ElementTrackTranslateCodec:         "TrackTranslateCodec",
	ElementTrackTranslateTrackID:       "TrackTranslateTrackID",
	ElementVideo:                       "Video",
	ElementFlagInterlaced:              "FlagInterlaced",
	ElementFieldOrder:                  "FieldOrder",
	ElementStereoMode:                  "StereoMode",
	ElementAlphaMode:                   "AlphaMode",
	ElementPixelWidth:                  "PixelWidth",
	ElementPixelHeight:                 "PixelHeight",
	ElementPixelCropBottom:             "PixelCropBottom",
	ElementPixelCropTop:                "PixelCropTop",
	ElementPixelCropLeft:               "PixelCropLeft",
	ElementPixelCropRight:              "PixelCropRight",
	ElementDisplayWidth:                "DisplayWidth",
	ElementDisplayHeight:               "DisplayHeight",
	ElementDisplayUnit:                 "DisplayUnit",
	ElementAspectRatioType:             "AspectRatioType",
	ElementColourSpace:                 "ColourSpace",
	ElementColour:                      "Colour",
	ElementMatrixCoefficients:          "MatrixCoefficients",
	ElementBitsPerChannel:              "BitsPerChannel",
	ElementChromaSubsamplingHorz:       "ChromaSubsamplingHorz",
	ElementChromaSubsamplingVert:       "ChromaSubsamplingVert",
	ElementCbSubsamplingHorz:           "CbSubsamplingHorz",
	ElementCbSubsamplingVert:           "CbSubsamplingVert",
	ElementChromaSitingHorz:            "ChromaSitingHorz",
	ElementChromaSitingVert:            "ChromaSitingVert",
	ElementRange:                       "Range",
	ElementTransferCharacteristics:     "TransferCharacteristics",
	ElementPrimaries:                   "Primaries",
	ElementMaxCLL:                      "MaxCLL",
	ElementMaxFALL:                     "MaxFALL",
	ElementMasteringMetadata:           "MasteringMetadata",
	ElementPrimaryRChromaticityX:       "PrimaryRChromaticityX",
	ElementPrimaryRChromaticityY:       "PrimaryRChromaticityY",
	ElementPrimaryGChromaticityX:       "PrimaryGChromaticityX",
	ElementPrimaryGChromaticityY:       "PrimaryGChromaticityY",
	ElementPrimaryBChromaticityX:       "PrimaryBChromaticityX",
	ElementPrimaryBChromaticityY:       "PrimaryBChromaticityY",
	ElementWhitePointChromaticityX:     "WhitePointChromaticityX",
	ElementWhitePointChromaticityY:     "WhitePointChromaticityY",
	ElementLuminanceMax:                "LuminanceMax",
	ElementLuminanceMin:                "LuminanceMin",
	ElementProjection:                  "Projection",
	ElementProjectionType:              "ProjectionType",
	ElementProjectionPrivate:           "ProjectionPrivate",
	ElementProjectionPoseYaw:           "ProjectionPoseYaw",
	ElementProjectionPosePitch:         "ProjectionPosePitch",
	ElementProjectionPoseRoll:          "ProjectionPoseRoll",
	ElementAudio:                       "Audio",
	ElementSamplingFrequency:           "SamplingFrequency",
	ElementOutputSamplingFrequency:     "OutputSamplingFrequency",
	ElementChannels:                    "Channels",
	ElementBitDepth:                    "BitDepth",
	ElementTrackOperation:              "TrackOperation",
	ElementTrackCombinePlanes:          "TrackCombinePlanes",
	ElementTrackPlane:                  "TrackPlane",
	ElementTrackPlaneUID:               "TrackPlaneUID",
	ElementTrackPlaneType:              "TrackPlaneType",
	ElementTrackJoinBlocks:             "TrackJoinBlocks",
	ElementTrackJoinUID:                "TrackJoinUID",
	ElementContentEncodings:            "ContentEncodings",
	ElementContentEncoding:             "ContentEncoding",
	ElementContentEncodingOrder:        "ContentEncodingOrder",
	ElementContentEncodingScope:        "ContentEncodingScope",
	ElementContentEncodingType:         "ContentEncodingType",
	ElementContentCompression:          "ContentCompression",
	ElementContentCompAlgo:             "ContentCompAlgo",
	ElementContentCompSettings:         "ContentCompSettings",
	ElementContentEncryption:           "ContentEncryption",
	ElementContentEncAlgo:              "ContentEncAlgo",
	ElementContentEncKeyID:             "ContentEncKeyID",
	ElementContentEncAESSettings:       "ContentEncAESSettings",
	ElementAESSettingsCipherMode:       "AESSettingsCipherMode",
	ElementContentSignature:            "ContentSignature",
	ElementContentSigKeyID:             "ContentSigKeyID",
	ElementContentSigAlgo:              "ContentSigAlgo",
	ElementContentSigHashAlgo:          "ContentSigHashAlgo",
	ElementCues:                        "Cues",
	ElementCuePoint:                    "CuePoint",
	ElementCueTime:                     "CueTime",
	ElementCueTrackPositions:           "CueTrackPositions",
	ElementCueTrack:                    "CueTrack",
	ElementCueClusterPosition:          "CueClusterPosition",
	ElementCueRelativePosition:         "CueRelativePosition",
	ElementCueDuration:                 "CueDuration",
	ElementCueBlockNumber:              "CueBlockNumber",
	ElementCueCodecState:               "CueCodecState",
	ElementCueReference:                "CueReference",
	ElementCueRefTime:                  "CueRefTime",
	ElementAttachments:                 "Attachments",
	ElementAttachedFile:                "AttachedFile",
	ElementFileDescription:             "FileDescription",
	ElementFileName:                    "FileName",
	ElementFileMimeType:                "FileMimeType",
	ElementFileData:                    "FileData",
	ElementFileUID:                     "FileUID",
	ElementChapters:                    "Chapters",
	ElementEditionEntry:                "EditionEntry",
	ElementEditionUID:                  "EditionUID",
	ElementEditionFlagHidden:           "EditionFlagHidden",
	ElementEditionFlagDefault:          "EditionFlagDefault",
	ElementEditionFlagOrdered:          "EditionFlagOrdered",
	ElementChapterAtom:                 "ChapterAtom",
	ElementChapterUID:                  "ChapterUID",
	ElementChapterStringUID:            "ChapterStringUID",
	ElementChapterTimeStart:            "ChapterTimeStart",
	ElementChapterTimeEnd:              "ChapterTimeEnd",
	ElementChapterFlagHidden:           "ChapterFlagHidden",
	ElementChapterFlagEnabled:          "ChapterFlagEnabled",
	ElementChapterSegmentUID:           "ChapterSegmentUID",
	ElementChapterSegmentEditionUID:    "ChapterSegmentEditionUID",
	ElementChapterPhysicalEquiv:        "ChapterPhysicalEquiv",
	ElementChapterTrack:                "ChapterTrack",
	ElementChapterTrackUID:             "ChapterTrackUID",
	ElementChapterDisplay:              "ChapterDisplay",
	ElementChapString:                  "ChapString",
	ElementChapLanguage:                "ChapLanguage",
	ElementChapLanguageIETF:            "ChapLanguageIETF",
	ElementChapCountry:                 "ChapCountry",
	ElementChapProcess:                 "ChapProcess",
	ElementChapProcessCodecID:          "ChapProcessCodecID",
	ElementChapProcessPrivate:          "ChapProcessPrivate",
	ElementChapProcessCommand:          "ChapProcessCommand",
	ElementChapProcessTime:             "ChapProcessTime",
	ElementChapProcessData:             "ChapProcessData",
	ElementTags:                        "Tags",
	ElementTag:                         "Tag",
	ElementTargets:                     "Targets",
	ElementTargetTypeValue:             "TargetTypeValue",
	ElementTargetType:                  "TargetType",
	ElementTagTrackUID:                 "TagTrackUID",
	ElementTagEditionUID:               "TagEditionUID",
	ElementTagChapterUID:               "TagChapterUID",
	ElementTagAttachmentUID:            "TagAttachmentUID",
	ElementSimpleTag:                   "SimpleTag",
	ElementTagName:                     "TagName",
	ElementTagLanguage:                 "TagLanguage",
	ElementTagLanguageIETF:             "TagLanguageIETF",
	ElementTagDefault:                  "TagDefault",
	ElementTagString:                   "TagString",
	ElementTagBinary:                   "TagBinary",
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

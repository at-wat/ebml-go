package ebml

// ElementType represents EBML Element type
type ElementType int

// EBML Element types
const (
	ElementRoot ElementType = iota
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
	ElementTimecodeScale
	ElementDuration
	ElementDateUTC
	ElementTitle
	ElementMuxingApp
	ElementWritingApp

	ElementCluster
	ElementTimecode
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
	ElementFlagForced
	ElementFlagLacing
	ElementDefaultDuration
	ElementName
	ElementLanguage
	ElementCodecID
	ElementCodecPrivate
	ElementCodecName
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
)

func (i ElementType) String() string {
	switch i {
	case ElementEBML:
		return "EBML"
	case ElementEBMLVersion:
		return "EBMLVersion"
	case ElementEBMLReadVersion:
		return "EBMLReadVersion"
	case ElementEBMLMaxIDLength:
		return "EBMLMaxIDLength"
	case ElementEBMLMaxSizeLength:
		return "EBMLMaxSizeLength"
	case ElementEBMLDocType:
		return "EBMLDocType"
	case ElementEBMLDocTypeVersion:
		return "EBMLDocTypeVersion"
	case ElementEBMLDocTypeReadVersion:
		return "EBMLDocTypeReadVersion"
	case ElementVoid:
		return "Void"
	case ElementSegment:
		return "Segment"
	case ElementSeekHead:
		return "SeekHead"
	case ElementSeek:
		return "Seek"
	case ElementSeekID:
		return "SeekID"
	case ElementSeekPosition:
		return "SeekPosition"
	case ElementInfo:
		return "Info"
	case ElementTimecodeScale:
		return "TimecodeScale"
	case ElementDuration:
		return "Duration"
	case ElementDateUTC:
		return "DateUTC"
	case ElementTitle:
		return "Title"
	case ElementMuxingApp:
		return "MuxingApp"
	case ElementWritingApp:
		return "WritingApp"
	case ElementCluster:
		return "Cluster"
	case ElementTimecode:
		return "Timecode"
	case ElementPrevSize:
		return "PrevSize"
	case ElementSimpleBlock:
		return "SimpleBlock"
	case ElementBlockGroup:
		return "BlockGroup"
	case ElementBlock:
		return "Block"
	case ElementBlockAdditions:
		return "BlockAdditions"
	case ElementBlockMore:
		return "BlockMore"
	case ElementBlockAddID:
		return "BlockAddID"
	case ElementBlockAdditional:
		return "BlockAdditional"
	case ElementBlockDuration:
		return "BlockDuration"
	case ElementReferenceBlock:
		return "ReferenceBlock"
	case ElementDiscardPadding:
		return "DiscardPadding"
	case ElementTracks:
		return "Tracks"
	case ElementTrackEntry:
		return "TrackEntry"
	case ElementTrackNumber:
		return "TrackNumber"
	case ElementTrackUID:
		return "TrackUID"
	case ElementTrackType:
		return "TrackType"
	case ElementFlagEnabled:
		return "FlagEnabled"
	case ElementFlagForced:
		return "FlagForced"
	case ElementFlagLacing:
		return "FlagLacing"
	case ElementDefaultDuration:
		return "DefaultDuration"
	case ElementName:
		return "Name"
	case ElementLanguage:
		return "Language"
	case ElementCodecID:
		return "CodecID"
	case ElementCodecPrivate:
		return "CodecPrivate"
	case ElementCodecName:
		return "CodecName"
	case ElementCodecDelay:
		return "CodecDelay"
	case ElementSeekPreRoll:
		return "SeekPreRoll"
	case ElementVideo:
		return "Video"
	case ElementFlagInterlaced:
		return "FlagInterlaced"
	case ElementStereoMode:
		return "StereoMode"
	case ElementAlphaMode:
		return "AlphaMode"
	case ElementPixelWidth:
		return "PixelWidth"
	case ElementPixelHeight:
		return "PixelHeight"
	case ElementPixelCropBottom:
		return "PixelCropBottom"
	case ElementPixelCropTop:
		return "PixelCropTop"
	case ElementPixelCropLeft:
		return "PixelCropLeft"
	case ElementPixelCropRight:
		return "PixelCropRight"
	case ElementDisplayWidth:
		return "DisplayWidth"
	case ElementDisplayHeight:
		return "DisplayHeight"
	case ElementDisplayUnit:
		return "DisplayUnit"
	case ElementAspectRatioType:
		return "AspectRatioType"
	case ElementAudio:
		return "Audio"
	case ElementSamplingFrequency:
		return "SamplingFrequency"
	case ElementOutputSamplingFrequency:
		return "OutputSamplingFrequency"
	case ElementChannels:
		return "Channels"
	case ElementBitDepth:
		return "BitDepth"
	case ElementContentEncodings:
		return "ContentEncodings"
	case ElementContentEncoding:
		return "ContentEncoding"
	case ElementContentEncodingOrder:
		return "ContentEncodingOrder"
	case ElementContentEncodingScope:
		return "ContentEncodingScope"
	case ElementContentEncodingType:
		return "ContentEncodingType"
	case ElementContentEncryption:
		return "ContentEncryption"
	case ElementContentEncAlgo:
		return "ContentEncAlgo"
	case ElementContentEncKeyID:
		return "ContentEncKeyID"
	case ElementContentEncAESSettings:
		return "ContentEncAESSettings"
	case ElementAESSettingsCipherMode:
		return "AESSettingsCipherMode"
	default:
		return "unknown"
	}
}

// ElementTypeFromString converts string to ElementType
func ElementTypeFromString(s string) (ElementType, error) {
	switch s {
	case "EBML":
		return ElementEBML, nil
	case "EBMLVersion":
		return ElementEBMLVersion, nil
	case "EBMLReadVersion":
		return ElementEBMLReadVersion, nil
	case "EBMLMaxIDLength":
		return ElementEBMLMaxIDLength, nil
	case "EBMLMaxSizeLength":
		return ElementEBMLMaxSizeLength, nil
	case "EBMLDocType":
		return ElementEBMLDocType, nil
	case "EBMLDocTypeVersion":
		return ElementEBMLDocTypeVersion, nil
	case "EBMLDocTypeReadVersion":
		return ElementEBMLDocTypeReadVersion, nil
	case "Void":
		return ElementVoid, nil
	case "Segment":
		return ElementSegment, nil
	case "SeekHead":
		return ElementSeekHead, nil
	case "Seek":
		return ElementSeek, nil
	case "SeekID":
		return ElementSeekID, nil
	case "SeekPosition":
		return ElementSeekPosition, nil
	case "Info":
		return ElementInfo, nil
	case "TimecodeScale":
		return ElementTimecodeScale, nil
	case "Duration":
		return ElementDuration, nil
	case "DateUTC":
		return ElementDateUTC, nil
	case "Title":
		return ElementTitle, nil
	case "MuxingApp":
		return ElementMuxingApp, nil
	case "WritingApp":
		return ElementWritingApp, nil
	case "Cluster":
		return ElementCluster, nil
	case "Timecode":
		return ElementTimecode, nil
	case "PrevSize":
		return ElementPrevSize, nil
	case "SimpleBlock":
		return ElementSimpleBlock, nil
	case "BlockGroup":
		return ElementBlockGroup, nil
	case "Block":
		return ElementBlock, nil
	case "BlockAdditions":
		return ElementBlockAdditions, nil
	case "BlockMore":
		return ElementBlockMore, nil
	case "BlockAddID":
		return ElementBlockAddID, nil
	case "BlockAdditional":
		return ElementBlockAdditional, nil
	case "BlockDuration":
		return ElementBlockDuration, nil
	case "ReferenceBlock":
		return ElementReferenceBlock, nil
	case "DiscardPadding":
		return ElementDiscardPadding, nil
	case "Tracks":
		return ElementTracks, nil
	case "TrackEntry":
		return ElementTrackEntry, nil
	case "TrackNumber":
		return ElementTrackNumber, nil
	case "TrackUID":
		return ElementTrackUID, nil
	case "TrackType":
		return ElementTrackType, nil
	case "FlagEnabled":
		return ElementFlagEnabled, nil
	case "FlagForced":
		return ElementFlagForced, nil
	case "FlagLacing":
		return ElementFlagLacing, nil
	case "DefaultDuration":
		return ElementDefaultDuration, nil
	case "Name":
		return ElementName, nil
	case "Language":
		return ElementLanguage, nil
	case "CodecID":
		return ElementCodecID, nil
	case "CodecPrivate":
		return ElementCodecPrivate, nil
	case "CodecName":
		return ElementCodecName, nil
	case "CodecDelay":
		return ElementCodecDelay, nil
	case "SeekPreRoll":
		return ElementSeekPreRoll, nil
	case "Video":
		return ElementVideo, nil
	case "FlagInterlaced":
		return ElementFlagInterlaced, nil
	case "StereoMode":
		return ElementStereoMode, nil
	case "AlphaMode":
		return ElementAlphaMode, nil
	case "PixelWidth":
		return ElementPixelWidth, nil
	case "PixelHeight":
		return ElementPixelHeight, nil
	case "PixelCropBottom":
		return ElementPixelCropBottom, nil
	case "PixelCropTop":
		return ElementPixelCropTop, nil
	case "PixelCropLeft":
		return ElementPixelCropLeft, nil
	case "PixelCropRight":
		return ElementPixelCropRight, nil
	case "DisplayWidth":
		return ElementDisplayWidth, nil
	case "DisplayHeight":
		return ElementDisplayHeight, nil
	case "DisplayUnit":
		return ElementDisplayUnit, nil
	case "AspectRatioType":
		return ElementAspectRatioType, nil
	case "Audio":
		return ElementAudio, nil
	case "SamplingFrequency":
		return ElementSamplingFrequency, nil
	case "OutputSamplingFrequency":
		return ElementOutputSamplingFrequency, nil
	case "Channels":
		return ElementChannels, nil
	case "BitDepth":
		return ElementBitDepth, nil
	case "ContentEncodings":
		return ElementContentEncodings, nil
	case "ContentEncoding":
		return ElementContentEncoding, nil
	case "ContentEncodingOrder":
		return ElementContentEncodingOrder, nil
	case "ContentEncodingScope":
		return ElementContentEncodingScope, nil
	case "ContentEncodingType":
		return ElementContentEncodingType, nil
	case "ContentEncryption":
		return ElementContentEncryption, nil
	case "ContentEncAlgo":
		return ElementContentEncAlgo, nil
	case "ContentEncKeyID":
		return ElementContentEncKeyID, nil
	case "ContentEncAESSettings":
		return ElementContentEncAESSettings, nil
	case "AESSettingsCipherMode":
		return ElementAESSettingsCipherMode, nil
	default:
		return 0, nil
	}
}

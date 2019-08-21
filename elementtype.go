package ebml

// ElementType represents EBML Element type
type ElementType int

// EBML Element types
const (
	ElementEBML ElementType = iota
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
	ElementSeekPosision

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
	case ElementSeekPosision:
		return "SeekPosision"
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

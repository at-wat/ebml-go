package ebml

type elementTable map[uint8]interface{}
type element struct {
	e   ElementType
	t   Type
	top bool
}

var table = elementTable{
	0x11: elementTable{
		0x4D: elementTable{
			0x9B: elementTable{
				0x74: element{ElementSeekHead, TypeMaster, true},
			},
		},
	},
	0x15: elementTable{
		0x49: elementTable{
			0xA9: elementTable{
				0x66: element{ElementInfo, TypeMaster, true},
			},
		},
	},
	0x18: elementTable{
		0x53: elementTable{
			0x80: elementTable{
				0x67: element{ElementSegment, TypeMaster, false},
			},
		},
	},
	0x16: elementTable{
		0x54: elementTable{
			0xAE: elementTable{
				0x6B: element{ElementTracks, TypeMaster, true},
			},
		},
	},
	0x1A: elementTable{
		0x45: elementTable{
			0xDF: elementTable{
				0xA3: element{ElementEBML, TypeMaster, false},
			},
		},
	},
	0x1F: elementTable{
		0x43: elementTable{
			0xB6: elementTable{
				0x75: element{ElementCluster, TypeMaster, true},
			},
		},
	},
	0x23: elementTable{
		0xE3: elementTable{
			0x83: element{ElementDefaultDuration, TypeUInt, false},
		},
	},
	0x2A: elementTable{
		0xD7: elementTable{
			0xB1: element{ElementTimecodeScale, TypeUInt, false},
		},
	},
	0x42: elementTable{
		0x86: element{ElementEBMLVersion, TypeUInt, false},
		0xF7: element{ElementEBMLReadVersion, TypeUInt, false},
		0xF2: element{ElementEBMLMaxIDLength, TypeUInt, false},
		0xF3: element{ElementEBMLMaxSizeLength, TypeUInt, false},
		0x82: element{ElementEBMLDocType, TypeString, false},
		0x87: element{ElementEBMLDocTypeVersion, TypeUInt, false},
		0x85: element{ElementEBMLDocTypeReadVersion, TypeUInt, false},
	},
	0x44: elementTable{
		0x61: element{ElementDateUTC, TypeDate, false},
		0x89: element{ElementDuration, TypeFloat, false},
	},
	0x4D: elementTable{
		0xBB: element{ElementSeek, TypeMaster, false},
		0x80: element{ElementMuxingApp, TypeString, false},
	},
	0x53: elementTable{
		0x6E: element{ElementName, TypeString, false},
	},
	0x56: elementTable{
		0xAA: element{ElementCodecDelay, TypeUInt, false},
		0xBB: element{ElementSeekPreRoll, TypeUInt, false},
	},
	0x57: elementTable{
		0x41: element{ElementWritingApp, TypeString, false},
	},
	0x63: elementTable{
		0xA2: element{ElementCodecPrivate, TypeBinary, false},
	},
	0x73: elementTable{
		0xC5: element{ElementTrackUID, TypeUInt, false},
	},
	0x83: element{ElementTrackType, TypeUInt, false},
	0x86: element{ElementCodecID, TypeString, false},
	0x9F: element{ElementChannels, TypeUInt, false},
	0xA3: element{ElementSimpleBlock, TypeBinary, false},
	0xAB: element{ElementPrevSize, TypeUInt, false},
	0xAE: element{ElementTrackEntry, TypeMaster, false},
	0xB0: element{ElementPixelWidth, TypeUInt, false},
	0xB5: element{ElementSamplingFrequency, TypeFloat, false},
	0xBA: element{ElementPixelHeight, TypeUInt, false},
	0xD7: element{ElementTrackNumber, TypeUInt, false},
	0xE0: element{ElementVideo, TypeMaster, false},
	0xE1: element{ElementAudio, TypeMaster, false},
	0xE7: element{ElementTimecode, TypeUInt, false},
	0xEC: element{ElementVoid, TypeMaster, false},
}

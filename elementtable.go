package ebml

type elementDef struct {
	b   []byte
	t   Type
	top bool
}
type elementTable map[ElementType]elementDef

var table = elementTable{
	ElementSeekHead:               elementDef{[]byte{0x11, 0x4D, 0x9B, 0x74}, TypeMaster, true},
	ElementInfo:                   elementDef{[]byte{0x15, 0x49, 0xA9, 0x66}, TypeMaster, true},
	ElementSegment:                elementDef{[]byte{0x18, 0x53, 0x80, 0x67}, TypeMaster, false},
	ElementTracks:                 elementDef{[]byte{0x16, 0x54, 0xAE, 0x6B}, TypeMaster, true},
	ElementEBML:                   elementDef{[]byte{0x1A, 0x45, 0xDF, 0xA3}, TypeMaster, false},
	ElementCluster:                elementDef{[]byte{0x1F, 0x43, 0xB6, 0x75}, TypeMaster, true},
	ElementDefaultDuration:        elementDef{[]byte{0x23, 0xE3, 0x83}, TypeUInt, false},
	ElementTimecodeScale:          elementDef{[]byte{0x2A, 0xD7, 0xB1}, TypeUInt, false},
	ElementEBMLVersion:            elementDef{[]byte{0x42, 0x86}, TypeUInt, false},
	ElementEBMLReadVersion:        elementDef{[]byte{0x42, 0xF7}, TypeUInt, false},
	ElementEBMLMaxIDLength:        elementDef{[]byte{0x42, 0xF2}, TypeUInt, false},
	ElementEBMLMaxSizeLength:      elementDef{[]byte{0x42, 0xF3}, TypeUInt, false},
	ElementEBMLDocType:            elementDef{[]byte{0x42, 0x82}, TypeString, false},
	ElementEBMLDocTypeVersion:     elementDef{[]byte{0x42, 0x87}, TypeUInt, false},
	ElementEBMLDocTypeReadVersion: elementDef{[]byte{0x42, 0x85}, TypeUInt, false},
	ElementDateUTC:                elementDef{[]byte{0x44, 0x61}, TypeDate, false},
	ElementDuration:               elementDef{[]byte{0x44, 0x89}, TypeFloat, false},
	ElementSeek:                   elementDef{[]byte{0x4D, 0xBB}, TypeMaster, false},
	ElementSeekID:                 elementDef{[]byte{0x53, 0xAB}, TypeBinary, false},
	ElementSeekPosition:           elementDef{[]byte{0x53, 0xAC}, TypeUInt, false},
	ElementMuxingApp:              elementDef{[]byte{0x4D, 0x80}, TypeString, false},
	ElementName:                   elementDef{[]byte{0x53, 0x6E}, TypeString, false},
	ElementCodecDelay:             elementDef{[]byte{0x56, 0xAA}, TypeUInt, false},
	ElementSeekPreRoll:            elementDef{[]byte{0x56, 0xBB}, TypeUInt, false},
	ElementWritingApp:             elementDef{[]byte{0x57, 0x41}, TypeString, false},
	ElementCodecPrivate:           elementDef{[]byte{0x63, 0xA2}, TypeBinary, false},
	ElementTrackUID:               elementDef{[]byte{0x73, 0xC5}, TypeUInt, false},
	ElementTrackType:              elementDef{[]byte{0x83}, TypeUInt, false},
	ElementCodecID:                elementDef{[]byte{0x86}, TypeString, false},
	ElementChannels:               elementDef{[]byte{0x9F}, TypeUInt, false},
	ElementSimpleBlock:            elementDef{[]byte{0xA3}, TypeBinary, false},
	ElementBlockGroup:             elementDef{[]byte{0xA0}, TypeMaster, false},
	ElementBlockDuration:          elementDef{[]byte{0x9B}, TypeUInt, false},
	ElementBlock:                  elementDef{[]byte{0xA1}, TypeBinary, false},
	ElementPrevSize:               elementDef{[]byte{0xAB}, TypeUInt, false},
	ElementTrackEntry:             elementDef{[]byte{0xAE}, TypeMaster, false},
	ElementPixelWidth:             elementDef{[]byte{0xB0}, TypeUInt, false},
	ElementSamplingFrequency:      elementDef{[]byte{0xB5}, TypeFloat, false},
	ElementPixelHeight:            elementDef{[]byte{0xBA}, TypeUInt, false},
	ElementTrackNumber:            elementDef{[]byte{0xD7}, TypeUInt, false},
	ElementVideo:                  elementDef{[]byte{0xE0}, TypeMaster, false},
	ElementAudio:                  elementDef{[]byte{0xE1}, TypeMaster, false},
	ElementTimecode:               elementDef{[]byte{0xE7}, TypeUInt, false},
	ElementVoid:                   elementDef{[]byte{0xEC}, TypeMaster, false},
}

type elementRevTable map[uint8]interface{}
type element struct {
	e   ElementType
	t   Type
	top bool
}

var revTable elementRevTable

func init() {
	revTable = make(elementRevTable)

	for k, v := range table {
		var p interface{}
		p = revTable
		for i := 0; i < len(v.b); i++ {
			b := v.b[i]
			if p.(elementRevTable)[b] == nil {
				p.(elementRevTable)[b] = make(elementRevTable)
			}
			if i == len(v.b)-1 {
				p.(elementRevTable)[b] = element{e: k, t: v.t, top: v.top}
			} else {
				p = p.(elementRevTable)[b]
			}
		}
	}
}

package ebml

import (
	"bytes"
	"testing"
)

func TestMarshal(t *testing.T) {
	type EBMLHeader struct {
		DocType            string `ebml:"EBMLDocType"`
		DocTypeVersion     uint64 `ebml:"EBMLDocTypeVersion"`
		DocTypeReadVersion uint64 `ebml:"EBMLDocTypeReadVersion"`
	}
	type TestEBML struct {
		Header EBMLHeader `ebml:"EBML"`
	}
	s := TestEBML{
		Header: EBMLHeader{
			DocType:            "webm",
			DocTypeVersion:     2,
			DocTypeReadVersion: 2,
		},
	}

	expected := []byte{
		0x1a, 0x45, 0xdf, 0xa3, // EBML
		0x90, // 0x10
		0x42, 0x82, 0x85, 0x77, 0x65, 0x62, 0x6d, 0x00,
		0x42, 0x87, 0x81, 0x02, 0x42, 0x85, 0x81, 0x02,
	}

	b, err := Marshal(&s)
	if err != nil {
		t.Fatalf("error: %+v\n", err)
	}

	if bytes.Compare(expected, b) != 0 {
		t.Errorf("Marshaled binary doesn't match:\n expected: %v,\n      got: %v", expected, b)
	}
}

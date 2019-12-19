package ebml

import (
	"bytes"
	"testing"
)

func TestElementType_Roundtrip(t *testing.T) {
	for e := ElementInvalid + 1; e < elementMax; e++ {
		s := e.String()
		if el, err := ElementTypeFromString(s); err != nil {
			t.Errorf("Failed to get ElementType from string: %v", err)
		} else if e != el {
			t.Errorf("Failed to roundtrip ElementType %d and string", e)
		}
	}
}

func TestElementType_Bytes(t *testing.T) {
	expected := []byte{0x18, 0x53, 0x80, 0x67}

	if !bytes.Equal(ElementSegment.Bytes(), expected) {
		t.Errorf("Unexpected bytes, expected: %v, got: %v", expected, ElementSegment.Bytes())
	}
	if ElementSegment.DataType() != DataTypeMaster {
		t.Errorf("Unexpected DataType, expected: %s, got: %s", DataTypeMaster, ElementSegment.DataType())
	}
}

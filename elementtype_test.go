package ebml

import (
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

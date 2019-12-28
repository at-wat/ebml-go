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
	"io"
	"testing"

	"github.com/at-wat/ebml-go/internal/errs"
)

func TestElementType_Roundtrip(t *testing.T) {
	for e := ElementInvalid + 1; e < elementMax; e++ {
		s := e.String()
		if el, err := ElementTypeFromString(s); err != nil {
			t.Errorf("Failed to get ElementType from string: '%v'", err)
		} else if e != el {
			t.Errorf("Failed to roundtrip ElementType %d and string", e)
		}
	}
	if elementMax.String() != "unknown" {
		t.Errorf("Invalid ElementType string should be 'unknown', got '%s'", elementMax.String())
	}
}

func TestElementType_Bytes(t *testing.T) {
	expected := []byte{0x18, 0x53, 0x80, 0x67}

	if !bytes.Equal(ElementSegment.Bytes(), expected) {
		t.Errorf("Expected bytes: '%v', got: '%v'", expected, ElementSegment.Bytes())
	}
	if ElementSegment.DataType() != DataTypeMaster {
		t.Errorf("Expected DataType: %s, got: %s", DataTypeMaster, ElementSegment.DataType())
	}
}

func TestElementType_InitReverseLookupTable(t *testing.T) {
	defer func() {
		err := recover()
		switch v := err.(type) {
		case error:
			if !errs.Is(v, io.ErrUnexpectedEOF) {
				t.Errorf("Expected initReverseLookupTable panic: '%v', got: '%v'", io.ErrUnexpectedEOF, v)
			}
		default:
			t.Errorf("initReverseLookupTable paniced with unexpected type %T", v)
		}
	}()

	revTb := make(elementRevTable)
	initReverseLookupTable(revTb, elementTable{
		ElementType(0): elementDef{}, // empty bytes representation
	})
	t.Fatal("initReverseLookupTable must panic if elementTable is broken.")
}

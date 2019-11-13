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

// Type represents EBML Element data type
type Type int

// EBML Element data types
const (
	TypeMaster Type = iota
	TypeInt
	TypeUInt
	TypeDate
	TypeFloat
	TypeBinary
	TypeString
	TypeBlock
)

func (t Type) String() string {
	switch t {
	case TypeMaster:
		return "Master"
	case TypeInt:
		return "Int"
	case TypeUInt:
		return "UInt"
	case TypeDate:
		return "Date"
	case TypeFloat:
		return "Float"
	case TypeBinary:
		return "Binary"
	case TypeString:
		return "String"
	case TypeBlock:
		return "Block"
	default:
		return "Unknown type"
	}
}

// Metadata represents a metadata (position) of the EBML element
type Metadata struct {
	Position uint64
}

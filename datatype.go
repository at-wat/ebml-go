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
	"reflect"
)

// DataType represents EBML Element data type.
type DataType int

// EBML Element data types.
const (
	DataTypeMaster DataType = iota
	DataTypeInt
	DataTypeUInt
	DataTypeDate
	DataTypeFloat
	DataTypeBinary
	DataTypeString
	DataTypeBlock
)

var dataTypeName = map[DataType]string{
	DataTypeMaster: "Master",
	DataTypeInt:    "Int",
	DataTypeUInt:   "UInt",
	DataTypeDate:   "Date",
	DataTypeFloat:  "Float",
	DataTypeBinary: "Binary",
	DataTypeString: "String",
	DataTypeBlock:  "Block",
}

func (t DataType) String() string {
	if name, ok := dataTypeName[t]; ok {
		return name
	}
	return "unknown"
}

func isConvertible(src, dst reflect.Type) bool {
	switch src.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch dst.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return true
		default:
			return false
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch dst.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return true
		default:
			return false
		}
	case reflect.Float32, reflect.Float64:
		switch dst.Kind() {
		case reflect.Float32, reflect.Float64:
			return true
		default:
			return false
		}
	}
	return false
}

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

// Package mkv provides the Matroska multimedia writer.
//
// The package implements block data writer for multi-track Matroska container.
package mkv

import (
	"time"
)

// EBMLHeader represents EBML header struct.
type EBMLHeader struct {
	EBMLVersion        uint64 `ebml:"EBMLVersion"`
	EBMLReadVersion    uint64 `ebml:"EBMLReadVersion"`
	EBMLMaxIDLength    uint64 `ebml:"EBMLMaxIDLength"`
	EBMLMaxSizeLength  uint64 `ebml:"EBMLMaxSizeLength"`
	DocType            string `ebml:"EBMLDocType"`
	DocTypeVersion     uint64 `ebml:"EBMLDocTypeVersion"`
	DocTypeReadVersion uint64 `ebml:"EBMLDocTypeReadVersion"`
}

// Info represents Info element struct.
type Info struct {
	TimecodeScale uint64    `ebml:"TimecodeScale"`
	MuxingApp     string    `ebml:"MuxingApp,omitempty"`
	WritingApp    string    `ebml:"WritingApp,omitempty"`
	Duration      float64   `ebml:"Duration,omitempty"`
	DateUTC       time.Time `ebml:"DateUTC,omitempty"`
}

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

package mkvcore

import (
	"github.com/at-wat/ebml-go"
)

type simpleBlockGroup struct {
	Block             []ebml.Block `ebml:"Block"`
	ReferencePriority uint64       `ebml:"ReferencePriority"`
}

type simpleBlockCluster struct {
	Timecode    uint64             `ebml:"Timecode"`
	PrevSize    uint64             `ebml:"PrevSize,omitempty"`
	SimpleBlock []ebml.Block       `ebml:"SimpleBlock"`
	BlockGroup  []simpleBlockGroup `ebml:"BlockGroup,omitempty"`
}

type seekFixed struct {
	SeekID       []byte  `ebml:"SeekID"`
	SeekPosition *uint64 `ebml:"SeekPosition,size=8"`
}

type seekHeadFixed struct {
	Seek []seekFixed `ebml:"Seek"`
}

type flexTracks struct {
	TrackEntry []interface{} `ebml:"TrackEntry"`
}

type flexSegment struct {
	SeekHead *seekHeadFixed       `ebml:"SeekHead,omitempty"`
	Info     interface{}          `ebml:"Info"`
	Tracks   flexTracks           `ebml:"Tracks"`
	Cluster  []simpleBlockCluster `ebml:"Cluster,size=unknown"`
}

type flexHeader struct {
	Header  interface{} `ebml:"EBML"`
	Segment flexSegment `ebml:"Segment,size=unknown"`
}

// TrackEntry is a TrackEntry struct with all mandatory elements and commonly used elements.
type TrackEntry struct {
	TrackNumber        uint64
	TrackUID           uint64
	TrackType          uint8
	FlagEnabled        uint8
	FlagDefault        uint8
	FlagForced         uint8
	FlagLacing         uint8
	MinCache           uint64
	DefaultDuration    uint64
	MaxBlockAdditionID uint64
	Name               string
	Language           string
	LanguageIETF       string
	CodecID            string
	CodecDecodeAll     uint8
	SeekPreRoll        uint64
}

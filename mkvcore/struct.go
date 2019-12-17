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

type simpleBlockCluster struct {
	Timecode    uint64       `ebml:"Timecode"`
	PrevSize    uint64       `ebml:"PrevSize,omitempty"`
	SimpleBlock []ebml.Block `ebml:"SimpleBlock"`
}

type seekFixed struct {
	SeekID       []byte  `ebml:"SeekID"`
	SeekPosition *uint64 `ebml:"SeekPosition,size=8"`
}

type seekHeadFixed struct {
	Seek []seekFixed `ebml:"Seek"`
}

type flexSegment struct {
	SeekHead *seekHeadFixed `ebml:"SeekHead,omitempty"`
	Info     interface{}    `ebml:"Info"`
	Tracks   struct {
		TrackEntry []interface{} `ebml:"TrackEntry"`
	} `ebml:"Tracks"`
	Cluster []simpleBlockCluster `ebml:"Cluster,size=unknown"`
}

type flexHeader struct {
	Header  interface{} `ebml:"EBML"`
	Segment flexSegment `ebml:"Segment,size=unknown"`
}

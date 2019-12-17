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
	"bytes"

	"github.com/at-wat/ebml-go"
)

func setSeekHead(header *flexHeader, opts ...ebml.MarshalOption) error {
	infoPos := new(uint64)
	tracksPos := new(uint64)
	header.Segment.SeekHead = &seekHeadFixed{}
	if header.Segment.Info != nil {
		header.Segment.SeekHead.Seek = append(header.Segment.SeekHead.Seek, seekFixed{
			SeekID:       ebml.ElementInfo.Bytes(),
			SeekPosition: infoPos,
		})
	}
	header.Segment.SeekHead.Seek = append(header.Segment.SeekHead.Seek, seekFixed{
		SeekID:       ebml.ElementTracks.Bytes(),
		SeekPosition: tracksPos,
	})

	var segmentPos uint64
	hook := func(e *ebml.Element) {
		switch e.Name {
		case "SeekHead":
			// SeekHead position is the top of the Segment contents.
			// Origin of the segment position is here.
			segmentPos = e.Position
		case "Info":
			*infoPos = e.Position - segmentPos
		case "Tracks":
			*tracksPos = e.Position - segmentPos
		}
	}

	optsWithHook := append([]ebml.MarshalOption{}, opts...)
	optsWithHook = append(optsWithHook, ebml.WithElementWriteHooks(hook))

	var buf bytes.Buffer
	if err := ebml.Marshal(header, &buf, optsWithHook...); err != nil {
		return err
	}

	return nil
}

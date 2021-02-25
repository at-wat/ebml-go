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
	"sync"
	"testing"

	"github.com/at-wat/ebml-go"
	"github.com/at-wat/ebml-go/internal/buffercloser"
)

func TestBlockReader(t *testing.T) {
	s := struct {
		Segment flexSegment `ebml:"Segment"`
	}{
		Segment: flexSegment{
			Tracks: flexTracks{TrackEntry: []interface{}{
				map[string]interface{}{"TrackNumber": uint(1)},
				map[string]interface{}{"TrackNumber": uint(2)},
			}},
			Cluster: []simpleBlockCluster{
				{
					Timecode: uint64(0),
					SimpleBlock: []ebml.Block{
						{
							TrackNumber: 1,
							Timecode:    int16(0),
							Keyframe:    false,
							Data:        [][]byte{{0x01, 0x02}},
						},
						{
							TrackNumber: 2,
							Timecode:    int16(10),
							Keyframe:    true,
							Data:        [][]byte{{0x03, 0x04, 0x05}},
						},
						{
							TrackNumber: 1,
							Timecode:    int16(30),
							Keyframe:    true,
							Data:        [][]byte{{0x06}},
						},
					},
				},
				{
					Timecode: uint64(30),
					PrevSize: uint64(39),
				},
			},
		},
	}
	buf := buffercloser.New()
	if err := ebml.Marshal(&s, buf); err != nil {
		t.Fatalf("Failed to marshal test data: '%v'", err)
	}
	buf.Close()

	ws, err := NewSimpleBlockReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("Failed to create BlockReader: '%v'", err)
	}

	if len(ws) != 2 {
		t.Fatalf("Number of the returned writer (%d) must be same as the number of TrackEntry (%d)", len(ws), 2)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// TODO: store test data as array.
	go func() {
		defer wg.Done()

		if buf, keyframe, timestamp, err := ws[0].Read(); err != nil {
			t.Errorf("Failed to Read: '%v'", err)
		} else if keyframe {
			t.Errorf("Expected keyframe: false, got: %v", keyframe)
		} else if timestamp != 0 {
			t.Errorf("Expected timestamp: 0, got: %v", timestamp)
		} else if bytes.Compare(buf, []byte{0x01, 0x02}) != 0 {
			t.Errorf("Expected bytes: [0x01, 0x02], got: %v", buf)
		}

		if buf, keyframe, timestamp, err := ws[0].Read(); err != nil {
			t.Errorf("Failed to Read: '%v'", err)
		} else if !keyframe {
			t.Errorf("Expected keyframe: true, got: %v", keyframe)
		} else if timestamp != 30 {
			t.Errorf("Expected timestamp: 30, got: %v", timestamp)
		} else if bytes.Compare(buf, []byte{0x06}) != 0 {
			t.Errorf("Expected bytes: [0x06], got: %v", buf)
		}
	}()

	go func() {
		defer wg.Done()

		if buf, keyframe, timestamp, err := ws[1].Read(); err != nil {
			t.Errorf("Failed to Read: '%v'", err)
		} else if !keyframe {
			t.Errorf("Expected keyframe: true, got: %v", keyframe)
		} else if timestamp != 10 {
			t.Errorf("Expected timestamp: 10, got: %v", timestamp)
		} else if bytes.Compare(buf, []byte{0x03, 0x04, 0x05}) != 0 {
			t.Errorf("Expected bytes: [0x03, 0x04, 0x05], got: %v", buf)
		}
	}()

	wg.Wait()
}

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

package ebml_test

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/at-wat/ebml-go"
	"github.com/at-wat/ebml-go/webm"
)

func TestMarshal_RoundtripWebM(t *testing.T) {
	webm0 := struct {
		Header  webm.EBMLHeader `ebml:"EBML"`
		Segment webm.Segment    `ebml:"Segment,size=unknown"`
	}{
		Header: webm.EBMLHeader{
			EBMLVersion:        1,
			EBMLReadVersion:    1,
			EBMLMaxIDLength:    4,
			EBMLMaxSizeLength:  8,
			DocType:            "webm",
			DocTypeVersion:     2,
			DocTypeReadVersion: 2,
		},
		Segment: webm.Segment{
			Info: webm.Info{
				TimecodeScale: 1000000, // 1ms
				MuxingApp:     "ebml-go example",
				WritingApp:    "ebml-go example",
				DateUTC:       time.Now().Truncate(time.Millisecond),
			},
			Tracks: webm.Tracks{
				TrackEntry: []webm.TrackEntry{
					{
						Name:            "Video",
						TrackNumber:     1,
						TrackUID:        12345,
						CodecID:         "V_VP8",
						TrackType:       1,
						DefaultDuration: 33333333,
						Video: &webm.Video{
							PixelWidth:  320,
							PixelHeight: 240,
						},
						CodecPrivate: []byte{0x01, 0x02},
					},
					{
						Name:            "Audio",
						TrackNumber:     2,
						TrackUID:        54321,
						CodecID:         "A_OPUS",
						TrackType:       2,
						DefaultDuration: 33333333,
						Audio: &webm.Audio{
							SamplingFrequency: 48000.0,
							Channels:          2,
						},
					},
				},
			},
			Cluster: []webm.Cluster{
				{
					Timecode: 0,
				},
				{
					Timecode: 1234567,
				},
			},
			Cues: &webm.Cues{
				CuePoint: []webm.CuePoint{
					{
						CueTime: 1,
						CueTrackPositions: []webm.CueTrackPosition{
							{CueTrack: 2, CueClusterPosition: 3, CueBlockNumber: 4},
						},
					},
				},
			},
		},
	}

	var b bytes.Buffer
	if err := ebml.Marshal(&webm0, &b); err != nil {
		t.Fatalf("Failed to Marshal: '%v'", err)
	}
	var webm1 struct {
		Header  webm.EBMLHeader `ebml:"EBML"`
		Segment webm.Segment    `ebml:"Segment,size=unknown"`
	}
	if err := ebml.Unmarshal(bytes.NewBuffer(b.Bytes()), &webm1); err != nil {
		t.Fatalf("Failed to Unmarshal: '%v'", err)
	}

	if !reflect.DeepEqual(webm0, webm1) {
		t.Errorf("Roundtrip result doesn't match original\nexpected: %+v\n     got: %+v", webm0, webm1)
	}
}

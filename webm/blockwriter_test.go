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

package webm

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/at-wat/ebml-go"
	"github.com/at-wat/ebml-go/internal/buffercloser"
	"github.com/at-wat/ebml-go/mkvcore"
)

func TestBlockWriter(t *testing.T) {
	buf := buffercloser.New()

	tracks := []TrackEntry{
		{
			Name:        "Video",
			TrackNumber: 1,
			TrackUID:    12345,
			CodecID:     "V_VP8",
			TrackType:   1,
			Video: &Video{
				PixelWidth:  320,
				PixelHeight: 240,
			},
		},
		{
			Name:        "Audio",
			TrackNumber: 2,
			TrackUID:    54321,
			CodecID:     "A_OPUS",
			TrackType:   2,
			Audio: &Audio{
				SamplingFrequency: 48000.0,
				Channels:          2,
			},
		},
	}
	ws, err := NewSimpleBlockWriter(buf, tracks)
	if err != nil {
		t.Fatalf("Failed to create BlockWriter: %v", err)
	}

	if len(ws) != len(tracks) {
		t.Fatalf("Number of the returned writer (%d) must be same as the number of TrackEntry (%d)", len(ws), len(tracks))
	}

	if n, err := ws[0].Write(false, 100, []byte{0x01, 0x02}); err != nil {
		t.Fatalf("Failed to Write: %v", err)
	} else if n != 2 {
		t.Errorf("Expected return value of BlockWriter.Write: 2, got: %d", n)
	}

	if n, err := ws[1].Write(true, 110, []byte{0x03, 0x04, 0x05}); err != nil {
		t.Fatalf("Failed to Write: %v", err)
	} else if n != 3 {
		t.Errorf("Expected return value of BlockWriter.Write: 3, got: %d", n)
	}

	// Ignored due to old timestamp
	if n, err := ws[0].Write(true, -32769, []byte{0x0A}); err != nil {
		t.Fatalf("Failed to Write: %v", err)
	} else if n != 1 {
		t.Errorf("Expected return value of BlockWriter.Write: 1, got: %d", n)
	}

	if n, err := ws[0].Write(true, 130, []byte{0x06}); err != nil {
		t.Fatalf("Failed to Write: %v", err)
	} else if n != 1 {
		t.Errorf("Expected return value of BlockWriter.Write: 1, got: %d", n)
	}

	ws[0].Close()
	ws[1].Close()
	select {
	case <-buf.Closed():
	default:
		t.Errorf("Base io.WriteCloser is not closed by BlockWriter")
	}

	expected := struct {
		Header  EBMLHeader `ebml:"EBML"`
		Segment Segment    `ebml:"Segment,size=unknown"`
	}{
		Header: *DefaultEBMLHeader,
		Segment: Segment{
			Info: *DefaultSegmentInfo,
			Tracks: Tracks{
				TrackEntry: tracks,
			},
			Cluster: []Cluster{
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
	var result struct {
		Header  EBMLHeader `ebml:"EBML"`
		Segment Segment    `ebml:"Segment,size=unknown"`
	}
	if err := ebml.Unmarshal(bytes.NewReader(buf.Bytes()), &result); err != nil {
		t.Fatalf("Failed to Unmarshal resultant binary: %v", err)
	}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Unexpected WebM data,\nexpected: %+v\n     got: %+v", expected, result)
	}
}

func TestBlockWriter_NewSimpleWriter(t *testing.T) {
	buf := buffercloser.New()

	tracks := []TrackEntry{
		{
			TrackNumber: 1,
			TrackUID:    2,
			CodecID:     "",
			TrackType:   1,
		},
	}

	// Check old API
	var ws []*FrameWriter
	var err error

	ws, err = NewSimpleWriter(
		buf, tracks,
		mkvcore.WithEBMLHeader(nil),
		mkvcore.WithSegmentInfo(nil),
		mkvcore.WithSeekHead(false),
	)
	if err != nil {
		t.Fatalf("Failed to create BlockWriter: %v", err)
	}

	if len(ws) != 1 {
		t.Fatalf("Number of the returned writer must be 1, but got %d", len(ws))
	}
	ws[0].Close()

	expectedBytes := []byte{
		0x18, 0x53, 0x80, 0x67, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0x16, 0x54, 0xAE, 0x6B, 0x8E,
		0xAE, 0x8C,
		0xD7, 0x81, 0x01,
		0x73, 0xC5, 0x81, 0x02,
		0x86, 0x80,
		0x83, 0x81, 0x01,
		0x1F, 0x43, 0xB6, 0x75, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xE7, 0x81, 0x00,
	}
	if !bytes.Equal(buf.Bytes(), expectedBytes) {
		t.Errorf("Unexpected WebM binary,\nexpected: %+v\n     got: %+v", expectedBytes, buf.Bytes())
	}
}

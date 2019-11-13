package webm

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/at-wat/ebml-go"
)

func TestSimpleWriter(t *testing.T) {
	buf := &bufferCloser{closed: make(chan struct{})}

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
			CodecID:     "V_OPUS",
			TrackType:   2,
			Audio: &Audio{
				SamplingFrequency: 48000.0,
				Channels:          2,
			},
		},
	}
	ws, err := NewSimpleWriter(buf, tracks)
	if err != nil {
		t.Fatalf("Failed to create SimpleWriter: %v", err)
	}

	if len(ws) != len(tracks) {
		t.Fatalf("Number of the returned writer (%d) must be same as the number of TrackEntry (%d)", len(ws), len(tracks))
	}

	if n, err := ws[0].Write(false, 100, []byte{0x01, 0x02}); err != nil {
		t.Fatalf("Failed to Write: %v", err)
	} else if n != 2 {
		t.Errorf("Unexpected return value of FrameWriter.Write, expected: 2, got: %d", n)
	}

	if n, err := ws[1].Write(true, 110, []byte{0x03, 0x04, 0x05}); err != nil {
		t.Fatalf("Failed to Write: %v", err)
	} else if n != 3 {
		t.Errorf("Unexpected return value of FrameWriter.Write, expected: 3, got: %d", n)
	}

	if n, err := ws[0].Write(true, 130, []byte{0x06}); err != nil {
		t.Fatalf("Failed to Write: %v", err)
	} else if n != 1 {
		t.Errorf("Unexpected return value of FrameWriter.Write, expected: 1, got: %d", n)
	}

	ws[0].Close()
	ws[1].Close()
	select {
	case <-buf.closed:
	default:
		t.Errorf("Base io.WriteCloser is not closed by SimpleWriter")
	}

	expected := struct {
		Header  EBMLHeader `ebml:"EBML"`
		Segment Segment    `ebml:"Segment,inf"`
	}{
		Header: defaultEBMLHeader,
		Segment: Segment{
			Info: defaultSegmentInfo,
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
							Data:        [][]byte{[]byte{0x01, 0x02}},
						},
						{
							TrackNumber: 2,
							Timecode:    int16(10),
							Keyframe:    true,
							Data:        [][]byte{[]byte{0x03, 0x04, 0x05}},
						},
						{
							TrackNumber: 1,
							Timecode:    int16(30),
							Keyframe:    true,
							Data:        [][]byte{[]byte{0x06}},
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
	defer func() {
		var result struct {
			Header  EBMLHeader `ebml:"EBML"`
			Segment Segment    `ebml:"Segment,inf"`
		}
		if err := ebml.Unmarshal(bytes.NewReader(buf.Bytes()), &result); err != nil {
			t.Fatalf("Failed to Unmarshal resultant binary: %v", err)
		}
		// Clear all metadata because metadata is not an interest of this test case
		clearAllMetadata(reflect.ValueOf(&result).Elem())
		if !reflect.DeepEqual(expected, result) {
			t.Errorf("Unexpected WebM data,\nexpected: %+v\n     got: %+v", expected, result)
		}
	}()
}

func clearAllMetadata(vo reflect.Value) {
	switch vo.Kind() {
	case reflect.Struct:
		for i := 0; i < vo.NumField(); i++ {
			if vo.Type().Field(i).Name == "Metadata" {
				vo.Field(i).Set(reflect.ValueOf(ebml.Metadata{}))
			} else {
				clearAllMetadata(vo.Field(i))
			}
		}
	case reflect.Ptr:
		if !vo.IsNil() {
			vo = vo.Elem()
			for i := 0; i < vo.NumField(); i++ {
				if vo.Type().Field(i).Name == "Metadata" {
					vo.Field(i).Set(reflect.ValueOf(ebml.Metadata{}))
				} else {
					clearAllMetadata(vo.Field(i))
				}
			}
		}
	case reflect.Slice:
		for i := 0; i < vo.Len(); i++ {
			clearAllMetadata(vo.Index(i))
		}
	}
}

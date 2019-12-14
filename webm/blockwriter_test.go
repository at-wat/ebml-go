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
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/at-wat/ebml-go"
)

func TestBlockWriter(t *testing.T) {
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
		t.Errorf("Unexpected return value of BlockWriter.Write, expected: 2, got: %d", n)
	}

	if n, err := ws[1].Write(true, 110, []byte{0x03, 0x04, 0x05}); err != nil {
		t.Fatalf("Failed to Write: %v", err)
	} else if n != 3 {
		t.Errorf("Unexpected return value of BlockWriter.Write, expected: 3, got: %d", n)
	}

	// Ignored due to old timestamp
	if n, err := ws[0].Write(true, -32769, []byte{0x0A}); err != nil {
		t.Fatalf("Failed to Write: %v", err)
	} else if n != 1 {
		t.Errorf("Unexpected return value of BlockWriter.Write, expected: 1, got: %d", n)
	}

	if n, err := ws[0].Write(true, 130, []byte{0x06}); err != nil {
		t.Fatalf("Failed to Write: %v", err)
	} else if n != 1 {
		t.Errorf("Unexpected return value of BlockWriter.Write, expected: 1, got: %d", n)
	}

	ws[0].Close()
	ws[1].Close()
	select {
	case <-buf.closed:
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

func TestBlockWriter_Options(t *testing.T) {
	buf := &bufferCloser{closed: make(chan struct{})}

	tracks := []TrackEntry{
		{
			TrackNumber: 1,
			TrackUID:    2,
			CodecID:     "",
			TrackType:   1,
		},
	}

	ws, err := NewSimpleBlockWriter(
		buf, tracks,
		WithEBMLHeader(nil),
		WithSegmentInfo(nil),
		WithMarshalOptions(ebml.WithDataSizeLen(2)),
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
		0x16, 0x54, 0xAE, 0x6B, 0x40, 0x14,
		0xAE, 0x40, 0x11,
		0xD7, 0x40, 0x01, 0x01,
		0x73, 0xC5, 0x40, 0x01, 0x02,
		0x86, 0x40, 0x01, 0x00,
		0x83, 0x40, 0x01, 0x01,
		0x1F, 0x43, 0xB6, 0x75, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xE7, 0x40, 0x01, 0x00,
	}
	if !bytes.Equal(buf.Bytes(), expectedBytes) {
		t.Errorf("Unexpected WebM binary,\nexpected: %+v\n     got: %+v", expectedBytes, buf.Bytes())
	}
}

func TestBlockWriter_FailingOptions(t *testing.T) {
	errDummy0 := errors.New("an error 0")
	errDummy1 := errors.New("an error 1")

	cases := map[string]struct {
		opts []BlockWriterOption
		err  error
	}{
		"WriterOptionError": {
			opts: []BlockWriterOption{
				func(*BlockWriterOptions) error { return errDummy0 },
			},
			err: errDummy0,
		},
		"MarshalOptionError": {
			opts: []BlockWriterOption{
				WithMarshalOptions(
					func(*ebml.MarshalOptions) error { return errDummy1 },
				),
			},
			err: errDummy1,
		},
		"MaxKeyframeIntervalOptionError": {
			opts: []BlockWriterOption{
				WithMaxKeyframeInterval(0, 0),
			},
			err: errInvalidTrackNumber,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			buf := &bufferCloser{closed: make(chan struct{})}
			_, err := NewSimpleBlockWriter(buf, []TrackEntry{}, c.opts...)
			if err != c.err {
				t.Errorf("Unexpected error, expected: %v, got: %v", c.err, err)
			}
		})
	}
}

type errorWriter struct {
	wrote chan struct{}
	err   error
	mu    sync.Mutex
}

func (w *errorWriter) setError(err error) {
	w.mu.Lock()
	w.err = err
	w.mu.Unlock()
}

func (w *errorWriter) Write(b []byte) (int, error) {
	select {
	case w.wrote <- struct{}{}:
	default:
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.err != nil {
		return 0, w.err
	}
	return len(b), nil
}

func (w *errorWriter) WaitWrite() bool {
	select {
	case <-w.wrote:
	case <-time.After(time.Second):
		return false
	}
	return true
}

func (w *errorWriter) Close() error {
	return nil
}

func TestBlockWriter_ErrorHandling(t *testing.T) {
	tracks := []TrackEntry{
		{
			TrackNumber: 1,
			TrackUID:    2,
			CodecID:     "",
			TrackType:   1,
		},
	}

	const (
		atBeginning int = iota
		atClusterWriting
		atFrameWriting
		atClosing
	)

	for name, errAt := range map[string]int{
		"ErrorAtBeginning":      atBeginning,
		"ErrorAtClusterWriting": atClusterWriting,
		"ErrorAtFrameWriting":   atFrameWriting,
		"ErrorAtClosing":        atClosing,
	} {
		t.Run(name, func(t *testing.T) {
			chFatal := make(chan error, 1)
			chError := make(chan error, 1)
			clearErr := func() {
				for {
					select {
					case <-chFatal:
					case <-chError:
					default:
						return
					}
				}
			}

			w := &errorWriter{wrote: make(chan struct{}, 1)}

			if errAt == atBeginning {
				w.setError(bytes.ErrTooLarge)
			}
			clearErr()
			ws, err := NewSimpleBlockWriter(
				w, tracks,
				WithOnErrorHandler(func(err error) { chError <- err }),
				WithOnFatalHandler(func(err error) { chFatal <- err }),
				WithBlockInterceptor(nil), // write without sorter
			)
			if err != nil {
				if errAt == atBeginning {
					if err != bytes.ErrTooLarge {
						t.Fatalf("Unexpected error, expected: %v, got: %v", bytes.ErrTooLarge, err)
					}
					return
				}
				t.Fatalf("Failed to create SimpleWriter: %v", err)
			}

			if len(ws) != 1 {
				t.Fatalf("Number of the returned writer must be 1, but got %d", len(ws))
			}

			if errAt == atClusterWriting {
				w.setError(bytes.ErrTooLarge)
			}
			clearErr()
			if _, err := ws[0].Write(false, 100, []byte{0x01, 0x02}); err != nil {
				t.Fatalf("Failed to Write: %v", err)
			}
			if errAt == atClusterWriting {
				select {
				case err := <-chFatal:
					if err != bytes.ErrTooLarge {
						t.Fatalf("Unexpected error, expected: %v, got: %v", bytes.ErrTooLarge, err)
					}
					return
				case err := <-chError:
					t.Fatalf("Unexpected error: %v", err)
				case <-time.After(time.Second):
					t.Fatal("Error is not emitted on write error")
				}
			}
			if !w.WaitWrite() {
				t.Fatal("Cluster is not written")
			}

			time.Sleep(50 * time.Millisecond)

			if errAt == atFrameWriting {
				w.setError(bytes.ErrTooLarge)
			}
			clearErr()
			if _, err := ws[0].Write(false, 110, []byte{0x01, 0x02}); err != nil {
				t.Fatalf("Failed to Write: %v", err)
			}
			if errAt == atFrameWriting {
				select {
				case err := <-chFatal:
					if err != bytes.ErrTooLarge {
						t.Fatalf("Unexpected error, expected: %v, got: %v", bytes.ErrTooLarge, err)
					}
					return
				case err := <-chError:
					t.Fatalf("Unexpected error: %v", err)
				case <-time.After(time.Second):
					t.Fatal("Error is not emitted on write error")
				}
			}
			if !w.WaitWrite() {
				t.Fatal("Second frame is not written")
			}

			// Very old frame
			clearErr()
			if _, err := ws[0].Write(true, -32769, []byte{0x0A}); err != nil {
				t.Fatalf("Failed to Write: %v", err)
			}
			select {
			case err := <-chError:
				if err != errIgnoreOldFrame {
					t.Errorf("Unexpected error, expected: %v, got: %v", errIgnoreOldFrame, err)
				}
			case err := <-chFatal:
				t.Fatalf("Unexpected fatal: %v", err)
			case <-time.After(time.Second):
				t.Fatal("Error is not emitted for old frame")
			}

			if errAt == atClosing {
				w.setError(bytes.ErrTooLarge)
			}
			clearErr()
			ws[0].Close()
			if errAt == atClosing {
				select {
				case err := <-chFatal:
					if err != bytes.ErrTooLarge {
						t.Fatalf("Unexpected error, expected: %v, got: %v", bytes.ErrTooLarge, err)
					}
					return
				case err := <-chError:
					t.Fatalf("Unexpected error: %v", err)
				case <-time.After(time.Second):
					t.Fatal("Error is not emitted on write error")
				}
			}
		})
	}
}

func TestBlockWriter_NewSimpleWriter(t *testing.T) {
	buf := &bufferCloser{closed: make(chan struct{})}

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
		WithEBMLHeader(nil),
		WithSegmentInfo(nil),
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
		0x16, 0x54, 0xAE, 0x6B, 0x8F,
		0xAE, 0x8D,
		0xD7, 0x81, 0x01,
		0x73, 0xC5, 0x81, 0x02,
		0x86, 0x81, 0x00,
		0x83, 0x81, 0x01,
		0x1F, 0x43, 0xB6, 0x75, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xE7, 0x81, 0x00,
	}
	if !bytes.Equal(buf.Bytes(), expectedBytes) {
		t.Errorf("Unexpected WebM binary,\nexpected: %+v\n     got: %+v", expectedBytes, buf.Bytes())
	}
}

func TestBlockWriter_WithMaxKeyframeInterval(t *testing.T) {
	buf := &bufferCloser{closed: make(chan struct{})}

	tracks := []TrackEntry{
		{
			TrackNumber: 1,
			TrackUID:    2,
			CodecID:     "",
			TrackType:   1,
		},
	}

	ws, err := NewSimpleBlockWriter(
		buf, tracks,
		WithEBMLHeader(nil),
		WithSegmentInfo(nil),
		WithMaxKeyframeInterval(1, 900*0x6FFF),
	)
	if err != nil {
		t.Fatalf("Failed to create BlockWriter: %v", err)
	}
	if len(ws) != 1 {
		t.Fatalf("Number of the returned writer must be 1, but got %d", len(ws))
	}

	for _, block := range []struct {
		keyframe bool
		timecode int64
		b        []byte
	}{
		{true, 0, []byte{0x01}},
		{false, 1, []byte{0x02}},
		{false, 0x1000, []byte{0x03}},
		{true, 0x1001, []byte{0x04}}, // This will be the head of the next cluster
	} {
		if _, err := ws[0].Write(block.keyframe, block.timecode, block.b); err != nil {
			t.Fatalf("Failed to Write: %v", err)
		}
	}

	ws[0].Close()

	expectedBytes := []byte{
		// Segment
		0x18, 0x53, 0x80, 0x67, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		// Tracks
		0x16, 0x54, 0xAE, 0x6B, 0x8F,
		0xAE, 0x8D,
		0xD7, 0x81, 0x01,
		0x73, 0xC5, 0x81, 0x02,
		0x86, 0x81, 0x00,
		0x83, 0x81, 0x01,
		// Cluster
		0x1F, 0x43, 0xB6, 0x75, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xE7, 0x81, 0x00,
		0xA3, 0x85, 0x81, 0x00, 0x00, 0x80, 0x01, // block 0
		0xA3, 0x85, 0x81, 0x00, 0x01, 0x00, 0x02, // block 1
		0xA3, 0x85, 0x81, 0x10, 0x00, 0x00, 0x03, // block 2
		// New cluster
		0x1F, 0x43, 0xB6, 0x75, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xE7, 0x82, 0x10, 0x01,
		0xAB, 0x81, 0x24,
		0xA3, 0x85, 0x81, 0x00, 0x00, 0x80, 0x04, // block 3
		// Finalization
		0x1F, 0x43, 0xB6, 0x75, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xE7, 0x82, 0x10, 0x01,
		0xAB, 0x81, 0x1A,
	}
	if !bytes.Equal(buf.Bytes(), expectedBytes) {
		t.Errorf("Unexpected WebM binary,\nexpected: %+v\n     got: %+v", expectedBytes, buf.Bytes())
	}
}

func TestBlockWriter_WithSeekHead(t *testing.T) {
	buf := &bufferCloser{closed: make(chan struct{})}

	tracks := []TrackEntry{
		{
			TrackNumber: 1,
			TrackUID:    2,
			CodecID:     "",
			TrackType:   1,
		},
	}

	ws, err := NewSimpleBlockWriter(
		buf, tracks,
		WithEBMLHeader(nil),
		WithSegmentInfo(&Info{TimecodeScale: 1000000}),
		WithSeekHead(),
	)
	if err != nil {
		t.Fatalf("Failed to create BlockWriter: %v", err)
	}
	if len(ws) != 1 {
		t.Fatalf("Number of the returned writer must be 1, but got %d", len(ws))
	}

	ws[0].Close()

	expectedBytes := []byte{
		// 1     2     3     4     5     6     7     8     9    10    11    12
		// Segment
		0x18, 0x53, 0x80, 0x67, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		// SeekHead
		0x11, 0x4D, 0x9B, 0x74, 0xBF,
		0x4D, 0xBB, 0x92,
		0x53, 0xAB, 0x84, 0x15, 0x49, 0xA9, 0x66, // Info
		0x53, 0xAC, 0x88, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x50,
		0x4D, 0xBB, 0x92,
		0x53, 0xAB, 0x84, 0x16, 0x54, 0xAE, 0x6B, // Tracks
		0x53, 0xAC, 0x88, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x5C,
		0x4D, 0xBB, 0x92,
		0x53, 0xAB, 0x84, 0x1F, 0x43, 0xB6, 0x75, // Cluster
		0x53, 0xAC, 0x88, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x70,
		// Info, pos: 80
		0x15, 0x49, 0xA9, 0x66, 0x87,
		0x2A, 0xD7, 0xB1, 0x83, 0x0F, 0x42, 0x40,
		// Tracks, pos: 92
		0x16, 0x54, 0xAE, 0x6B, 0x8F,
		0xAE, 0x8D,
		0xD7, 0x81, 0x01,
		0x73, 0xC5, 0x81, 0x02,
		0x86, 0x81, 0x00,
		0x83, 0x81, 0x01,
		// Cluster, pos: 112
		0x1F, 0x43, 0xB6, 0x75, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xE7, 0x81, 0x00,
	}
	if !bytes.Equal(buf.Bytes(), expectedBytes) {
		t.Errorf("Unexpected WebM binary,\nexpected: %+v\n     got: %+v", expectedBytes, buf.Bytes())
	}
}

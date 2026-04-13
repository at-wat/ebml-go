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
	"errors"
	"io"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/at-wat/ebml-go"
	"github.com/at-wat/ebml-go/internal/buffercloser"
	"github.com/at-wat/ebml-go/internal/errs"
)

func TestBlockWriter(t *testing.T) {
	buf := buffercloser.New()

	tracks := []TrackDescription{
		{TrackNumber: 1},
		{TrackNumber: 2},
	}

	blockSorter, err := NewMultiTrackBlockSorter(WithMaxDelayedPackets(10), WithSortRule(BlockSorterDropOutdated))
	if err != nil {
		t.Fatalf("Failed to create MultiTrackBlockSorter: %v", err)
	}

	ws, err := NewSimpleBlockWriter(buf, tracks,
		WithBlockInterceptor(blockSorter))
	if err != nil {
		t.Fatalf("Failed to create BlockWriter: '%v'", err)
	}

	if len(ws) != len(tracks) {
		t.Fatalf("Number of the returned writer (%d) must be same as the number of TrackEntry (%d)", len(ws), len(tracks))
	}

	if n, err := ws[1].Write(true, 110, []byte{0x03, 0x04, 0x05}); err != nil {
		t.Fatalf("Failed to Write: '%v'", err)
	} else if n != 3 {
		t.Errorf("Expected return value of BlockWriter.Write: 3, got: %d", n)
	}

	if n, err := ws[0].Write(false, 100, []byte{0x01, 0x02}); err != nil {
		t.Fatalf("Failed to Write: '%v'", err)
	} else if n != 2 {
		t.Errorf("Expected return value of BlockWriter.Write: 2, got: %d", n)
	}

	// Ignored due to old timestamp
	if n, err := ws[0].Write(true, -32769, []byte{0x0A}); err != nil {
		t.Fatalf("Failed to Write: '%v'", err)
	} else if n != 1 {
		t.Errorf("Expected return value of BlockWriter.Write: 1, got: %d", n)
	}

	if n, err := ws[0].Write(true, 130, []byte{0x06}); err != nil {
		t.Fatalf("Failed to Write: '%v'", err)
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
		Segment flexSegment `ebml:"Segment,size=unknown"`
	}{
		Segment: flexSegment{
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
	var result struct {
		Segment flexSegment `ebml:"Segment,size=unknown"`
	}
	if err := ebml.Unmarshal(bytes.NewReader(buf.Bytes()), &result); err != nil {
		t.Fatalf("Failed to Unmarshal resultant binary: '%v'", err)
	}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Unexpected data,\nexpected: %+v\n     got: %+v", expected, result)
	}
}

func TestBlockWriter_Options(t *testing.T) {
	buf := buffercloser.New()

	ws, err := NewSimpleBlockWriter(
		buf,
		[]TrackDescription{{TrackNumber: 1}},
		WithEBMLHeader(&struct {
			DocTypeVersion uint64 `ebml:"EBMLDocTypeVersion"`
		}{}),
		WithSegmentInfo(nil),
		WithMarshalOptions(ebml.WithDataSizeLen(2)),
		WithSeekHead(false),
	)
	if err != nil {
		t.Fatalf("Failed to create BlockWriter: '%v'", err)
	}

	if len(ws) != 1 {
		t.Fatalf("Number of the returned writer must be 1, got %d", len(ws))
	}
	ws[0].Close()

	expectedBytes := []byte{
		0x1A, 0x45, 0xDF, 0xA3, 0x40, 0x05,
		0x42, 0x87, 0x40, 0x01, 0x00,
		0x18, 0x53, 0x80, 0x67, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0x16, 0x54, 0xAE, 0x6B, 0x40, 0x00,
		0x1F, 0x43, 0xB6, 0x75, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xE7, 0x40, 0x01, 0x00,
	}
	if !bytes.Equal(buf.Bytes(), expectedBytes) {
		t.Errorf("Unexpected binary,\nexpected: %+v\n     got: %+v", expectedBytes, buf.Bytes())
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
				BlockWriterOptionFn(func(*BlockWriterOptions) error { return errDummy0 }),
			},
			err: errDummy0,
		},
		"MarshalOptionError": {
			opts: []BlockWriterOption{
				WithMarshalOptions(
					func(*ebml.MarshalOptions) error { return errDummy1 },
				),
				WithSeekHead(false),
			},
			err: errDummy1,
		},
		"MarshalOptionErrorWithSeekHead": {
			opts: []BlockWriterOption{
				WithMarshalOptions(
					func(*ebml.MarshalOptions) error {
						return errDummy1
					},
				),
			},
			err: errDummy1,
		},
		"MaxKeyframeIntervalOptionError": {
			opts: []BlockWriterOption{
				WithMaxKeyframeInterval(0, 0),
			},
			err: ErrInvalidTrackNumber,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			buf := buffercloser.New()
			_, err := NewSimpleBlockWriter(buf, []TrackDescription{}, c.opts...)
			if !errs.Is(err, c.err) {
				t.Errorf("Expected error: '%v', got: '%v'", c.err, err)
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
				w,
				[]TrackDescription{{TrackNumber: 1}},
				WithOnErrorHandler(func(err error) { chError <- err }),
				WithOnFatalHandler(func(err error) { chFatal <- err }),
			)
			if err != nil {
				if errAt == atBeginning {
					if !errs.Is(err, bytes.ErrTooLarge) {
						t.Fatalf("Expected error: '%v', got: '%v'", bytes.ErrTooLarge, err)
					}
					return
				}
				t.Fatalf("Failed to create SimpleWriter: '%v'", err)
			}

			if len(ws) != 1 {
				t.Fatalf("Number of the returned writer must be 1, got %d", len(ws))
			}

			if errAt == atClusterWriting {
				w.setError(bytes.ErrTooLarge)
			}
			clearErr()
			if _, err := ws[0].Write(false, 100, []byte{0x01, 0x02}); err != nil {
				t.Fatalf("Failed to Write: '%v'", err)
			}
			if errAt == atClusterWriting {
				select {
				case err := <-chFatal:
					if !errs.Is(err, bytes.ErrTooLarge) {
						t.Fatalf("Expected error: '%v', got: '%v'", bytes.ErrTooLarge, err)
					}
					return
				case err := <-chError:
					t.Fatalf("Unexpected error: '%v'", err)
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
				t.Fatalf("Failed to Write: '%v'", err)
			}
			if errAt == atFrameWriting {
				select {
				case err := <-chFatal:
					if !errs.Is(err, bytes.ErrTooLarge) {
						t.Fatalf("Expected error: '%v', got: '%v'", bytes.ErrTooLarge, err)
					}
					return
				case err := <-chError:
					t.Fatalf("Unexpected error: '%v'", err)
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
				t.Fatalf("Failed to Write: '%v'", err)
			}
			select {
			case err := <-chError:
				if !errs.Is(err, ErrIgnoreOldFrame) {
					t.Errorf("Expected error: '%v', got: '%v'", ErrIgnoreOldFrame, err)
				}
			case err := <-chFatal:
				t.Fatalf("Unexpected fatal: '%v'", err)
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
					if !errs.Is(err, bytes.ErrTooLarge) {
						t.Fatalf("Expected error: '%v', got: '%v'", bytes.ErrTooLarge, err)
					}
					return
				case err := <-chError:
					t.Fatalf("Unexpected error: '%v'", err)
				case <-time.After(time.Second):
					t.Fatal("Error is not emitted on write error")
				}
			}
		})
	}
}

func TestBlockWriter_WithMaxKeyframeInterval(t *testing.T) {
	buf := buffercloser.New()

	ws, err := NewSimpleBlockWriter(
		buf,
		[]TrackDescription{{TrackNumber: 1}},
		WithEBMLHeader(nil),
		WithSegmentInfo(nil),
		WithMaxKeyframeInterval(1, 900*0x6FFF),
		WithSeekHead(false),
	)
	if err != nil {
		t.Fatalf("Failed to create BlockWriter: '%v'", err)
	}
	if len(ws) != 1 {
		t.Fatalf("Number of the returned writer must be 1, got %d", len(ws))
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
			t.Fatalf("Failed to Write: '%v'", err)
		}
	}

	ws[0].Close()

	expectedBytes := []byte{
		// Segment
		0x18, 0x53, 0x80, 0x67, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		// Tracks
		0x16, 0x54, 0xAE, 0x6B, 0x80,
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
		t.Errorf("Unexpected binary,\nexpected: %+v\n     got: %+v", expectedBytes, buf.Bytes())
	}
}

func TestBlockWriter_WithSeekHead(t *testing.T) {
	t.Run("GenerateSeekHead", func(t *testing.T) {
		buf := buffercloser.New()

		ws, err := NewSimpleBlockWriter(
			buf,
			[]TrackDescription{{TrackNumber: 1}},
			WithEBMLHeader(nil),
			WithSegmentInfo(&struct {
				TimecodeScale uint64 `ebml:"TimecodeScale"`
			}{TimecodeScale: 1000000}),
			WithSeekHead(true),
		)
		if err != nil {
			t.Fatalf("Failed to create BlockWriter: '%v'", err)
		}
		if len(ws) != 1 {
			t.Fatalf("Number of the returned writer must be 1, got %d", len(ws))
		}

		ws[0].Close()

		expectedBytes := []byte{
			// 1     2     3     4     5     6     7     8     9    10    11    12
			// Segment
			0x18, 0x53, 0x80, 0x67, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			// SeekHead
			0x11, 0x4D, 0x9B, 0x74, 0xAA,
			0x4D, 0xBB, 0x92,
			0x53, 0xAB, 0x84, 0x15, 0x49, 0xA9, 0x66, // Info
			0x53, 0xAC, 0x88, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2F,
			0x4D, 0xBB, 0x92,
			0x53, 0xAB, 0x84, 0x16, 0x54, 0xAE, 0x6B, // Tracks
			0x53, 0xAC, 0x88, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x3B,
			// Info, pos: 47
			0x15, 0x49, 0xA9, 0x66, 0x87,
			0x2A, 0xD7, 0xB1, 0x83, 0x0F, 0x42, 0x40,
			// Tracks, pos: 59
			0x16, 0x54, 0xAE, 0x6B, 0x80,
			// Cluster
			0x1F, 0x43, 0xB6, 0x75, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xE7, 0x81, 0x00,
		}
		if !bytes.Equal(buf.Bytes(), expectedBytes) {
			t.Errorf("Unexpected binary,\nexpected: %+v\n     got: %+v", expectedBytes, buf.Bytes())
		}
	})
	t.Run("InvalidHeader", func(t *testing.T) {
		buf := buffercloser.New()

		_, err := NewSimpleBlockWriter(
			buf,
			[]TrackDescription{{TrackNumber: 1}},
			WithSegmentInfo(&struct {
				Invalid uint64 `ebml:"InvalidA"`
			}{}),
			WithSeekHead(true),
		)
		if !errs.Is(err, ebml.ErrUnknownElementName) {
			t.Errorf("Expected error: '%v', got: '%v'", ebml.ErrUnknownElementName, err)
		}
	})
}

func BenchmarkBlockWriter_InitFinalize(b *testing.B) {
	tracks := []TrackDescription{
		{TrackNumber: 1},
	}

	for i := 0; i < b.N; i++ {
		buf := buffercloser.New()

		blockSorter, err := NewMultiTrackBlockSorter(WithMaxDelayedPackets(10), WithSortRule(BlockSorterDropOutdated))
		if err != nil {
			b.Fatalf("Failed to create MultiTrackBlockSorter: %v", err)
		}

		ws, err := NewSimpleBlockWriter(buf, tracks,
			WithBlockInterceptor(blockSorter),
		)
		if err != nil {
			b.Fatalf("Failed to create BlockWriter: %v", err)
		}
		for _, w := range ws {
			w.Close()
		}
	}
}

func BenchmarkBlockWriter_SimpleBlock(b *testing.B) {
	tracks := []TrackDescription{
		{TrackNumber: 1},
	}

	buf := buffercloser.New()

	blockSorter, err := NewMultiTrackBlockSorter(WithMaxDelayedPackets(10), WithSortRule(BlockSorterDropOutdated))
	if err != nil {
		b.Fatalf("Failed to create MultiTrackBlockSorter: %v", err)
	}

	ws, err := NewSimpleBlockWriter(buf, tracks,
		WithBlockInterceptor(blockSorter),
	)
	if err != nil {
		b.Fatalf("Failed to create BlockWriter: %v", err)
	}

	data := []byte{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, w := range ws {
			if _, err := w.Write(true, int64(i*20), data); err != nil {
				b.Fatalf("Failed to Write: %v", err)
			}
		}
	}
	b.StopTimer()
	for _, w := range ws {
		w.Close()
	}
}

// seekableBuffer implements io.WriteCloser and io.WriteSeeker for testing.
type seekableBuffer struct {
	buf    []byte
	pos    int
	closed chan struct{}
}

func newSeekableBuffer() *seekableBuffer {
	return &seekableBuffer{
		closed: make(chan struct{}),
	}
}

func (b *seekableBuffer) Write(p []byte) (int, error) {
	end := b.pos + len(p)
	if end > len(b.buf) {
		b.buf = append(b.buf, make([]byte, end-len(b.buf))...)
	}
	copy(b.buf[b.pos:], p)
	b.pos = end
	return len(p), nil
}

func (b *seekableBuffer) Seek(offset int64, whence int) (int64, error) {
	var newPos int64
	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos = int64(b.pos) + offset
	case io.SeekEnd:
		newPos = int64(len(b.buf)) + offset
	}
	b.pos = int(newPos)
	return newPos, nil
}

func (b *seekableBuffer) Close() error {
	close(b.closed)
	return nil
}

func (b *seekableBuffer) Closed() <-chan struct{} {
	return b.closed
}

func (b *seekableBuffer) Bytes() []byte {
	return b.buf
}

// cuesTestSegment is the EBML segment structure used across Cues tests.
type cuesTestSegment struct {
	SeekHead *struct {
		Seek []struct {
			SeekID       []byte `ebml:"SeekID"`
			SeekPosition uint64 `ebml:"SeekPosition"`
		} `ebml:"Seek"`
	} `ebml:"SeekHead"`
	Info interface{} `ebml:"Info"`
	Cues *struct {
		CuePoint []struct {
			CueTime           uint64 `ebml:"CueTime"`
			CueTrackPositions []struct {
				CueTrack           uint64 `ebml:"CueTrack"`
				CueClusterPosition uint64 `ebml:"CueClusterPosition"`
			} `ebml:"CueTrackPositions"`
		} `ebml:"CuePoint"`
	} `ebml:"Cues"`
	Tracks  flexTracks           `ebml:"Tracks"`
	Cluster []simpleBlockCluster `ebml:"Cluster,size=unknown"`
}

func TestBlockWriter_WithCues(t *testing.T) {
	t.Run("ValidationRequiresSeekHead", func(t *testing.T) {
		buf := newSeekableBuffer()
		_, err := NewSimpleBlockWriter(
			buf,
			[]TrackDescription{{TrackNumber: 1}},
			WithEBMLHeader(nil),
			WithSegmentInfo(nil),
			WithCues(1024),
			// No WithSeekHead
		)
		if !errs.Is(err, ErrCuesRequiresSeekHead) {
			t.Errorf("Expected error: '%v', got: '%v'", ErrCuesRequiresSeekHead, err)
		}
	})

	t.Run("ValidationRequiresSeeker", func(t *testing.T) {
		buf := buffercloser.New() // not an io.WriteSeeker
		_, err := NewSimpleBlockWriter(
			buf,
			[]TrackDescription{{TrackNumber: 1}},
			WithEBMLHeader(nil),
			WithSegmentInfo(nil),
			WithSeekHead(true),
			WithCues(1024),
		)
		if !errs.Is(err, ErrCuesRequiresSeeker) {
			t.Errorf("Expected error: '%v', got: '%v'", ErrCuesRequiresSeeker, err)
		}
	})

	t.Run("CuesWritten", func(t *testing.T) {
		buf := newSeekableBuffer()
		ws, err := NewSimpleBlockWriter(
			buf,
			[]TrackDescription{{TrackNumber: 1}},
			WithEBMLHeader(nil),
			WithSegmentInfo(&struct {
				TimecodeScale uint64 `ebml:"TimecodeScale"`
			}{TimecodeScale: 1000000}),
			WithSeekHead(true),
			WithCues(4096),
		)
		if err != nil {
			t.Fatalf("Failed to create BlockWriter: '%v'", err)
		}
		if len(ws) != 1 {
			t.Fatalf("Number of the returned writer must be 1, got %d", len(ws))
		}

		// Write blocks that span multiple clusters
		// Each cluster boundary creates a CuePoint
		for i := 0; i < 5; i++ {
			if _, err := ws[0].Write(true, int64(i)*0x8000, []byte{0x01}); err != nil {
				t.Fatalf("Failed to Write: '%v'", err)
			}
		}

		ws[0].Close()
		<-buf.Closed()

		// Unmarshal and verify Cues are present
		var result struct {
			Segment cuesTestSegment `ebml:"Segment,size=unknown"`
		}
		if err := ebml.Unmarshal(bytes.NewReader(buf.Bytes()), &result); err != nil {
			t.Fatalf("Failed to Unmarshal: '%v'", err)
		}

		if result.Segment.Cues == nil {
			t.Fatal("Expected Cues to be present, got nil")
		}

		// Should have 5 CuePoints (one for each cluster)
		if got := len(result.Segment.Cues.CuePoint); got != 5 {
			t.Errorf("Expected 5 CuePoints, got %d", got)
		}

		// Verify CueTimes are increasing
		for i := 1; i < len(result.Segment.Cues.CuePoint); i++ {
			prev := result.Segment.Cues.CuePoint[i-1].CueTime
			curr := result.Segment.Cues.CuePoint[i].CueTime
			if curr <= prev {
				t.Errorf("CuePoint[%d].CueTime (%d) <= CuePoint[%d].CueTime (%d)", i, curr, i-1, prev)
			}
		}

		// Verify CueClusterPositions are increasing
		for i := 1; i < len(result.Segment.Cues.CuePoint); i++ {
			prev := result.Segment.Cues.CuePoint[i-1].CueTrackPositions[0].CueClusterPosition
			curr := result.Segment.Cues.CuePoint[i].CueTrackPositions[0].CueClusterPosition
			if curr <= prev {
				t.Errorf("CuePoint[%d].CueClusterPosition (%d) <= CuePoint[%d].CueClusterPosition (%d)", i, curr, i-1, prev)
			}
		}

		// Verify SeekHead contains Cues entry
		if result.Segment.SeekHead == nil {
			t.Fatal("Expected SeekHead to be present, got nil")
		}
		foundCues := false
		cuesID := ebml.ElementCues.Bytes()
		for _, seek := range result.Segment.SeekHead.Seek {
			if bytes.Equal(seek.SeekID, cuesID) {
				foundCues = true
				break
			}
		}
		if !foundCues {
			t.Error("SeekHead does not contain Cues entry")
		}
	})

	t.Run("CuesOverflow", func(t *testing.T) {
		buf := newSeekableBuffer()
		ws, err := NewSimpleBlockWriter(
			buf,
			[]TrackDescription{{TrackNumber: 1}},
			WithEBMLHeader(nil),
			WithSegmentInfo(nil),
			WithSeekHead(true),
			WithCues(32), // Tiny reserved space - will overflow
		)
		if err != nil {
			t.Fatalf("Failed to create BlockWriter: '%v'", err)
		}

		// Write blocks to create multiple clusters
		for i := 0; i < 5; i++ {
			if _, err := ws[0].Write(true, int64(i)*0x8000, []byte{0x01}); err != nil {
				t.Fatalf("Failed to Write: '%v'", err)
			}
		}
		ws[0].Close()
		<-buf.Closed()

		// File should still be valid, just without Cues
		var result struct {
			Segment struct {
				Tracks  flexTracks           `ebml:"Tracks"`
				Cluster []simpleBlockCluster `ebml:"Cluster,size=unknown"`
			} `ebml:"Segment,size=unknown"`
		}
		if err := ebml.Unmarshal(bytes.NewReader(buf.Bytes()), &result); err != nil {
			t.Fatalf("Failed to Unmarshal: '%v'", err)
		}
		// Should have clusters (file is valid)
		if len(result.Segment.Cluster) == 0 {
			t.Error("Expected clusters in output")
		}
	})

	t.Run("PositionVerification", func(t *testing.T) {
		buf := newSeekableBuffer()
		ws, err := NewSimpleBlockWriter(
			buf,
			[]TrackDescription{{TrackNumber: 1}},
			WithEBMLHeader(nil),
			WithSegmentInfo(&struct {
				TimecodeScale uint64 `ebml:"TimecodeScale"`
			}{TimecodeScale: 1000000}),
			WithSeekHead(true),
			WithCues(4096),
		)
		if err != nil {
			t.Fatalf("Failed to create BlockWriter: '%v'", err)
		}

		// Write 3 frames, each triggering a new cluster (timecodes far apart)
		for i := 0; i < 3; i++ {
			if _, err := ws[0].Write(true, int64(i)*0x8000, []byte{0x01}); err != nil {
				t.Fatalf("Failed to Write: '%v'", err)
			}
		}
		ws[0].Close()
		<-buf.Closed()

		data := buf.Bytes()

		// Find Segment data start by locating the Segment element ID
		segmentID := []byte{0x18, 0x53, 0x80, 0x67}
		segmentIdx := bytes.Index(data, segmentID)
		if segmentIdx < 0 {
			t.Fatal("Segment element not found")
		}
		// Segment uses unknown size: 4 byte ID + 8 byte size VINT = 12 byte header
		segmentDataStart := segmentIdx + 4 + 8

		// Unmarshal to get SeekHead and Cues data
		var result struct {
			Segment cuesTestSegment `ebml:"Segment,size=unknown"`
		}
		if err := ebml.Unmarshal(bytes.NewReader(data), &result); err != nil {
			t.Fatalf("Failed to Unmarshal: '%v'", err)
		}

		if result.Segment.Cues == nil {
			t.Fatal("Cues not found in output")
		}
		if result.Segment.SeekHead == nil {
			t.Fatal("SeekHead not found in output")
		}

		clusterElementID := []byte{0x1F, 0x43, 0xB6, 0x75}

		// SeekHead Cues position points to actual Cues element
		t.Run("SeekHeadCuesPosition", func(t *testing.T) {
			cuesElementID := []byte{0x1C, 0x53, 0xBB, 0x6B}
			var seekHeadCuesPos uint64
			for _, seek := range result.Segment.SeekHead.Seek {
				if bytes.Equal(seek.SeekID, ebml.ElementCues.Bytes()) {
					seekHeadCuesPos = seek.SeekPosition
					break
				}
			}
			cuesAbsolutePos := segmentDataStart + int(seekHeadCuesPos)

			if cuesAbsolutePos+4 > len(data) {
				t.Fatalf("Cues position %d is beyond file size %d", cuesAbsolutePos, len(data))
			}
			actualCuesID := data[cuesAbsolutePos : cuesAbsolutePos+4]
			if !bytes.Equal(actualCuesID, cuesElementID) {
				t.Errorf("SeekHead Cues position points to bytes %X, expected Cues element ID %X", actualCuesID, cuesElementID)
			}
		})

		// CueClusterPositions point to actual Cluster elements
		t.Run("CueClusterPositions", func(t *testing.T) {
			for i, cp := range result.Segment.Cues.CuePoint {
				if len(cp.CueTrackPositions) == 0 {
					t.Errorf("CuePoint[%d] has no CueTrackPositions", i)
					continue
				}
				clusterRelPos := cp.CueTrackPositions[0].CueClusterPosition
				clusterAbsPos := segmentDataStart + int(clusterRelPos)

				if clusterAbsPos+4 > len(data) {
					t.Errorf("CuePoint[%d] Cluster position %d is beyond file size %d", i, clusterAbsPos, len(data))
					continue
				}
				actualID := data[clusterAbsPos : clusterAbsPos+4]
				if !bytes.Equal(actualID, clusterElementID) {
					t.Errorf("CuePoint[%d] CueClusterPosition points to bytes %X, expected Cluster element ID %X",
						i, actualID, clusterElementID)
				}
			}
		})

		// Find all Cluster positions by scanning binary and compare
		t.Run("ClusterPositionCrossCheck", func(t *testing.T) {
			var actualClusterPositions []int
			for i := 0; i <= len(data)-4; i++ {
				if bytes.Equal(data[i:i+4], clusterElementID) {
					relPos := i - segmentDataStart
					actualClusterPositions = append(actualClusterPositions, relPos)
				}
			}

			// We should have 3 streaming clusters + 1 terminal cluster = 4 total
			// and 3 CuePoints (one for each streaming cluster)
			if len(result.Segment.Cues.CuePoint) != 3 {
				t.Errorf("Expected 3 CuePoints, got %d", len(result.Segment.Cues.CuePoint))
			}

			// Verify each CueClusterPosition matches a real cluster
			for i, cp := range result.Segment.Cues.CuePoint {
				clusterPos := int(cp.CueTrackPositions[0].CueClusterPosition)
				found := false
				for _, actual := range actualClusterPositions {
					if actual == clusterPos {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("CuePoint[%d] CueClusterPosition %d does not match any Cluster position. Actual positions: %v",
						i, clusterPos, actualClusterPositions)
				}
			}
		})
	})

	t.Run("WithEBMLHeader", func(t *testing.T) {
		// Verify positions are correct when a full EBML header is present
		// (matching real-world usage through the webm package)
		buf := newSeekableBuffer()
		ws, err := NewSimpleBlockWriter(
			buf,
			[]TrackDescription{{TrackNumber: 1}},
			WithEBMLHeader(&struct {
				EBMLVersion        uint64 `ebml:"EBMLVersion"`
				EBMLReadVersion    uint64 `ebml:"EBMLReadVersion"`
				EBMLMaxIDLength    uint64 `ebml:"EBMLMaxIDLength"`
				EBMLMaxSizeLength  uint64 `ebml:"EBMLMaxSizeLength"`
				DocType            string `ebml:"EBMLDocType"`
				DocTypeVersion     uint64 `ebml:"EBMLDocTypeVersion"`
				DocTypeReadVersion uint64 `ebml:"EBMLDocTypeReadVersion"`
			}{1, 1, 4, 8, "webm", 4, 2}),
			WithSegmentInfo(&struct {
				TimecodeScale uint64 `ebml:"TimecodeScale"`
			}{TimecodeScale: 1000000}),
			WithSeekHead(true),
			WithCues(4096),
		)
		if err != nil {
			t.Fatalf("Failed to create BlockWriter: '%v'", err)
		}

		for i := 0; i < 3; i++ {
			if _, err := ws[0].Write(true, int64(i)*0x8000, []byte{0x01}); err != nil {
				t.Fatalf("Failed to Write: '%v'", err)
			}
		}
		ws[0].Close()
		<-buf.Closed()

		data := buf.Bytes()

		// Find Segment data start
		segmentID := []byte{0x18, 0x53, 0x80, 0x67}
		segmentIdx := bytes.Index(data, segmentID)
		if segmentIdx < 0 {
			t.Fatal("Segment element not found")
		}
		segmentDataStart := segmentIdx + 4 + 8

		// Verify SeekHead Cues position points to Cues element
		cuesElementID := []byte{0x1C, 0x53, 0xBB, 0x6B}
		clusterElementID := []byte{0x1F, 0x43, 0xB6, 0x75}

		var result struct {
			Header  interface{}     `ebml:"EBML"`
			Segment cuesTestSegment `ebml:"Segment,size=unknown"`
		}
		if err := ebml.Unmarshal(bytes.NewReader(data), &result); err != nil {
			t.Fatalf("Failed to Unmarshal: '%v'", err)
		}

		if result.Segment.Cues == nil {
			t.Fatal("Cues not found")
		}

		// Verify SeekHead Cues position
		var seekHeadCuesPos uint64
		for _, seek := range result.Segment.SeekHead.Seek {
			if bytes.Equal(seek.SeekID, ebml.ElementCues.Bytes()) {
				seekHeadCuesPos = seek.SeekPosition
				break
			}
		}
		cuesAbsPos := segmentDataStart + int(seekHeadCuesPos)
		if cuesAbsPos+4 > len(data) || !bytes.Equal(data[cuesAbsPos:cuesAbsPos+4], cuesElementID) {
			t.Errorf("SeekHead Cues position %d (abs %d) doesn't point to Cues element", seekHeadCuesPos, cuesAbsPos)
		}

		// Verify CueClusterPositions point to Cluster elements
		for i, cp := range result.Segment.Cues.CuePoint {
			clusterAbsPos := segmentDataStart + int(cp.CueTrackPositions[0].CueClusterPosition)
			if clusterAbsPos+4 > len(data) || !bytes.Equal(data[clusterAbsPos:clusterAbsPos+4], clusterElementID) {
				t.Errorf("CuePoint[%d] position %d (abs %d) doesn't point to Cluster",
					i, cp.CueTrackPositions[0].CueClusterPosition, clusterAbsPos)
			}
		}
	})
}

func TestWriteVoidElement(t *testing.T) {
	for _, size := range []int{9, 100, 4096, 51200} {
		var buf bytes.Buffer
		if err := writeVoidElement(&buf, size); err != nil {
			t.Fatalf("writeVoidElement(%d) failed: %v", size, err)
		}
		if buf.Len() != size {
			t.Errorf("writeVoidElement(%d): got %d bytes, want %d", size, buf.Len(), size)
		}
		b := buf.Bytes()
		if b[0] != 0xEC {
			t.Errorf("writeVoidElement(%d): first byte = 0x%02X, want 0xEC", size, b[0])
		}
		if b[1] != 0x01 {
			t.Errorf("writeVoidElement(%d): VINT marker = 0x%02X, want 0x01 (8-byte VINT)", size, b[1])
		}
	}
}

// durationInfo is a test segmentInfo that implements durationSettable.
type durationInfo struct {
	TimecodeScale uint64  `ebml:"TimecodeScale"`
	Duration      float64 `ebml:"Duration,omitempty"`
}

func (i *durationInfo) SetDuration(d float64) { i.Duration = d }

func TestBlockWriter_Duration(t *testing.T) {
	t.Run("DurationWritten", func(t *testing.T) {
		buf := newSeekableBuffer()
		ws, err := NewSimpleBlockWriter(
			buf,
			[]TrackDescription{{TrackNumber: 1}},
			WithEBMLHeader(nil),
			WithSegmentInfo(&durationInfo{TimecodeScale: 1000000}),
			WithSeekHead(true),
			WithCues(4096),
		)
		if err != nil {
			t.Fatalf("Failed to create BlockWriter: '%v'", err)
		}

		// Write blocks spanning multiple clusters.
		// tc0=0, lastTc=3*0x8000=98304; expected Duration=98304.0
		for i := 0; i < 4; i++ {
			if _, err := ws[0].Write(true, int64(i)*0x8000, []byte{0x01}); err != nil {
				t.Fatalf("Failed to Write: '%v'", err)
			}
		}
		ws[0].Close()
		<-buf.Closed()

		data := buf.Bytes()

		// Unmarshal and verify Duration via EBML
		var result struct {
			Segment struct {
				Info struct {
					Duration float64 `ebml:"Duration"`
				} `ebml:"Info"`
			} `ebml:"Segment,size=unknown"`
		}
		if err := ebml.Unmarshal(bytes.NewReader(data), &result); err != nil {
			t.Fatalf("Failed to Unmarshal: '%v'", err)
		}
		expected := float64(3 * 0x8000)
		if result.Segment.Info.Duration != expected {
			t.Errorf("Expected Duration %v, got %v", expected, result.Segment.Info.Duration)
		}

	})

	t.Run("WithoutSettable", func(t *testing.T) {
		// segmentInfo without SetDuration method — no Duration element should appear
		buf := newSeekableBuffer()
		ws, err := NewSimpleBlockWriter(
			buf,
			[]TrackDescription{{TrackNumber: 1}},
			WithEBMLHeader(nil),
			WithSegmentInfo(&struct {
				TimecodeScale uint64 `ebml:"TimecodeScale"`
			}{TimecodeScale: 1000000}),
			WithSeekHead(true),
			WithCues(4096),
		)
		if err != nil {
			t.Fatalf("Failed to create BlockWriter: '%v'", err)
		}

		for i := 0; i < 3; i++ {
			if _, err := ws[0].Write(true, int64(i)*0x8000, []byte{0x01}); err != nil {
				t.Fatalf("Failed to Write: '%v'", err)
			}
		}
		ws[0].Close()
		<-buf.Closed()

		// Duration element with 8-byte VINT should not be present
		durationWithVINT := []byte{0x44, 0x89, 0x88}
		if bytes.Contains(buf.Bytes(), durationWithVINT) {
			t.Error("Duration element should not be present when segmentInfo doesn't implement durationSettable")
		}
	})

	t.Run("NoFrames", func(t *testing.T) {
		buf := newSeekableBuffer()
		ws, err := NewSimpleBlockWriter(
			buf,
			[]TrackDescription{{TrackNumber: 1}},
			WithEBMLHeader(nil),
			WithSegmentInfo(&durationInfo{TimecodeScale: 1000000}),
			WithSeekHead(true),
			WithCues(4096),
		)
		if err != nil {
			t.Fatalf("Failed to create BlockWriter: '%v'", err)
		}

		// Close immediately without writing any frames
		ws[0].Close()
		<-buf.Closed()

		data := buf.Bytes()
		var result struct {
			Segment struct {
				Info struct {
					Duration float64 `ebml:"Duration"`
				} `ebml:"Info"`
			} `ebml:"Segment,size=unknown"`
		}
		if err := ebml.Unmarshal(bytes.NewReader(data), &result); err != nil {
			t.Fatalf("Failed to Unmarshal: '%v'", err)
		}
		if result.Segment.Info.Duration != 0.0 {
			t.Errorf("Expected Duration 0.0 when no frames written, got %v", result.Segment.Info.Duration)
		}
	})
}

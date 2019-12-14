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
	"io"
	"sync"
)

// BlockMuxer is a interface of WebM block stream muxer.
type BlockMuxer interface {
	// Filter reads blocks of each track, filters, and writes.
	Filter(r []BlockReader, w []BlockWriter)
}

type filterWriter struct {
	trackNumber uint64
	ch          chan *frame
}

type filterReader struct {
	ch chan *frame
}

func (w *filterWriter) Write(keyframe bool, timestamp int64, b []byte) (int, error) {
	w.ch <- &frame{
		trackNumber: w.trackNumber,
		keyframe:    keyframe,
		timestamp:   timestamp,
		b:           b,
	}
	return len(b), nil
}

func (r *filterReader) Read() ([]byte, bool, int64, error) {
	select {
	case frame, ok := <-r.ch:
		if !ok {
			return nil, false, 0, io.EOF
		}
		return frame.b, frame.keyframe, frame.timestamp, nil
	}
}

func (r *filterReader) close() {
	close(r.ch)
}

// NewMultiTrackBlockSorter creates BlockMuxer, which sorts blocks on multiple tracks by timestamp.
// The index of TrackEntry sorts blocks with the same timestamp.
// Place the audio track before the video track to meet WebM Muxer Guidelines.
func NewMultiTrackBlockSorter(maxDelay int) BlockMuxer {
	return &multiTrackBlockSorter{maxDelay: maxDelay}
}

type multiTrackBlockSorter struct {
	maxDelay int
}

func (s *multiTrackBlockSorter) Filter(r []BlockReader, w []BlockWriter) {
	var wg sync.WaitGroup
	wg.Add(len(r))

	ch := make(chan *frame)
	for i, r := range r {
		go func(i int, r BlockReader) {
			for {
				var err error
				f := &frame{trackNumber: uint64(i)}
				if f.b, f.keyframe, f.timestamp, err = r.Read(); err != nil {
					wg.Done()
					return
				}
				ch <- f
			}
		}(i, r)
	}

	closed := make(chan struct{})
	go func() {
		wg.Wait()
		close(closed)
	}()

	var tDone int64
	buf := make([]*frameBuffer, len(r))
	for i := range buf {
		buf[i] = &frameBuffer{}
	}

	flush := func(all bool) {
		nChReq := 1
		if !all {
			nChReq = len(r)
		}
		for {
			var tOldest int64
			var nCh, nMax int
			var bOldest *frameBuffer
			for _, b := range buf {
				if n := b.Size(); n > 0 {
					nCh++
					if f := b.Head(); f.timestamp < tOldest || bOldest == nil {
						tOldest = f.timestamp
						bOldest = b
					}
					if n > nMax {
						nMax = n
					}
				}
			}
			if nCh >= nChReq || nMax > s.maxDelay {
				fOldest := bOldest.Pop()
				w[fOldest.trackNumber].Write(fOldest.keyframe, fOldest.timestamp, fOldest.b)
				tDone = fOldest.timestamp
			} else {
				break
			}
		}
	}

	for {
		select {
		case d := <-ch:
			if d.timestamp >= tDone {
				buf[d.trackNumber].Push(d)
				flush(false)
			}
		case <-closed:
			flush(true)
			return
		}
	}
}

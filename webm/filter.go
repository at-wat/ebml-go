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
	Filter(r []BlockReader, w []BlockWriter) error
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

// NewMultiTrackBlockSorter create BlockMuxer which sorts blocks on multiple tracks by timestamp.
func NewMultiTrackBlockSorter(maxDelay int) BlockMuxer {
	return &multiTrackBlockSorter{maxDelay: maxDelay}
}

type multiTrackBlockSorter struct {
	maxDelay int
}

func (s *multiTrackBlockSorter) Filter(r []BlockReader, w []BlockWriter) error {
	var wg sync.WaitGroup

	type frameWithTrackID struct {
		frame *frame
		i     int
	}
	ch := make(chan frameWithTrackID)
	for i, r := range r {
		wg.Add(1)
		go func(i int, r BlockReader) {
			for {
				// Read one
				f := &frame{}
				var err error
				if f.b, f.keyframe, f.timestamp, err = r.Read(); err != nil {
					wg.Done()
					return
				}
				ch <- frameWithTrackID{f, i}
			}
		}(i, r)
	}

	closed := make(chan struct{})
	go func() {
		wg.Wait()
		close(closed)
	}()

	var tDone int64
	buf := make([][]frameWithTrackID, len(r))

	flush := func(all bool) {
		var nChReq int
		if !all {
			nChReq = len(r)
		}
		for {
			var tOldest int64
			var nCh int
			iOldest := -1
			for _, b := range buf {
				if len(b) > 0 {
					nCh++
					if b[0].frame.timestamp < tOldest || iOldest < 0 {
						tOldest = b[0].frame.timestamp
						iOldest = b[0].i
					}
				}
			}
			if iOldest < 0 {
				break
			}
			nOldest := len(buf[iOldest])
			if nCh >= nChReq || nOldest > s.maxDelay {
				d := buf[iOldest][0]
				buf[iOldest][0].frame = nil // for quick GC
				if nOldest == 1 {
					buf[iOldest] = nil
				} else {
					buf[iOldest] = buf[iOldest][1:]
				}
				w[iOldest].Write(d.frame.keyframe, d.frame.timestamp, d.frame.b)
				tDone = d.frame.timestamp
			} else {
				break
			}
		}
	}

	for {
		select {
		case d := <-ch:
			if d.frame.timestamp >= tDone {
				buf[d.i] = append(buf[d.i], d)
				flush(false)
			}
		case <-closed:
			flush(true)
			return nil
		}
	}
}

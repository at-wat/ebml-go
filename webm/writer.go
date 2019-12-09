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
	"errors"
	"io"
	"sync"

	"github.com/at-wat/ebml-go"
)

var (
	// DefaultEBMLHeader is the default EBML header and is used by NewSimpleWriter.
	DefaultEBMLHeader = &EBMLHeader{
		EBMLVersion:        1,
		EBMLReadVersion:    1,
		EBMLMaxIDLength:    4,
		EBMLMaxSizeLength:  8,
		DocType:            "webm",
		DocTypeVersion:     2,
		DocTypeReadVersion: 2,
	}
	// DefaultSegmentInfo is the default Segment.Info and is used by NewSimpleWriter.
	DefaultSegmentInfo = &Info{
		TimecodeScale: 1000000, // 1ms
		MuxingApp:     "ebml-go.webm.SimpleWriter",
		WritingApp:    "ebml-go.webm.SimpleWriter",
	}
)

var (
	errIgnoreOldFrame = errors.New("too old frame")
)

// NewSimpleWriter creates FrameWriter for each track specified as tracks argument.
// Resultant WebM is written to given io.WriteCloser.
// io.WriteCloser will be closed automatically; don't close it by yourself.
func NewSimpleWriter(w0 io.WriteCloser, tracks []TrackEntry, opts ...SimpleWriterOption) ([]*FrameWriter, error) {
	options := &SimpleWriterOptions{
		ebmlHeader:  DefaultEBMLHeader,
		segmentInfo: DefaultSegmentInfo,
		onFatal: func(err error) {
			panic(err)
		},
	}
	for _, o := range opts {
		if err := o(options); err != nil {
			return nil, err
		}
	}

	w := &writerWithSizeCount{w: w0}

	type FlexSegment struct {
		SeekHead interface{} `ebml:"SeekHead,omitempty"`
		Info     interface{} `ebml:"Info"`
		Tracks   Tracks      `ebml:"Tracks"`
		Cluster  []Cluster   `ebml:"Cluster"`
	}

	header := struct {
		Header  interface{} `ebml:"EBML"`
		Segment FlexSegment `ebml:"Segment,size=unknown"`
	}{
		Header: options.ebmlHeader,
		Segment: FlexSegment{
			SeekHead: options.seekHead,
			Info:     options.segmentInfo,
			Tracks: Tracks{
				TrackEntry: tracks,
			},
		},
	}
	if err := ebml.Marshal(&header, w, options.marshalOpts...); err != nil {
		return nil, err
	}

	w.Clear()

	ch := make(chan *frame)
	fin := make(chan struct{}, len(tracks)-1)
	wg := sync.WaitGroup{}
	var ws []*FrameWriter

	for _, t := range tracks {
		wg.Add(1)
		ws = append(ws, &FrameWriter{
			trackNumber: t.TrackNumber,
			f:           ch,
			wg:          &wg,
			fin:         fin,
		})
	}

	closed := make(chan struct{})
	go func() {
		wg.Wait()
		close(closed)
	}()

	go func() {
		const invalidTimestamp = int64(0x7FFFFFFFFFFFFFFF)
		tc0 := invalidTimestamp
		tc1 := invalidTimestamp
		lastTc := int64(0)

		defer func() {
			// Finalize WebM
			if tc0 == invalidTimestamp {
				// No data written
				tc0 = 0
			}
			cluster := struct {
				Cluster Cluster `ebml:"Cluster,size=unknown"`
			}{
				Cluster: Cluster{
					Timecode: uint64(lastTc - tc0),
					PrevSize: uint64(w.Size()),
				},
			}
			if err := ebml.Marshal(&cluster, w, options.marshalOpts...); err != nil {
				if options.onFatal != nil {
					options.onFatal(err)
				}
			}
			w.Close()
			<-fin // read one data to release blocked Close()
		}()

	L_WRITE:
		for {
			select {
			case <-closed:
				break L_WRITE
			case f := <-ch:
				if tc0 == invalidTimestamp {
					tc0 = f.timestamp
				}
				lastTc = f.timestamp
				tc := f.timestamp - tc1
				if tc >= 0x7FFF || tc1 == invalidTimestamp {
					// Create new Cluster
					tc1 = f.timestamp
					tc = 0

					cluster := struct {
						Cluster Cluster `ebml:"Cluster,size=unknown"`
					}{
						Cluster: Cluster{
							Timecode: uint64(tc1 - tc0),
							PrevSize: uint64(w.Size()),
						},
					}
					w.Clear()
					if err := ebml.Marshal(&cluster, w, options.marshalOpts...); err != nil {
						if options.onFatal != nil {
							options.onFatal(err)
						}
						return
					}
				}
				if tc <= -0x7FFF {
					// Ignore too old frame
					if options.onError != nil {
						options.onError(errIgnoreOldFrame)
					}
					continue
				}

				b := struct {
					Block ebml.Block `ebml:"SimpleBlock"`
				}{
					ebml.Block{
						TrackNumber: f.trackNumber,
						Timecode:    int16(tc),
						Keyframe:    f.keyframe,
						Data:        [][]byte{f.b},
					},
				}
				// Write SimpleBlock to the file
				if err := ebml.Marshal(&b, w, options.marshalOpts...); err != nil {
					options.onError(err)
					return
				}
			}
		}
	}()

	return ws, nil
}

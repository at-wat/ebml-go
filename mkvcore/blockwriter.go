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
	"errors"
	"io"
	"sync"

	"github.com/at-wat/ebml-go"
)

// ErrIgnoreOldFrame means that a frame has too old timestamp and ignored.
var ErrIgnoreOldFrame = errors.New("too old frame")

type blockWriter struct {
	trackNumber uint64
	f           chan *frame
	wg          *sync.WaitGroup
	fin         chan struct{}
}

type frame struct {
	trackNumber uint64
	keyframe    bool
	timestamp   int64
	b           []byte
}

func (w *blockWriter) Write(keyframe bool, timestamp int64, b []byte) (int, error) {
	w.f <- &frame{
		trackNumber: w.trackNumber,
		keyframe:    keyframe,
		timestamp:   timestamp,
		b:           b,
	}
	return len(b), nil
}

func (w *blockWriter) Close() error {
	w.wg.Done()

	// If it is the last writer, block until closing output writer.
	w.fin <- struct{}{}

	return nil
}

// TrackDescription stores track number and its TrackEntry struct.
type TrackDescription struct {
	TrackNumber uint64
	TrackEntry  interface{}
}

// NewSimpleBlockWriter creates BlockWriteCloser for each track specified as tracks argument.
// Blocks will be written to the writer as EBML SimpleBlocks.
// Given io.WriteCloser will be closed automatically; don't close it by yourself.
// Frames written to each track must be sorted by their timestamp.
func NewSimpleBlockWriter(w0 io.WriteCloser, tracks []TrackDescription, opts ...BlockWriterOption) ([]BlockWriteCloser, error) {
	options := &BlockWriterOptions{
		BlockReadWriterOptions: BlockReadWriterOptions{
			onFatal: func(err error) {
				panic(err)
			},
		},
		ebmlHeader:  nil,
		segmentInfo: nil,
		interceptor: nil,
		seekHead:    false,
	}
	for _, o := range opts {
		if err := o.ApplyToBlockWriterOptions(options); err != nil {
			return nil, err
		}
	}

	w := &writerWithSizeCount{w: w0}

	header := flexHeader{
		Header: options.ebmlHeader,
		Segment: flexSegment{
			Info: options.segmentInfo,
		},
	}
	for _, t := range tracks {
		header.Segment.Tracks.TrackEntry = append(header.Segment.Tracks.TrackEntry, t.TrackEntry)
	}
	if options.seekHead {
		if err := setSeekHead(&header, options.marshalOpts...); err != nil {
			return nil, err
		}
	}
	if err := ebml.Marshal(&header, w, options.marshalOpts...); err != nil {
		return nil, err
	}

	w.Clear()

	ch := make(chan *frame)
	fin := make(chan struct{}, len(tracks)-1)
	wg := sync.WaitGroup{}
	var ws []BlockWriteCloser
	var fw []BlockWriter
	var fr []BlockReader

	for _, t := range tracks {
		wg.Add(1)
		var chSrc chan *frame
		if options.interceptor == nil {
			chSrc = ch
		} else {
			chSrc = make(chan *frame)
			fr = append(fr, &filterReader{chSrc})
			fw = append(fw, &filterWriter{t.TrackNumber, ch})
		}
		ws = append(ws, &blockWriter{
			trackNumber: t.TrackNumber,
			f:           chSrc,
			wg:          &wg,
			fin:         fin,
		})
	}

	filterFlushed := make(chan struct{})
	if options.interceptor != nil {
		go func() {
			options.interceptor.Intercept(fr, fw)
			close(filterFlushed)
		}()
	} else {
		close(filterFlushed)
	}

	closed := make(chan struct{})
	go func() {
		wg.Wait()
		for _, c := range fr {
			c.(*filterReader).close()
		}
		<-filterFlushed
		close(closed)
	}()

	tNextCluster := 0x7FFF - options.maxKeyframeInterval

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
				Cluster simpleBlockCluster `ebml:"Cluster,size=unknown"`
			}{
				Cluster: simpleBlockCluster{
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
				if tc1 == invalidTimestamp || tc >= 0x7FFF || (f.trackNumber == options.mainTrackNumber && tc >= tNextCluster && f.keyframe) {
					// Create new Cluster
					tc1 = f.timestamp
					tc = 0

					cluster := struct {
						Cluster simpleBlockCluster `ebml:"Cluster,size=unknown"`
					}{
						Cluster: simpleBlockCluster{
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
						options.onError(ErrIgnoreOldFrame)
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
					if options.onFatal != nil {
						options.onFatal(err)
					}
					return
				}
			}
		}
	}()

	return ws, nil
}

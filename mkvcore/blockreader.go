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
	"fmt"
	"io"
	"reflect"
	"sync"

	"github.com/at-wat/ebml-go"
)

type ReadableBlock struct {
	Header      interface{}
	trackNumber uint64
	f           chan *frame
	wg          *sync.WaitGroup
}

func (r *ReadableBlock) Read() (b []byte, keyframe bool, timestamp int64, err error) {
	frame := <-r.f
	return frame.b, frame.keyframe, frame.timestamp, nil
}

func (r *ReadableBlock) Close() error {
	r.wg.Done()

	return nil
}

// NewSimpleBlockReader creates BlockReadCloser for each track specified as tracks argument.
func NewSimpleBlockReader(r0 io.Reader, options BlockReaderOptions) ([]ReadableBlock, error) {
	if options.onFatal == nil {
		options.onFatal = func(err error) {
			panic(err)
		}
	}

	r := &readerWithSizeCount{r: r0}

	var header flexHeader
	if err := ebml.Unmarshal(r, &header, options.unmarshalOpts...); err != nil {
		return nil, err
	}

	r.Clear()

	ch := make(chan *frame)
	wg := sync.WaitGroup{}
	var ws []ReadableBlock
	var fw []BlockWriter
	var fr []BlockReader

	fmt.Printf("track entry %+v\n", header)
	for _, t := range header.Segment.Tracks.TrackEntry {
		wg.Add(1)
		var chSrc chan *frame
		chSrc = make(chan *frame)
		fr = append(fr, &filterReader{chSrc})
		trackNumber := reflect.ValueOf(t).FieldByName("TrackNumber").Uint()
		fw = append(fw, &filterWriter{trackNumber, ch})
		ws = append(ws, ReadableBlock{
			Header:      t,
			trackNumber: trackNumber,
			f:           chSrc,
			wg:          &wg,
		})
	}

	closed := make(chan struct{})
	go func() {
		wg.Wait()
		for _, c := range fr {
			c.(*filterReader).close()
		}
		close(closed)
	}()

	go func() {
		for {
			var b struct {
				Block ebml.Block `ebml:"SimpleBlock"`
			}
			// Read SimpleBlock from the file
			err := ebml.Unmarshal(r, &b, options.unmarshalOpts...)
			if err == nil {
				fmt.Printf("block %+v\n", b)
				frame := &frame{
					trackNumber: b.Block.TrackNumber,
					keyframe:    b.Block.Keyframe,
					// This is not exactly the original timestamp, but the encoding process loses the initial offset.
					timestamp: int64(b.Block.Timecode),
					b:         b.Block.Data[0],
				}
				for _, track := range ws {
					if track.trackNumber == b.Block.TrackNumber {
						track.f <- frame
					}
				}
				continue
			}
			// Maybe it's a cluster.
			var cluster struct {
				Cluster simpleBlockCluster `ebml:"Cluster,size=unknown"`
			}
			err = ebml.Unmarshal(r, &cluster, options.unmarshalOpts...)
			if err == nil {
				// Ignore this, no useful information.
				continue
			}
			// Give up.
			if options.onFatal != nil {
				options.onFatal(err)
			}
			return
		}
	}()

	return ws, nil
}

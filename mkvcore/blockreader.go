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
	"io"

	"github.com/at-wat/ebml-go"
)

type blockReader struct {
	f      chan *frame
	closed chan struct{}
}

func (r *blockReader) Read() (b []byte, keyframe bool, timestamp int64, err error) {
	frame, ok := <-r.f
	if !ok {
		return nil, false, 0, io.EOF
	}
	return frame.b, frame.keyframe, frame.timestamp, nil
}

func (r *blockReader) Close() error {
	close(r.closed)
	return nil
}

// NewSimpleBlockReader creates BlockReadCloser for each track specified as tracks argument.
func NewSimpleBlockReader(r io.Reader, opts ...BlockReaderOption) ([]BlockReadCloser, error) {
	options := &BlockReaderOptions{
		BlockReadWriterOptions: BlockReadWriterOptions{
			onFatal: func(err error) {
				panic(err)
			},
		},
	}
	for _, o := range opts {
		if err := o.ApplyToBlockReaderOptions(options); err != nil {
			return nil, err
		}
	}

	var header struct {
		Segment struct {
			Tracks struct {
				TrackEntry []struct {
					TrackNumber uint64
				}
			} `ebml:"Tracks,stop"`
		}
	}
	switch err := ebml.Unmarshal(r, &header, options.unmarshalOpts...); err {
	case ebml.ErrReadStopped:
	default:
		return nil, err
	}

	var ws []BlockReadCloser
	br := make(map[uint64]*blockReader)

	for _, t := range header.Segment.Tracks.TrackEntry {
		r := &blockReader{
			f:      make(chan *frame),
			closed: make(chan struct{}),
		}
		ws = append(ws, r)
		br[t.TrackNumber] = r
	}

	type blockGroup struct {
		Block             ebml.Block
		ReferencePriority uint64
	}
	type clusterReader struct {
		Timecode    uint64
		SimpleBlock chan ebml.Block
		BlockGroup  chan blockGroup
	}
	blockCh := make(chan ebml.Block)
	blockGroupCh := make(chan blockGroup)
	c := struct {
		Cluster clusterReader
	}{
		Cluster: clusterReader{
			SimpleBlock: blockCh,
			BlockGroup:  blockGroupCh,
		},
	}
	go func() {
		blockCh := blockCh
		blockGroupCh := blockGroupCh
	L_READ:
		for {
			var b *ebml.Block
			select {
			case block, ok := <-blockCh:
				if !ok {
					blockCh = nil
					if blockGroupCh == nil {
						break L_READ
					}
					continue
				}
				b = &block
			case bg, ok := <-blockGroupCh:
				if !ok {
					blockGroupCh = nil
					if blockCh == nil {
						break L_READ
					}
					continue
				}
				b = &bg.Block
				b.Keyframe = bg.ReferencePriority != 0
			}
			r := br[b.TrackNumber]
			for l := range b.Data {
				frame := &frame{
					trackNumber: b.TrackNumber,
					keyframe:    b.Keyframe,
					timestamp:   int64(c.Cluster.Timecode) + int64(b.Timecode),
					b:           b.Data[l],
				}
				select {
				case r.f <- frame:
				case <-r.closed:
				}
			}
		}
		for k := range br {
			close(br[k].f)
		}
	}()
	go func() {
		defer func() {
			close(blockCh)
			close(blockGroupCh)
		}()
		if err := ebml.Unmarshal(r, &c, options.unmarshalOpts...); err != nil {
			if options.onFatal != nil {
				options.onFatal(err)
			}
		}
	}()

	return ws, nil
}

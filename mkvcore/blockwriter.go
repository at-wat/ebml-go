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
	"encoding/binary"
	"errors"
	"io"
	"math"
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

	// Validate Cues requirements
	var seeker io.WriteSeeker
	if options.cuesReservedSize > 0 {
		if !options.seekHead {
			return nil, ErrCuesRequiresSeekHead
		}
		var ok bool
		seeker, ok = w0.(io.WriteSeeker)
		if !ok {
			return nil, ErrCuesRequiresSeeker
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

	// When Cues are enabled and segmentInfo supports it, set a placeholder
	// Duration so the marshaler includes the element in the output.
	// We'll overwrite it with the real value at finalization.
	if options.cuesReservedSize > 0 {
		if ds, ok := options.segmentInfo.(durationSettable); ok {
			ds.SetDuration(math.SmallestNonzeroFloat64)
		}
	}

	var durationElementPos uint64
	var segmentDataStart uint64
	if options.seekHead {
		var err error
		segmentDataStart, durationElementPos, err = setSeekHead(&header, options.cuesReservedSize > 0, options.marshalOpts...)
		if err != nil {
			return nil, err
		}
	}
	if err := ebml.Marshal(&header, w, options.marshalOpts...); err != nil {
		return nil, err
	}

	// Reserve space for Cues by writing a Void element
	var cuesReservedStart uint64
	var posAfterHeader uint64
	if options.cuesReservedSize > 0 {
		cuesReservedStart = uint64(w.Size())
		if err := writeVoidElement(w, options.cuesReservedSize); err != nil {
			return nil, err
		}
	}
	posAfterHeader = uint64(w.Size())

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

		// Cues tracking state
		runningPos := posAfterHeader
		var cuePoints []cuePoint

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

			// Write Cues to the reserved space
			if options.cuesReservedSize > 0 && len(cuePoints) > 0 {
				writeCuesToReserved(
					seeker, cuePoints, options.marshalOpts,
					cuesReservedStart, options.cuesReservedSize,
					options.onFatal,
				)
			}

			// Overwrite the placeholder Duration with the real value.
			// Duration element layout: 2-byte ID (0x44 0x89) + 1-byte VINT (0x88) + 8-byte float64.
			if durationElementPos > 0 && seeker != nil {
				duration := float64(lastTc - tc0)
				var buf [8]byte
				binary.BigEndian.PutUint64(buf[:], math.Float64bits(duration))
				if _, err := seeker.Seek(int64(durationElementPos)+3, io.SeekStart); err != nil {
					if options.onFatal != nil {
						options.onFatal(err)
					}
				} else if _, err := seeker.Write(buf[:]); err != nil {
					if options.onFatal != nil {
						options.onFatal(err)
					}
				} else if _, err := seeker.Seek(0, io.SeekEnd); err != nil {
					if options.onFatal != nil {
						options.onFatal(err)
					}
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

					// Collect CuePoint before Clear
					if options.cuesReservedSize > 0 {
						runningPos += uint64(w.Size())
						cuePoints = append(cuePoints, cuePoint{
							CueTime: uint64(tc1 - tc0),
							CueTrackPositions: []cueTrackPosition{{
								CueTrack:           f.trackNumber,
								CueClusterPosition: runningPos - segmentDataStart,
							}},
						})
					}

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

// writeVoidElement writes an EBML Void element of exactly totalSize bytes.
// Always uses 8-byte VINT for simplicity; callers must ensure totalSize >= 9.
func writeVoidElement(w io.Writer, totalSize int) error {
	buf := make([]byte, totalSize)
	buf[0] = 0xEC // Void Element ID
	dataSize := uint64(totalSize - 9)
	buf[1] = 0x01
	buf[2] = byte(dataSize >> 48)
	buf[3] = byte(dataSize >> 40)
	buf[4] = byte(dataSize >> 32)
	buf[5] = byte(dataSize >> 24)
	buf[6] = byte(dataSize >> 16)
	buf[7] = byte(dataSize >> 8)
	buf[8] = byte(dataSize)
	_, err := w.Write(buf)
	return err
}

// writeCuesToReserved marshals Cues and writes them into the reserved Void space.
func writeCuesToReserved(
	seeker io.WriteSeeker,
	cuePoints []cuePoint,
	marshalOpts []ebml.MarshalOption,
	reservedStart uint64, reservedSize int,
	onFatal func(error),
) {
	cuesData := struct {
		Cues cues `ebml:"Cues"`
	}{
		Cues: cues{CuePoint: cuePoints},
	}
	var buf bytes.Buffer
	if err := ebml.Marshal(&cuesData, &buf, marshalOpts...); err != nil {
		if onFatal != nil {
			onFatal(err)
		}
		return
	}

	cuesBytes := buf.Bytes()
	remaining := reservedSize - len(cuesBytes)
	if remaining < 0 || (remaining > 0 && remaining < 9) {
		// Cues don't fit, or leftover space is too small for a Void element
		// (need 9 bytes minimum: 1-byte ID + 8-byte VINT).
		// Skip silently — the file remains valid, just not seekable.
		return
	}

	if _, err := seeker.Seek(int64(reservedStart), io.SeekStart); err != nil {
		if onFatal != nil {
			onFatal(err)
		}
		return
	}
	if _, err := seeker.Write(cuesBytes); err != nil {
		if onFatal != nil {
			onFatal(err)
		}
		return
	}

	if remaining > 0 {
		if err := writeVoidElement(seeker, remaining); err != nil {
			if onFatal != nil {
				onFatal(err)
			}
			return
		}
	}

	// Seek back to end of file
	if _, err := seeker.Seek(0, io.SeekEnd); err != nil {
		if onFatal != nil {
			onFatal(err)
		}
	}
}

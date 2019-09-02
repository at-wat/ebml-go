package webm

import (
	"io"
	"sync"

	"github.com/at-wat/ebml-go"
)

// NewSimpleWriter creates FrameWriter for each frame type specified by tracks argument.
// Resultant WebM is written to given io.WriteCloser.
// io.WriteCloser will be closed automatically; don't close it by yourself.
func NewSimpleWriter(w0 io.WriteCloser, tracks []TrackEntry) ([]*FrameWriter, error) {
	w := &writerWithSizeCount{w: w0}

	header := struct {
		Header  EBMLHeader `ebml:"EBML"`
		Segment Segment    `ebml:"Segment,inf"`
	}{
		Header: EBMLHeader{
			EBMLVersion:        1,
			EBMLReadVersion:    1,
			EBMLMaxIDLength:    4,
			EBMLMaxSizeLength:  8,
			DocType:            "webm",
			DocTypeVersion:     2,
			DocTypeReadVersion: 2,
		},
		Segment: Segment{
			Info: Info{
				TimecodeScale: 1000000, // 1ms
				MuxingApp:     "ebml-go.webm.SimpleWriter",
				WritingApp:    "ebml-go.webm.SimpleWriter",
			},
			Tracks: Tracks{
				TrackEntry: tracks,
			},
		},
	}
	if err := ebml.Marshal(&header, w); err != nil {
		return nil, err
	}

	w.Clear()
	cluster := struct {
		Cluster Cluster `ebml:"Cluster,inf"`
	}{
		Cluster: Cluster{
			Timecode: 0,
		},
	}
	if err := ebml.Marshal(&cluster, w); err != nil {
		return nil, err
	}

	ch := make(chan *frame)
	wg := sync.WaitGroup{}
	var ws []*FrameWriter

	for _, t := range tracks {
		wg.Add(1)
		ws = append(ws, &FrameWriter{
			trackNumber: t.TrackNumber,
			f:           ch,
			wg:          &wg,
		})
	}

	closed := make(chan struct{})
	go func() {
		wg.Wait()
		close(closed)
	}()

	go func() {
		tc0 := int64(0xFFFFFFFF)
		tc1 := int64(0xFFFFFFFF)
		lastTc := int64(0)

		defer func() {
			// Finalize WebM
			cluster := struct {
				Cluster Cluster `ebml:"Cluster,inf"`
			}{
				Cluster: Cluster{
					Timecode: uint64(lastTc),
					PrevSize: uint64(w.Size()),
				},
			}
			w.Clear()
			if err := ebml.Marshal(&cluster, w); err != nil {
				// TODO: output error
				panic(err)
			}
			w.Close()
		}()

	L_WRITE:
		for {
			select {
			case <-closed:
				break L_WRITE
			case f := <-ch:
				if tc0 == 0xFFFFFFFF {
					tc0 = f.timestamp
				}
				lastTc = f.timestamp
				tc := f.timestamp - tc1
				if tc >= 0x7FFF || tc1 == 0xFFFFFFFF {
					// Create new Cluster
					tc1 := f.timestamp
					tc = 0

					cluster := struct {
						Cluster Cluster `ebml:"Cluster,inf"`
					}{
						Cluster: Cluster{
							Timecode: uint64(tc1 - tc0),
							PrevSize: uint64(w.Size()),
						},
					}
					w.Clear()
					if err := ebml.Marshal(&cluster, w); err != nil {
						// TODO: output error
						panic(err)
					}
				}
				if tc <= -0x7FFF {
					// Ignore too old frame
					// TODO: output error
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
				if err := ebml.Marshal(&b, w); err != nil {
					// TODO: output error
					panic(err)
				}
			}
		}
	}()

	return ws, nil
}

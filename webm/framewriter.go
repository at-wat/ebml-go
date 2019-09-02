package webm

import (
	"sync"
)

// FrameWriter is a stream frame.
// It's simillar to io.Writer, but having additional arguments keyframe flag and timestamp.
type FrameWriter struct {
	trackNumber uint64
	f           chan *frame
	wg          *sync.WaitGroup
}

type frame struct {
	trackNumber uint64
	keyframe    bool
	timestamp   int64
	b           []byte
}

// Write writes a stream frame to the connected WebM writer.
// timestamp is in millisecond.
func (w *FrameWriter) Write(keyframe bool, timestamp int64, b []byte) (int, error) {
	w.f <- &frame{
		trackNumber: w.trackNumber,
		keyframe:    keyframe,
		timestamp:   timestamp,
		b:           b,
	}
	return len(b), nil
}

// Close closes a stream frame writer.
// Output WebM will be closed after closing all FrameWriter.
func (w *FrameWriter) Close() {
	w.wg.Done()
}

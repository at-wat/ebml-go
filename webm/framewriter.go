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
	"sync"
)

// FrameWriter is a stream frame writer.
// It's simillar to io.Writer, but having additional arguments keyframe flag and timestamp.
type FrameWriter struct {
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
func (w *FrameWriter) Close() error {
	w.wg.Done()

	// If it is the last writer, block until closing output writer.
	w.fin <- struct{}{}

	return nil
}

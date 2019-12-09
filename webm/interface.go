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

// Writer is a stream frame writer interface.
type Writer interface {
	// Write a frame to the connected WebM writer.
	// timestamp is in millisecond.
	Write(keyframe bool, timestamp int64, b []byte) (int, error)
}

// Closer is a stream frame closer interface.
type Closer interface {
	// Close the stream frame writer.
	// Output WebM will be closed after closing all FrameWriter.
	Close() error
}

// WriteCloser groups Writer and Closer.
type WriteCloser interface {
	Writer
	Closer
}

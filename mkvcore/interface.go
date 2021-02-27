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

// BlockWriter is a Matroska block writer interface.
type BlockWriter interface {
	// Write a block to the connected Matroska writer.
	// timestamp is in millisecond.
	Write(keyframe bool, timestamp int64, b []byte) (int, error)
}

// BlockReader is a Matroska block reader interface.
type BlockReader interface {
	// Read a block from the connected Matroska reader.
	Read() (b []byte, keyframe bool, timestamp int64, err error)
}

// BlockCloser is a Matroska closer interface.
type BlockCloser interface {
	// Close the stream frame writer.
	// Output Matroska will be closed after closing all FrameWriter.
	Close() error
}

// BlockWriteCloser groups Writer and Closer.
type BlockWriteCloser interface {
	BlockWriter
	BlockCloser
}

// BlockReadCloser groups Reader and Closer.
type BlockReadCloser interface {
	BlockReader
	BlockCloser
}

// TrackEntryGetter is a interface to get TrackEntry.
type TrackEntryGetter interface {
	TrackEntry() TrackEntry
}

// BlockReadCloserWithTrackEntry groups BlockReadCloser and TrackEntryGetter.
type BlockReadCloserWithTrackEntry interface {
	BlockReadCloser
	TrackEntryGetter
}

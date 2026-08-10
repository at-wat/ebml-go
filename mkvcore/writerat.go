// Copyright 2026 The ebml-go authors.
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
)

type writerAtBySeeker struct {
	io.WriteSeeker
}

// WriteAt emulates WriterAt using WriteSeeker.
// Genuine WriterAt must be goroutine-safe if the ranges do not overlap.
// This is for BlockWriter internal use and not goroutine-safe.
func (w *writerAtBySeeker) WriteAt(p []byte, off int64) (int, error) {
	if _, err := w.Seek(off, io.SeekStart); err != nil {
		return 0, err
	}
	n, err := w.Write(p)
	if err != nil {
		return n, err
	}
	if _, err := w.Seek(0, io.SeekEnd); err != nil {
		return n, err
	}
	return n, nil
}

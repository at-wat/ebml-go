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
)

type writerWithSizeCount struct {
	size int
	w    io.WriteCloser
}

func (w *writerWithSizeCount) Write(b []byte) (int, error) {
	w.size += len(b)
	return w.w.Write(b)
}

func (w *writerWithSizeCount) Clear() {
	w.size = 0
}

func (w *writerWithSizeCount) Close() error {
	return w.w.Close()
}

func (w *writerWithSizeCount) Size() int {
	return w.size
}

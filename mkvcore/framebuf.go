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

type frameBuffer struct {
	buf []*frame
}

func (b *frameBuffer) Push(f *frame) {
	b.buf = append(b.buf, f)
}
func (b *frameBuffer) Head() *frame {
	if len(b.buf) == 0 {
		return nil
	}
	return b.buf[0]
}
func (b *frameBuffer) Tail() *frame {
	if len(b.buf) == 0 {
		return nil
	}
	return b.buf[len(b.buf)-1]
}
func (b *frameBuffer) Pop() *frame {
	n := len(b.buf)
	if n == 0 {
		return nil
	}
	head := b.buf[0]
	b.buf[0] = nil

	if n == 1 {
		b.buf = nil
	} else {
		b.buf = b.buf[1:]
	}
	return head
}
func (b *frameBuffer) Size() int {
	return len(b.buf)
}

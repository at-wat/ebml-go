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
	"reflect"
	"testing"
)

func TestFrameBuffer(t *testing.T) {
	buf := &frameBuffer{}

	if n := buf.Size(); n != 0 {
		t.Errorf("Size() must return 0 at beginning, got %d", n)
	}
	if h := buf.Head(); h != nil {
		t.Errorf("Head() must return nil at beginning, got %v", h)
	}

	frames := []frame{
		{trackNumber: 2},
		{trackNumber: 3},
	}
	buf.Push(&frames[0])
	buf.Push(&frames[1])

	if n := buf.Size(); n != 2 {
		t.Errorf("Size() must return 2 after pushing two frames, got %d", n)
	}
	if h := buf.Head(); !reflect.DeepEqual(*h, frames[0]) {
		t.Errorf("Head() must return first frame, expected: %v, got %v", frames[0].trackNumber, *h)
	}

	if h := buf.Pop(); !reflect.DeepEqual(*h, frames[0]) {
		t.Errorf("Pop() must return first frame, expected: %v, got %v", frames[0].trackNumber, *h)
	}
	if n := buf.Size(); n != 1 {
		t.Errorf("Size() must return 1 after popping one frames, got %d", n)
	}
}

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
	"sync"
	"testing"
)

func TestMultiTrackBlockSorter(t *testing.T) {
	wg := sync.WaitGroup{}
	f := NewMultiTrackBlockSorter(2)

	chOut := make(chan *frame)
	ch := []chan *frame{
		make(chan *frame),
		make(chan *frame),
	}

	w := []BlockWriter{
		&filterWriter{0, chOut},
		&filterWriter{1, chOut},
	}
	r := []BlockReader{
		&filterReader{ch[0]},
		&filterReader{ch[1]},
	}

	var frames []frame
	wg.Add(1)
	go func() {
		for f := range chOut {
			frames = append(frames, *f)
		}
		wg.Done()
	}()

	go func() {
		ch[0] <- &frame{0, false, 10, []byte{1}}
		ch[0] <- &frame{0, false, 11, []byte{2}}
		ch[1] <- &frame{1, false, 9, []byte{3}}
		ch[0] <- &frame{0, false, 16, []byte{4}}
		ch[0] <- &frame{0, false, 17, []byte{5}}
		ch[0] <- &frame{0, false, 18, []byte{6}}
		ch[1] <- &frame{1, false, 15, []byte{7}} // drop due to maxDelay=2
		ch[1] <- &frame{1, false, 18, []byte{8}}
		close(ch[0])
		close(ch[1])
	}()

	if err := f.Filter(r, w); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	close(chOut)
	wg.Wait()

	framesExpected := []frame{
		{1, false, 9, []byte{3}},
		{0, false, 10, []byte{1}},
		{0, false, 11, []byte{2}},
		{0, false, 16, []byte{4}},
		{0, false, 17, []byte{5}},
		{0, false, 18, []byte{6}},
		{1, false, 18, []byte{8}},
	}
	if !reflect.DeepEqual(framesExpected, frames) {
		t.Errorf("Unexpected sort result, \nexpected: %v, \n     got: %v", framesExpected, frames)
	}
}

func BenchmarkMultiTrackBlockSorter(b *testing.B) {
	f := NewMultiTrackBlockSorter(2)

	chOut := make(chan *frame)
	ch := []chan *frame{
		make(chan *frame),
		make(chan *frame),
	}

	w := []BlockWriter{
		&filterWriter{0, chOut},
		&filterWriter{1, chOut},
	}
	r := []BlockReader{
		&filterReader{ch[0]},
		&filterReader{ch[1]},
	}

	go func() {
		for range chOut {
		}
	}()

	go func() {
		for i := 0; i < b.N; i++ {
			ch[0] <- &frame{0, false, int64(i), []byte{1, 2, 3, 4}}
			ch[1] <- &frame{1, false, int64(i) + 5, []byte{2, 3, 4, 5}}
		}
		close(ch[0])
		close(ch[1])
	}()

	b.ResetTimer()
	if err := f.Filter(r, w); err != nil {
		b.Fatalf("Unexpected error: %v", err)
	}
	close(chOut)
}

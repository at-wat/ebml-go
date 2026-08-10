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
	"bytes"
	"errors"
	"io"
	"testing"
)

type dummySeeker struct {
	writeFunc func([]byte) (int, error)
	seekFunc  func(int64, int) (int64, error)
}

func (s *dummySeeker) Write(p []byte) (int, error)           { return s.writeFunc(p) }
func (s *dummySeeker) Seek(off int64, wh int) (int64, error) { return s.seekFunc(off, wh) }

func TestWriterAtBySeeker(t *testing.T) {
	errDummy := errors.New("dummy")

	t.Run("WriteAt", func(t *testing.T) {
		var iSeek, iWrite int
		ws := &dummySeeker{
			writeFunc: func(p []byte) (int, error) {
				iWrite++
				if iWrite != 1 {
					t.Fatal("Unexpected Write call")
				}
				expected := []byte{1, 2}
				if !bytes.Equal(p, expected) {
					t.Errorf("Expected to write %v, got %v", expected, p)
				}
				return len(p), nil
			},
			seekFunc: func(off int64, wh int) (int64, error) {
				iSeek++
				switch iSeek {
				case 1:
					if wh != io.SeekStart {
						t.Fatalf("First seek must be relative to SeekStart(0), got %d", wh)
					}
					if off != 3 {
						t.Fatalf("Expected first seek offset of 3, got %d", off)
					}
					return 3, nil
				case 2:
					if wh != io.SeekEnd {
						t.Fatalf("First seek must be relative to SeekEnd(2), got %d", wh)
					}
					if off != 0 {
						t.Fatalf("Expected second seek offset of 0, got %d", off)
					}
					return 5, nil
				default:
					t.Fatal("Unexpected Seek call")
				}
				panic("must not reach here")
			},
		}
		w := &writerAtBySeeker{
			WriteSeeker: ws,
		}

		n, err := w.WriteAt([]byte{1, 2}, 3)
		if err != nil {
			t.Fatal(err)
		}
		if n != 2 {
			t.Errorf("Expected n: 2, got: %d", n)
		}
	})

	t.Run("ErrorPropagation", func(t *testing.T) {
		testCases := map[string]struct {
			w   func() *dummySeeker
			n   int
			err error
		}{
			"WriteError": {
				w: func() *dummySeeker {
					return &dummySeeker{
						writeFunc: func([]byte) (int, error) { return 0, errDummy },
						seekFunc:  func(int64, int) (int64, error) { return 0, nil },
					}
				},
				n:   0,
				err: errDummy,
			},
			"FirstSeekError": {
				w: func() *dummySeeker {
					return &dummySeeker{
						seekFunc: func(int64, int) (int64, error) { return 0, errDummy },
					}
				},
				n:   0,
				err: errDummy,
			},
			"SecondSeekError": {
				w: func() *dummySeeker {
					var i int
					return &dummySeeker{
						writeFunc: func(p []byte) (int, error) { return len(p), nil },
						seekFunc: func(int64, int) (int64, error) {
							i++
							if i == 2 {
								return 0, errDummy
							}
							return 0, nil
						},
					}
				},
				n:   0,
				err: errDummy,
			},
		}
		for name, testCase := range testCases {
			testCase := testCase
			t.Run(name, func(t *testing.T) {
				w := &writerAtBySeeker{
					WriteSeeker: testCase.w(),
				}
				n, err := w.WriteAt([]byte{0}, 1)
				if n != testCase.n {
					t.Errorf("Expected n: %d, got: %d", testCase.n, n)
				}
				if !errors.Is(err, testCase.err) {
					t.Errorf("Expected error: '%v', got: '%v'", testCase.err, err)
				}
			})
		}
	})
}

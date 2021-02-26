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
	"bytes"
	"errors"
	"io"
	"sync"
	"testing"

	"github.com/at-wat/ebml-go"
	"github.com/at-wat/ebml-go/internal/buffercloser"
	"github.com/at-wat/ebml-go/internal/errs"
)

func TestBlockReader(t *testing.T) {
	type testMkvHeader struct {
		Segment flexSegment `ebml:"Segment"`
	}
	testCases := map[string]struct {
		input    testMkvHeader
		expected [][]frame
	}{
		"TwoTracks": {
			input: testMkvHeader{
				Segment: flexSegment{
					Tracks: flexTracks{TrackEntry: []interface{}{
						map[string]interface{}{"TrackNumber": uint(1)},
						map[string]interface{}{"TrackNumber": uint(2)},
					}},
					Cluster: []simpleBlockCluster{
						{
							Timecode: uint64(100),
							SimpleBlock: []ebml.Block{
								{
									TrackNumber: 1,
									Timecode:    int16(-10),
									Keyframe:    false,
									Data:        [][]byte{{0x01, 0x02}},
								},
								{
									TrackNumber: 2,
									Timecode:    int16(10),
									Keyframe:    true,
									Data:        [][]byte{{0x03, 0x04, 0x05}},
								},
								{
									TrackNumber: 1,
									Timecode:    int16(30),
									Keyframe:    true,
									Data:        [][]byte{{0x06}},
								},
							},
						},
						{
							Timecode: uint64(30),
							PrevSize: uint64(39),
						},
					},
				},
			},
			expected: [][]frame{
				{
					{keyframe: false, timestamp: 90, b: []byte{0x01, 0x02}},
					{keyframe: true, timestamp: 130, b: []byte{0x06}},
				},
				{
					{keyframe: true, timestamp: 110, b: []byte{0x03, 0x04, 0x05}},
				},
			},
		},
		"NoBlock": {
			input: testMkvHeader{
				Segment: flexSegment{
					Tracks: flexTracks{TrackEntry: []interface{}{
						map[string]interface{}{"TrackNumber": uint(1)},
						map[string]interface{}{"TrackNumber": uint(2)},
					}},
					Cluster: []simpleBlockCluster{},
				},
			},
			expected: [][]frame{{}, {}},
		},
		"NoCluster": {
			input: testMkvHeader{
				Segment: flexSegment{
					Tracks: flexTracks{TrackEntry: []interface{}{
						map[string]interface{}{"TrackNumber": uint(1)},
						map[string]interface{}{"TrackNumber": uint(2)},
					}},
				},
			},
			expected: [][]frame{{}, {}},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			buf := buffercloser.New()
			if err := ebml.Marshal(&testCase.input, buf); err != nil {
				t.Fatalf("Failed to marshal test data: '%v'", err)
			}
			buf.Close()

			ws, err := NewSimpleBlockReader(bytes.NewReader(buf.Bytes()))
			if err != nil {
				t.Fatalf("Failed to create BlockReader: '%v'", err)
			}

			if len(ws) != len(testCase.expected) {
				t.Fatalf("Number of the returned writer (%d) must be same as the number of TrackEntry (%d)", len(ws), len(testCase.expected))
			}

			var wg sync.WaitGroup
			wg.Add(len(testCase.expected))

			for i, dd := range testCase.expected {
				i, dd := i, dd
				go func() {
					defer wg.Done()

					for _, d := range dd {
						buf, keyframe, timestamp, err := ws[i].Read()
						if err != nil {
							t.Errorf("Failed to Read: '%v'", err)
						}
						if keyframe != d.keyframe {
							t.Errorf("Expected keyframe: %v, got: %v", d.keyframe, keyframe)
						}
						if timestamp != d.timestamp {
							t.Errorf("Expected timestamp: %v, got: %v", d.timestamp, timestamp)
						}
						if !bytes.Equal(buf, d.b) {
							t.Errorf("Expected bytes: %v, got: %v", d.b, buf)
						}
					}
					if _, _, _, err := ws[i].Read(); err != io.EOF {
						t.Errorf("Expected: EOF, got: %v", err)
					}
					if err := ws[i].Close(); err != nil {
						t.Errorf("Unexpected error: %v", err)
					}
				}()
			}

			wg.Wait()
		})
	}
}

func TestBlockReader_FailingOptions(t *testing.T) {
	errDummy0 := errors.New("an error 0")
	errDummy1 := errors.New("an error 1")

	cases := map[string]struct {
		opts []BlockReaderOption
		err  error
	}{
		"ReaderOptionError": {
			opts: []BlockReaderOption{
				BlockReaderOptionFn(func(*BlockReaderOptions) error { return errDummy0 }),
			},
			err: errDummy0,
		},
		"UnmarshalOptionError": {
			opts: []BlockReaderOption{
				WithUnmarshalOptions(
					func(*ebml.UnmarshalOptions) error { return errDummy1 },
				),
			},
			err: errDummy1,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			buf := bytes.NewReader([]byte{})
			_, err := NewSimpleBlockReader(buf, c.opts...)
			if !errs.Is(err, c.err) {
				t.Errorf("Expected error: '%v', got: '%v'", c.err, err)
			}
		})
	}
}

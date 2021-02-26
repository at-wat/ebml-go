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
	"time"

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

			rs, err := NewSimpleBlockReader(bytes.NewReader(buf.Bytes()))
			if err != nil {
				t.Fatalf("Failed to create BlockReader: '%v'", err)
			}

			if len(rs) != len(testCase.expected) {
				t.Fatalf("Number of the returned writer (%d) must be same as the number of TrackEntry (%d)", len(rs), len(testCase.expected))
			}

			var wg sync.WaitGroup
			wg.Add(len(testCase.expected))

			for i, dd := range testCase.expected {
				i, dd := i, dd
				go func() {
					defer wg.Done()

					for _, d := range dd {
						buf, keyframe, timestamp, err := rs[i].Read()
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
					if _, _, _, err := rs[i].Read(); err != io.EOF {
						t.Errorf("Expected: EOF, got: %v", err)
					}
					if err := rs[i].Close(); err != nil {
						t.Errorf("Unexpected error: %v", err)
					}
				}()
			}

			wg.Wait()
		})
	}
}

func TestBlockReader_Close(t *testing.T) {
	type testMkvHeader struct {
		Segment flexSegment `ebml:"Segment"`
	}
	input := testMkvHeader{
		Segment: flexSegment{
			Tracks: flexTracks{TrackEntry: []interface{}{
				map[string]interface{}{"TrackNumber": uint(1)},
				map[string]interface{}{"TrackNumber": uint(2)},
			}},
			Cluster: []simpleBlockCluster{
				{
					SimpleBlock: []ebml.Block{
						{TrackNumber: 1, Data: [][]byte{{0x01}}},
						{TrackNumber: 2, Data: [][]byte{{0x02}}},
						{TrackNumber: 1, Data: [][]byte{{0x03}}},
					},
				},
			},
		},
	}

	buf := buffercloser.New()
	if err := ebml.Marshal(&input, buf); err != nil {
		t.Fatalf("Failed to marshal test data: '%v'", err)
	}
	buf.Close()

	rs, err := NewSimpleBlockReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("Failed to create BlockReader: '%v'", err)
	}

	if len(rs) != 2 {
		t.Fatalf("Number of the returned writer (%d) must be same as the number of TrackEntry (%d)", len(rs), 2)
	}

	if err := rs[0].Close(); err != nil {
		t.Fatalf("Unexpected Close error: '%v'", err)
	}

	errTimeout := errors.New("timeout")
	readWithTimeout := func(i int) error {
		errCh := make(chan error)
		go func() {
			_, _, _, err := rs[i].Read()
			errCh <- err
		}()

		select {
		case err := <-errCh:
			return err
		case <-time.After(time.Second):
			return errTimeout
		}
	}

	if err := readWithTimeout(1); err != nil {
		t.Fatalf("Unexpected Read error: '%v'", err)
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

func TestBlockReader_WithUnmarshalOptions(t *testing.T) {
	testCases := map[string]struct {
		opts     []BlockReaderOption
		err      error
		nReaders int
	}{
		"Default": {
			err: ebml.ErrUnknownElement,
		},
		"IgnoreUnknown": {
			opts: []BlockReaderOption{
				WithUnmarshalOptions(ebml.WithIgnoreUnknown(true)),
			},
			nReaders: 1,
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			testBinary := []byte{
				0x18, 0x53, 0x80, 0x67, 0xFF, // Segment
				0x16, 0x54, 0xae, 0x6b, 0x95, // Tracks
				0x81, 0x81, // 0x81 is not defined in Matroska v4
				0xae, 0x83, // TrackEntry[0]
				0xd7, 0x81, 0x01, // TrackNumber=1
				0x1F, 0x43, 0xB6, 0x75, 0xFF, // Cluster
				0xE7, 0x81, 0x00, // Timecode
				0xA3, 0x86, 0x81, 0x00, 0x00, 0x88, 0xAA, 0xCC, // SimpleBlock
			}

			rs, err := NewSimpleBlockReader(
				bytes.NewReader(testBinary),
				testCase.opts...,
			)
			if !errs.Is(err, testCase.err) {
				if testCase.err != nil {
					t.Fatalf("Expected error: '%v', got: '%v'", testCase.err, err)
				} else {
					t.Fatalf("Unexpected error: '%v'", err)
				}
			}

			if len(rs) != testCase.nReaders {
				t.Fatalf("Number of the returned writer (%d) must be same as the number of TrackEntry (%d)", len(rs), 1)
			}
		})
	}
}

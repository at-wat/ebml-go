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

package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"os/signal"

	"github.com/at-wat/ebml-go/webm"
)

func main() {
	w, err := os.OpenFile("test.webm", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}

	ws, err := webm.NewSimpleBlockWriter(w,
		[]webm.TrackEntry{
			{
				Name:            "Video",
				TrackNumber:     1,
				TrackUID:        12345,
				CodecID:         "V_VP8",
				TrackType:       1,
				DefaultDuration: 33333333,
				Video: &webm.Video{
					PixelWidth:  320,
					PixelHeight: 240,
				},
			},
		})
	if err != nil {
		panic(err)
	}

	fmt.Print("Run following command to send RTP stream to this example:\n" +
		"$ gst-launch-1.0 videotestsrc" +
		" ! video/x-raw,width=320,height=240,framerate=30/1" +
		" ! vp8enc target-bitrate=4000" +
		" ! rtpvp8pay ! udpsink host=localhost port=4000\n\n" +
		"Waiting first keyframe...\n")

	// Listen UDP RTP packets
	addr, err := net.ResolveUDPAddr("udp", ":4000")
	if err != nil {
		panic(err)
	}
	pc, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}

	closed := make(chan os.Signal, 1)
	signal.Notify(closed, os.Interrupt)
	go func() {
		<-closed
		pc.Close()
	}()

	var frame []byte
	var keyframe bool
	var keyframeCnt int
	var tcRawLast int64 = -1
	var tcRawBase int64
	buffer := make([]byte, 1522)

	for {
		n, _, err := pc.ReadFrom(buffer)
		if err != nil {
			fmt.Println(err)
			break
		}
		if n < 14 {
			fmt.Print("RTP packet size is too small.\n")
			continue
		}

		// RTP descriptor and header must be fully parsed in the real application.
		vp8Desc := buffer[12]

		// *Extended control bits present flag* is not supported in this example
		if vp8Desc&0x80 != 0 {
			panic("Incoming VP8 payload descriptor has extended fields which is not supported by this example!")
		}

		// *Start of VP8 partition flag*
		if vp8Desc&0x10 != 0 {
			if keyframe {
				keyframeCnt++
			}
			if len(frame) > 0 && keyframeCnt > 0 {
				tcRaw := int64(binary.BigEndian.Uint32(buffer[4:8]))
				if tcRawLast == -1 {
					tcRawLast = tcRaw
				}
				if tcRaw < 0x10000 && tcRawLast > 0xFFFFFFFF-0x10000 {
					// counter overflow
					tcRawBase += 0x100000000
				} else if tcRawLast < 0x10000 && tcRaw > 0xFFFFFFFF-0x10000 {
					// counter underflow
					tcRawBase -= 0x100000000
				}
				tcRawLast = tcRaw

				tc := (tcRaw + tcRawBase) / 90 // VP8 timestamp rate is 90000.
				fmt.Printf("RTP frame received. (len: %d, timestamp: %d, keyframe: %v)\n", len(frame), tc, keyframe)
				ws[0].Write(keyframe, int64(tc), frame)
			}
			frame = []byte{}
			keyframe = false

			// RTP header 12 bytes, VP8 payload descriptor 1 byte.
			vp8Header := buffer[13]
			if vp8Header&0x01 == 0 {
				keyframe = true
			}
		}

		frame = append(frame, buffer[13:n]...)
	}

	fmt.Printf("\nFinalizing webm...\n")
	ws[0].Close()
}

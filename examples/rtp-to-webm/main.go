package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"os/signal"

	"github.com/at-wat/ebml-go"
	"github.com/at-wat/ebml-go/webm"
)

func main() {
	header := struct {
		Header  webm.EBMLHeader `ebml:"EBML"`
		Segment webm.Segment    `ebml:"Segment,inf"`
	}{
		Header: webm.EBMLHeader{
			EBMLVersion:        1,
			EBMLReadVersion:    1,
			EBMLMaxIDLength:    4,
			EBMLMaxSizeLength:  8,
			DocType:            "webm",
			DocTypeVersion:     2,
			DocTypeReadVersion: 2,
		},
		Segment: webm.Segment{
			Info: webm.Info{
				TimecodeScale: 1000000, // 1ms
				MuxingApp:     "ebml-go example",
				WritingApp:    "ebml-go example",
			},
			Tracks: webm.Tracks{
				TrackEntry: []webm.TrackEntry{
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
				},
			},
		},
	}
	clusterHead := struct {
		Cluster webm.Cluster `ebml:"Cluster,inf"`
	}{
		Cluster: webm.Cluster{
			Timecode: 0,
		},
	}

	w, err := os.OpenFile("test.webm", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	if err := ebml.Marshal(&header, w); err != nil {
		panic(err)
	}

	// Store cluster start position to make it seekable later
	w.Sync()
	clusterHeadOffset, err := w.Seek(0, 1)
	if err != nil {
		panic(err)
	}
	if err := ebml.Marshal(&clusterHead, w); err != nil {
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
	defer pc.Close()

	var ts0 uint32
	var ts uint32
	buffer := make([]byte, 1522)

	var frame []byte
	var keyframe bool
	var keyframeCnt int

	closed := make(chan os.Signal)
	signal.Notify(closed, os.Interrupt)
	go func() {
		<-closed
		pc.Close()
	}()

	for {
		n, _, err := pc.ReadFrom(buffer)
		if err != nil {
			fmt.Println(err)
			break
		}
		if n < 12 {
			fmt.Print("RTP packet size is too small.\n")
			continue
		}

		// RTP Header must be fully parsed in the real application.
		// And, 32bit counter overflow must be treated in real.
		tsAbs := binary.BigEndian.Uint32(buffer[4:8]) / 90 // VP8 ts rate is 90000.

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
			if len(frame) > 0 {
				if keyframeCnt > 0 {
					if ts0 == 0 {
						ts0 = tsAbs
					}
					ts = tsAbs - ts0
					if ts >= 0x8000 {
						fmt.Print("Cluster is full.\n")
						break
					}
					fmt.Printf("RTP frame received. (len: %d, timestamp: %d, keyframe: %v)\n", len(frame), tsAbs, keyframe)
					b := struct {
						Block ebml.Block `ebml:"SimpleBlock"`
					}{
						ebml.Block{
							TrackNumber: 1,
							Timecode:    int16(ts),
							Keyframe:    keyframe,
							Data:        [][]byte{frame},
						},
					}
					// Write SimpleBlock to the file
					if err := ebml.Marshal(&b, w); err != nil {
						panic(err)
					}
				}
			}
			frame = []byte{}
			keyframe = false
		}

		// RTP header 12 bytes, VP8 payload descriptor 1 byte.
		vp8Header := buffer[13]
		if vp8Header&0x01 == 0 {
			keyframe = true
		}

		frame = append(frame, buffer[13:n]...)
	}

	// Calculate cluster size and finalize
	w.Sync()
	clusterTailOffset, err := w.Seek(0, 1)
	if err != nil {
		panic(err)
	}
	l := uint64(clusterTailOffset - clusterHeadOffset)
	fmt.Printf("\nFinalizing webm... %d milliseconds, %d bytes\n", ts, l)

	finalizer := struct {
		Cluster webm.Cluster `ebml:"Cluster,inf"`
	}{
		Cluster: webm.Cluster{
			Timecode: uint64(ts),
			PrevSize: l,
		},
	}
	if err := ebml.Marshal(&finalizer, w); err != nil {
		panic(err)
	}
}

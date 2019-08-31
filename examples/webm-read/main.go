package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/at-wat/ebml-go"
)

func main() {
	r, err := os.Open("sample.webm")
	if err != nil {
		panic(err)
	}
	defer r.Close()

	var ret struct {
		Header struct {
			EBMLVersion        uint64 `ebml:"EBMLVersion"`
			EBMLReadVersion    uint64 `ebml:"EBMLReadVersion"`
			EBMLMaxIDLength    uint64 `ebml:"EBMLMaxIDLength"`
			EBMLMaxSizeLength  uint64 `ebml:"EBMLMaxSizeLength"`
			DocType            string `ebml:"EBMLDocType"`
			DocTypeVersion     uint64 `ebml:"EBMLDocTypeVersion"`
			DocTypeReadVersion uint64 `ebml:"EBMLDocTypeReadVersion"`
		} `ebml:"EBML"`
		Segment struct {
			SeekHead struct {
				Seek []struct {
					SeekID       []byte `ebml:"SeekID"`
					SeekPosition uint64 `ebml:"SeekPosition"`
				} `ebml:"Seek"`
			} `ebml:"SeekHead"`
			Info struct {
				TimecodeScale uint64    `ebml:"TimecodeScale"`
				MuxingApp     string    `ebml:"MuxingApp"`
				WritingApp    string    `ebml:"WritingApp"`
				Duration      float64   `ebml:"Duration"`
				DateUTC       time.Time `ebml:"DateUTC"`
			} `ebml:"Info"`
			Tracks struct {
				TrackEntry []struct {
					Name            string `ebml:"Name"`
					TrackNumber     uint64 `ebml:"TrackNumber"`
					TrackUID        uint64 `ebml:"TrackUID"`
					CodecID         string `ebml:"CodecID"`
					CodecPrivate    []byte `ebml:"CodecPrivate"`
					CodecDelay      uint64 `ebml:"CodecDelay"`
					TrackType       uint64 `ebml:"TrackType"`
					DefaultDuration uint64 `ebml:"DefaultDuration"`
					SeekPreRoll     uint64 `ebml:"SeekPreRoll"`
					Audio           struct {
						SamplingFrequency float64 `ebml:"SamplingFrequency"`
						Channels          uint64  `ebml:"Channels"`
					} `ebml:"Audio"`
					Video struct {
						PixedWidth  uint64 `ebml:"PixelWidth"`
						PixedHeight uint64 `ebml:"PixelHeight"`
					} `ebml:"Video"`
				} `ebml:"TrackEntry"`
			} `ebml:"Tracks"`
			Cluster []struct {
				Timecode   uint64 `ebml:"Timecode"`
				BlockGroup []struct {
					BlockDuration uint64       `ebml:"BlockDuration"`
					Block         []ebml.Block `ebml:"Block"`
				} `ebml:"BlockGroup"`
				SimpleBlock []ebml.Block `ebml:"SimpleBlock"`
			} `ebml:"Cluster"`
		} `ebml:"Segment"`
	}
	if err := ebml.Unmarshal(r, &ret); err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	j, err := json.MarshalIndent(ret, "", "  ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Printf("%s\n", string(j))

	w, err := os.OpenFile("copy.webm", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer w.Close()
	if err := ebml.Marshal(&ret, w); err != nil {
		panic(err)
	}
>>>>>>> Stashed changes
}

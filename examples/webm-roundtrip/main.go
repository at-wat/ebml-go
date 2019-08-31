package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/at-wat/ebml-go"
	"github.com/at-wat/ebml-go/webm"
)

func main() {
	r, err := os.Open("sample.webm")
	if err != nil {
		panic(err)
	}
	defer r.Close()

	var ret struct {
		Header  webm.EBMLHeader    `ebml:"EBML"`
		Segment webm.SegmentStream `ebml:"Segment,inf"`
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
}

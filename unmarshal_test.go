package ebml

import (
	"bytes"
	"fmt"
)

func ExampleUnmarshal() {
	TestBinary := []byte{
		0x1a, 0x45, 0xdf, 0xa3, // EBML
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, // 0x10
		0x42, 0x82, 0x85, 0x77, 0x65, 0x62, 0x6d, 0x00,
		0x42, 0x87, 0x81, 0x02, 0x42, 0x85, 0x81, 0x02,
	}
	type TestEBML struct {
		EBML struct {
			EBMLDocType            string
			EBMLDocTypeVersion     uint64
			EBMLDocTypeReadVersion uint64
		}
	}

	r := bytes.NewReader(TestBinary)

	var ret TestEBML
	if err := Unmarshal(r, &ret); err != nil {
		fmt.Printf("error: %+v\n", err)
	}
	fmt.Println(ret)

	// Output: {{webm 2 2}}
}

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

var (
	// DefaultEBMLHeader is the default EBML header and is used by writer.
	DefaultEBMLHeader = &EBMLHeader{
		EBMLVersion:        1,
		EBMLReadVersion:    1,
		EBMLMaxIDLength:    4,
		EBMLMaxSizeLength:  8,
		DocType:            "webm",
		DocTypeVersion:     2,
		DocTypeReadVersion: 2,
	}
	// DefaultSegmentInfo is the default Segment.Info and is used by writer.
	DefaultSegmentInfo = &Info{
		TimecodeScale: 1000000, // 1ms
		MuxingApp:     "ebml-go.webm.FrameWriter",
		WritingApp:    "ebml-go.webm.FrameWriter",
	}
)

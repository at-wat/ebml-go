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

package ebml

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

type structTag struct {
	name      string
	size      uint64
	omitEmpty bool
}

var (
	errEmptyTag   = errors.New("Empty tag in tag string")
	errInvalidTag = errors.New("Invalid tag in tag string")
)

func parseTag(rawtag string) (*structTag, error) {
	tag := &structTag{}

	ts := strings.Split(rawtag, ",")
	if len(ts) == 0 {
		return tag, nil
	}

	for i, t := range ts {
		kv := strings.SplitN(t, "=", 2)

		if len(kv) == 1 {
			if i == 0 {
				tag.name = kv[0]
			} else {
				switch kv[0] {
				case "":
					return nil, errEmptyTag
				case "omitempty":
					tag.omitEmpty = true
				case "inf":
					os.Stderr.WriteString("Deprecated: \"inf\" tag is replaced by \"size=unknown\"\n")
					tag.size = sizeUnknown
				default:
					return nil, errInvalidTag
				}
			}
			continue
		}

		switch kv[0] {
		case "":
			return nil, errEmptyTag
		case "size":
			if kv[1] == "unknown" {
				tag.size = sizeUnknown
			} else {
				s, err := strconv.Atoi(kv[1])
				if err != nil {
					return nil, err
				}
				tag.size = uint64(s)
			}
		default:
			return nil, errInvalidTag
		}
	}
	return tag, nil
}

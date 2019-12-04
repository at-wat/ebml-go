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
		if len(t) == 0 {
			return nil, errEmptyTag
		}
		kv := strings.SplitN(t, "=", 2)

		if len(kv) == 1 {
			if i == 0 {
				tag.name = kv[0]
			} else {
				switch kv[0] {
				case "omitempty":
					tag.omitEmpty = true
				case "inf":
					os.Stderr.WriteString("Deprecated: \"inf\" tag is replaced by \"size=unknown\"\n")
					tag.size = sizeInf
				default:
					return nil, errInvalidTag
				}
			}
			continue
		}

		switch kv[0] {
		case "size":
			if kv[1] == "unknown" {
				tag.size = sizeInf
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

package ebml

import (
	"reflect"
	"testing"
)

func TestParseTag(t *testing.T) {
	cases := map[string]struct {
		input    string
		expected *structTag
		err      error
	}{
		"Empty": {
			"",
			&structTag{}, nil,
		},
		"Name": {
			"Name123",
			&structTag{name: "Name123"}, nil,
		},
		"OmitEmpty": {
			"Name123,omitempty",
			&structTag{name: "Name123", omitEmpty: true}, nil,
		},
		"OmitEmptyWithDefaultName": {
			",omitempty",
			&structTag{omitEmpty: true}, nil,
		},
		"Size": {
			"Name123,size=45",
			&structTag{name: "Name123", size: 45}, nil,
		},
		"UnknownSize": {
			"Name123,size=unknown",
			&structTag{name: "Name123", size: sizeUnknown}, nil,
		},
		"UnknownSizeDeprecated": {
			"Name123,inf",
			&structTag{name: "Name123", size: sizeUnknown}, nil,
		},
		"InvalidTag": {
			"Name,invalidtag",
			nil, errInvalidTag,
		},
		"InvalidTagWithValue": {
			"Name,invalidtag=1",
			nil, errInvalidTag,
		},
		"EmptyTag": {
			"Name,",
			nil, errEmptyTag,
		},
		"EmptyTagWithValue": {
			"Name,=45",
			nil, errEmptyTag,
		},
		"TwoEmptyTags": {
			"Name,,",
			nil, errEmptyTag,
		},
	}
	for n, c := range cases {
		t.Run(n, func(t *testing.T) {
			tag, err := parseTag(c.input)
			if err != c.err {
				t.Errorf("Unexpected error, expected: %v, got: %v", c.err, err)
			}
			if (c.expected == nil) != (tag == nil) {
				t.Errorf("Unexpected output nil-ness, expected: %v, got: %v", c.expected == nil, tag == nil)
			} else if tag != nil && !reflect.DeepEqual(*c.expected, *tag) {
				t.Errorf("Unexpected output, expected: %v, got: %v", *c.expected, *tag)
			}
		})
	}
}

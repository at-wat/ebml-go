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
	"reflect"
	"strconv"
	"testing"

	"github.com/at-wat/ebml-go/internal/errs"
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
			&structTag{name: "Name123", size: SizeUnknown}, nil,
		},
		"UnknownSizeDeprecated": {
			"Name123,inf",
			&structTag{name: "Name123", size: SizeUnknown}, nil,
		},
		"InvalidSize": {
			"Name123,size=a",
			nil, strconv.ErrSyntax,
		},
		"InvalidTag": {
			"Name,invalidtag",
			nil, ErrInvalidTag,
		},
		"InvalidTagWithValue": {
			"Name,invalidtag=1",
			nil, ErrInvalidTag,
		},
		"EmptyTag": {
			"Name,",
			nil, ErrEmptyTag,
		},
		"EmptyTagWithValue": {
			"Name,=45",
			nil, ErrEmptyTag,
		},
		"TwoEmptyTags": {
			"Name,,",
			nil, ErrEmptyTag,
		},
	}
	for n, c := range cases {
		t.Run(n, func(t *testing.T) {
			tag, err := parseTag(c.input)
			if !errs.Is(err, c.err) {
				t.Errorf("Expected error: '%v', got: '%v'", c.err, err)
			}
			if (c.expected == nil) != (tag == nil) {
				t.Errorf("Expected output nil-ness: %v, got: %v", c.expected == nil, tag == nil)
			} else if tag != nil && !reflect.DeepEqual(*c.expected, *tag) {
				t.Errorf("Expected output: %v, got: %v", *c.expected, *tag)
			}
		})
	}
}

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

package errs

import (
	"errors"
	"testing"
)

type dummyError113 struct {
	Err error
}

func (e *dummyError113) Error() string {
	return "dummy: " + e.Err.Error()
}

func (e *dummyError113) Unwrap() error {
	return e.Err
}

type dummyError113Is struct {
	Err error
}

func (e *dummyError113Is) Error() string {
	return "dummy: " + e.Err.Error()
}

func (e *dummyError113Is) Is(target error) bool {
	return e.Err == target
}

func TestIs(t *testing.T) {
	errBase := errors.New("an error")
	errOther := errors.New("other error")
	errChained113Base := &dummyError113{errBase}
	errChained113Other := &dummyError113{errOther}
	errChained113Nil := &dummyError113{}
	errChained113IsBase := &dummyError113Is{errBase}
	errChained113IsOther := &dummyError113Is{errOther}
	errChained113IsNil := &dummyError113Is{}

	cases := []struct {
		err    error
		target error
		is     bool
	}{
		{nil, nil, true},
		{errBase, errBase, true},
		{errChained113Base, errBase, true},
		{errChained113IsBase, errBase, true},
		{errOther, errBase, false},
		{nil, errBase, false},
		{errBase, nil, false},
		{errChained113Other, errBase, false},
		{errChained113IsOther, errBase, false},
		{errChained113Nil, errBase, false},
		{errChained113IsNil, errBase, false},
	}

	for _, c := range cases {
		if Is(c.err, c.target) != c.is {
			if c.is {
				t.Errorf("Expected '%v' is '%v', but is not", c.err, c.target)
			} else {
				t.Errorf("Expected '%v' is not '%v', but is", c.err, c.target)
			}
		}
	}
}

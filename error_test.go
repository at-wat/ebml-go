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
	"testing"

	"github.com/at-wat/ebml-go/internal/errs"
)

type dummyError struct {
	Err error
}

func (e *dummyError) Error() string {
	return e.Err.Error()
}

func TestError(t *testing.T) {
	errBase := errors.New("an error")
	errOther := errors.New("an another error")
	errChained := wrapErrorf(errBase, "info")
	errDoubleChained := wrapErrorf(errChained, "info")
	errChainedNil := wrapErrorf(nil, "info")
	errChainedOther := wrapErrorf(errOther, "info")
	err112Chained := wrapErrorf(&dummyError{errBase}, "info")
	err112Nil := wrapErrorf(&dummyError{nil}, "info")
	errStr := "info: an error"

	t.Run("ErrorsIs", func(t *testing.T) {
		if !errs.Is(errChained, errBase) {
			t.Errorf("Wrapped error '%v' doesn't chain '%v'", errChained, errBase)
		}
	})

	t.Run("Is", func(t *testing.T) {
		if !errChained.(*Error).Is(errChained) {
			t.Errorf("Wrapped error '%v' doesn't match its-self", errChained)
		}
		if !errChained.(*Error).Is(errBase) {
			t.Errorf("Wrapped error '%v' doesn't match '%v'", errChained, errBase)
		}
		if !errDoubleChained.(*Error).Is(errBase) {
			t.Errorf("Wrapped error '%v' doesn't match '%v'", errDoubleChained, errBase)
		}
		if !err112Chained.(*Error).Is(errBase) {
			t.Errorf("Wrapped error '%v' doesn't match '%v'",
				err112Chained, errBase)
		}
		if !errChainedNil.(*Error).Is(nil) {
			t.Errorf("Nil chained error '%v' doesn't match 'nil'", errChainedNil)
		}

		if errChainedNil.(*Error).Is(errBase) {
			t.Errorf("Wrapped error '%v' unexpectedly matched '%v'",
				errChainedNil, errBase)
		}
		if errChainedOther.(*Error).Is(errBase) {
			t.Errorf("Wrapped error '%v' unexpectedly matched '%v'",
				errChainedOther, errBase)
		}
		if err112Nil.(*Error).Is(errBase) {
			t.Errorf("Wrapped error '%v' unexpectedly matched '%v'",
				errChainedOther, errBase)
		}
	})

	if errChained.Error() != errStr {
		t.Errorf("Error string expected: %s, got: %s", errStr, errChained.Error())
	}
	if errChained.(*Error).Unwrap() != errBase {
		t.Errorf("Unwrapped error expected: %s, got: %s", errBase, errChained.(*Error).Unwrap())
	}
}

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
	"fmt"
)

// Error records an EBML handling error.
type Error struct {
	Err     error
	Failure string
}

func (e *Error) Error() string {
	return e.Failure + ": " + e.Err.Error()
}

// Unwrap returns the base error.
func (e *Error) Unwrap() error {
	return e.Err
}

// Is reports whether chained error contains target.
//
// Deprecated: Only for API compatibility. Will be removed in the future release to rely only on Unwrap().
func (e *Error) Is(target error) bool {
	if e == target || e.Err == target {
		return true
	}
	if is, ok := e.Err.(interface{ Is(error) bool }); ok {
		return is.Is(target)
	}
	return false
}

func wrapError(err error, failure string) error {
	return &Error{
		Failure: failure,
		Err:     err,
	}
}

func wrapErrorf(err error, failureFmt string, v ...interface{}) error {
	return wrapError(err, fmt.Sprintf(failureFmt, v...))
}

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
	"reflect"
)

// Error records a failed parsing.
type Error struct {
	Err     error
	Failure string
}

func (e *Error) Error() string {
	// TODO: migrate to fmt.Sprintf %w once Go1.12 reaches EOL.
	return e.Failure + ": " + e.Err.Error()
}

// Unwrap returns the reason of the failure.
// This is for Go1.13 error unwrapping.
func (e *Error) Unwrap() error {
	return e.Err
}

// Is reports whether chained error contains target.
// This is for Go1.13 error unwrapping.
func (e *Error) Is(target error) bool {
	err := e.Err

	switch target {
	case e:
		return true
	case nil:
		return err == nil
	}
	for {
		switch err {
		case nil:
			return false
		case target:
			return true
		}
		x, ok := err.(interface{ Unwrap() error })
		if !ok {
			// Some stdlibs haven't have error unwrapper yet.
			// Check err.Err field if exposed.
			if reflect.TypeOf(err).Kind() == reflect.Ptr {
				e := reflect.ValueOf(err).Elem().FieldByName("Err")
				if e.IsValid() {
					e2, ok := e.Interface().(error)
					if !ok {
						return false
					}
					err = e2
					continue
				}
			}
			return false
		}
		err = x.Unwrap()
	}
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

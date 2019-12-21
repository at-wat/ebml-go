// +build !go1.13

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
	"reflect"
)

// Is compares error type. Works like Go1.13 errors.Is().
func Is(err, target error) bool {
	if target == nil {
		return err == nil
	}
	for {
		if err == nil {
			return false
		}
		if err == target {
			return true
		}
		x, ok := err.(interface{ Unwrap() error })
		if !ok {
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
		if err = x.Unwrap(); err == nil {
			return false
		}
	}
}

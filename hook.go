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

// Element represents an EBML element.
type Element struct {
	Value    interface{}
	Name     string
	Type     ElementType
	Position uint64
	Size     uint64
	Parent   *Element
}

func withElementMap(m map[string][]*Element) func(*Element) {
	return func(elem *Element) {
		key := elem.Name
		e := elem
		for {
			if e.Parent == nil {
				break
			}
			e = e.Parent
			key = fmt.Sprintf("%s.%s", e.Name, key)
		}
		elements, ok := m[key]
		if !ok {
			elements = make([]*Element, 0)
		}
		elements = append(elements, elem)
		m[key] = elements
	}
}

func elementPositionMap(m map[string][]*Element) map[string][]uint64 {
	pm := make(map[string][]uint64)
	for key, elements := range m {
		for _, e := range elements {
			pm[key] = append(pm[key], e.Position)
		}
	}
	return pm
}

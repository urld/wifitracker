// Copyright (c) 2016, David Url
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tracker

import "encoding/json"

// A Set is a unordered collection of unique string elements.
type Set struct {
	set map[string]bool
}

// Add an element to the Set.
func (s *Set) Add(element string) {
	if element != "" {
		if s.set == nil {
			s.set = make(map[string]bool)
		}
		s.set[element] = true
	}
}

// MarshalJSON returns the JSON encoding of the Set.
// The Set is converted to a JSON Array, and empty strings are ignored.
func (s *Set) MarshalJSON() ([]byte, error) {
	elements := make([]string, 0, len(s.set))
	for element := range s.set {
		if element != "" {
			elements = append(elements, element)
		}
	}
	setJSON, err := json.Marshal(elements)
	return setJSON, err
}

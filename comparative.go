/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package diff

import "reflect"

// Comparative ...
type Comparative struct {
	A, B *reflect.Value
}

// ComparativeList : stores indexed comparative
type ComparativeList map[interface{}]*Comparative

// NewComparativeList : returns a new comparative list
func NewComparativeList() *ComparativeList {
	cl := make(ComparativeList)
	return &cl
}

func (cl *ComparativeList) addA(k interface{}, v *reflect.Value) {
	if (*cl)[k] == nil {
		(*cl)[k] = &Comparative{}
	}
	(*cl)[k].A = v
}

func (cl *ComparativeList) addB(k interface{}, v *reflect.Value) {
	if (*cl)[k] == nil {
		(*cl)[k] = &Comparative{}
	}
	(*cl)[k].B = v
}

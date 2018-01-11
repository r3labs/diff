/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package diff

import (
	"reflect"
)

func (cl *Changelog) diffMap(path []string, a, b reflect.Value) error {
	if a.Kind() != b.Kind() {
		return ErrTypeMismatch
	}

	c := NewComparativeList()

	for _, k := range a.MapKeys() {
		ae := a.MapIndex(k)
		c.addA(k.Interface(), &ae)
	}

	for _, k := range b.MapKeys() {
		be := b.MapIndex(k)
		c.addB(k.Interface(), &be)
	}

	return cl.diffComparative(path, c)
}

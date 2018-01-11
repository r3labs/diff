/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package diff

import (
	"reflect"
)

func (cl *Changelog) diffComparative(path []string, c *ComparativeList) error {
	for k, v := range *c {
		fpath := append(path, idstring(k))

		if v.A != nil && v.B == nil {
			cl.add(DELETE, fpath, v.A.Interface(), nil)
		}

		if v.A == nil && v.B != nil {
			cl.add(CREATE, fpath, nil, v.B.Interface())
		}

		if v.A != nil && v.B != nil {
			err := cl.diff(fpath, *v.A, *v.B)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func comparative(a, b reflect.Value) bool {
	if a.Len() > 0 {
		ae := a.Index(0)
		ak := getFinalValue(ae)

		if ak.Kind() == reflect.Struct {
			if identifier(ak) != nil {
				return true
			}
		}
	}

	if b.Len() > 0 {
		be := b.Index(0)
		bk := getFinalValue(be)

		if bk.Kind() == reflect.Struct {
			if identifier(bk) != nil {
				return true
			}
		}
	}

	return false
}

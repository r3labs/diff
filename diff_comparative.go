/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package diff

import (
	"reflect"
)

func (cl *Changelog) diffComparative(path []string, c *ComparativeList) error {
	for _, k := range c.keys {
		fpath := append(path, idstring(k))

		if c.m[k].A != nil && c.m[k].B == nil {
			cl.add(DELETE, fpath, c.m[k].A.Interface(), nil)
		}

		if c.m[k].A == nil && c.m[k].B != nil {
			cl.add(CREATE, fpath, nil, c.m[k].B.Interface())
		}

		if c.m[k].A != nil && c.m[k].B != nil {
			err := cl.diff(fpath, *c.m[k].A, *c.m[k].B)
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

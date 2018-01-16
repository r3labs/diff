/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package diff

import "reflect"

func (cl *Changelog) diffStruct(path []string, a, b reflect.Value) error {
	if a.Kind() == reflect.Invalid {
		return cl.structValues(CREATE, path, b)
	}

	if b.Kind() == reflect.Invalid {
		return cl.structValues(DELETE, path, a)
	}

	for i := 0; i < a.NumField(); i++ {
		field := a.Type().Field(i)
		tname := tagName(field)

		if tname == "-" || hasTagOption(field, "immutable") {
			continue
		}

		af := a.Field(i)
		bf := b.FieldByName(field.Name)

		fpath := append(path, tname)

		err := cl.diff(fpath, af, bf)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cl *Changelog) structValues(t string, path []string, a reflect.Value) error {
	if t != CREATE && t != DELETE {
		return ErrInvalidChangeType
	}

	if a.Kind() == reflect.Ptr {
		a = reflect.Indirect(a)
	}

	if a.Kind() != reflect.Struct {
		return ErrTypeMismatch
	}

	x := reflect.New(a.Type()).Elem()

	for i := 0; i < a.NumField(); i++ {

		field := a.Type().Field(i)
		tname := tagName(field)

		if tname == "-" {
			continue
		}

		af := a.Field(i)
		xf := x.FieldByName(field.Name)

		fpath := append(path, tname)

		err := cl.diff(fpath, xf, af)
		if err != nil {
			return err
		}
	}

	for i := 0; i < len(*cl); i++ {
		(*cl)[i] = swapChange(t, (*cl)[i])
	}

	return nil
}

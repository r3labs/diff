/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package diff

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
)

var (
	// ErrTypeMismatch : Compared types do not match
	ErrTypeMismatch = errors.New("types do not match")
	// ErrInvalidChangeType : The specified change values are not unsupported
	ErrInvalidChangeType = errors.New("change type must be one of 'create' or 'delete'")
)

const (
	// CREATE : represents when an element has been added
	CREATE = "create"
	// UPDATE : represents when an element has been updated
	UPDATE = "update"
	// DELETE : represents when an element has been removed
	DELETE = "delete"
)

// Changelog : stores a list of changed items
type Changelog []Change

// Change : stores information about a changed item
type Change struct {
	Type string      `json:"type"`
	Path []string    `json:"path"`
	From interface{} `json:"from"`
	To   interface{} `json:"to"`
}

// Changed : returns true if both values differ
func Changed(a, b interface{}) bool {
	cl, _ := Diff(a, b)
	return len(cl) > 0
}

// Diff : returns a changelog of all mutated values from both
func Diff(a, b interface{}) (Changelog, error) {
	var cl Changelog

	return cl, cl.diff([]string{}, reflect.ValueOf(a), reflect.ValueOf(b))
}

// StructValues : gets all values from a struct
// values are stored as "created" or "deleted" entries in the changelog,
// depending on the change type specified
func StructValues(t string, s interface{}) (Changelog, error) {
	var cl Changelog

	if t != CREATE && t != DELETE {
		return cl, ErrInvalidChangeType
	}

	a := reflect.ValueOf(s)

	if a.Kind() != reflect.Struct {
		return cl, ErrTypeMismatch
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

		err := cl.diff([]string{tname}, xf, af)
		if err != nil {
			return cl, err
		}
	}

	for i := 0; i < len(cl); i++ {
		cl[i] = swapChange(t, cl[i])
	}

	return cl, nil
}

// Filter : filter changes based on path. Paths may contain valid regexp to match items
func (cl *Changelog) Filter(path []string) Changelog {
	var ncl Changelog

	for _, c := range *cl {
		if pathmatch(path, c.Path) {
			ncl = append(ncl, c)
		}
	}

	return ncl
}

func (cl *Changelog) diff(path []string, a, b reflect.Value) error {
	var err error

	if a.Kind() != b.Kind() {
		return errors.New("types do not match")
	}

	switch a.Kind() {
	case reflect.Struct:
		err = cl.diffStruct(path, a, b)
	case reflect.Slice:
		err = cl.diffSlice(path, a, b)
	case reflect.String:
		err = cl.diffString(path, a, b)
	case reflect.Bool:
		err = cl.diffBool(path, a, b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		err = cl.diffInt(path, a, b)
	case reflect.Map:
		err = cl.diffMap(path, a, b)
	case reflect.Ptr:
		err = cl.diffPtr(path, a, b)
	case reflect.Interface:
		err = cl.diffInterface(path, a, b)
	default:
		err = errors.New("unsupported type: " + a.Kind().String())
	}

	return err
}

func (cl *Changelog) add(t string, path []string, from, to interface{}) {
	(*cl) = append((*cl), Change{
		Type: t,
		Path: path,
		From: from,
		To:   to,
	})
}

func tag(v reflect.Value, i int) string {
	return v.Type().Field(i).Tag.Get("diff")
}

func tagName(f reflect.StructField) string {
	t := f.Tag.Get("diff")

	parts := strings.Split(t, ",")
	if len(parts) < 1 {
		return "-"
	}

	return parts[0]
}

func identifier(v reflect.Value) interface{} {
	for i := 0; i < v.NumField(); i++ {
		if hasTagOption(v.Type().Field(i), "identifier") {
			return v.Field(i).Interface()
		}
	}

	return nil
}

func hasTagOption(f reflect.StructField, opt string) bool {
	parts := strings.Split(f.Tag.Get("diff"), ",")
	if len(parts) < 2 {
		return false
	}

	for _, option := range parts[1:] {
		if option == opt {
			return true
		}
	}

	return false
}

func swapChange(t string, c Change) Change {
	nc := Change{
		Type: t,
		Path: c.Path,
	}

	switch t {
	case CREATE:
		nc.To = c.To
	case DELETE:
		nc.From = c.To
	}

	return nc
}

func idstring(v interface{}) string {
	switch v.(type) {
	case string:
		return v.(string)
	case int:
		return strconv.Itoa(v.(int))
	default:
		return ""
	}
}

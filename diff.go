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

// Changed ...
func Changed(a, b interface{}) bool {
	cl, _ := Diff(a, b)
	return len(cl) > 0
}

// Diff ...
func Diff(a, b interface{}) (Changelog, error) {
	var cl Changelog

	return cl, cl.diff([]string{}, reflect.ValueOf(a), reflect.ValueOf(b))
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
	case reflect.Int:
		err = cl.diffInt(path, a, b)
	case reflect.Map:
		err = cl.diffMap(path, a, b)
	case reflect.Ptr:
		err = cl.diffPtr(path, a, b)
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

func tagName(v reflect.Value, i int) string {
	t := tag(v, i)

	parts := strings.Split(t, ",")
	if len(parts) < 1 {
		return ""
	}

	return parts[0]
}

func identifier(v reflect.Value) interface{} {
	for i := 0; i < v.NumField(); i++ {
		t := tag(v, i)

		parts := strings.Split(t, ",")
		if len(parts) < 1 {
			continue
		}

		if parts[1] == "identifier" {
			return v.Field(i).Interface()
		}
	}

	return nil
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

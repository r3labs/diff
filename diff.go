/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package diff

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
)

const (
	// CREATE represents when an element has been added
	CREATE = "create"
	// UPDATE represents when an element has been updated
	UPDATE = "update"
	// DELETE represents when an element has been removed
	DELETE = "delete"
)

// DiffType represents an enum with all the supported diff types
type DiffType uint8

const (
	UNSUPPORTED DiffType = iota
	STRUCT
	SLICE
	ARRAY
	STRING
	BOOL
	INT
	UINT
	FLOAT
	MAP
	PTR
	INTERFACE
)

func (t DiffType) String() string {
	switch t {
	case STRUCT:
		return "STRUCT"
	case SLICE:
		return "SLICE"
	case ARRAY:
		return "ARRAY"
	case STRING:
		return "STRING"
	case BOOL:
		return "BOOL"
	case INT:
		return "INT"
	case UINT:
		return "UINT"
	case FLOAT:
		return "FLOAT"
	case MAP:
		return "MAP"
	case PTR:
		return "PTR"
	case INTERFACE:
		return "INTERFACE"
	default:
		return "UNSUPPORTED"
	}
}

// DiffFunc represents the built-in diff functions
type DiffFunc func([]string, reflect.Value, reflect.Value, interface{}) error

// Differ a configurable diff instance
type Differ struct {
	TagName                string
	SliceOrdering          bool
	DisableStructValues    bool
	customValueDiffers     []ValueDiffer
	cl                     Changelog
	AllowTypeMismatch      bool
	DiscardParent          bool
	StructMapKeys          bool
	FlattenEmbeddedStructs bool
	ConvertCompatibleTypes bool
	Filter                 FilterFunc
}

// Changelog stores a list of changed items
type Changelog []Change

// Change stores information about a changed item
type Change struct {
	Type   string      `json:"type"`
	Path   []string    `json:"path"`
	From   interface{} `json:"from"`
	To     interface{} `json:"to"`
	parent interface{} `json:"parent"`
}

// ValueDiffer is an interface for custom differs
type ValueDiffer interface {
	Match(a, b reflect.Value) bool
	Diff(dt DiffType, df DiffFunc, cl *Changelog, path []string, a, b reflect.Value, parent interface{}) error
	InsertParentDiffer(dfunc func(path []string, a, b reflect.Value, p interface{}) error)
}

// Changed returns true if both values differ
func Changed(a, b interface{}) bool {
	cl, _ := Diff(a, b)
	return len(cl) > 0
}

// Diff returns a changelog of all mutated values from both
func Diff(a, b interface{}, opts ...func(d *Differ) error) (Changelog, error) {
	d, err := NewDiffer(opts...)
	if err != nil {
		return nil, err
	}
	return d.Diff(a, b)
}

// NewDiffer creates a new configurable diffing object
func NewDiffer(opts ...func(d *Differ) error) (*Differ, error) {
	d := Differ{
		TagName:       "diff",
		DiscardParent: false,
	}

	for _, opt := range opts {
		err := opt(&d)
		if err != nil {
			return nil, err
		}
	}

	return &d, nil
}

// FilterFunc is a function that determines whether to descend into a struct field.
// parent is the struct being examined and field is a field on that struct. path
// is the path to the field from the root of the diff.
type FilterFunc func(path []string, parent reflect.Type, field reflect.StructField) bool

// StructValues gets all values from a struct
// values are stored as "created" or "deleted" entries in the changelog,
// depending on the change type specified
func StructValues(t string, path []string, s interface{}) (Changelog, error) {
	d := Differ{
		TagName:       "diff",
		DiscardParent: false,
	}

	v := reflect.ValueOf(s)

	return d.cl, d.structValues(t, path, v)
}

// FilterOut filter out the changes based on path. Paths may contain valid regexp to match items
func (cl *Changelog) FilterOut(path []string) Changelog {
	var ncl Changelog

	for _, c := range *cl {
		if !pathmatch(path, c.Path) {
			ncl = append(ncl, c)
		}
	}

	return ncl
}

// Filter filter changes based on path. Paths may contain valid regexp to match items
func (cl *Changelog) Filter(path []string) Changelog {
	var ncl Changelog

	for _, c := range *cl {
		if pathmatch(path, c.Path) {
			ncl = append(ncl, c)
		}
	}

	return ncl
}

func (d *Differ) getDiffType(a, b reflect.Value) (DiffType, DiffFunc) {
	switch {
	case are(a, b, reflect.Struct, reflect.Invalid):
		return STRUCT, d.diffStruct
	case are(a, b, reflect.Slice, reflect.Invalid):
		return SLICE, d.diffSlice
	case are(a, b, reflect.Array, reflect.Invalid):
		return ARRAY, d.diffSlice
	case are(a, b, reflect.String, reflect.Invalid):
		return STRING, d.diffString
	case are(a, b, reflect.Bool, reflect.Invalid):
		return BOOL, d.diffBool
	case are(a, b, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Invalid):
		return INT, d.diffInt
	case are(a, b, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Invalid):
		return UINT, d.diffUint
	case are(a, b, reflect.Float32, reflect.Float64, reflect.Invalid):
		return FLOAT, d.diffFloat
	case are(a, b, reflect.Map, reflect.Invalid):
		return MAP, d.diffMap
	case are(a, b, reflect.Ptr, reflect.Invalid):
		return PTR, d.diffPtr
	case are(a, b, reflect.Interface, reflect.Invalid):
		return INTERFACE, d.diffInterface
	default:
		return UNSUPPORTED, nil
	}
}

// Diff returns a changelog of all mutated values from both
func (d *Differ) Diff(a, b interface{}) (Changelog, error) {
	// reset the state of the diff
	d.cl = Changelog{}

	return d.cl, d.diff([]string{}, reflect.ValueOf(a), reflect.ValueOf(b), nil)
}

func (d *Differ) diff(path []string, a, b reflect.Value, parent interface{}) error {

	//look and see if we need to discard the parent
	if parent != nil {
		if d.DiscardParent || reflect.TypeOf(parent).Kind() != reflect.Struct {
			parent = nil
		}
	}

	// check if types match or are
	if invalid(a, b) {
		if d.AllowTypeMismatch {
			d.cl.Add(UPDATE, path, a.Interface(), b.Interface())
			return nil
		}
		return ErrTypeMismatch
	}

	// get the diff type and the corresponding built-int diff function to handle this type
	diffType, diffFunc := d.getDiffType(a, b)

	// first go through custom diff functions
	if len(d.customValueDiffers) > 0 {
		for _, vd := range d.customValueDiffers {
			if vd.Match(a, b) {
				err := vd.Diff(diffType, diffFunc, &d.cl, path, a, b, parent)
				if err != nil {
					return err
				}
				return nil
			}
		}
	}

	// then built-in diff functions
	if diffType == UNSUPPORTED {
		return errors.New("unsupported type: " + a.Kind().String())
	}

	return diffFunc(path, a, b, parent)
}

func (cl *Changelog) Add(t string, path []string, ftco ...interface{}) {
	change := Change{
		Type: t,
		Path: path,
		From: ftco[0],
		To:   ftco[1],
	}
	if len(ftco) > 2 {
		change.parent = ftco[2]
	}
	(*cl) = append((*cl), change)
}

func tagName(tag string, f reflect.StructField) string {
	t := f.Tag.Get(tag)

	parts := strings.Split(t, ",")
	if len(parts) < 1 {
		return "-"
	}

	return parts[0]
}

func identifier(tag string, v reflect.Value) interface{} {
	if v.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < v.NumField(); i++ {
		if hasTagOption(tag, v.Type().Field(i), "identifier") {
			return v.Field(i).Interface()
		}
	}

	return nil
}

func hasTagOption(tag string, f reflect.StructField, opt string) bool {
	parts := strings.Split(f.Tag.Get(tag), ",")
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

func idComplex(v interface{}) string {
	switch v := v.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	default:
		b, err := msgpack.Marshal(v)
		if err != nil {
			panic(err)
		}
		return string(b)
	}

}
func idstring(v interface{}) string {
	switch v := v.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	default:
		return fmt.Sprint(v)
	}
}

func invalid(a, b reflect.Value) bool {
	if a.Kind() == b.Kind() {
		return false
	}

	if a.Kind() == reflect.Invalid {
		return false
	}
	if b.Kind() == reflect.Invalid {
		return false
	}

	return true
}

func are(a, b reflect.Value, kinds ...reflect.Kind) bool {
	var amatch, bmatch bool

	for _, k := range kinds {
		if a.Kind() == k {
			amatch = true
		}
		if b.Kind() == k {
			bmatch = true
		}
	}

	return amatch && bmatch
}

func AreType(a, b reflect.Value, types ...reflect.Type) bool {
	var amatch, bmatch bool

	for _, t := range types {
		if a.Kind() != reflect.Invalid {
			if a.Type() == t {
				amatch = true
			}
		}
		if b.Kind() != reflect.Invalid {
			if b.Type() == t {
				bmatch = true
			}
		}
	}

	return amatch && bmatch
}

func copyAppend(src []string, elems ...string) []string {
	dst := make([]string, len(src)+len(elems))
	copy(dst, src)
	for i := len(src); i < len(src)+len(elems); i++ {
		dst[i] = elems[i-len(src)]
	}
	return dst
}

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package diff_test

import (
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zajko/diff/v3"
)

var currentTime = time.Now()

var struct1 = tmstruct{Bar: 1, Foo: "one"}
var struct2 = tmstruct{Bar: 2, Foo: "two"}

type tistruct struct {
	Name  string `diff:"name,identifier"`
	Value int    `diff:"value"`
}

type nestedstruct struct {
	Somestruct `diff:"tistruct,nestedIdentifier"`
}

type Somestruct struct {
	Name  string `diff:"name,identifier"`
	Value int    `diff:"value"`
}
type tuistruct struct {
	Value int `diff:"value"`
}

type tnstruct struct {
	Slice []tmstruct `diff:"slice"`
}

type tmstruct struct {
	Foo string `diff:"foo"`
	Bar int    `diff:"bar"`
}

type Embedded struct {
	Foo string `diff:"foo"`
	Bar int    `diff:"bar"`
}

type embedstruct struct {
	Embedded
	Baz bool `diff:"baz"`
}

type customTagStruct struct {
	Foo string `json:"foo"`
	Bar int    `json:"bar"`
}

type privateValueStruct struct {
	Public  string
	Private *sync.RWMutex
}

type privateMapStruct struct {
	set map[string]interface{}
}

type CustomStringType string
type CustomIntType int
type customTypeStruct struct {
	Foo CustomStringType `diff:"foo"`
	Bar CustomIntType    `diff:"bar"`
}

type tstruct struct {
	ID              string            `diff:"id,immutable"`
	Name            string            `diff:"name"`
	Value           int               `diff:"value"`
	Bool            bool              `diff:"bool"`
	Values          []string          `diff:"values"`
	Map             map[string]string `diff:"map"`
	Time            time.Time         `diff:"time"`
	Pointer         *string           `diff:"pointer"`
	Ignored         bool              `diff:"-"`
	Identifiables   []tistruct        `diff:"identifiables"`
	Unidentifiables []tuistruct       `diff:"unidentifiables"`
	Nested          tnstruct          `diff:"nested"`
	private         int               `diff:"private"`
}

func sptr(s string) *string {
	return &s
}

func TestDiff(t *testing.T) {
	cases := []struct {
		Name      string
		A, B      interface{}
		Changelog diff.Changelog
		Error     error
	}{
		{
			"uint-slice-insert", []uint{1, 2, 3}, []uint{1, 2, 3, 4},
			diff.Changelog{
				diff.Change{Type: diff.CREATE, Path: []string{"3"}, To: uint(4)},
			},
			nil,
		},
		{
			"uint-array-insert", [3]uint{1, 2, 3}, [4]uint{1, 2, 3, 4},
			diff.Changelog{
				diff.Change{Type: diff.CREATE, Path: []string{"3"}, To: uint(4)},
			},
			nil,
		},
		{
			"int-slice-insert", []int{1, 2, 3}, []int{1, 2, 3, 4},
			diff.Changelog{
				diff.Change{Type: diff.CREATE, Path: []string{"3"}, To: 4},
			},
			nil,
		},
		{
			"int-array-insert", [3]int{1, 2, 3}, [4]int{1, 2, 3, 4},
			diff.Changelog{
				diff.Change{Type: diff.CREATE, Path: []string{"3"}, To: 4},
			},
			nil,
		},
		{
			"uint-slice-delete", []uint{1, 2, 3}, []uint{1, 3},
			diff.Changelog{
				diff.Change{Type: diff.DELETE, Path: []string{"1"}, From: uint(2)},
			},
			nil,
		},
		{
			"uint-array-delete", [3]uint{1, 2, 3}, [2]uint{1, 3},
			diff.Changelog{
				diff.Change{Type: diff.DELETE, Path: []string{"1"}, From: uint(2)},
			},
			nil,
		},
		{
			"int-slice-delete", []int{1, 2, 3}, []int{1, 3},
			diff.Changelog{
				diff.Change{Type: diff.DELETE, Path: []string{"1"}, From: 2},
			},
			nil,
		},
		{
			"uint-slice-insert-delete", []uint{1, 2, 3}, []uint{1, 3, 4},
			diff.Changelog{
				diff.Change{Type: diff.DELETE, Path: []string{"1"}, From: uint(2)},
				diff.Change{Type: diff.CREATE, Path: []string{"2"}, To: uint(4)},
			},
			nil,
		},
		{
			"uint-slice-array-delete", [3]uint{1, 2, 3}, [3]uint{1, 3, 4},
			diff.Changelog{
				diff.Change{Type: diff.DELETE, Path: []string{"1"}, From: uint(2)},
				diff.Change{Type: diff.CREATE, Path: []string{"2"}, To: uint(4)},
			},
			nil,
		},
		{
			"int-slice-insert-delete", []int{1, 2, 3}, []int{1, 3, 4},
			diff.Changelog{
				diff.Change{Type: diff.DELETE, Path: []string{"1"}, From: 2},
				diff.Change{Type: diff.CREATE, Path: []string{"2"}, To: 4},
			},
			nil,
		},
		{
			"string-slice-insert", []string{"1", "2", "3"}, []string{"1", "2", "3", "4"},
			diff.Changelog{
				diff.Change{Type: diff.CREATE, Path: []string{"3"}, To: "4"},
			},
			nil,
		},
		{
			"string-array-insert", [3]string{"1", "2", "3"}, [4]string{"1", "2", "3", "4"},
			diff.Changelog{
				diff.Change{Type: diff.CREATE, Path: []string{"3"}, To: "4"},
			},
			nil,
		},
		{
			"string-slice-delete", []string{"1", "2", "3"}, []string{"1", "3"},
			diff.Changelog{
				diff.Change{Type: diff.DELETE, Path: []string{"1"}, From: "2"},
			},
			nil,
		},
		{
			"string-slice-delete", [3]string{"1", "2", "3"}, [2]string{"1", "3"},
			diff.Changelog{
				diff.Change{Type: diff.DELETE, Path: []string{"1"}, From: "2"},
			},
			nil,
		},
		{
			"string-slice-insert-delete", []string{"1", "2", "3"}, []string{"1", "3", "4"},
			diff.Changelog{
				diff.Change{Type: diff.DELETE, Path: []string{"1"}, From: "2"},
				diff.Change{Type: diff.CREATE, Path: []string{"2"}, To: "4"},
			},
			nil,
		},
		{
			"string-array-insert-delete", [3]string{"1", "2", "3"}, [3]string{"1", "3", "4"},
			diff.Changelog{
				diff.Change{Type: diff.DELETE, Path: []string{"1"}, From: "2"},
				diff.Change{Type: diff.CREATE, Path: []string{"2"}, To: "4"},
			},
			nil,
		},
		{
			"comparable-slice-insert", []tistruct{{"one", 1}}, []tistruct{{"one", 1}, {"two", 2}},
			diff.Changelog{
				diff.Change{Type: diff.CREATE, Path: []string{"two", "name"}, To: "two"},
				diff.Change{Type: diff.CREATE, Path: []string{"two", "value"}, To: 2},
			},
			nil,
		},
		{
			"comparable-array-insert", [1]tistruct{{"one", 1}}, [2]tistruct{{"one", 1}, {"two", 2}},
			diff.Changelog{
				diff.Change{Type: diff.CREATE, Path: []string{"two", "name"}, To: "two"},
				diff.Change{Type: diff.CREATE, Path: []string{"two", "value"}, To: 2},
			},
			nil,
		},
		{
			"comparable-slice-delete", []tistruct{{"one", 1}, {"two", 2}}, []tistruct{{"one", 1}},
			diff.Changelog{
				diff.Change{Type: diff.DELETE, Path: []string{"two", "name"}, From: "two"},
				diff.Change{Type: diff.DELETE, Path: []string{"two", "value"}, From: 2},
			},
			nil,
		},
		{
			"comparable-array-delete", [2]tistruct{{"one", 1}, {"two", 2}}, [1]tistruct{{"one", 1}},
			diff.Changelog{
				diff.Change{Type: diff.DELETE, Path: []string{"two", "name"}, From: "two"},
				diff.Change{Type: diff.DELETE, Path: []string{"two", "value"}, From: 2},
			},
			nil,
		},
		{
			"comparable-slice-update", []tistruct{{"one", 1}}, []tistruct{{"one", 50}},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"one", "value"}, From: 1, To: 50},
			},
			nil,
		},
		{
			"comparable-array-update", [1]tistruct{{"one", 1}}, [1]tistruct{{"one", 50}},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"one", "value"}, From: 1, To: 50},
			},
			nil,
		},
		{
			"map-slice-insert", []map[string]string{{"test": "123"}}, []map[string]string{{"test": "123", "tset": "456"}},
			diff.Changelog{
				diff.Change{Type: diff.CREATE, Path: []string{"0", "tset"}, To: "456"},
			},
			nil,
		},
		{
			"map-array-insert", [1]map[string]string{{"test": "123"}}, [1]map[string]string{{"test": "123", "tset": "456"}},
			diff.Changelog{
				diff.Change{Type: diff.CREATE, Path: []string{"0", "tset"}, To: "456"},
			},
			nil,
		},
		{
			"map-slice-update", []map[string]string{{"test": "123"}}, []map[string]string{{"test": "456"}},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"0", "test"}, From: "123", To: "456"},
			},
			nil,
		},
		{
			"map-array-update", [1]map[string]string{{"test": "123"}}, [1]map[string]string{{"test": "456"}},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"0", "test"}, From: "123", To: "456"},
			},
			nil,
		},
		{
			"map-slice-delete", []map[string]string{{"test": "123", "tset": "456"}}, []map[string]string{{"test": "123"}},
			diff.Changelog{
				diff.Change{Type: diff.DELETE, Path: []string{"0", "tset"}, From: "456"},
			},
			nil,
		},
		{
			"map-array-delete", [1]map[string]string{{"test": "123", "tset": "456"}}, [1]map[string]string{{"test": "123"}},
			diff.Changelog{
				diff.Change{Type: diff.DELETE, Path: []string{"0", "tset"}, From: "456"},
			},
			nil,
		},
		{
			"map-interface-slice-update", []map[string]interface{}{{"test": nil}}, []map[string]interface{}{{"test": "456"}},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"0", "test"}, From: nil, To: "456"},
			},
			nil,
		},
		{
			"map-interface-array-update", [1]map[string]interface{}{{"test": nil}}, [1]map[string]interface{}{{"test": "456"}},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"0", "test"}, From: nil, To: "456"},
			},
			nil,
		},
		{
			"map-nil", map[string]string{"one": "test"}, nil,
			diff.Changelog{
				diff.Change{Type: diff.DELETE, Path: []string{"\xa3one"}, From: "test", To: nil},
			},
			nil,
		},
		{
			"nil-map", nil, map[string]string{"one": "test"},
			diff.Changelog{
				diff.Change{Type: diff.CREATE, Path: []string{"\xa3one"}, From: nil, To: "test"},
			},
			nil,
		},
		{
			"nested-map-insert", map[string]map[string]string{"a": {"test": "123"}}, map[string]map[string]string{"a": {"test": "123", "tset": "456"}},
			diff.Changelog{
				diff.Change{Type: diff.CREATE, Path: []string{"a", "tset"}, To: "456"},
			},
			nil,
		},
		{
			"nested-map-interface-insert", map[string]map[string]interface{}{"a": {"test": "123"}}, map[string]map[string]interface{}{"a": {"test": "123", "tset": "456"}},
			diff.Changelog{
				diff.Change{Type: diff.CREATE, Path: []string{"a", "tset"}, To: "456"},
			},
			nil,
		},
		{
			"nested-map-update", map[string]map[string]string{"a": {"test": "123"}}, map[string]map[string]string{"a": {"test": "456"}},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"a", "test"}, From: "123", To: "456"},
			},
			nil,
		},
		{
			"nested-map-delete", map[string]map[string]string{"a": {"test": "123"}}, map[string]map[string]string{"a": {}},
			diff.Changelog{
				diff.Change{Type: diff.DELETE, Path: []string{"a", "test"}, From: "123", To: nil},
			},
			nil,
		},
		{
			"nested-slice-insert", map[string][]int{"a": {1, 2, 3}}, map[string][]int{"a": {1, 2, 3, 4}},
			diff.Changelog{
				diff.Change{Type: diff.CREATE, Path: []string{"a", "3"}, To: 4},
			},
			nil,
		},
		{
			"nested-array-insert", map[string][3]int{"a": {1, 2, 3}}, map[string][4]int{"a": {1, 2, 3, 4}},
			diff.Changelog{
				diff.Change{Type: diff.CREATE, Path: []string{"a", "3"}, To: 4},
			},
			nil,
		},
		{
			"nested-slice-update", map[string][]int{"a": {1, 2, 3}}, map[string][]int{"a": {1, 4, 3}},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"a", "1"}, From: 2, To: 4},
			},
			nil,
		},
		{
			"nested-array-update", map[string][3]int{"a": {1, 2, 3}}, map[string][3]int{"a": {1, 4, 3}},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"a", "1"}, From: 2, To: 4},
			},
			nil,
		},
		{
			"nested-slice-delete", map[string][]int{"a": {1, 2, 3}}, map[string][]int{"a": {1, 3}},
			diff.Changelog{
				diff.Change{Type: diff.DELETE, Path: []string{"a", "1"}, From: 2, To: nil},
			},
			nil,
		},
		{
			"nested-array-delete", map[string][3]int{"a": {1, 2, 3}}, map[string][2]int{"a": {1, 3}},
			diff.Changelog{
				diff.Change{Type: diff.DELETE, Path: []string{"a", "1"}, From: 2, To: nil},
			},
			nil,
		},
		{
			"struct-string-update", tstruct{Name: "one"}, tstruct{Name: "two"},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"name"}, From: "one", To: "two"},
			},
			nil,
		},
		{
			"struct-int-update", tstruct{Value: 1}, tstruct{Value: 50},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"value"}, From: 1, To: 50},
			},
			nil,
		},
		{
			"struct-bool-update", tstruct{Bool: true}, tstruct{Bool: false},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"bool"}, From: true, To: false},
			},
			nil,
		},
		{
			"struct-time-update", tstruct{}, tstruct{Time: currentTime},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"time"}, From: time.Time{}, To: currentTime},
			},
			nil,
		},
		{
			"struct-map-update", tstruct{Map: map[string]string{"test": "123"}}, tstruct{Map: map[string]string{"test": "456"}},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"map", "test"}, From: "123", To: "456"},
			},
			nil,
		},
		{
			"struct-string-pointer-update", tstruct{Pointer: sptr("test")}, tstruct{Pointer: sptr("test2")},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"pointer"}, From: "test", To: "test2"},
			},
			nil,
		},
		{
			"struct-nil-string-pointer-update", tstruct{Pointer: nil}, tstruct{Pointer: sptr("test")},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"pointer"}, From: nil, To: sptr("test")},
			},
			nil,
		},
		{
			"struct-generic-slice-insert", tstruct{Values: []string{"one"}}, tstruct{Values: []string{"one", "two"}},
			diff.Changelog{
				diff.Change{Type: diff.CREATE, Path: []string{"values", "1"}, From: nil, To: "two"},
			},
			nil,
		},
		{
			"struct-identifiable-slice-insert", tstruct{Identifiables: []tistruct{{"one", 1}}}, tstruct{Identifiables: []tistruct{{"one", 1}, {"two", 2}}},
			diff.Changelog{
				diff.Change{Type: diff.CREATE, Path: []string{"identifiables", "two", "name"}, From: nil, To: "two"},
				diff.Change{Type: diff.CREATE, Path: []string{"identifiables", "two", "value"}, From: nil, To: 2},
			},
			nil,
		},
		{
			"struct-generic-slice-delete", tstruct{Values: []string{"one", "two"}}, tstruct{Values: []string{"one"}},
			diff.Changelog{
				diff.Change{Type: diff.DELETE, Path: []string{"values", "1"}, From: "two", To: nil},
			},
			nil,
		},
		{
			"struct-identifiable-slice-delete", tstruct{Identifiables: []tistruct{{"one", 1}, {"two", 2}}}, tstruct{Identifiables: []tistruct{{"one", 1}}},
			diff.Changelog{
				diff.Change{Type: diff.DELETE, Path: []string{"identifiables", "two", "name"}, From: "two", To: nil},
				diff.Change{Type: diff.DELETE, Path: []string{"identifiables", "two", "value"}, From: 2, To: nil},
			},
			nil,
		},
		{
			"struct-unidentifiable-slice-insert-delete", tstruct{Unidentifiables: []tuistruct{{1}, {2}, {3}}}, tstruct{Unidentifiables: []tuistruct{{5}, {2}, {3}, {4}}},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"unidentifiables", "0", "value"}, From: 1, To: 5},
				diff.Change{Type: diff.CREATE, Path: []string{"unidentifiables", "3", "value"}, From: nil, To: 4},
			},
			nil,
		},
		{
			"struct-with-private-value", privateValueStruct{Public: "one", Private: new(sync.RWMutex)}, privateValueStruct{Public: "two", Private: new(sync.RWMutex)},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"Public"}, From: "one", To: "two"},
			},
			nil,
		},
		{
			"mismatched-values-struct-map", map[string]string{"test": "one"}, &tstruct{Identifiables: []tistruct{{"one", 1}}},
			diff.Changelog{},
			diff.ErrTypeMismatch,
		},
		{
			"omittable", tstruct{Ignored: false}, tstruct{Ignored: true},
			diff.Changelog{},
			nil,
		},
		{
			"slice", &tstruct{}, &tstruct{Nested: tnstruct{Slice: []tmstruct{{"one", 1}, {"two", 2}}}},
			diff.Changelog{
				diff.Change{Type: diff.CREATE, Path: []string{"nested", "slice", "0", "foo"}, From: nil, To: "one"},
				diff.Change{Type: diff.CREATE, Path: []string{"nested", "slice", "0", "bar"}, From: nil, To: 1},
				diff.Change{Type: diff.CREATE, Path: []string{"nested", "slice", "1", "foo"}, From: nil, To: "two"},
				diff.Change{Type: diff.CREATE, Path: []string{"nested", "slice", "1", "bar"}, From: nil, To: 2},
			},
			nil,
		},
		{
			"slice-duplicate-items", []int{1}, []int{1, 1},
			diff.Changelog{
				diff.Change{Type: diff.CREATE, Path: []string{"1"}, From: nil, To: 1},
			},
			nil,
		},
		{
			"mixed-slice-map", []map[string]interface{}{{"name": "name1", "type": []string{"null", "string"}}}, []map[string]interface{}{{"name": "name1", "type": []string{"null", "int"}}, {"name": "name2", "type": []string{"null", "string"}}},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"0", "type", "1"}, From: "string", To: "int"},
				diff.Change{Type: diff.CREATE, Path: []string{"1", "\xa4name"}, From: nil, To: "name2"},
				diff.Change{Type: diff.CREATE, Path: []string{"1", "\xa4type"}, From: nil, To: []string{"null", "string"}},
			},
			nil,
		},
		{
			"map-string-pointer-create",
			map[string]*tmstruct{"one": &struct1},
			map[string]*tmstruct{"one": &struct1, "two": &struct2},
			diff.Changelog{
				diff.Change{Type: diff.CREATE, Path: []string{"two", "foo"}, From: nil, To: "two"},
				diff.Change{Type: diff.CREATE, Path: []string{"two", "bar"}, From: nil, To: 2},
			},
			nil,
		},
		{
			"map-string-pointer-delete",
			map[string]*tmstruct{"one": &struct1, "two": &struct2},
			map[string]*tmstruct{"one": &struct1},
			diff.Changelog{
				diff.Change{Type: diff.DELETE, Path: []string{"two", "foo"}, From: "two", To: nil},
				diff.Change{Type: diff.DELETE, Path: []string{"two", "bar"}, From: 2, To: nil},
			},
			nil,
		},
		{
			"private-struct-field",
			tstruct{private: 1},
			tstruct{private: 4},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"private"}, From: int64(1), To: int64(4)},
			},
			nil,
		},
		{
			"embedded-struct-field",
			embedstruct{Embedded{Foo: "a", Bar: 2}, true},
			embedstruct{Embedded{Foo: "b", Bar: 3}, false},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"foo"}, From: "a", To: "b"},
				diff.Change{Type: diff.UPDATE, Path: []string{"bar"}, From: 2, To: 3},
				diff.Change{Type: diff.UPDATE, Path: []string{"baz"}, From: true, To: false},
			},
			nil,
		},
		{
			"custom-tags",
			customTagStruct{Foo: "abc", Bar: 3},
			customTagStruct{Foo: "def", Bar: 4},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"foo"}, From: "abc", To: "def"},
				diff.Change{Type: diff.UPDATE, Path: []string{"bar"}, From: 3, To: 4},
			},
			nil,
		},
		{
			"custom-types",
			customTypeStruct{Foo: "a", Bar: 1},
			customTypeStruct{Foo: "b", Bar: 2},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"foo"}, From: CustomStringType("a"), To: CustomStringType("b")},
				diff.Change{Type: diff.UPDATE, Path: []string{"bar"}, From: CustomIntType(1), To: CustomIntType(2)},
			},
			nil,
		},
		{
			"struct-private-map-create", privateMapStruct{set: map[string]interface{}{"1": struct{}{}, "2": struct{}{}}}, privateMapStruct{set: map[string]interface{}{"1": struct{}{}, "2": struct{}{}, "3": struct{}{}}},
			diff.Changelog{
				diff.Change{Type: diff.CREATE, Path: []string{"set", "3"}, From: nil, To: struct{}{}},
			},
			nil,
		},
		{
			"struct-private-map-delete", privateMapStruct{set: map[string]interface{}{"1": struct{}{}, "2": struct{}{}, "3": struct{}{}}}, privateMapStruct{set: map[string]interface{}{"1": struct{}{}, "2": struct{}{}}},
			diff.Changelog{
				diff.Change{Type: diff.DELETE, Path: []string{"set", "3"}, From: struct{}{}, To: nil},
			},
			nil,
		},
		{
			"struct-private-map-nil-values", privateMapStruct{set: map[string]interface{}{"1": nil, "2": nil}}, privateMapStruct{set: map[string]interface{}{"1": nil, "2": nil, "3": nil}},
			diff.Changelog{
				diff.Change{Type: diff.CREATE, Path: []string{"set", "3"}, From: nil, To: nil},
			},
			nil,
		},
		{
			"slice-of-struct-with-slice",
			[]tnstruct{{[]tmstruct{struct1, struct2}}, {[]tmstruct{struct2, struct2}}},
			[]tnstruct{{[]tmstruct{struct2, struct2}}, {[]tmstruct{struct2, struct1}}},
			diff.Changelog{},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {

			var options []func(d *diff.Differ) error
			switch tc.Name {
			case "mixed-slice-map", "nil-map", "map-nil":
				options = append(options, diff.StructMapKeySupport())
			case "embedded-struct-field":
				options = append(options, diff.FlattenEmbeddedStructs())
			case "custom-tags":
				options = append(options, diff.TagName("json"))
			}
			cl, err := diff.Diff(tc.A, tc.B, options...)

			assert.Equal(t, tc.Error, err)
			require.Equal(t, len(tc.Changelog), len(cl))

			for i, c := range cl {
				assert.Equal(t, tc.Changelog[i].Type, c.Type)
				assert.Equal(t, tc.Changelog[i].Path, c.Path)
				assert.Equal(t, tc.Changelog[i].From, c.From)
				assert.Equal(t, tc.Changelog[i].To, c.To)
			}
		})
	}
}

func TestDiffSliceOrdering(t *testing.T) {
	cases := []struct {
		Name      string
		A, B      interface{}
		Changelog diff.Changelog
		Error     error
	}{
		{
			"int-slice-insert-in-middle", []int{1, 2, 4}, []int{1, 2, 3, 4},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"2"}, From: 4, To: 3},
				diff.Change{Type: diff.CREATE, Path: []string{"3"}, To: 4},
			},
			nil,
		},
		{
			"int-slice-delete", []int{1, 2, 3}, []int{1, 3},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"1"}, From: 2, To: 3},
				diff.Change{Type: diff.DELETE, Path: []string{"2"}, From: 3},
			},
			nil,
		},
		{
			"int-slice-insert-delete", []int{1, 2, 3}, []int{1, 3, 4},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"1"}, From: 2, To: 3},
				diff.Change{Type: diff.UPDATE, Path: []string{"2"}, From: 3, To: 4},
			},
			nil,
		},
		{
			"int-slice-reorder", []int{1, 2, 3}, []int{1, 3, 2},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"1"}, From: 2, To: 3},
				diff.Change{Type: diff.UPDATE, Path: []string{"2"}, From: 3, To: 2},
			},
			nil,
		},
		{
			"string-slice-delete", []string{"1", "2", "3"}, []string{"1", "3"},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"1"}, From: "2", To: "3"},
				diff.Change{Type: diff.DELETE, Path: []string{"2"}, From: "3"},
			},
			nil,
		},
		{
			"string-slice-insert-delete", []string{"1", "2", "3"}, []string{"1", "3", "4"},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"1"}, From: "2", To: "3"},
				diff.Change{Type: diff.UPDATE, Path: []string{"2"}, From: "3", To: "4"},
			},
			nil,
		},
		{
			"string-slice-reorder", []string{"1", "2", "3"}, []string{"1", "3", "2"},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"1"}, From: "2", To: "3"},
				diff.Change{Type: diff.UPDATE, Path: []string{"2"}, From: "3", To: "2"},
			},
			nil,
		},
		{
			"nested-slice-delete", map[string][]int{"a": {1, 2, 3}}, map[string][]int{"a": {1, 3}},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"a", "1"}, From: 2, To: 3},
				diff.Change{Type: diff.DELETE, Path: []string{"a", "2"}, From: 3},
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			d, err := diff.NewDiffer(diff.SliceOrdering(true))
			require.Nil(t, err)
			cl, err := d.Diff(tc.A, tc.B)

			assert.Equal(t, tc.Error, err)
			require.Equal(t, len(tc.Changelog), len(cl))

			for i, c := range cl {
				assert.Equal(t, tc.Changelog[i].Type, c.Type)
				assert.Equal(t, tc.Changelog[i].Path, c.Path)
				assert.Equal(t, tc.Changelog[i].From, c.From)
				assert.Equal(t, tc.Changelog[i].To, c.To)
			}
		})
	}

}

func TestFilter(t *testing.T) {
	cases := []struct {
		Name     string
		Filter   []string
		Expected [][]string
	}{
		{"simple", []string{"item-1", "subitem"}, [][]string{{"item-1", "subitem"}}},
		{"regex", []string{"item-*"}, [][]string{{"item-1", "subitem"}, {"item-2", "subitem"}}},
	}

	cl := diff.Changelog{
		{Path: []string{"item-1", "subitem"}},
		{Path: []string{"item-2", "subitem"}},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			ncl := cl.Filter(tc.Filter)
			assert.Len(t, ncl, len(tc.Expected))
			for i, e := range tc.Expected {
				assert.Equal(t, e, ncl[i].Path)
			}
		})
	}
}

func TestFilterOut(t *testing.T) {
	cases := []struct {
		Name     string
		Filter   []string
		Expected [][]string
	}{
		{"simple", []string{"item-1", "subitem"}, [][]string{{"item-2", "subitem"}}},
		{"regex", []string{"item-*"}, [][]string{}},
	}

	cl := diff.Changelog{
		{Path: []string{"item-1", "subitem"}},
		{Path: []string{"item-2", "subitem"}},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			ncl := cl.FilterOut(tc.Filter)
			assert.Len(t, ncl, len(tc.Expected))
			for i, e := range tc.Expected {
				assert.Equal(t, e, ncl[i].Path)
			}
		})
	}
}

func TestStructValues(t *testing.T) {
	cases := []struct {
		Name       string
		ChangeType string
		X          interface{}
		Changelog  diff.Changelog
		Error      error
	}{
		{
			"struct-create", diff.CREATE, tstruct{ID: "xxxxx", Name: "something", Value: 1, Values: []string{"one", "two", "three"}},
			diff.Changelog{
				diff.Change{Type: diff.CREATE, Path: []string{"id"}, From: nil, To: "xxxxx"},
				diff.Change{Type: diff.CREATE, Path: []string{"name"}, From: nil, To: "something"},
				diff.Change{Type: diff.CREATE, Path: []string{"value"}, From: nil, To: 1},
				diff.Change{Type: diff.CREATE, Path: []string{"values", "0"}, From: nil, To: "one"},
				diff.Change{Type: diff.CREATE, Path: []string{"values", "1"}, From: nil, To: "two"},
				diff.Change{Type: diff.CREATE, Path: []string{"values", "2"}, From: nil, To: "three"},
			},
			nil,
		},
		{
			"struct-delete", diff.DELETE, tstruct{ID: "xxxxx", Name: "something", Value: 1, Values: []string{"one", "two", "three"}},
			diff.Changelog{
				diff.Change{Type: diff.DELETE, Path: []string{"id"}, From: "xxxxx", To: nil},
				diff.Change{Type: diff.DELETE, Path: []string{"name"}, From: "something", To: nil},
				diff.Change{Type: diff.DELETE, Path: []string{"value"}, From: 1, To: nil},
				diff.Change{Type: diff.DELETE, Path: []string{"values", "0"}, From: "one", To: nil},
				diff.Change{Type: diff.DELETE, Path: []string{"values", "1"}, From: "two", To: nil},
				diff.Change{Type: diff.DELETE, Path: []string{"values", "2"}, From: "three", To: nil},
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			cl, err := diff.StructValues(tc.ChangeType, []string{}, tc.X)

			assert.Equal(t, tc.Error, err)
			assert.Equal(t, len(tc.Changelog), len(cl))

			for i, c := range cl {
				assert.Equal(t, tc.Changelog[i].Type, c.Type)
				assert.Equal(t, tc.Changelog[i].Path, c.Path)
				assert.Equal(t, tc.Changelog[i].From, c.From)
				assert.Equal(t, tc.Changelog[i].To, c.To)
			}
		})
	}
}

func TestDifferReuse(t *testing.T) {
	d, err := diff.NewDiffer()
	require.Nil(t, err)

	cl, err := d.Diff([]string{"1", "2", "3"}, []string{"1"})
	require.Nil(t, err)

	require.Len(t, cl, 2)

	assert.Equal(t, "2", cl[0].From)
	assert.Equal(t, nil, cl[0].To)
	assert.Equal(t, "3", cl[1].From)
	assert.Equal(t, nil, cl[1].To)

	cl, err = d.Diff([]string{"a", "b"}, []string{"a", "c"})
	require.Nil(t, err)

	require.Len(t, cl, 1)

	assert.Equal(t, "b", cl[0].From)
	assert.Equal(t, "c", cl[0].To)
}

func TestDiffingOptions(t *testing.T) {
	d, err := diff.NewDiffer(diff.SliceOrdering(false))
	require.Nil(t, err)

	assert.False(t, d.SliceOrdering)

	cl, err := d.Diff([]int{1, 2, 3}, []int{1, 3, 2})
	require.Nil(t, err)

	assert.Len(t, cl, 0)

	d, err = diff.NewDiffer(diff.SliceOrdering(true))
	require.Nil(t, err)

	assert.True(t, d.SliceOrdering)

	cl, err = d.Diff([]int{1, 2, 3}, []int{1, 3, 2})
	require.Nil(t, err)

	assert.Len(t, cl, 2)

	// some other options..
}

func TestDiffPrivateField(t *testing.T) {
	cl, err := diff.Diff(tstruct{private: 1}, tstruct{private: 3})
	require.Nil(t, err)
	assert.Len(t, cl, 1)
}

type testType string
type testTypeDiffer struct {
	DiffFunc (func(path []string, a, b reflect.Value, p interface{}) error)
}

func (o *testTypeDiffer) InsertParentDiffer(dfunc func(path []string, a, b reflect.Value, p interface{}) error) {
	o.DiffFunc = dfunc
}

func (o *testTypeDiffer) Match(a, b reflect.Value) bool {
	return diff.AreType(a, b, reflect.TypeOf(testType("")))
}
func (o *testTypeDiffer) Diff(dt diff.DiffType, df diff.DiffFunc, cl *diff.Changelog, path []string, a, b reflect.Value, parent interface{}) error {
	if a.String() != "custom" && b.String() != "match" {
		cl.Add(diff.UPDATE, path, a.Interface(), b.Interface())
	}
	return nil
}

func TestCustomDiffer(t *testing.T) {
	type custom struct {
		T testType
	}

	d, err := diff.NewDiffer(
		diff.CustomValueDiffers(
			&testTypeDiffer{},
		),
	)
	require.Nil(t, err)

	cl, err := d.Diff(custom{"custom"}, custom{"match"})
	require.Nil(t, err)

	assert.Len(t, cl, 0)

	d, err = diff.NewDiffer(
		diff.CustomValueDiffers(
			&testTypeDiffer{},
		),
	)
	require.Nil(t, err)

	cl, err = d.Diff(custom{"same"}, custom{"same"})
	require.Nil(t, err)

	assert.Len(t, cl, 1)
}

type testStringInterceptorDiffer struct {
	DiffFunc (func(path []string, a, b reflect.Value, p interface{}) error)
}

func (o *testStringInterceptorDiffer) InsertParentDiffer(dfunc func(path []string, a, b reflect.Value, p interface{}) error) {
	o.DiffFunc = dfunc
}

func (o *testStringInterceptorDiffer) Match(a, b reflect.Value) bool {
	return diff.AreType(a, b, reflect.TypeOf(testType("")))
}
func (o *testStringInterceptorDiffer) Diff(dt diff.DiffType, df diff.DiffFunc, cl *diff.Changelog, path []string, a, b reflect.Value, parent interface{}) error {
	if dt.String() == "STRING" {
		// intercept the data
		aValue, aOk := a.Interface().(testType)
		bValue, bOk := b.Interface().(testType)

		if aOk && bOk {
			if aValue == "avalue" {
				aValue = testType(strings.ToUpper(string(aValue)))
				a = reflect.ValueOf(aValue)
			}

			if bValue == "bvalue" {
				bValue = testType(strings.ToUpper(string(aValue)))
				b = reflect.ValueOf(bValue)
			}
		}
	}

	// continue the diff logic passing the updated a/b values
	return df(path, a, b, parent)
}

func TestStringInterceptorDiffer(t *testing.T) {
	d, err := diff.NewDiffer(
		diff.CustomValueDiffers(
			&testStringInterceptorDiffer{},
		),
	)
	require.Nil(t, err)

	cl, err := d.Diff(testType("avalue"), testType("bvalue"))
	require.Nil(t, err)

	assert.Len(t, cl, 0)
}

type RecursiveTestStruct struct {
	Id       int
	Children []RecursiveTestStruct
}

type recursiveTestStructDiffer struct {
	DiffFunc (func(path []string, a, b reflect.Value, p interface{}) error)
}

func (o *recursiveTestStructDiffer) InsertParentDiffer(dfunc func(path []string, a, b reflect.Value, p interface{}) error) {
	o.DiffFunc = dfunc
}

func (o *recursiveTestStructDiffer) Match(a, b reflect.Value) bool {
	return diff.AreType(a, b, reflect.TypeOf(RecursiveTestStruct{}))
}

func (o *recursiveTestStructDiffer) Diff(dt diff.DiffType, df diff.DiffFunc, cl *diff.Changelog, path []string, a, b reflect.Value, parent interface{}) error {
	if a.Kind() == reflect.Invalid {
		cl.Add(diff.CREATE, path, nil, b.Interface())
		return nil
	}
	if b.Kind() == reflect.Invalid {
		cl.Add(diff.DELETE, path, a.Interface(), nil)
		return nil
	}
	var awt, bwt RecursiveTestStruct
	awt, _ = a.Interface().(RecursiveTestStruct)
	bwt, _ = b.Interface().(RecursiveTestStruct)
	if awt.Id != bwt.Id {
		cl.Add(diff.UPDATE, path, a.Interface(), b.Interface())
	}
	for i := 0; i < a.NumField(); i++ {
		field := a.Type().Field(i)
		tname := field.Name
		if tname != "Children" {
			continue
		}
		af := a.Field(i)
		bf := b.FieldByName(field.Name)
		fpath := copyAppend(path, tname)
		err := o.DiffFunc(fpath, af, bf, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestRecursiveCustomDiffer(t *testing.T) {
	treeA := RecursiveTestStruct{
		Id:       1,
		Children: []RecursiveTestStruct{},
	}

	treeB := RecursiveTestStruct{
		Id: 1,
		Children: []RecursiveTestStruct{
			{
				Id:       4,
				Children: []RecursiveTestStruct{},
			},
		},
	}
	d, err := diff.NewDiffer(
		diff.CustomValueDiffers(
			&recursiveTestStructDiffer{},
		),
	)
	require.Nil(t, err)
	cl, err := d.Diff(treeA, treeB)
	require.Nil(t, err)
	assert.Len(t, cl, 1)
}

func TestHandleDifferentTypes(t *testing.T) {
	cases := []struct {
		Name               string
		A, B               interface{}
		Changelog          diff.Changelog
		Error              error
		HandleTypeMismatch bool
	}{
		{
			"type-change-not-allowed-error",
			1, "1",
			nil,
			diff.ErrTypeMismatch,
			false,
		},
		{
			"type-change-not-allowed-error-struct",
			struct {
				p1 string
				p2 int
			}{"1", 1},
			struct {
				p1 string
				p2 string
			}{"1", "1"},
			nil,
			diff.ErrTypeMismatch,
			false,
		},
		{
			"type-change-allowed",
			1, "1",
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{}, From: 1, To: "1"},
			},
			nil,
			true,
		},
		{
			"type-change-allowed-struct",
			struct {
				P1 string
				P2 int
				P3 map[string]string
			}{"1", 1, map[string]string{"1": "1"}},
			struct {
				P1 string
				P2 string
				P3 string
			}{"1", "1", "1"},
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"P2"}, From: 1, To: "1"},
				diff.Change{Type: diff.UPDATE, Path: []string{"P3"}, From: map[string]string{"1": "1"}, To: "1"},
			},
			nil,
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			d, err := diff.NewDiffer(diff.AllowTypeMismatch(tc.HandleTypeMismatch))
			require.Nil(t, err)
			cl, err := d.Diff(tc.A, tc.B)

			assert.Equal(t, tc.Error, err)
			require.Equal(t, len(tc.Changelog), len(cl))

			for i, c := range cl {
				assert.Equal(t, tc.Changelog[i].Type, c.Type)
				assert.Equal(t, tc.Changelog[i].Path, c.Path)
				assert.Equal(t, tc.Changelog[i].From, c.From)
				assert.Equal(t, tc.Changelog[i].To, c.To)
			}
		})
	}
}

func TestNestedIdentifier_Equal(t *testing.T) {
	a := nestedstruct{
		Somestruct{
			Name:  "abc",
			Value: 100,
		},
	}
	b := nestedstruct{
		Somestruct{
			Name:  "abc2",
			Value: 101,
		},
	}
	d, err := diff.NewDiffer()
	require.Nil(t, err)
	slice1 := []nestedstruct{a, b}
	slice2 := []nestedstruct{b, a}
	cl, err := d.Diff(slice1, slice2)
	require.Nil(t, err)
	require.Empty(t, cl)
}

func TestNestedIdentifier_FindsAppropriateChanges(t *testing.T) {
	a := nestedstruct{
		Somestruct{
			Name:  "abc",
			Value: 100,
		},
	}
	b := nestedstruct{
		Somestruct{
			Name:  "abc2",
			Value: 101,
		},
	}
	a2 := nestedstruct{
		Somestruct{
			Name:  "abc",
			Value: 103,
		},
	}
	b2 := nestedstruct{
		Somestruct{
			Name:  "abc2",
			Value: 104,
		},
	}
	d, err := diff.NewDiffer()
	require.Nil(t, err)
	slice1 := []nestedstruct{a, b}
	slice2 := []nestedstruct{b2, a2}
	cl, err := d.Diff(slice1, slice2)
	require.Nil(t, err)
	require.Equal(t, len(cl), 2)
}

func copyAppend(src []string, elems ...string) []string {
	dst := make([]string, len(src)+len(elems))
	copy(dst, src)
	for i := len(src); i < len(src)+len(elems); i++ {
		dst[i] = elems[i-len(src)]
	}
	return dst
}

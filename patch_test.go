package diff

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPatch(t *testing.T) {
	cases := []struct {
		Name      string
		A, B      interface{}
		Changelog Changelog
		Error     error
	}{
		{
			"uint-slice-insert", &[]uint{1, 2, 3}, &[]uint{1, 2, 3, 4},
			Changelog{
				Change{Type: CREATE, Path: []string{"3"}, To: uint(4)},
			},
			nil,
		},
		{
			"int-slice-insert", &[]int{1, 2, 3}, &[]int{1, 2, 3, 4},
			Changelog{
				Change{Type: CREATE, Path: []string{"3"}, To: 4},
			},
			nil,
		},
		{
			"uint-slice-delete", &[]uint{1, 2, 3}, &[]uint{1, 3},
			Changelog{
				Change{Type: DELETE, Path: []string{"1"}, From: uint(2)},
			},
			nil,
		},
		{
			"int-slice-delete", &[]int{1, 2, 3}, &[]int{1, 3},
			Changelog{
				Change{Type: DELETE, Path: []string{"1"}, From: 2},
			},
			nil,
		},
		{
			"uint-slice-insert-delete", &[]uint{1, 2, 3}, &[]uint{1, 3, 4},
			Changelog{
				Change{Type: DELETE, Path: []string{"1"}, From: uint(2)},
				Change{Type: CREATE, Path: []string{"2"}, To: uint(4)},
			},
			nil,
		},
		{
			"int-slice-insert-delete", &[]int{1, 2, 3}, &[]int{1, 3, 4},
			Changelog{
				Change{Type: DELETE, Path: []string{"1"}, From: 2},
				Change{Type: CREATE, Path: []string{"2"}, To: 4},
			},
			nil,
		},
		{
			"string-slice-insert", &[]string{"1", "2", "3"}, &[]string{"1", "2", "3", "4"},
			Changelog{
				Change{Type: CREATE, Path: []string{"3"}, To: "4"},
			},
			nil,
		},
		{
			"string-slice-delete", &[]string{"1", "2", "3"}, &[]string{"1", "3"},
			Changelog{
				Change{Type: DELETE, Path: []string{"1"}, From: "2"},
			},
			nil,
		},
		{
			"string-slice-insert-delete", &[]string{"1", "2", "3"}, &[]string{"1", "3", "4"},
			Changelog{
				Change{Type: DELETE, Path: []string{"1"}, From: "2"},
				Change{Type: CREATE, Path: []string{"2"}, To: "4"},
			},
			nil,
		},
		// {
		// 	"comparable-slice-insert", &[]tistruct{{"one", 1}}, &[]tistruct{{"one", 1}, {"two", 2}},
		// 	Changelog{
		// 		Change{Type: CREATE, Path: []string{"two", "name"}, To: "two"},
		// 		Change{Type: CREATE, Path: []string{"two", "value"}, To: 2},
		// 	},
		// 	nil,
		// },
		// {
		// 	"comparable-slice-delete", &[]tistruct{{"one", 1}, {"two", 2}}, &[]tistruct{{"one", 1}},
		// 	Changelog{
		// 		Change{Type: DELETE, Path: []string{"two", "name"}, From: "two"},
		// 		Change{Type: DELETE, Path: []string{"two", "value"}, From: 2},
		// 	},
		// 	nil,
		// },
		{
			"comparable-slice-update", &[]tistruct{{"one", 1}}, &[]tistruct{{"one", 50}},
			Changelog{
				Change{Type: UPDATE, Path: []string{"one", "value"}, From: 1, To: 50},
			},
			nil,
		},
		// {
		// 	"map-slice-insert", &[]map[string]string{{"test": "123"}}, &[]map[string]string{{"test": "123", "tset": "456"}},
		// 	Changelog{
		// 		Change{Type: CREATE, Path: []string{"0", "tset"}, To: "456"},
		// 	},
		// 	nil,
		// },
		// {
		// 	"map-slice-update", &[]map[string]string{{"test": "123"}}, &[]map[string]string{{"test": "456"}},
		// 	Changelog{
		// 		Change{Type: UPDATE, Path: []string{"0", "test"}, From: "123", To: "456"},
		// 	},
		// 	nil,
		// },
		// {
		// 	"map-slice-delete", &[]map[string]string{{"test": "123", "tset": "456"}}, &[]map[string]string{{"test": "123"}},
		// 	Changelog{
		// 		Change{Type: DELETE, Path: []string{"0", "tset"}, From: "456"},
		// 	},
		// 	nil,
		// },
		// {
		// 	"map-interface-slice-update", &[]map[string]interface{}{{"test": nil}}, &[]map[string]interface{}{{"test": "456"}},
		// 	Changelog{
		// 		Change{Type: UPDATE, Path: []string{"0", "test"}, From: nil, To: "456"},
		// 	},
		// 	nil,
		// },
		// {
		// 	"map-nil", &map[string]string{"one": "test"}, nil,
		// 	Changelog{
		// 		Change{Type: DELETE, Path: []string{"\xa3one"}, From: "test", To: nil},
		// 	},
		// 	nil,
		// },
		// {
		// 	"nil-map", nil, &map[string]string{"one": "test"},
		// 	Changelog{
		// 		Change{Type: CREATE, Path: []string{"\xa3one"}, From: nil, To: "test"},
		// 	},
		// 	nil,
		// },
		// {
		// 	"nested-map-insert", &map[string]map[string]string{"a": {"test": "123"}}, &map[string]map[string]string{"a": {"test": "123", "tset": "456"}},
		// 	Changelog{
		// 		Change{Type: CREATE, Path: []string{"a", "tset"}, To: "456"},
		// 	},
		// 	nil,
		// },
		// {
		// 	"nested-map-interface-insert", &map[string]map[string]interface{}{"a": {"test": "123"}}, &map[string]map[string]interface{}{"a": {"test": "123", "tset": "456"}},
		// 	Changelog{
		// 		Change{Type: CREATE, Path: []string{"a", "tset"}, To: "456"},
		// 	},
		// 	nil,
		// },
		// {
		// 	"nested-map-update", &map[string]map[string]string{"a": {"test": "123"}}, &map[string]map[string]string{"a": {"test": "456"}},
		// 	Changelog{
		// 		Change{Type: UPDATE, Path: []string{"a", "test"}, From: "123", To: "456"},
		// 	},
		// 	nil,
		// },
		// {
		// 	"nested-map-delete", &map[string]map[string]string{"a": {"test": "123"}}, &map[string]map[string]string{"a": {}},
		// 	Changelog{
		// 		Change{Type: DELETE, Path: []string{"a", "test"}, From: "123", To: nil},
		// 	},
		// 	nil,
		// },
		// {
		// 	"nested-slice-insert", &map[string][]int{"a": {1, 2, 3}}, &map[string][]int{"a": {1, 2, 3, 4}},
		// 	Changelog{
		// 		Change{Type: CREATE, Path: []string{"a", "3"}, To: 4},
		// 	},
		// 	nil,
		// },
		// {
		// 	"nested-slice-update", &map[string][]int{"a": {1, 2, 3}}, &map[string][]int{"a": {1, 4, 3}},
		// 	Changelog{
		// 		Change{Type: UPDATE, Path: []string{"a", "1"}, From: 2, To: 4},
		// 	},
		// 	nil,
		// },
		// {
		// 	"nested-slice-delete", &map[string][]int{"a": {1, 2, 3}}, &map[string][]int{"a": {1, 3}},
		// 	Changelog{
		// 		Change{Type: DELETE, Path: []string{"a", "1"}, From: 2, To: nil},
		// 	},
		// 	nil,
		// },
		{
			"struct-string-update", &tstruct{Name: "one"}, &tstruct{Name: "two"},
			Changelog{
				Change{Type: UPDATE, Path: []string{"name"}, From: "one", To: "two"},
			},
			nil,
		},
		{
			"struct-int-update", &tstruct{Value: 1}, &tstruct{Value: 50},
			Changelog{
				Change{Type: UPDATE, Path: []string{"value"}, From: 1, To: 50},
			},
			nil,
		},
		{
			"struct-bool-update", &tstruct{Bool: true}, &tstruct{Bool: false},
			Changelog{
				Change{Type: UPDATE, Path: []string{"bool"}, From: true, To: false},
			},
			nil,
		},
		{
			"struct-time-update", &tstruct{}, &tstruct{Time: currentTime},
			Changelog{
				Change{Type: UPDATE, Path: []string{"time"}, From: time.Time{}, To: currentTime},
			},
			nil,
		},
		// {
		// 	"struct-map-update", &tstruct{Map: map[string]string{"test": "123"}}, &tstruct{Map: map[string]string{"test": "456"}},
		// 	Changelog{
		// 		Change{Type: UPDATE, Path: []string{"map", "test"}, From: "123", To: "456"},
		// 	},
		// 	nil,
		// },
		// {
		// 	"struct-string-pointer-update", &tstruct{Pointer: sptr("test")}, &tstruct{Pointer: sptr("test2")},
		// 	Changelog{
		// 		Change{Type: UPDATE, Path: []string{"pointer"}, From: "test", To: "test2"},
		// 	},
		// 	nil,
		// },
		{
			"struct-nil-string-pointer-update", &tstruct{Pointer: nil}, &tstruct{Pointer: sptr("test")},
			Changelog{
				Change{Type: UPDATE, Path: []string{"pointer"}, From: nil, To: sptr("test")},
			},
			nil,
		},
		{
			"struct-generic-slice-insert", &tstruct{Values: []string{"one"}}, &tstruct{Values: []string{"one", "two"}},
			Changelog{
				Change{Type: CREATE, Path: []string{"values", "1"}, From: nil, To: "two"},
			},
			nil,
		},
		// {
		// 	"struct-identifiable-slice-insert", &tstruct{Identifiables: []tistruct{{"one", 1}}}, &tstruct{Identifiables: []tistruct{{"one", 1}, {"two", 2}}},
		// 	Changelog{
		// 		Change{Type: CREATE, Path: []string{"identifiables", "two", "name"}, From: nil, To: "two"},
		// 		Change{Type: CREATE, Path: []string{"identifiables", "two", "value"}, From: nil, To: 2},
		// 	},
		// 	nil,
		// },
		{
			"struct-generic-slice-delete", &tstruct{Values: []string{"one", "two"}}, &tstruct{Values: []string{"one"}},
			Changelog{
				Change{Type: DELETE, Path: []string{"values", "1"}, From: "two", To: nil},
			},
			nil,
		},
		// {
		// 	"struct-identifiable-slice-delete", &tstruct{Identifiables: []tistruct{{"one", 1}, {"two", 2}}}, &tstruct{Identifiables: []tistruct{{"one", 1}}},
		// 	Changelog{
		// 		Change{Type: DELETE, Path: []string{"identifiables", "two", "name"}, From: "two", To: nil},
		// 		Change{Type: DELETE, Path: []string{"identifiables", "two", "value"}, From: 2, To: nil},
		// 	},
		// 	nil,
		// },
		{
			"struct-unidentifiable-slice-insert-delete", &tstruct{Unidentifiables: []tuistruct{{1}, {2}, {3}}}, &tstruct{Unidentifiables: []tuistruct{{5}, {2}, {3}, {4}}},
			Changelog{
				Change{Type: UPDATE, Path: []string{"unidentifiables", "0", "value"}, From: 1, To: 5},
				Change{Type: CREATE, Path: []string{"unidentifiables", "3", "value"}, From: nil, To: 4},
			},
			nil,
		},
		{
			"slice", &tstruct{}, &tstruct{Nested: tnstruct{Slice: []tmstruct{{"one", 1}, {"two", 2}}}},
			Changelog{
				Change{Type: CREATE, Path: []string{"nested", "slice", "0", "foo"}, From: nil, To: "one"},
				Change{Type: CREATE, Path: []string{"nested", "slice", "0", "bar"}, From: nil, To: 1},
				Change{Type: CREATE, Path: []string{"nested", "slice", "1", "foo"}, From: nil, To: "two"},
				Change{Type: CREATE, Path: []string{"nested", "slice", "1", "bar"}, From: nil, To: 2},
			},
			nil,
		},
		{
			"slice-duplicate-items", &[]int{1}, &[]int{1, 1},
			Changelog{
				Change{Type: CREATE, Path: []string{"1"}, From: nil, To: 1},
			},
			nil,
		},
		// {
		// 	"mixed-slice-map", &[]map[string]interface{}{{"name": "name1", "type": []string{"null", "string"}}}, &[]map[string]interface{}{{"name": "name1", "type": []string{"null", "int"}}, {"name": "name2", "type": []string{"null", "string"}}},
		// 	Changelog{
		// 		Change{Type: UPDATE, Path: []string{"0", "type", "1"}, From: "string", To: "int"},
		// 		Change{Type: CREATE, Path: []string{"1", "\xa4name"}, From: nil, To: "name2"},
		// 		Change{Type: CREATE, Path: []string{"1", "\xa4type"}, From: nil, To: []string{"null", "string"}},
		// 	},
		// 	nil,
		// },
		// {
		// 	"map-string-pointer-create",
		// 	&map[string]*tmstruct{"one": &struct1},
		// 	&map[string]*tmstruct{"one": &struct1, "two": &struct2},
		// 	Changelog{
		// 		Change{Type: CREATE, Path: []string{"two", "foo"}, From: nil, To: "two"},
		// 		Change{Type: CREATE, Path: []string{"two", "bar"}, From: nil, To: 2},
		// 	},
		// 	nil,
		// },
		// {
		// 	"map-string-pointer-delete",
		// 	&map[string]*tmstruct{"one": &struct1, "two": &struct2},
		// 	&map[string]*tmstruct{"one": &struct1},
		// 	Changelog{
		// 		Change{Type: DELETE, Path: []string{"two", "foo"}, From: "two", To: nil},
		// 		Change{Type: DELETE, Path: []string{"two", "bar"}, From: 2, To: nil},
		// 	},
		// 	nil,
		// },
		// {
		// 	"private-struct-field",
		// 	&tstruct{private: 1},
		// 	&tstruct{private: 4},
		// 	Changelog{
		// 		Change{Type: UPDATE, Path: []string{"private"}, From: int64(1), To: int64(4)},
		// 	},
		// 	nil,
		// },
		{
			"embedded-struct-field",
			&embedstruct{Tmstruct{Foo: "a", Bar: 2}, true},
			&embedstruct{Tmstruct{Foo: "b", Bar: 3}, false},
			Changelog{
				Change{Type: UPDATE, Path: []string{"foo"}, From: "a", To: "b"},
				Change{Type: UPDATE, Path: []string{"bar"}, From: 2, To: 3},
				Change{Type: UPDATE, Path: []string{"baz"}, From: true, To: false},
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {

			var options []func(d *Differ) error
			switch tc.Name {
			case "mixed-slice-map", "nil-map", "map-nil":
				options = append(options, StructMapKeySupport())
			}
			d, err := NewDiffer(options...)
			if err != nil {
				panic(err)
			}
			pl := d.Patch(tc.Changelog, &tc.A)

			assert.Equal(t, tc.B, tc.A)
			require.Equal(t, len(tc.Changelog), len(pl))
		})
	}
}

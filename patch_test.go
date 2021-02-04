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
		{
			"comparable-slice-update", &[]tistruct{{"one", 1}}, &[]tistruct{{"one", 50}},
			Changelog{
				Change{Type: UPDATE, Path: []string{"one", "value"}, From: 1, To: 50},
			},
			nil,
		},
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
		{
			"struct-generic-slice-delete", &tstruct{Values: []string{"one", "two"}}, &tstruct{Values: []string{"one"}},
			Changelog{
				Change{Type: DELETE, Path: []string{"values", "1"}, From: "two", To: nil},
			},
			nil,
		},
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
		{
			"embedded-struct-field",
			&embedstruct{Embedded{Foo: "a", Bar: 2}, true},
			&embedstruct{Embedded{Foo: "b", Bar: 3}, false},
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
			case "embedded-struct-field":
				options = append(options, FlattenEmbeddedStructs(true))
			}
			d, err := NewDiffer(options...)
			if err != nil {
				panic(err)
			}
			pl := d.Patch(tc.Changelog, tc.A)

			assert.Equal(t, tc.B, tc.A)
			require.Equal(t, len(tc.Changelog), len(pl))
		})
	}
}

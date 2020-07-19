package diff

import (
	"fmt"
	"reflect"
)

func ExampleDiff() {
	type Tag struct {
		Name  string `diff:"name,identifier"`
		Value string `diff:"value"`
	}

	type Fruit struct {
		ID        int      `diff:"id"`
		Name      string   `diff:"name"`
		Healthy   bool     `diff:"healthy"`
		Nutrients []string `diff:"nutrients"`
		Tags      []Tag    `diff:"tags"`
	}

	a := Fruit{
		ID:      1,
		Name:    "Green Apple",
		Healthy: true,
		Nutrients: []string{
			"vitamin c",
			"vitamin d",
		},
		Tags: []Tag{
			{
				Name:  "kind",
				Value: "fruit",
			},
		},
	}

	b := Fruit{
		ID:      2,
		Name:    "Red Apple",
		Healthy: true,
		Nutrients: []string{
			"vitamin c",
			"vitamin d",
			"vitamin e",
		},
		Tags: []Tag{
			{
				Name:  "popularity",
				Value: "high",
			},
			{
				Name:  "kind",
				Value: "fruit",
			},
		},
	}

	changelog, err := Diff(a, b)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", changelog)
	// Produces: diff.Changelog{diff.Change{Type:"update", Path:[]string{"id"}, From:1, To:2}, diff.Change{Type:"update", Path:[]string{"name"}, From:"Green Apple", To:"Red Apple"}, diff.Change{Type:"create", Path:[]string{"nutrients", "2"}, From:interface {}(nil), To:"vitamin e"}, diff.Change{Type:"create", Path:[]string{"tags", "popularity"}, From:interface {}(nil), To:main.Tag{Name:"popularity", Value:"high"}}}
}

func ExampleFilter() {
	type Tag struct {
		Name  string `diff:"name,identifier"`
		Value string `diff:"value"`
	}

	type Fruit struct {
		ID        int      `diff:"id"`
		Name      string   `diff:"name"`
		Healthy   bool     `diff:"healthy"`
		Nutrients []string `diff:"nutrients"`
		Tags      []Tag    `diff:"tags"`
	}

	a := Fruit{
		ID:      1,
		Name:    "Green Apple",
		Healthy: true,
		Nutrients: []string{
			"vitamin c",
			"vitamin d",
		},
	}

	b := Fruit{
		ID:      2,
		Name:    "Red Apple",
		Healthy: true,
		Nutrients: []string{
			"vitamin c",
			"vitamin d",
			"vitamin e",
		},
	}

	d, err := NewDiffer(Filter(func(path []string, parent reflect.Type, field reflect.StructField) bool {
		return field.Name != "Name"
	}))
	if err != nil {
		panic(err)
	}

	changelog, err := d.Diff(a, b)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", changelog)
	// Output: diff.Changelog{diff.Change{Type:"update", Path:[]string{"id"}, From:1, To:2}, diff.Change{Type:"create", Path:[]string{"nutrients", "2"}, From:interface {}(nil), To:"vitamin e"}}
}

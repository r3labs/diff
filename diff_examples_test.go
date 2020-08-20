package diff

import (
	"fmt"
	"reflect"
)

//Try to do a bunch of stuff that will result in some or all failures
//when trying to apply either a valid or invalid changelog
func ExamplePatchWithErrors(){

	type Fruit struct {
		ID        int      `diff:"ID" json:"Identifier"`
		Name      string   `diff:"name"`
		Healthy   bool     `diff:"-"`
		Nutrients []string `diff:"nutrients"`
		Labels    map[string]int `diff:"labs"`
	}

	type Bat struct {
		ID string `diff:"ID"`
		Name string `diff:"-"`
	}

	a := Fruit{
		ID:      1,
		Name:    "Green Apple",
		Healthy: true,
		Nutrients: []string{
			"vitamin a",
			"vitamin b",
			"vitamin c",
			"vitamin d",
		},
		Labels: make(map[string]int),
	}
	a.Labels["likes"] = 10
	a.Labels["colors"] = 2

	b := Fruit{
		ID:      2,
		Name:    "Red Apple",
		Healthy: true,
		Nutrients: []string{
			"vitamin c",
			"vitamin d",
			"vitamin e",
		},
		Labels: make(map[string]int),
	}
	b.Labels["forests"] = 1223
	b.Labels["colors"] = 1222

	c := Fruit{
		Labels: make(map[string]int),
		Nutrients: []string{
			"vitamin c",
			"vitamin d",
			"vitamin a",
		},
	}
	c.Labels["likes"] = 21
	c.Labels["colors"] = 42

	d := Bat{
		ID: "first",
		Name: "second",
	}

	changelog, err := Diff(a, b)
	if err != nil {
		panic(err)
	}

	//This fails in total because c is not assignable (passed by value)
	patchLog := Patch(changelog, c)

	//this also demonstrated the nested errors with 'next'
	errors := patchLog[7].Errors.(*DiffError)

	//we can also continue to nest errors if we like
	message := errors.WithCause(NewError("This is a custom message")).
		              WithCause(fmt.Errorf("this is an error from somewhere else but still compatible")).
		              Error()

	//invoke a few failures, i.e. bad changelog
	changelog[2].Path[1] = "bad index"
	changelog[3].Path[0] = "bad struct field"

	patchLog = Patch(changelog, &c)

	patchLog, _ = Merge(a,nil, &c)

	patchLog, _ = Merge(a, d, &c)

	//try patching a string
	patchLog = Patch(changelog, message)

	//test an invalid change value
	var bad *ChangeValue
	if bad.IsValid() {
		fmt.Print("this should never happen")
	}

	//Output:
}

//ExampleMerge demonstrates how to use the Merge function
func ExampleMerge() {
	type Fruit struct {
		ID        int      `diff:"ID" json:"Identifier"`
		Name      string   `diff:"name"`
		Healthy   bool     `diff:"healthy"`
		Nutrients []string `diff:"nutrients,create,omitunequal"`
		Labels    map[string]int `diff:"labs,create"`
	}

	a := Fruit{
		ID:      1,
		Name:    "Green Apple",
		Healthy: true,
		Nutrients: []string{
			"vitamin a",
			"vitamin b",
			"vitamin c",
			"vitamin d",
		},
		Labels: make(map[string]int),
	}
	a.Labels["likes"] = 10
	a.Labels["colors"] = 2

	b := Fruit{
		ID:      2,
		Name:    "Red Apple",
		Healthy: true,
		Nutrients: []string{
			"vitamin c",
			"vitamin d",
			"vitamin e",
		},
		Labels: make(map[string]int),
	}
	b.Labels["forests"] = 1223
	b.Labels["colors"] = 1222

	c := Fruit{
		Labels: make(map[string]int),
		Nutrients: []string{
			"vitamin a",
			"vitamin c",
			"vitamin d",
		},
	}
	c.Labels["likes"] = 21
	c.Labels["colors"] = 42

	//the only error that can happen here comes from the diff step
	patchLog, _ := Merge(a,b, &c)

	//Note that unlike our patch version we've not included 'create' in the
	//tag for nutrients. This will omit "vitamin e" from ending up in c
	fmt.Printf("%#v", len(patchLog))

	//Output: 8
}

//ExamplePatch demonstrates how to use the Patch function
func ExamplePatch(){

	type Fruit struct {
		ID        int      `diff:"ID" json:"Identifier"`
		Name      string   `diff:"name"`
		Healthy   bool     `diff:"healthy"`
		Nutrients []string `diff:"nutrients,create,omitunequal"`
		Labels    map[string]int `diff:"labs,create"`
	}

	a := Fruit{
		ID:      1,
		Name:    "Green Apple",
		Healthy: true,
		Nutrients: []string{
			"vitamin a",
			"vitamin b",
			"vitamin c",
			"vitamin d",
		},
		Labels: make(map[string]int),
	}
	a.Labels["likes"] = 10
	a.Labels["colors"] = 2

	b := Fruit{
		ID:      2,
		Name:    "Red Apple",
		Healthy: true,
		Nutrients: []string{
			"vitamin c",
			"vitamin d",
			"vitamin e",
		},
		Labels: make(map[string]int),
	}
	b.Labels["forests"] = 1223
	b.Labels["colors"] = 1222

	c := Fruit{
		Labels: make(map[string]int),
		Nutrients: []string{
			"vitamin a",
			"vitamin c",
			"vitamin d",
		},
	}
	c.Labels["likes"] = 21

	changelog, err := Diff(a, b)
	if err != nil {
		panic(err)
	}

	patchLog := Patch(changelog, &c)

	fmt.Printf("%#v", len(patchLog))

	//Output: 8
}

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

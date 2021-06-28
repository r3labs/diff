package diff

import (
	"fmt"
	"math/big"
	"reflect"
)

//Try to do a bunch of stuff that will result in some or all failures
//when trying to apply either a valid or invalid changelog
func ExamplePatchWithErrors() {

	type Fruit struct {
		ID        int            `diff:"ID" json:"Identifier"`
		Name      string         `diff:"name"`
		Healthy   bool           `diff:"-"`
		Nutrients []string       `diff:"nutrients"`
		Labels    map[string]int `diff:"labs"`
	}

	type Bat struct {
		ID   string `diff:"ID"`
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
		ID:   "first",
		Name: "second",
	}

	changelog, err := Diff(a, b)
	if err != nil {
		panic(err)
	}

	//This fails in total because c is not assignable (passed by Value)
	patchLog := Patch(changelog, c)

	//this also demonstrated the nested errors with 'next'

	errors := patchLog[0].Errors.(*DiffError)

	//we can also continue to nest errors if we like
	message := errors.WithCause(NewError("This is a custom message")).
		WithCause(fmt.Errorf("this is an error from somewhere else but still compatible")).
		Error()

	//invoke a few failures, i.e. bad changelog
	changelog[2].Path[1] = "bad index"
	changelog[3].Path[0] = "bad struct field"

	patchLog = Patch(changelog, &c)

	patchLog, _ = Merge(a, nil, &c)

	patchLog, _ = Merge(a, d, &c)

	//try patching a string
	patchLog = Patch(changelog, message)

	//test an invalid change Value
	var bad *ChangeValue
	if bad.IsValid() {
		fmt.Print("this should never happen")
	}

	//Output:
}

//ExampleMerge demonstrates how to use the Merge function
func ExampleMerge() {
	type Fruit struct {
		ID        int            `diff:"ID" json:"Identifier"`
		Name      string         `diff:"name"`
		Healthy   bool           `diff:"healthy"`
		Nutrients []string       `diff:"nutrients,create,omitunequal"`
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
	patchLog, _ := Merge(a, b, &c)

	//Note that unlike our patch version we've not included 'create' in the
	//tag for nutrients. This will omit "vitamin e" from ending up in c
	fmt.Printf("%#v", len(patchLog))

	//Output: 8
}

//ExamplePrimitiveSlice demonstrates working with arrays and primitive values
func ExamplePrimitiveSlice() {

	sla := []string{
		"this",
		"is",
		"a",
		"simple",
	}

	slb := []string{
		"slice",
		"That",
		"can",
		"be",
		"diff'ed",
	}

	slc := []string{
		"ok",
	}

	patch, err := Diff(sla, slb, StructMapKeySupport())
	if err != nil {
		fmt.Print("failed to diff sla and slb")
	}
	cl := Patch(patch, &slc)

	//now the other way, round
	sla = []string{
		"slice",
		"That",
		"can",
		"be",
		"diff'ed",
	}
	slb = []string{
		"this",
		"is",
		"a",
		"simple",
	}

	patch, err = Diff(sla, slb)
	if err != nil {
		fmt.Print("failed to diff sla and slb")
	}
	cl = Patch(patch, &slc)

	//and finally a clean view
	sla = []string{
		"slice",
		"That",
		"can",
		"be",
		"diff'ed",
	}
	slb = []string{}

	patch, err = Diff(sla, slb)
	if err != nil {
		fmt.Print("failed to diff sla and slb")
	}
	cl = Patch(patch, &slc)

	fmt.Printf("%d changes made to string array; %v", len(cl), slc)

	//Output: 5 changes made to string array; [simple a]
}

//ExampleComplexMapPatch demonstrates how to use the Patch function for complex slices
//NOTE: There is a potential pitfall here, take a close look at b[2]. If patching the
//      original, the operation will work intuitively however, in a merge situation we
//      may not get everything we expect because it's a true diff between a and b and
//      the diff log will not contain enough information to fully recreate b from an
//      empty slice. This is exemplified in that the test "colors" is dropped in element
//      3 of c. Change "colors" to "color" and see what happens. Keep in mind this only
//      happens when we need to allocate a new complex element. In normal operations we
//      fix for this by keeping a copy of said element in the diff log (as parent) and
//      allocate such an element as a whole copy prior to applying any updates?
//
//      The new default is to carry this information forward, we invoke this pitfall
//      by creating such a situation and explicitly telling diff to discard the parent
//      In memory constrained environments if the developer is careful, they can use
//      the discard feature but unless you REALLY understand what's happening here, use
//      the default.
func ExampleComplexSlicePatch() {

	type Content struct {
		Text   string `diff:",create"`
		Number int    `diff:",create"`
	}
	type Attributes struct {
		Labels []Content `diff:",create"`
	}

	a := Attributes{
		Labels: []Content{
			{
				Text:   "likes",
				Number: 10,
			},
			{
				Text:   "forests",
				Number: 10,
			},
			{
				Text:   "colors",
				Number: 2,
			},
		},
	}

	b := Attributes{
		Labels: []Content{
			{
				Text:   "forests",
				Number: 14,
			},
			{
				Text:   "location",
				Number: 0x32,
			},
			{
				Text:   "colors",
				Number: 1222,
			},
			{
				Text:   "trees",
				Number: 34,
			},
		},
	}
	c := Attributes{}

	changelog, err := Diff(a, b, DiscardComplexOrigin(), StructMapKeySupport())
	if err != nil {
		panic(err)
	}

	patchLog := Patch(changelog, &c)

	fmt.Printf("Patched %d entries and encountered %d errors", len(patchLog), patchLog.ErrorCount())

	//Output: Patched 7 entries and encountered 4 errors
}

//ExampleComplexMapPatch demonstrates how to use the Patch function for complex slices.
func ExampleComplexMapPatch() {

	type Key struct {
		Value  string
		weight int
	}
	type Content struct {
		Text        string
		Number      float64
		WholeNumber int
	}
	type Attributes struct {
		Labels map[Key]Content
	}

	a := Attributes{
		Labels: make(map[Key]Content),
	}
	a.Labels[Key{Value: "likes"}] = Content{
		WholeNumber: 10,
		Number:      23.4,
	}

	a.Labels[Key{Value: "colors"}] = Content{
		WholeNumber: 2,
	}

	b := Attributes{
		Labels: make(map[Key]Content),
	}
	b.Labels[Key{Value: "forests"}] = Content{
		Text: "Sherwood",
	}
	b.Labels[Key{Value: "colors"}] = Content{
		Number: 1222,
	}
	b.Labels[Key{Value: "latitude"}] = Content{
		Number: 38.978797,
	}
	b.Labels[Key{Value: "longitude"}] = Content{
		Number: -76.490986,
	}

	//c := Attributes{}
	c := Attributes{
		Labels: make(map[Key]Content),
	}
	c.Labels[Key{Value: "likes"}] = Content{
		WholeNumber: 210,
		Number:      23.4453,
	}

	changelog, err := Diff(a, b)
	if err != nil {
		panic(err)
	}

	patchLog := Patch(changelog, &c)

	fmt.Printf("%#v", len(patchLog))

	//Output: 7
}

//ExamplePatch demonstrates how to use the Patch function
func ExamplePatch() {

	type Key struct {
		value  string
		weight int
	}
	type Cycle struct {
		Name  string `diff:"name,create"`
		Count int    `diff:"count,create"`
	}
	type Fruit struct {
		ID        int           `diff:"ID" json:"Identifier"`
		Name      string        `diff:"name"`
		Healthy   bool          `diff:"healthy"`
		Nutrients []string      `diff:"nutrients,create,omitunequal"`
		Labels    map[Key]Cycle `diff:"labs,create"`
		Cycles    []Cycle       `diff:"cycles,immutable"`
		Weights   []int
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
		Labels: make(map[Key]Cycle),
	}
	a.Labels[Key{value: "likes"}] = Cycle{
		Count: 10,
	}
	a.Labels[Key{value: "colors"}] = Cycle{
		Count: 2,
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
		Labels: make(map[Key]Cycle),
		Weights: []int{
			1,
			2,
			3,
			4,
		},
	}
	b.Labels[Key{value: "forests"}] = Cycle{
		Count: 1223,
	}
	b.Labels[Key{value: "colors"}] = Cycle{
		Count: 1222,
	}

	c := Fruit{
		//Labels: make(map[string]int),
		Nutrients: []string{
			"vitamin a",
			"vitamin c",
			"vitamin d",
		},
	}
	//c.Labels["likes"] = 21

	d := a
	d.Cycles = []Cycle{
		Cycle{
			Name:  "First",
			Count: 45,
		},
		Cycle{
			Name:  "Third",
			Count: 4,
		},
	}
	d.Nutrients = append(d.Nutrients, "minerals")

	changelog, err := Diff(a, b)
	if err != nil {
		panic(err)
	}

	patchLog := Patch(changelog, &c)

	changelog, _ = Diff(a, d)
	patchLog = Patch(changelog, &c)

	fmt.Printf("%#v", len(patchLog))

	//Output: 1
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
	// Output: diff.Changelog{diff.Change{Type:"update", Path:[]string{"id"}, From:1, To:2, parent:diff.Fruit{ID:1, Name:"Green Apple", Healthy:true, Nutrients:[]string{"vitamin c", "vitamin d"}, Tags:[]diff.Tag(nil)}}, diff.Change{Type:"create", Path:[]string{"nutrients", "2"}, From:interface {}(nil), To:"vitamin e", parent:interface {}(nil)}}
}

func ExamplePrivatePtr() {
	type number struct {
		value *big.Int
		exp   int32
	}
	a := number{}
	b := number{value: big.NewInt(111)}

	changelog, err := Diff(a, b)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", changelog)
	// Output: diff.Changelog{diff.Change{Type:"update", Path:[]string{"value"}, From:interface {}(nil), To:111, parent:diff.number{value:(*big.Int)(nil), exp:0}}}
}

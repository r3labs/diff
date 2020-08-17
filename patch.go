package diff

import (
	"reflect"
	"strconv"
)

/**
	This is a method of applying a changelog to a value or struct. change logs
    should be generated with Diff and never manually created. This DOES NOT
    apply fuzzy logic as would be in the case of a text patch. It does however
    have a few additional features added to our struct tags.

	1) create. This tag on a struct field indicates that the patch should
       create the value if it's not there. I.e. if it's nil. This works for
       pointers, maps and slices.

	2) omitunequal. Generally, you don't want to do this, the expectation is
       that if an item isn't there, you want to add it. For example, if your
       diff shows an array element at index 6 is a string 'hello' but your target
       only has 3 elements, none of them matching... you want to add 'hello'
       regardless of the index. (think in a distributed context, another process
       may have deleted more than one entry and 'hello' may no longer be in that
       indexed spot.

       So given this scenario, the default behavior is to scan for the previous
       value and replace it anyway, or simply append the new value. For maps the
       default behavior is to simply add the key if it doesn't match.

       However, if you don't like the default behavior, and add the omitunequal
       tag to your struct, patch will *NOT* update an array or map with the key
       or array value unless they key or index contains a 'match' to the
       previous value. In which case it will skip over that change.

    Patch is implemented as a best effort algorithm. That means you can receive
    multiple nested errors and still successfully have a modified target. This
    may even be acceptable depending on your use case. So keep in mind, just
    because err != nil *DOESN'T* mean that the patch didn't accomplish your goal
    in setting those changes that are actually available. For example, you may
    diff two structs of the same type, then attempt to apply to an entirely
    different struct that is similar in constitution (think interface here) and
    you may in fact get all of the values populated you wished to anyway.
 */

// Merge is a convenience function that diffs, the original and changed items
// and merges said changes with target all in one call.
func Merge(original interface{}, changed interface{}, target interface{}) (PatchLog, error) {
	if cl, err := Diff(original, changed); err == nil {
		return Patch(cl, target), nil
	}else{
		return nil, err
	}
}

//Patch... the missing feature.
func Patch(cl Changelog, target interface{}) (ret PatchLog) {
	for _, c := range cl {
		ret = append(ret, NewPatchLogEntry(patch(NewChangeValue(c, target))))
	}
	return ret
}

//patch apply an individual change as opposed to a change log
func patch(c *ChangeValue) (cv *ChangeValue){

	//Can be messy to clean up when doing reflection
	defer func() {
		if r := recover(); r != nil {
			cv.AddError(NewError("cannot set values on target",
				NewError("passed by value not reference")))
			cv.SetFlag(FlagInvalidTarget)
		}
	}()
	cv = c
	cv.val = cv.val.Elem()

	//resolve where we're actually going to set this value
	cv.index = -1
	if cv.Kind() == reflect.Struct { //substitute and solve for t (path)
		for _, p := range cv.change.Path {
			if cv.Kind() == reflect.Slice {
				//field better be an index of the slice
				if cv.index, cv.err = strconv.Atoi(p); cv.err != nil {
					return cv.AddError(NewErrorf("invalid index in path: %s", p).
						WithCause(cv.err))
				}
				break //last item in a path that's a slice is it's index
			}
			if cv.Kind() == reflect.Map {
				keytype := cv.KeyType()
				cv.key = reflect.ValueOf(p)
				cv.key.Convert(keytype)
				break //same as a slice
			}
			cv = renderTargetField(cv, p)
		}
	}

	//we have to know that the new element we're trying to set is valid
	if !cv.CanSet(){
		cv.SetFlag(FlagInvalidTarget)
		return cv.AddError(NewError("cannot set values on target",
			NewError("passed by value not reference")))
	}

	switch cv.change.Type {
	case DELETE:
		deleteOperation(cv)
	case UPDATE, CREATE:
		updateOperation(cv)
	}
	return cv
}

//renderTargetField this interrogates the path and returns the correct value to
//change. Note that his is a recursion, t is not the same value as ret
func renderTargetField(t *ChangeValue, field string) (ret *ChangeValue) {
	ret = t
	//substitute and solve for t (path)
	switch t.val.Kind() {
	case reflect.Struct:
		for i := 0; i < t.val.NumField(); i++ {
			f := t.val.Type().Field(i)
			tname := tagName("diff", f)
			if tname == "-" || hasTagOption("diff", f, "immutable") {
				continue
			}
			if tname == field || f.Name == field{
				ret = &ChangeValue{
					val:    t.val.Field(i),
					change: t.change,
				}
				ret.ClearFlags()
				if hasTagOption("diff", f, "create") {
					ret.SetFlag(OptionCreate)
				}
				if hasTagOption("diff", f, "omitunequal"){
					ret.SetFlag(OptionOmitUnequal)
				}
			}
		}
	default:
		ret = &ChangeValue{
			val:    t.val,
			change: t.change,
		}
	}
	if !ret.IsValid() {
		ret.AddError(NewErrorf("Unable to access path value %v. Target field is invalid", field))
	}
	return
}

//deleteOperation takes out some of the cyclomatic complexity from the patch fuction
func deleteOperation(cv *ChangeValue) {
	switch cv.Kind() {
	case reflect.Slice:
		var x reflect.Value
		if cv.Len() > cv.index {
			x = cv.Index(cv.index)
		}
		found := true
		if !x.IsValid() || !reflect.DeepEqual(x.Interface(), cv.change.From) {
			found = false
			if !cv.HasFlag(OptionOmitUnequal){
				cv.AddError(NewErrorf("value index %d is invalid", cv.index).
					WithCause(NewError("scanning for value index")))
				for i := 0; i < cv.Len(); i++ {
					x = cv.Index(i)
					if reflect.DeepEqual(x, cv.change.From) {
						cv.AddError(NewErrorf("value changed index to %d", i))
						found = true
						cv.index = i
					}
				}
			}
		}
		if x.IsValid() && found{
			cv.val.Index(cv.index).Set(cv.val.Index(cv.Len() - 1))
			cv.val.Set(cv.val.Slice(0, cv.Len() - 1))
			cv.SetFlag(FlagDeleted)
		}else{
			cv.SetFlag(FlagIgnored)
			cv.AddError(NewError("Unable to find matching slice index entry"))
		}
	case reflect.Map:
		if !reflect.DeepEqual(cv.change.From, cv.Interface()) &&
			cv.HasFlag(OptionOmitUnequal){
			cv.SetFlag(FlagIgnored)
			cv.AddError(NewError("target change doesn't match original"))
			return
		}
		if cv.IsNil() {
			cv.SetFlag(FlagIgnored)
			cv.AddError(NewError("target has nil map nothing to delete"))
			return
		}else{
			cv.SetFlag(FlagDeleted)
			cv.SetMapValue(cv.key, reflect.Value{})
		}
	default:
		cv.Set(reflect.Zero(cv.Type()))
		cv.SetFlag(FlagDeleted)
	}
}

//updateOperation takes out some of the cyclomatic complexity from the patch fuction
func updateOperation(cv *ChangeValue) {
	switch cv.Kind() {
	case reflect.Slice:
		var x reflect.Value
		if cv.Len() > cv.index {
			x = cv.Index(cv.index)
		}
		found := true
		if !x.IsValid() || !reflect.DeepEqual(x.Interface(), cv.change.From) {
			found = false
			if !cv.HasFlag(OptionOmitUnequal){
				cv.AddError(NewErrorf("value index %d is invalid", cv.index).
						    WithCause(NewError("scanning for value index")))
				for i := 0; i < cv.Len(); i++ {
					x = cv.Index(i)
					if reflect.DeepEqual(x, cv.change.From) {
						cv.AddError(NewErrorf("value changed index to %d", i))
						found = true
					}
				}
			}
		}
		if x.IsValid() && found{
			x.Set(reflect.ValueOf(cv.change.To))
			cv.SetFlag(FlagUpdated)
		}else if cv.HasFlag(OptionCreate) && cv.change.Type == CREATE {
			cv.Set(reflect.Append(cv.val, reflect.ValueOf(cv.change.To)))
			cv.SetFlag(FlagCreated)
		}else{
			cv.AddError(NewError("Unable to find matching slice index entry"))
			cv.SetFlag(FlagIgnored)
		}
	case reflect.Map:
		if !reflect.DeepEqual(cv.change.From, cv.Interface()) &&
			cv.HasFlag(OptionOmitUnequal){
			cv.SetFlag(FlagIgnored)
			cv.AddError(NewError("target change doesn't match original"))
			return
		}
		if cv.IsNil() {
			if cv.HasFlag(OptionCreate) {
				nm := reflect.MakeMap(cv.Type())
				nm.SetMapIndex(cv.key, reflect.ValueOf(cv.change.To))
				cv.Set(nm)
				cv.SetFlag(FlagCreated)
			}else{
				cv.SetFlag(FlagIgnored)
				cv.AddError(NewError("target has nil map and create not set"))
				return
			}
		}else{
			cv.SetFlag(FlagUpdated)
			cv.SetMapValue(cv.key, reflect.ValueOf(cv.change.To))
		}
	default:
		if !reflect.DeepEqual(cv.change.From, cv.Interface()) &&
			cv.HasFlag(OptionOmitUnequal){
			cv.SetFlag(FlagIgnored)
			cv.AddError(NewError("target change doesn't match original"))
			return
		}
		cv.Set(reflect.ValueOf(cv.change.To))
		cv.SetFlag(FlagUpdated)
	}
}


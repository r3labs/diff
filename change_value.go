package diff

import (
	"reflect"
)

//Not strictly necessary but migh be nice in some cases
//go:generate stringer -type=PatchFlags
type PatchFlags uint32
const (
	OptionCreate PatchFlags = 1 << iota
	OptionOmitUnequal
	FlagInvalidTarget
	FlagApplied
	FlagFailed
	FlagCreated
	FlagIgnored
	FlagDeleted
	FlagUpdated
)

//ChangeValue is a specialized struct for monitoring patching
type ChangeValue struct {
	val    reflect.Value
	flags  PatchFlags
	change *Change
	err    error
	index  int
	key    reflect.Value
}

//PatchLogEntry defines how a DiffLog entry was applied
type PatchLogEntry struct{
	Path  []string    `json:"path"`
	From  interface{} `json:"from"`
	To    interface{} `json:"to"`
	Flags PatchFlags  `json:"flags"`
	Errors error      `json:"errors"`
}
type PatchLog []PatchLogEntry

//NewChangeValue idiomatic constructor
func NewChangeValue(c Change, target interface{}) *ChangeValue{
	return &ChangeValue{
		val:    reflect.ValueOf(target),
		change: &c,
	}
}

//NewPatchLogEntry converts our complicated reflection based struct to
//a simpler format for the consumer
func NewPatchLogEntry(change *ChangeValue) PatchLogEntry {
	return PatchLogEntry{
		Path: change.change.Path,
		From: change.change.From,
		To: change.change.To,
		Flags: change.flags,
		Errors: change.err,
	}
}

// Sets a flag on the node and saves the change
func (c *ChangeValue) SetFlag(flag PatchFlags) {
	if c != nil {
		c.flags = c.flags|flag
	}
}

//ClearFlag Clears a flag on the node and saves the change
func (c *ChangeValue) ClearFlags(){
	if c != nil {c.flags = 0}
}

//HasFlag indicates if a flag is set on the node. returns false if node is bad
func (c *ChangeValue) HasFlag(flag PatchFlags) bool {
	return (c.flags & flag) != 0
}

//CanSet echos the reflection can set
func (c ChangeValue) CanSet() bool {
	return c.val.CanSet()
}

//IsValid echo for is valid
func (c *ChangeValue) IsValid() bool {
	if c != nil {return c.val.IsValid() || !c.HasFlag(FlagInvalidTarget)}
	return false
}

//Len echo for len
func (c ChangeValue) Len() int {
	return c.val.Len()
}

//Kind echos the reflection kind
func (c ChangeValue) Kind() reflect.Kind {
	return c.val.Kind()
}

//Type echos Type
func (c ChangeValue) Type() reflect.Type {
	return c.val.Type()
}

//Set echos reflect set
func (c *ChangeValue) Set(value reflect.Value){
	if c != nil {
		defer func() {
			if r := recover(); r != nil {
				c.SetFlag(FlagFailed)
			}
		}()
		c.val.Set(value)
		c.SetFlag(FlagApplied)
	}
}

//SetMapValue is used to set a map value
func (c *ChangeValue) SetMapValue(key reflect.Value, value reflect.Value){
	if c != nil {
		defer func() {
			if r := recover(); r != nil {
				c.SetFlag(FlagFailed)
			}
		}()
		c.val.SetMapIndex(key, value)
	}
}

//Index echo for index
func (c ChangeValue) Index(i int) reflect.Value {
	return c.val.Index(i)
}

//Interface gets the interface for the value
func (c ChangeValue) Interface() interface{} {
	return c.val.Interface()
}

//IsNil echo for is nil
func (c ChangeValue) IsNil() bool {
	return c.val.IsNil()
}

//KeyType returns the key type of a map if it is one
func (c ChangeValue) KeyType() reflect.Type {
	return c.Type().Key()
}

//AddError appends errors to this change value
func (c *ChangeValue) AddError(err error) *ChangeValue{
	if c != nil {c.err = err}
	return c
}
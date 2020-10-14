package diff

import "reflect"

/**
	Types are being split out to more closely follow the library structure already
    in place. Keeps the file simpler as well.
*/

//patchStruct - handles the rendering of a struct field
func (c *ChangeValue) patchStruct() {

	field := c.change.Path[c.pos]

	for i := 0; i < c.target.NumField(); i++ {
		f := c.target.Type().Field(i)
		tname := tagName("diff", f)
		if tname == "-" || hasTagOption("diff", f, "immutable") {
			c.SetFlag(OptionImmutable)
			continue
		}
		if tname == field || f.Name == field {
			x := c.target.Field(i)
			if hasTagOption("diff", f, "nocreate") {
				c.SetFlag(OptionNoCreate)
			}
			if hasTagOption("diff", f, "omitunequal") {
				c.SetFlag(OptionOmitUnequal)
			}
			c.swap(&x)
			break
		}
	}
}

//track and zero out struct members
func (c *ChangeValue) deleteStructEntry() {

	//deleting a struct value set's it to the 'basic' type
	c.Set(reflect.Zero(c.target.Type()))
}

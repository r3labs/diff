package diff

import "reflect"

/**
	Types are being split out to more closely follow the library structure already
    in place. Keeps the file simpler as well.
*/

//patchStruct - handles the rendering of a struct field
func (d *Differ) patchStruct(c *ChangeValue) {

	field := c.change.Path[c.pos]

	for i := 0; i < c.target.NumField(); i++ {
		f := c.target.Type().Field(i)
		tname := tagName(d.TagName, f)
		if tname == "-" {
			continue
		}
		if tname == field || f.Name == field {
			x := c.target.Field(i)
			if hasTagOption(d.TagName, f, "nocreate") {
				c.SetFlag(OptionNoCreate)
			}
			if hasTagOption(d.TagName, f, "omitunequal") {
				c.SetFlag(OptionOmitUnequal)
			}
			if hasTagOption(d.TagName, f, "immutable") {
				c.SetFlag(OptionImmutable)
			}
			c.swap(&x)
			break
		}
	}
}

//track and zero out struct members
func (d *Differ) deleteStructEntry(c *ChangeValue) {

	//deleting a struct value set's it to the 'basic' type
	c.Set(reflect.Zero(c.target.Type()))
}

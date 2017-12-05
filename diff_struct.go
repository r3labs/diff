package diff

import "reflect"

func (cl *Changelog) diffStruct(path []string, a, b reflect.Value) error {
	if a.Kind() != b.Kind() {
		return ErrTypeMismatch
	}

	for i := 0; i < a.NumField(); i++ {
		name := a.Type().Field(i).Name
		tname := tagName(a, i)

		if tname == "-" {
			continue
		}

		af := a.Field(i)
		bf := b.FieldByName(name)

		fpath := append(path, tname)

		err := cl.diff(fpath, af, bf)
		if err != nil {
			return err
		}
	}

	return nil
}

package diff

import "reflect"

func (cl *Changelog) diffStruct(path []string, a, b reflect.Value) error {
	if a.Kind() != b.Kind() {
		return ErrTypeMismatch
	}

	for i := 0; i < a.NumField(); i++ {
		name := a.Type().Field(i).Name

		af := a.Field(i)
		bf := b.FieldByName(name)

		fpath := append(path, tagName(a, i))

		err := cl.diff(fpath, af, bf)
		if err != nil {
			return err
		}
	}

	return nil
}

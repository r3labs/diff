package diff

import "reflect"

func (cl *Changelog) diffBool(path []string, a, b reflect.Value) error {
	if a.Kind() != b.Kind() {
		return ErrTypeMismatch
	}

	if a.Bool() != b.Bool() {
		cl.add(UPDATE, path, a.Interface(), b.Interface())
	}

	return nil
}

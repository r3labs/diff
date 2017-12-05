package diff

import "reflect"

func (cl *Changelog) diffInt(path []string, a, b reflect.Value) error {
	if a.Kind() != b.Kind() {
		return ErrTypeMismatch
	}

	if a.Int() != b.Int() {
		cl.add(UPDATE, path, a.Interface(), b.Interface())
	}

	return nil
}

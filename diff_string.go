package diff

import "reflect"

func (cl *Changelog) diffString(path []string, a, b reflect.Value) error {
	if a.Kind() != b.Kind() {
		return ErrTypeMismatch
	}

	if a.String() != b.String() {
		cl.add(UPDATE, path, a.Interface(), b.Interface())
	}

	return nil
}

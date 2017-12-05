package diff

import "reflect"

func (cl *Changelog) diffMap(path []string, a, b reflect.Value) error {
	if a.Kind() != b.Kind() {
		return ErrTypeMismatch
	}

	if !reflect.DeepEqual(a.Interface(), b.Interface()) {
		cl.add(UPDATE, path, a.Interface(), b.Interface())
	}

	return nil
}

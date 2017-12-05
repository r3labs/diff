package diff

import (
	"reflect"
	"strconv"
)

func (cl *Changelog) diffSlice(path []string, a, b reflect.Value) error {
	if a.Kind() != b.Kind() {
		return ErrTypeMismatch
	}

	if comparative(a, b) {
		return cl.diffSliceComparative(path, a, b)
	}

	return cl.diffSliceGeneric(path, a, b)
}

func (cl *Changelog) diffSliceGeneric(path []string, a, b reflect.Value) error {
	for i := 0; i < a.Len(); i++ {
		ae := a.Index(i)
		fpath := append(path, strconv.Itoa(i))

		if !sliceHas(b, ae) {
			cl.add(DELETE, fpath, ae.Interface(), nil)
		}
	}

	for i := 0; i < b.Len(); i++ {
		be := b.Index(i)
		fpath := append(path, strconv.Itoa(i))

		if !sliceHas(a, be) {
			cl.add(CREATE, fpath, nil, be.Interface())
		}
	}

	return nil
}

func (cl *Changelog) diffSliceComparative(path []string, a, b reflect.Value) error {
	c := NewComparativeList()

	for i := 0; i < a.Len(); i++ {
		ae := a.Index(i)

		id := identifier(ae)
		if id != nil {
			c.addA(id, &ae)
		}
	}

	for i := 0; i < b.Len(); i++ {
		be := b.Index(i)

		id := identifier(be)
		if id != nil {
			c.addB(id, &be)
		}
	}

	for k, v := range *c {
		fpath := append(path, idstring(k))

		if v.A != nil && v.B == nil {
			cl.add(DELETE, fpath, v.A.Interface(), nil)
		}

		if v.A == nil && v.B != nil {
			cl.add(CREATE, fpath, nil, v.B.Interface())
		}

		if v.A != nil && v.B != nil {
			err := cl.diff(fpath, *v.A, *v.B)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func sliceHas(s, v reflect.Value) bool {
	for i := 0; i < s.Len(); i++ {
		x := s.Index(i)
		if reflect.DeepEqual(x.Interface(), v.Interface()) {
			return true
		}
	}

	return false
}

func comparative(a, b reflect.Value) bool {
	if a.Len() > 0 {
		ae := a.Index(0)
		if ae.Kind() == reflect.Struct {
			if identifier(ae) != nil {
				return true
			}
		}
	}

	if b.Len() > 0 {
		be := b.Index(0)
		if be.Kind() == reflect.Struct {
			if identifier(be) != nil {
				return true
			}
		}
	}

	return false
}

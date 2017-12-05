package diff

import "reflect"

// Comparative ...
type Comparative struct {
	A, B *reflect.Value
}

// ComparativeList : stores indexed comparative
type ComparativeList map[interface{}]*Comparative

// NewComparativeList : returns a new comparative list
func NewComparativeList() *ComparativeList {
	cl := make(ComparativeList)
	return &cl
}

func (cl *ComparativeList) addA(k interface{}, v *reflect.Value) {
	if (*cl)[k] == nil {
		(*cl)[k] = &Comparative{}
	}
	(*cl)[k].A = v
}

func (cl *ComparativeList) addB(k interface{}, v *reflect.Value) {
	if (*cl)[k] == nil {
		(*cl)[k] = &Comparative{}
	}
	(*cl)[k].B = v
}

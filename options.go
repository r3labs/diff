package diff

// SliceOrdering determines whether the ordering of items in a slice results in a change
func SliceOrdering(enabled bool) func(d *Differ) error {
	return func(d *Differ) error {
		d.SliceOrdering = true
		return nil
	}
}

package cards

type resourceBase struct {
	value int
}

func (r *resourceBase) Value() int {
	return r.value
}

package cards

type resourceBase struct {
	value int
}

func newResourceBase(value int) *resourceBase {
	return &resourceBase{
		value: value,
	}
}

func (r *resourceBase) Value() int {
	return r.value
}

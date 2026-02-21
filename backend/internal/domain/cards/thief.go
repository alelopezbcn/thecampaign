package cards

type Thief interface {
	Card
	CanSteal()
}

type thief struct {
	*cardBase
}

func NewThief(id string) *thief {
	return &thief{
		cardBase: newCardBase(id, "Thief"),
	}
}

func (t *thief) CanSteal() {}

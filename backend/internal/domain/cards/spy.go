package cards

type Spy interface {
	Card
	CanSpy()
}

type spy struct {
	*cardBase
}

func NewSpy(id string) *spy {
	return &spy{
		cardBase: newCardBase(id, "Spy"),
	}
}

func (s *spy) CanSpy() {}

package cards

type Sabotage interface {
	Card
	CanSabotage()
}

type sabotage struct {
	*cardBase
}

func NewSabotage(id string) *sabotage {
	return &sabotage{
		cardBase: newCardBase(id, "Sabotage"),
	}
}

func (s *sabotage) CanSabotage() {}

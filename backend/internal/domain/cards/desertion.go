package cards

const DesertionMaxHP = 5

type Desertion interface {
	Card
	CanDesertion()
}

type desertionCard struct {
	*cardBase
}

func NewDesertionCard(id string) *desertionCard {
	return &desertionCard{
		cardBase: newCardBase(id, "Desertion"),
	}
}

func (d *desertionCard) CanDesertion() {}

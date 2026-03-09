package cards

const TreasonMaxHP = 5

type Treason interface {
	Card
	CanTreason()
}

type treasonCard struct {
	*cardBase
}

func NewTreason(id string) *treasonCard {
	return &treasonCard{
		cardBase: newCardBase(id, "Treason"),
	}
}

func (d *treasonCard) CanTreason() {}

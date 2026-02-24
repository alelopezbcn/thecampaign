package cards

// Fortress is a card that protects a castle from one catapult attack.
type Fortress interface {
	Card
	IsFortressCard() bool // marker — distinguishes Fortress from the base Card interface
}

type fortress struct {
	*cardBase
}

func NewFortress(id string) *fortress {
	return &fortress{
		cardBase: newCardBase(id, "Fortress"),
	}
}

func (f *fortress) IsFortressCard() bool { return true }

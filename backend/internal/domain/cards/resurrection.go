package cards

// Resurrection is a card that retrieves a random fallen warrior from the cemetery
// and places it on the player's (or ally's) field during the Attack phase.
type Resurrection interface {
	Card
	IsResurrectionCard() bool // marker — distinguishes Resurrection from the base Card interface
}

type resurrection struct {
	*cardBase
}

func NewResurrection(id string) *resurrection {
	return &resurrection{
		cardBase: newCardBase(id, "Resurrection"),
	}
}

func (r *resurrection) IsResurrectionCard() bool { return true }

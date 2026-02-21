package cards

type CastleTarget interface {
	RemoveGold(position int) (Resource, error)
}

type Catapult interface {
	Card
	Attack(castle CastleTarget, position int) (Resource, error)
}

type catapult struct {
	*cardBase
}

func NewCatapultCard(id string) *catapult {
	return &catapult{
		cardBase: newCardBase(id, "Catapult"),
	}
}
func (c *catapult) Attack(castle CastleTarget, position int) (Resource, error) {
	g, err := castle.RemoveGold(position)
	if err != nil {
		return nil, err
	}

	return g, nil
}

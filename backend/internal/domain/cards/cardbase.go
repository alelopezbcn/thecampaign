package cards

import (
	"strings"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

const (
	WarriorMaxHealth      = 20
	DragonMaxHealth       = 20
	SpecialPowerMaxHealth = 10
)

type cardBase struct {
	id                      string
	name                    string
	player                  ports.Player
	cardMovedToPileObserver ports.CardMovedToPileObserver
}

func newCardBase(id string, name string) *cardBase {
	return &cardBase{
		id:   strings.ToUpper(id),
		name: name,
	}
}

func (c *cardBase) GetID() string {
	return c.id
}
func (c *cardBase) AddCardMovedToPileObserver(observer ports.CardMovedToPileObserver) {
	c.cardMovedToPileObserver = observer
}
func (c *cardBase) GetCardMovedToPileObserver() ports.CardMovedToPileObserver {
	return c.cardMovedToPileObserver
}

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
	id                        string
	name                      string
	player                    ports.Player
	cardToBeDiscardedObserver ports.CardToBeDiscardedObserver
	messageObserver           ports.MessageObserver
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
func (c *cardBase) AssignedToPlayer(player ports.Player) {
	c.player = player
	c.cardToBeDiscardedObserver = player.(ports.CardToBeDiscardedObserver)
}
func (c *cardBase) GetCardToBeDiscardedObserver() ports.CardToBeDiscardedObserver {
	return c.cardToBeDiscardedObserver
}
func (c *cardBase) AddMessageObserver(o ports.MessageObserver) {
	c.messageObserver = o
}
func (c *cardBase) GetMessageObserver() ports.MessageObserver {
	return c.messageObserver
}

package domain

const (
	WarriorHealth      = 20
	DragonHealth       = 20
	SpecialPowerHealth = 10
)

type Card interface {
	GetID() string
	AssignedToPlayer(player Player)
	String() string
	GetCardToBeDiscardedObserver() CardToBeDiscardedObserver
	AddMessageObserver(o MessageObserver)
	GetMessageObserver() MessageObserver
}

type cardBase struct {
	id                        string
	name                      string
	player                    Player
	cardToBeDiscardedObserver CardToBeDiscardedObserver
	messageObserver           MessageObserver
}

func (c *cardBase) GetID() string {
	return c.id
}
func (c *cardBase) AssignedToPlayer(player Player) {
	c.player = player
	c.cardToBeDiscardedObserver = player.(CardToBeDiscardedObserver)
}
func (c *cardBase) GetCardToBeDiscardedObserver() CardToBeDiscardedObserver {
	return c.cardToBeDiscardedObserver
}
func (c *cardBase) AddMessageObserver(o MessageObserver) {
	c.messageObserver = o
}
func (c *cardBase) GetMessageObserver() MessageObserver {
	return c.messageObserver
}

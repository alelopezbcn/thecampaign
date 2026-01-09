package domain

const (
	WarriorHealth      = 20
	DragonHealth       = 20
	SpecialPowerHealth = 10
)

type Card interface {
	GetID() string
	SetPlayer(player *Player)
	String() string
	AddCardToBeDiscardedObserver(o CardToBeDiscardedObserver)
	GetCardToBeDiscardedObserver() CardToBeDiscardedObserver
	AddMessageObserver(o MessageObserver)
	GetMessageObserver() MessageObserver
}

type cardBase struct {
	id                        string
	name                      string
	player                    *Player
	cardToBeDiscardedObserver CardToBeDiscardedObserver
	messageObserver           MessageObserver
}

func (c *cardBase) GetID() string {
	return c.id
}
func (c *cardBase) SetPlayer(player *Player) {
	c.player = player
}
func (c *cardBase) AddCardToBeDiscardedObserver(o CardToBeDiscardedObserver) {
	c.cardToBeDiscardedObserver = o
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

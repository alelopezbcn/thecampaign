package cards

import (
	"strings"
)

type Card interface {
	GetID() string
	Name() string
	AddCardMovedToPileObserver(observer CardMovedToPileObserver)
	GetCardMovedToPileObserver() CardMovedToPileObserver
}

type TradeCard interface {
	CanBeTraded() bool
}

type CardMovedToPileObserver interface {
	OnCardMovedToPile(card Card)
}

type WarriorDeadObserver interface {
	OnWarriorDead(card Warrior)
}

type cardBase struct {
	id                      string
	name                    string
	cardMovedToPileObserver CardMovedToPileObserver
	warriorDeadObserver     WarriorDeadObserver
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

func (c *cardBase) Name() string {
	return c.name
}
func (c *cardBase) AddCardMovedToPileObserver(observer CardMovedToPileObserver) {
	c.cardMovedToPileObserver = observer
}
func (c *cardBase) GetCardMovedToPileObserver() CardMovedToPileObserver {
	return c.cardMovedToPileObserver
}

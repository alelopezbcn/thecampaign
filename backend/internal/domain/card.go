package domain

import (
	"errors"
	"fmt"
	"strings"
)

const (
	WarriorHealth       = 20
	DragonHealth        = 20
	SpecialMPowerHealth = 10
)

type iCard interface {
	GetID() string
	GetValue() int
	SetPlayer(player *Player)
	String() string
	AffectedBy() []iCard
	AddObserver(o WarriorDeadObserver)
}

type card struct {
	ID         string
	Name       string
	Value      int
	affectedBy []iCard
	Player     *Player
	Observer   WarriorDeadObserver
}

func (c *card) GetID() string {
	return c.ID
}

func (c *card) GetValue() int {
	return c.Value
}

func (c *card) SetPlayer(player *Player) {
	c.Player = player
}
func (c *card) AffectedBy() []iCard {
	return c.affectedBy
}

func (c *card) AddObserver(o WarriorDeadObserver) {
	c.Observer = o
}

type attacker interface {
	Attack(target, weapon iCard) error
}

type attackable interface {
	ReceiveDamage(amount int, weapon iCard)
}

type weapon interface {
	iCard
	CanAttack()
}

type resource interface {
	iCard
	CanBuy()
}

type warrior interface {
	iCard
	attacker
	attackable
}

type warriorCard struct {
	card
}

func (w *warriorCard) ReceiveDamage(amount int, weapon iCard) {
	w.Value -= amount
	w.affectedBy = append(w.affectedBy, weapon)

	if w.Value <= 0 {
		w.Observer.OnWarriorDead(w.Player, w)

	}
}
func (w *warriorCard) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("%s (%s)", w.Name, w.ID))
	if w.Value > 0 {
		sb.WriteString(fmt.Sprintf(" - Value: %d", w.Value))
	}
	if w.affectedBy != nil && len(w.affectedBy) > 0 {
		for _, card := range w.affectedBy {
			sb.WriteString(fmt.Sprintf("\n     * %s", card.String()))
		}
	}
	return sb.String()
}

type knightCard struct {
	warriorCard
}

func newKnightCard(id string) warrior {
	return &knightCard{
		warriorCard: warriorCard{
			card: card{
				ID:         strings.ToUpper(id),
				Name:       "Knight",
				Value:      WarriorHealth,
				affectedBy: []iCard{},
			},
		},
	}
}
func (k *knightCard) Attack(targetCard, weaponCard iCard) error {

	if _, ok := targetCard.(attackable); !ok {
		return fmt.Errorf("target cannot be attacked")
	}
	if _, ok := weaponCard.(weapon); !ok {
		return fmt.Errorf("card is not a weapon")
	}
	_, ok := weaponCard.(*swordCard)
	if !ok {
		return errors.New("knight can only attack with sword")
	}

	multiplier := 1
	if _, ok := targetCard.(*archerCard); ok {
		multiplier = 2
	}

	damage := weaponCard.GetValue() * multiplier
	targetCard.(attackable).ReceiveDamage(damage, weaponCard)

	return nil
}

type archerCard struct {
	warriorCard
}

func newArcherCard(id string) warrior {
	return &archerCard{
		warriorCard: warriorCard{
			card: card{
				ID:         strings.ToUpper(id),
				Name:       "Archer",
				Value:      WarriorHealth,
				affectedBy: []iCard{},
			},
		},
	}
}
func (a *archerCard) Attack(targetCard, weaponCard iCard) error {
	if _, ok := targetCard.(attackable); !ok {
		return fmt.Errorf("target cannot be attacked")
	}
	if _, ok := weaponCard.(weapon); !ok {
		return fmt.Errorf("card is not a weapon")
	}
	_, ok := weaponCard.(*arrowCard)
	if !ok {
		return errors.New("archer can only attack with arrow")
	}

	multiplier := 1
	if _, ok := targetCard.(*mageCard); ok {
		multiplier = 2
	}

	damage := weaponCard.GetValue() * multiplier
	targetCard.(attackable).ReceiveDamage(damage, weaponCard)

	return nil
}

type mageCard struct {
	warriorCard
}

func newMageCard(id string) warrior {
	return &mageCard{
		warriorCard: warriorCard{
			card: card{
				ID:         strings.ToUpper(id),
				Name:       "Mage",
				Value:      WarriorHealth,
				affectedBy: []iCard{},
			},
		},
	}
}
func (m *mageCard) Attack(targetCard, weaponCard iCard) error {
	if _, ok := targetCard.(attackable); !ok {
		return fmt.Errorf("target cannot be attacked")
	}
	if _, ok := weaponCard.(weapon); !ok {
		return fmt.Errorf("card is not a weapon")
	}
	_, ok := weaponCard.(*poisonCard)
	if !ok {
		return errors.New("mage can only attack with poison")
	}

	multiplier := 1
	if _, ok := targetCard.(*knightCard); ok {
		multiplier = 2
	}

	damage := weaponCard.GetValue() * multiplier
	targetCard.(attackable).ReceiveDamage(damage, weaponCard)

	return nil
}

type dragonCard struct {
	warriorCard
}

func newDragonCard(id string) warrior {
	return &dragonCard{
		warriorCard: warriorCard{
			card: card{
				ID:         strings.ToUpper(id),
				Name:       "Dragon",
				Value:      DragonHealth,
				affectedBy: []iCard{},
			},
		},
	}
}
func (d *dragonCard) Attack(targetCard, weaponCard iCard) error {
	if _, ok := targetCard.(attackable); !ok {
		return fmt.Errorf("target cannot be attacked")
	}
	if _, ok := weaponCard.(weapon); !ok {
		return fmt.Errorf("card is not a weapon")
	}

	multiplier := 1

	switch weaponCard.(type) {
	case *swordCard:
		if _, ok := targetCard.(*archerCard); ok {
			multiplier = 2
		}
	case *arrowCard:
		if _, ok := targetCard.(*mageCard); ok {
			multiplier = 2
		}
	case *poisonCard:
		if _, ok := targetCard.(*knightCard); ok {
			multiplier = 2
		}
	}

	damage := weaponCard.GetValue() * multiplier
	targetCard.(attackable).ReceiveDamage(damage, weaponCard)

	return nil
}

type swordCard struct {
	card
}

func newSwordCard(id string, value int) weapon {
	return &swordCard{
		card: card{
			ID:         strings.ToUpper(id),
			Name:       "Sword",
			Value:      value,
			affectedBy: []iCard{},
		},
	}
}
func (s *swordCard) CanAttack() {}
func (s *swordCard) String() string {
	return fmt.Sprintf("%d %s (%s)", s.Value, s.Name, s.ID)
}

type arrowCard struct {
	card
}

func newArrowCard(id string, value int) weapon {
	return &arrowCard{
		card: card{
			ID:         strings.ToUpper(id),
			Name:       "Arrow",
			Value:      value,
			affectedBy: []iCard{},
		},
	}
}
func (a *arrowCard) CanAttack() {}
func (a *arrowCard) String() string {
	return fmt.Sprintf("%d %s (%s)", a.Value, a.Name, a.ID)
}

type poisonCard struct {
	card
}

func newPoisonCard(id string, value int) weapon {
	{
		return &poisonCard{
			card: card{
				ID:         strings.ToUpper(id),
				Name:       "Poison",
				Value:      value,
				affectedBy: []iCard{},
			},
		}
	}
}
func (p *poisonCard) CanAttack() {}
func (p *poisonCard) String() string {
	return fmt.Sprintf("%d %s (%s)", p.Value, p.Name, p.ID)
}

type specialPowerCard struct {
	card
}

func newSpecialPowerCard(id string) *specialPowerCard {
	return &specialPowerCard{
		card: card{
			ID:         strings.ToUpper(id),
			Name:       "Special Power",
			Value:      SpecialMPowerHealth,
			affectedBy: []iCard{},
		},
	}
}
func (s *specialPowerCard) Attack(_, _ iCard) error {
	return errors.New("special power attack not implemented yet")
}
func (s *specialPowerCard) ReceiveDamage(amount int) {
	s.Value -= amount
}
func (s *specialPowerCard) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("%s (%s)", s.Name, s.ID))
	if s.Value > 0 {
		sb.WriteString(fmt.Sprintf(" - Value: %d", s.Value))
	}
	if s.affectedBy != nil && len(s.affectedBy) > 0 {
		for _, card := range s.affectedBy {
			sb.WriteString(fmt.Sprintf("\n  * %s", card.String()))
		}
	}
	return sb.String()
}

type spyCard struct {
	card
}

func newSpyCard(id string) iCard {
	return &spyCard{
		card: card{
			ID:   strings.ToUpper(id),
			Name: "Spy",
		},
	}
}
func (s *spyCard) String() string {
	return fmt.Sprintf("%s (%s)", s.Name, s.ID)
}

type thiefCard struct {
	card
}

func newThiefCard(id string) iCard {
	return &thiefCard{
		card: card{
			ID:   strings.ToUpper(id),
			Name: "Thief",
		},
	}
}
func (t *thiefCard) String() string {
	return fmt.Sprintf("%s (%s)", t.Name, t.ID)
}

type goldCard struct {
	card
}

func newGoldCard(id string, value int) resource {
	return &goldCard{
		card: card{
			ID:    strings.ToUpper(id),
			Name:  "Gold Coin",
			Value: value,
		},
	}
}
func (g *goldCard) CanBuy() {}
func (g *goldCard) String() string {
	return fmt.Sprintf("%d %s (%s)", g.Value, g.Name, g.ID)
}

type catapultCard struct {
	card
}

func newCatapultCard(id string) *catapultCard {
	return &catapultCard{
		card: card{
			ID:   strings.ToUpper(id),
			Name: "Catapult",
		},
	}
}
func (c *catapultCard) Attack(castle *Castle, position int) (resource, error) {
	gold, err := castle.RemoveGold(position)
	if err != nil {
		return nil, err
	}

	return gold, nil
}
func (c *catapultCard) String() string {
	return fmt.Sprintf("%s (%s)", c.Name, c.ID)
}

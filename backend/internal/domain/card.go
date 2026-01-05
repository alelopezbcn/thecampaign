package domain

import (
	"errors"
	"fmt"
	"strings"
)

const (
	WarriorHealth     = 10
	DragonHealth      = 20
	MaxHandSize       = 7
	SpecialMoveHealth = 10
)

type card struct {
	ID         string
	Name       string
	Value      int
	AffectedBy []iCard
}

func (c *card) GetID() string {
	return c.ID
}

func (c *card) GetValue() int {
	return c.Value
}

type iCard interface {
	GetID() string
	GetValue() int
	IsWarrior() bool
	IsResource() bool
	Attack(target, weapon iCard) error
	String() string
}

type knightCard struct {
	card
}

func newKnightCard(id string) *knightCard {
	return &knightCard{
		card: card{
			ID:         id,
			Name:       "Knight",
			Value:      WarriorHealth,
			AffectedBy: []iCard{},
		},
	}
}
func (k *knightCard) IsWarrior() bool {
	return true
}
func (k *knightCard) IsResource() bool {
	return false
}
func (k *knightCard) Attack(target, weapon iCard) error {
	return nil
}
func (k *knightCard) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("%s (%s)", k.Name, k.ID))
	if k.Value > 0 {
		sb.WriteString(fmt.Sprintf(" - Value: %d", k.Value))
	}
	if k.AffectedBy != nil && len(k.AffectedBy) > 0 {
		for _, card := range k.AffectedBy {
			sb.WriteString(fmt.Sprintf("\n  * %s", card.String()))
		}
	}
	return sb.String()
}

type archerCard struct {
	card
}

func newArcherCard(id string) *archerCard {
	return &archerCard{
		card: card{
			ID:         id,
			Name:       "Archer",
			Value:      WarriorHealth,
			AffectedBy: []iCard{},
		},
	}
}
func (a *archerCard) IsWarrior() bool {
	return true
}
func (a *archerCard) IsResource() bool {
	return false
}
func (a *archerCard) Attack(target, weapon iCard) error {
	return nil
}
func (a *archerCard) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("%s (%s)", a.Name, a.ID))
	if a.Value > 0 {
		sb.WriteString(fmt.Sprintf(" - Value: %d", a.Value))
	}
	if a.AffectedBy != nil && len(a.AffectedBy) > 0 {
		for _, card := range a.AffectedBy {
			sb.WriteString(fmt.Sprintf("\n  * %s", card.String()))
		}
	}
	return sb.String()
}

type mageCard struct {
	card
}

func newMageCard(id string) *mageCard {
	return &mageCard{
		card: card{
			ID:         id,
			Name:       "Mage",
			Value:      WarriorHealth,
			AffectedBy: []iCard{},
		},
	}
}
func (m *mageCard) IsWarrior() bool {
	return true
}
func (m *mageCard) IsResource() bool {
	return false
}
func (m *mageCard) Attack(target, weapon iCard) error {
	return errors.New("not implemented")
}
func (m *mageCard) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("%s (%s)", m.Name, m.ID))
	if m.Value > 0 {
		sb.WriteString(fmt.Sprintf(" - Value: %d", m.Value))
	}
	if m.AffectedBy != nil && len(m.AffectedBy) > 0 {
		for _, card := range m.AffectedBy {
			sb.WriteString(fmt.Sprintf("\n  * %s", card.String()))
		}
	}
	return sb.String()
}

type swordCard struct {
	card
}

func newSwordCard(id string, value int) *swordCard {
	return &swordCard{
		card: card{
			ID:         id,
			Name:       "Sword",
			Value:      value,
			AffectedBy: []iCard{},
		},
	}
}
func (s *swordCard) IsWarrior() bool {
	return false
}
func (s *swordCard) IsResource() bool {
	return false
}
func (s *swordCard) Attack(_, _ iCard) error {
	return errors.New("swords can perform attack")
}
func (s *swordCard) String() string {
	return fmt.Sprintf("%d %s (%s)", s.Value, s.Name, s.ID)
}

type arrowCard struct {
	card
}

func newArrowCard(id string, value int) *arrowCard {
	return &arrowCard{
		card: card{
			ID:         id,
			Name:       "Arrow",
			Value:      value,
			AffectedBy: []iCard{},
		},
	}
}
func (a *arrowCard) IsWarrior() bool {
	return false
}
func (a *arrowCard) IsResource() bool {
	return false
}
func (a *arrowCard) Attack(_, _ iCard) error {
	return errors.New("arrow can perform attack")
}
func (a *arrowCard) String() string {
	return fmt.Sprintf("%d %s (%s)", a.Value, a.Name, a.ID)
}

type poisonCard struct {
	card
}

func newPoisonCard(id string, value int) *poisonCard {
	{
		return &poisonCard{
			card: card{
				ID:         id,
				Name:       "Poison",
				Value:      value,
				AffectedBy: []iCard{},
			},
		}
	}
}
func (p *poisonCard) IsWarrior() bool {
	return false
}
func (p *poisonCard) IsResource() bool {
	return false
}
func (p *poisonCard) Attack(_, _ iCard) error {
	return errors.New("poison can perform attack")
}
func (p *poisonCard) String() string {
	return fmt.Sprintf("%d %s (%s)", p.Value, p.Name, p.ID)
}

type dragonCard struct {
	card
}

func newDragonCard(id string) *dragonCard {
	return &dragonCard{
		card: card{
			ID:         id,
			Name:       "Dragon",
			Value:      DragonHealth,
			AffectedBy: []iCard{},
		},
	}
}
func (d *dragonCard) IsWarrior() bool {
	return false
}
func (d *dragonCard) IsResource() bool {
	return false
}
func (d *dragonCard) Attack(target, weapon iCard) error {
	return errors.New("dragon attack not implemented yet")
}
func (d *dragonCard) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("%s (%s)", d.Name, d.ID))
	if d.Value > 0 {
		sb.WriteString(fmt.Sprintf(" - Value: %d", d.Value))
	}
	if d.AffectedBy != nil && len(d.AffectedBy) > 0 {
		for _, card := range d.AffectedBy {
			sb.WriteString(fmt.Sprintf("\n  * %s", card.String()))
		}
	}
	return sb.String()
}

type specialMoveCard struct {
	card
}

func newSpecialMoveCard(id string) *specialMoveCard {
	return &specialMoveCard{
		card: card{
			ID:         id,
			Name:       "Special Move",
			Value:      SpecialMoveHealth,
			AffectedBy: []iCard{},
		},
	}
}
func (s *specialMoveCard) IsWarrior() bool {
	return false
}
func (s *specialMoveCard) IsResource() bool {
	return false
}
func (s *specialMoveCard) Attack(_, _ iCard) error {
	return errors.New("special move attack not implemented yet")
}
func (s *specialMoveCard) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("%s (%s)", s.Name, s.ID))
	if s.Value > 0 {
		sb.WriteString(fmt.Sprintf(" - Value: %d", s.Value))
	}
	if s.AffectedBy != nil && len(s.AffectedBy) > 0 {
		for _, card := range s.AffectedBy {
			sb.WriteString(fmt.Sprintf("\n  * %s", card.String()))
		}
	}
	return sb.String()
}

type spyCard struct {
	card
}

func newSpyCard(id string) *spyCard {
	return &spyCard{
		card: card{
			ID:   id,
			Name: "Spy",
		},
	}
}
func (s *spyCard) IsWarrior() bool {
	return false
}
func (s *spyCard) IsResource() bool {
	return false
}
func (s *spyCard) Attack(_, _ iCard) error {
	return errors.New("spy cannot attack")
}
func (s *spyCard) String() string {
	return fmt.Sprintf("%s (%s)", s.Name, s.ID)
}

type thiefCard struct {
	card
}

func newThiefCard(id string) *thiefCard {
	return &thiefCard{
		card: card{
			ID:   id,
			Name: "Thief",
		},
	}
}
func (t *thiefCard) IsWarrior() bool {
	return false
}
func (t *thiefCard) IsResource() bool {
	return false
}
func (t *thiefCard) Attack(_, _ iCard) error {
	return errors.New("thief cannot attack")
}
func (t *thiefCard) String() string {
	return fmt.Sprintf("%s (%s)", t.Name, t.ID)
}

type goldCard struct {
	card
}

func newGoldCard(id string, value int) *goldCard {
	return &goldCard{
		card: card{
			ID:    id,
			Name:  "Gold Coin",
			Value: value,
		},
	}
}
func (g *goldCard) IsWarrior() bool {
	return false
}
func (g *goldCard) IsResource() bool {
	return true
}
func (g *goldCard) Attack(_, _ iCard) error {
	return errors.New("money cannot attack")
}
func (g *goldCard) String() string {
	return fmt.Sprintf("%d %s (%s)", g.Value, g.Name, g.ID)
}

type catapultCard struct {
	card
}

func newCatapultCard(id string) *catapultCard {
	return &catapultCard{
		card: card{
			ID:   id,
			Name: "Catapult",
		},
	}
}
func (c *catapultCard) IsWarrior() bool {
	return false
}
func (c *catapultCard) IsResource() bool {
	return false
}
func (c *catapultCard) Attack(target, weapon iCard) error {
	return errors.New("catapult attack not implemented yet")
}
func (c *catapultCard) String() string {
	return fmt.Sprintf("%s (%s)", c.Name, c.ID)
}

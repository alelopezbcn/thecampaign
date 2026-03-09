package gameactions

import (
	"errors"
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// forgeGame declares the minimum Game surface needed by forgeAction
type forgeGame interface {
	GamePlayers
	GameTurn
	GameTurnFlags
	GameHistory
	GameStatusProvider
}

type forgeAction struct {
	playerName   string
	cardID1      string
	cardID2      string
	currentPhase types.PhaseType

	weapon1 cards.Weapon
	weapon2 cards.Weapon
}

func NewForgeAction(playerName, cardID1, cardID2 string) *forgeAction {
	return &forgeAction{
		playerName: playerName,
		cardID1:    cardID1,
		cardID2:    cardID2,
	}
}

func (a *forgeAction) PlayerName() string { return a.playerName }

var forgeableTypes = []types.WeaponType{
	types.SwordWeaponType,
	types.ArrowWeaponType,
	types.PoisonWeaponType,
}

func (a *forgeAction) Validate(g Game) error {
	if g.TurnState().HasForged {
		return errors.New("already forged this turn")
	}

	p := g.CurrentPlayer()

	card1, ok := p.GetCardFromHand(a.cardID1)
	if !ok {
		return fmt.Errorf("card not in hand: %s", a.cardID1)
	}
	card2, ok := p.GetCardFromHand(a.cardID2)
	if !ok {
		return fmt.Errorf("card not in hand: %s", a.cardID2)
	}

	w1, ok := card1.(cards.Weapon)
	if !ok {
		return fmt.Errorf("card %s is not a weapon", a.cardID1)
	}
	w2, ok := card2.(cards.Weapon)
	if !ok {
		return fmt.Errorf("card %s is not a weapon", a.cardID2)
	}

	if w1.Type() != w2.Type() {
		return fmt.Errorf("cannot forge different weapon types: %s and %s", w1.Type(), w2.Type())
	}

	forgeable := false
	for _, ft := range forgeableTypes {
		if w1.Type() == ft {
			forgeable = true
			break
		}
	}
	if !forgeable {
		return fmt.Errorf("weapon type %s cannot be forged", w1.Type())
	}

	a.weapon1 = w1
	a.weapon2 = w2
	return nil
}

func (a *forgeAction) Execute(g Game) (*Result, func() gamestatus.GameStatus, error) {
	return a.execute(g)
}

func (a *forgeAction) execute(g forgeGame) (*Result, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()
	result := &Result{}

	if _, err := p.RemoveFromHand(a.cardID1, a.cardID2); err != nil {
		return result, nil, fmt.Errorf("removing weapons for forging failed: %w", err)
	}

	forgedDamage := a.weapon1.DamageAmount() + a.weapon2.DamageAmount()
	forgedID := "forged_" + a.cardID1 + "_" + a.cardID2
	newWeapon := createWeaponByType(a.weapon1.Type(), forgedID, forgedDamage)

	p.TakeCards(newWeapon)

	// Override the observer set by TakeCards so that when the forged weapon is
	// discarded (e.g. the warrior that used it dies), the original component
	// cards are discarded separately instead of the combined card.
	// Must be set after TakeCards, which would otherwise overwrite it.
	// Recursive forging is handled automatically: if a component is itself a
	// forged weapon, its own unforgeObserver will further split it on discard.
	newWeapon.AddCardMovedToPileObserver(&unforgeObserver{
		components: []cards.Weapon{a.weapon1, a.weapon2},
	})

	g.AddHistory(fmt.Sprintf("%s forged a %s %d", p.Name(), newWeapon.Name(), forgedDamage),
		types.CategoryAction)

	g.SetHasForged(true)
	g.SetCanForge(false)

	result.Action = types.LastActionForge
	a.currentPhase = g.CurrentAction()

	statusFn := func() gamestatus.GameStatus {
		return g.Status(p, newWeapon)
	}

	return result, statusFn, nil
}

func (a *forgeAction) NextPhase() types.PhaseType {
	return a.currentPhase
}

// unforgeObserver implements CardMovedToPileObserver for forged weapons.
// When the forged weapon is discarded, it recursively discards each original
// component card via its own observer instead of discarding the combined card.
type unforgeObserver struct {
	components []cards.Weapon
}

func (o *unforgeObserver) OnCardMovedToPile(_ cards.Card) {
	for _, c := range o.components {
		c.GetCardMovedToPileObserver().OnCardMovedToPile(c)
	}
}

func createWeaponByType(wt types.WeaponType, id string, damage int) cards.Weapon {
	switch wt {
	case types.SwordWeaponType:
		return cards.NewSword(id, damage)
	case types.ArrowWeaponType:
		return cards.NewArrow(id, damage)
	case types.PoisonWeaponType:
		return cards.NewPoison(id, damage)
	}
	return nil
}

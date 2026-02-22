package game

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type attackAction struct {
	playerName       string
	targetPlayerName string
	targetID         string
	weaponID         string

	currentPlayer board.Player
	target        cards.Attackable
	weapon        cards.Weapon
}

func NewAttackAction(playerName, targetPlayerName, targetID, weaponID string) *attackAction {
	return &attackAction{
		playerName:       playerName,
		targetPlayerName: targetPlayerName,
		targetID:         targetID,
		weaponID:         weaponID,
	}
}

func (a *attackAction) PlayerName() string { return a.playerName }

func (a *attackAction) Validate(g Game) error {
	if g.CurrentAction() != types.PhaseTypeAttack {
		return fmt.Errorf("cannot attack in the %s phase", g.CurrentAction())
	}

	targetPlayer, err := g.GetTargetPlayer(a.playerName, a.targetPlayerName)
	if err != nil {
		return err
	}

	targetCard, ok := targetPlayer.GetCardFromField(a.targetID)
	if !ok {
		return fmt.Errorf("target card not in enemy field: %s", a.targetID)
	}

	p := g.CurrentPlayer()
	a.currentPlayer = p
	weaponCard, ok := p.GetCardFromHand(a.weaponID)
	if !ok {
		return fmt.Errorf("weapon card not in hand: %s", a.weaponID)
	}

	a.target, ok = targetCard.(cards.Attackable)
	if !ok {
		return fmt.Errorf("the target card cannot be attacked")
	}

	a.weapon, ok = weaponCard.(cards.Weapon)
	if !ok {
		return fmt.Errorf("the card is not a weapon")
	}

	switch a.weapon.Type() {
	case types.SwordWeaponType:
		if !a.currentPlayer.Field().HasKnight() && !a.currentPlayer.Field().HasDragon() {
			return fmt.Errorf("sword weapon cannot be used")
		}
	case types.ArrowWeaponType:
		if !a.currentPlayer.Field().HasArcher() && !a.currentPlayer.Field().HasDragon() {
			return fmt.Errorf("arrow weapon cannot be used")
		}
	case types.PoisonWeaponType:
		if !a.currentPlayer.Field().HasMage() && !a.currentPlayer.Field().HasDragon() {
			return fmt.Errorf("poison weapon cannot be used")
		}
	}

	return nil
}

func (a *attackAction) Execute(g Game) (*GameActionResult, func() gamestatus.GameStatus, error) {
	err := a.target.BeAttacked(a.weapon)
	if err != nil {
		result := &GameActionResult{}
		return result, nil, fmt.Errorf("attack action failed: %w", err)
	}

	a.currentPlayer.RemoveFromHand(a.weaponID)

	g.AddHistory(fmt.Sprintf("%s attacked %s with %s",
		a.currentPlayer.Name(), a.target.String(), a.weapon.String()),
		types.CategoryAction)

	result := &GameActionResult{
		Action:             types.LastActionAttack,
		AttackWeaponID:     a.weaponID,
		AttackTargetID:     a.targetID,
		AttackTargetPlayer: a.targetPlayerName,
	}
	statusFn := func() gamestatus.GameStatus {
		return g.GameStatusProvider().Get(a.currentPlayer, g)
	}

	return result, statusFn, nil
}

func (a *attackAction) NextPhase() types.PhaseType {
	return types.PhaseTypeSpySteal
}

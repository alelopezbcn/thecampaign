package domain

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type AttackAction struct {
	playerName       string
	targetPlayerName string
	targetID         string
	weaponID         string
	target           ports.Attackable
	weapon           ports.Weapon
}

func NewAttackAction(playerName, targetPlayerName, targetID, weaponID string) *AttackAction {
	return &AttackAction{
		playerName:       playerName,
		targetPlayerName: targetPlayerName,
		targetID:         targetID,
		weaponID:         weaponID,
	}
}

func (a *AttackAction) PlayerName() string { return a.playerName }

func (a *AttackAction) Validate(g *Game) error {
	if g.currentAction != types.ActionTypeAttack {
		return fmt.Errorf("cannot attack in the %s phase", g.currentAction)
	}

	targetPlayer, err := g.getTargetPlayer(a.playerName, a.targetPlayerName)
	if err != nil {
		return err
	}

	targetCard, ok := targetPlayer.GetCardFromField(a.targetID)
	if !ok {
		return fmt.Errorf("target card not in enemy field: %s", a.targetID)
	}

	p := g.CurrentPlayer()
	weaponCard, ok := p.GetCardFromHand(a.weaponID)
	if !ok {
		return fmt.Errorf("weapon card not in hand: %s", a.weaponID)
	}

	a.target, ok = targetCard.(ports.Attackable)
	if !ok {
		return fmt.Errorf("the target cardBase cannot be attacked")
	}

	a.weapon, ok = weaponCard.(ports.Weapon)
	if !ok {
		return fmt.Errorf("the card is not a weapon")
	}

	return nil
}

func (a *AttackAction) Execute(g *Game) (*GameActionResult, func() GameStatus, error) {
	p := g.CurrentPlayer()

	if err := p.Attack(a.target, a.weapon); err != nil {
		result := &GameActionResult{}
		return result, nil, fmt.Errorf("attack action failed: %w", err)
	}

	g.addToHistory(fmt.Sprintf("%s attacked %s with %s",
		p.Name(), a.target.String(), a.weapon.String()),
		types.CategoryAction)

	result := &GameActionResult{
		Action:             types.LastActionAttack,
		AttackWeaponID:     a.weaponID,
		AttackTargetID:     a.targetID,
		AttackTargetPlayer: a.targetPlayerName,
	}
	statusFn := func() GameStatus {
		return g.GameStatusProvider.Get(p, g)
	}

	return result, statusFn, nil
}

func (a *AttackAction) NextPhase() types.ActionType {
	return types.ActionTypeAttack
}

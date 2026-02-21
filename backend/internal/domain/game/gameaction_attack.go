package game

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type attackAction struct {
	playerName       string
	targetPlayerName string
	targetID         string
	weaponID         string
	target           cards.Attackable
	weapon           cards.Weapon
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

func (a *attackAction) Validate(g *Game) error {
	if g.currentAction != types.PhaseTypeAttack {
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

	a.target, ok = targetCard.(cards.Attackable)
	if !ok {
		return fmt.Errorf("the target card cannot be attacked")
	}

	a.weapon, ok = weaponCard.(cards.Weapon)
	if !ok {
		return fmt.Errorf("the card is not a weapon")
	}

	return nil
}

func (a *attackAction) Execute(g *Game) (*GameActionResult, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()

	if err := p.Attack(a.target, a.weapon); err != nil {
		result := &GameActionResult{}
		return result, nil, fmt.Errorf("attack action failed: %w", err)
	}

	g.AddHistory(fmt.Sprintf("%s attacked %s with %s",
		p.Name(), a.target.String(), a.weapon.String()),
		types.CategoryAction)

	result := &GameActionResult{
		Action:             types.LastActionAttack,
		AttackWeaponID:     a.weaponID,
		AttackTargetID:     a.targetID,
		AttackTargetPlayer: a.targetPlayerName,
	}
	statusFn := func() gamestatus.GameStatus {
		return g.gameStatusProvider.Get(p, g)
	}

	return result, statusFn, nil
}

func (a *attackAction) NextPhase() types.PhaseType {
	return types.PhaseTypeSpySteal
}

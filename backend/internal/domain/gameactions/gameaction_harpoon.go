package gameactions

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// harpoonGame declares the minimum Game surface needed by harpoonAction
type harpoonGame interface {
	GamePlayers
	GameTurn
	GameHistory
	GameStatusProvider
}

type harpoonAction struct {
	playerName       string
	targetPlayerName string
	targetID         string
	weaponID         string

	dragon  cards.Dragon
	harpoon cards.Harpoon
}

func NewHarpoonAction(playerName, targetPlayerName, targetID, weaponID string) *harpoonAction {
	return &harpoonAction{
		playerName:       playerName,
		targetPlayerName: targetPlayerName,
		weaponID:         weaponID,
		targetID:         targetID,
	}
}

func (a *harpoonAction) PlayerName() string { return a.playerName }

func (a *harpoonAction) Validate(g Game) error {
	if g.CurrentAction() != types.PhaseTypeAttack {
		return fmt.Errorf("cannot use harpoon in the %s phase",
			g.CurrentAction())
	}

	targetPlayer, err := g.GetTargetPlayer(a.playerName, a.targetPlayerName)
	if err != nil {
		return err
	}

	targetCard, ok := targetPlayer.GetCardFromField(a.targetID)
	if !ok {
		return fmt.Errorf("dragon card not in enemy field: %s", a.targetID)
	}

	a.dragon, ok = targetCard.(cards.Dragon)
	if !ok {
		return fmt.Errorf("the target card is not a dragon")
	}

	p := g.CurrentPlayer()
	// Look up the specific harpoon card by the ID provided by the client.
	// This avoids the mismatch possible in FFA5 where both HA1 and HA2 exist.
	harpoonCard, ok := p.GetCardFromHand(a.weaponID)
	if !ok {
		return fmt.Errorf("player does not have a harpoon to use")
	}
	harpoon, ok := harpoonCard.(cards.Harpoon)
	if !ok {
		return fmt.Errorf("card %s is not a harpoon", a.weaponID)
	}

	a.harpoon = harpoon

	return nil
}

func (a *harpoonAction) Execute(g Game) (*Result, func() gamestatus.GameStatus, error) {
	return a.execute(g)
}

func (a *harpoonAction) execute(g harpoonGame) (*Result, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()

	if err := a.harpoon.Attack(a.dragon); err != nil {
		result := &Result{}
		return result, nil, fmt.Errorf("harpoon action failed: %w", err)
	}

	if _, err := p.RemoveFromHand(a.harpoon.GetID()); err != nil {
		result := &Result{}
		return result, nil, fmt.Errorf("removing harpoon from hand failed: %w", err)
	}

	g.AddHistory(fmt.Sprintf("%s used harpoon on %s",
		a.playerName, a.dragon.String()), types.CategoryAction)

	result := &Result{
		Action:             types.LastActionHarpoon,
		AttackWeaponID:     a.weaponID,
		AttackTargetID:     a.targetID,
		AttackTargetPlayer: a.targetPlayerName,
	}
	statusFn := func() gamestatus.GameStatus {
		return g.Status(p)
	}

	return result, statusFn, nil
}

func (a *harpoonAction) NextPhase() types.PhaseType {
	return types.PhaseTypeSpySteal
}

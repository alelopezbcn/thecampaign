package game

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

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

func (a *harpoonAction) Validate(g *Game) error {
	if g.currentAction != types.PhaseTypeAttack {
		return fmt.Errorf("cannot use harpoon in the %s phase",
			g.currentAction)
	}

	targetPlayer, err := g.getTargetPlayer(a.playerName, a.targetPlayerName)
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
	harpoon, ok := board.HasCardTypeInHand[cards.Harpoon](p)
	if !ok {
		return fmt.Errorf("player does not have a harpoon to use")
	}

	a.harpoon = harpoon

	return nil
}

func (a *harpoonAction) Execute(g *Game) (*GameActionResult, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()

	if err := a.harpoon.Attack(a.dragon); err != nil {
		result := &GameActionResult{}
		return result, nil, fmt.Errorf("harpoon action failed: %w", err)
	}

	if _, err := p.RemoveFromHand(a.weaponID); err != nil {
		result := &GameActionResult{}
		return result, nil, fmt.Errorf("removing harpoon from hand failed: %w", err)
	}

	g.AddHistory(fmt.Sprintf("%s used harpoon on %s",
		a.playerName, a.dragon.String()), types.CategoryAction)

	result := &GameActionResult{
		Action: types.LastActionHarpoon,
	}
	statusFn := func() gamestatus.GameStatus {
		return g.gameStatusProvider.Get(p, g)
	}

	return result, statusFn, nil
}

func (a *harpoonAction) NextPhase() types.PhaseType {
	return types.PhaseTypeSpySteal
}

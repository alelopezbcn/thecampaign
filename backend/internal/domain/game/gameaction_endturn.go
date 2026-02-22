package game

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type endTurnPhaseAction struct {
	playerName string
	expired    bool
}

func NewEndTurnPhaseAction(playerName string, expired bool) *endTurnPhaseAction {
	return &endTurnPhaseAction{playerName: playerName, expired: expired}
}

func (a *endTurnPhaseAction) PlayerName() string { return a.playerName }

func (a *endTurnPhaseAction) Validate(g *game) error {
	return nil
}

func (a *endTurnPhaseAction) Execute(g *game) (*GameActionResult, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()

	if a.expired {
		g.AddHistory(fmt.Sprintf("%s's turn expired", p.Name()),
			types.CategoryTurnExpired)
	} else {
		g.AddHistory(fmt.Sprintf("%s ended their turn", p.Name()),
			types.CategoryEndTurn)
	}

	result := &GameActionResult{Action: types.LastActionEndTurn}
	g.switchTurn()

	statusFn := func() gamestatus.GameStatus {
		return g.gameStatusProvider.Get(g.CurrentPlayer(), g)
	}

	return result, statusFn, nil
}

func (a *endTurnPhaseAction) NextPhase() types.PhaseType {
	return types.PhaseTypeDrawCard
}

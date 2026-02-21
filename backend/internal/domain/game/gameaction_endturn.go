package game

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type EndTurnPhaseAction struct {
	playerName string
	expired    bool
}

func NewEndTurnPhaseAction(playerName string, expired bool) *EndTurnPhaseAction {
	return &EndTurnPhaseAction{playerName: playerName, expired: expired}
}

func (a *EndTurnPhaseAction) PlayerName() string { return a.playerName }

func (a *EndTurnPhaseAction) Validate(g *Game) error {
	return nil
}

func (a *EndTurnPhaseAction) Execute(g *Game) (*GameActionResult, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()

	if a.expired {
		g.addToHistory(fmt.Sprintf("%s's turn expired", p.Name()),
			types.CategoryTurnExpired)
	} else {
		g.addToHistory(fmt.Sprintf("%s ended their turn", p.Name()),
			types.CategoryEndTurn)
	}

	result := &GameActionResult{Action: types.LastActionEndTurn}
	g.switchTurn()

	statusFn := func() gamestatus.GameStatus {
		return g.gameStatusProvider.Get(g.CurrentPlayer(), g)
	}

	return result, statusFn, nil
}

func (a *EndTurnPhaseAction) NextPhase() types.PhaseType {
	return types.PhaseTypeDrawCard
}

package gameactions

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// endTurnGame declares the minimum Game surface needed by endTurnPhaseAction
type endTurnGame interface {
	GamePlayers
	GameTurn
	GameHistory
	GameStatusProvider
}

type endTurnPhaseAction struct {
	playerName string
	expired    bool
}

func NewEndTurnPhaseAction(playerName string, expired bool) *endTurnPhaseAction {
	return &endTurnPhaseAction{playerName: playerName, expired: expired}
}

func (a *endTurnPhaseAction) PlayerName() string { return a.playerName }

func (a *endTurnPhaseAction) Validate(g Game) error {
	return nil
}

func (a *endTurnPhaseAction) Execute(g Game) (*Result, func() gamestatus.GameStatus, error) {
	return a.execute(g)
}

func (a *endTurnPhaseAction) execute(g endTurnGame) (*Result, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()

	if a.expired {
		g.AddHistory(fmt.Sprintf("%s's turn expired", p.Name()),
			types.CategoryTurnExpired)
	} else {
		g.AddHistory(fmt.Sprintf("%s ended their turn", p.Name()),
			types.CategoryEndTurn)
	}

	result := &Result{Action: types.LastActionEndTurn}
	g.SwitchTurn()

	statusFn := func() gamestatus.GameStatus {
		return g.Status(g.CurrentPlayer())
	}

	return result, statusFn, nil
}

func (a *endTurnPhaseAction) NextPhase() types.PhaseType {
	return types.PhaseTypeDrawCard
}

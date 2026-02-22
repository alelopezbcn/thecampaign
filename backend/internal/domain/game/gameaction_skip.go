package game

import (
	"errors"

	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type skipPhaseAction struct {
	playerName string
	nextPhase  types.PhaseType
}

func NewSkipPhaseAction(playerName string) *skipPhaseAction {
	return &skipPhaseAction{playerName: playerName}
}

func (a *skipPhaseAction) PlayerName() string { return a.playerName }

func (a *skipPhaseAction) Validate(g *game) error {
	switch g.currentAction {
	case types.PhaseTypeAttack:
		a.nextPhase = types.PhaseTypeSpySteal
	case types.PhaseTypeSpySteal:
		a.nextPhase = types.PhaseTypeBuy
	case types.PhaseTypeBuy:
		a.nextPhase = types.PhaseTypeConstruct
	case types.PhaseTypeConstruct:
		a.nextPhase = types.PhaseTypeEndTurn
	default:
		return errors.New("cannot skip this phase")
	}

	return nil
}

func (a *skipPhaseAction) Execute(g *game) (*GameActionResult, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()

	result := &GameActionResult{Action: types.LastActionSkip}

	statusFn := func() gamestatus.GameStatus {
		return g.gameStatusProvider.Get(p, g)
	}

	return result, statusFn, nil
}

func (a *skipPhaseAction) NextPhase() types.PhaseType {
	return a.nextPhase
}

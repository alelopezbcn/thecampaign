package game

import (
	"errors"

	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type SkipPhaseAction struct {
	playerName string
	nextPhase  types.PhaseType
}

func NewSkipPhaseAction(playerName string) *SkipPhaseAction {
	return &SkipPhaseAction{playerName: playerName}
}

func (a *SkipPhaseAction) PlayerName() string { return a.playerName }

func (a *SkipPhaseAction) Validate(g *Game) error {
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

func (a *SkipPhaseAction) Execute(g *Game) (*GameActionResult, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()

	result := &GameActionResult{Action: types.LastActionSkip}

	statusFn := func() gamestatus.GameStatus {
		return g.GameStatusProvider.Get(p, g)
	}

	return result, statusFn, nil
}

func (a *SkipPhaseAction) NextPhase() types.PhaseType {
	return a.nextPhase
}

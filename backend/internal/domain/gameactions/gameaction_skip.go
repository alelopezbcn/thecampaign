package gameactions

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

func (a *skipPhaseAction) Validate(g Game) error {
	switch g.CurrentAction() {
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

func (a *skipPhaseAction) Execute(g Game) (*Result, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()

	result := &Result{Action: types.LastActionSkip}

	statusFn := func() gamestatus.GameStatus {
		return g.Status(p)
	}

	return result, statusFn, nil
}

func (a *skipPhaseAction) NextPhase() types.PhaseType {
	return a.nextPhase
}

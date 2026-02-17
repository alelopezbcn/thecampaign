package domain

import (
	"errors"

	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type SkipPhaseAction struct {
	playerName string
	nextPhase  types.ActionType
}

func NewSkipPhaseAction(playerName string) *SkipPhaseAction {
	return &SkipPhaseAction{playerName: playerName}
}

func (a *SkipPhaseAction) PlayerName() string { return a.playerName }

func (a *SkipPhaseAction) Validate(g *Game) error {
	switch g.currentAction {
	case types.ActionTypeAttack:
		a.nextPhase = types.ActionTypeSpySteal
	case types.ActionTypeSpySteal:
		a.nextPhase = types.ActionTypeBuy
	case types.ActionTypeBuy:
		a.nextPhase = types.ActionTypeConstruct
	case types.ActionTypeConstruct:
		a.nextPhase = types.ActionTypeEndTurn
	default:
		return errors.New("cannot skip this phase")
	}

	return nil
}

func (a *SkipPhaseAction) Execute(g *Game) (*ActionResult, func() GameStatus, error) {
	p := g.CurrentPlayer()

	result := &ActionResult{Action: types.LastActionSkip}

	statusFn := func() GameStatus {
		return g.GameStatusProvider.Get(p, g)
	}

	return result, statusFn, nil
}

func (a *SkipPhaseAction) NextPhase() types.ActionType {
	return a.nextPhase
}

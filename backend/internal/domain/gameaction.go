package domain

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type GameAction interface {
	PlayerName() string
	Validate(g *Game) error
	Execute(g *Game) (*GameActionResult, func() GameStatus, error)
	NextPhase() types.PhaseType
}

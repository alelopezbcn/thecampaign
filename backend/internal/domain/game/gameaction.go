package game

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type GameAction interface {
	PlayerName() string
	Validate(g *game) error
	Execute(g *game) (*GameActionResult, func() gamestatus.GameStatus, error)
	NextPhase() types.PhaseType
}

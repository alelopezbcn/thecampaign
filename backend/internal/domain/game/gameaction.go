package game

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type GameAction interface {
	PlayerName() string
	Validate(g Game) error
	Execute(g Game) (*GameActionResult, func() gamestatus.GameStatus, error)
	NextPhase() types.PhaseType
}

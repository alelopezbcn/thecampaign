package domain

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type GameAction interface {
	PlayerName() string
	Validate(g *Game) error
	Execute(g *Game) (*ActionResult, func() GameStatus, error)
	NextPhase() types.ActionType
}

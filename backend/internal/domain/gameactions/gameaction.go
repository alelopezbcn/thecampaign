package gameactions

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type Game interface {
	GamePlayers
	GameTurn
	GameTurnFlags
	GameStatusProvider
	GameCards
	GameHistory
	GameBoard
}

type GameAction interface {
	PlayerName() string
	Validate(g Game) error
	Execute(g Game) (*Result, func() gamestatus.GameStatus, error)
	NextPhase() types.PhaseType
}

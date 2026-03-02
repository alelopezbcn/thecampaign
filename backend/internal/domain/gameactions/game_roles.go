package gameactions

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// GamePlayers — player queries
type GamePlayers interface {
	CurrentPlayer() board.Player
	GetPlayer(name string) board.Player
	GetTargetPlayer(playerName string, targetPlayerName string) (board.Player, error)
	PlayerIndex(name string) int
	SameTeam(i, j int) bool
	Allies(playerIdx int) []board.Player
	Enemies(playerIdx int) []board.Player
}

// GameTurn — turn state and control
type GameTurn interface {
	CurrentAction() types.PhaseType
	TurnState() types.TurnState
	SwitchTurn()
}

// GameTurnFlags — mutable turn flags
type GameTurnFlags interface {
	SetHasMovedWarrior(value bool)
	SetHasTraded(value bool)
	SetCanMoveWarrior(value bool)
	SetCanTrade(value bool)
	SetHasForged(value bool)
	SetCanForge(value bool)
}

// GameStatusProvider — state snapshots
type GameStatusProvider interface {
	Status(viewer board.Player, newCards ...cards.Card) gamestatus.GameStatus
	StatusWithModal(viewer board.Player, modalCards []cards.Card) gamestatus.GameStatus
}

// GameCards — card operations
type GameCards interface {
	DrawCards(p board.Player, count int) (cards []cards.Card, err error)
	OnCardMovedToPile(card cards.Card)
}

// GameHistory — event log
type GameHistory interface {
	AddHistory(msg string, cat types.Category)
}

// GameBoard — board access
type GameBoard interface {
	Board() board.Board
}

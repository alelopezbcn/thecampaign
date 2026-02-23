package gameactions

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type Game interface {
	CurrentPlayer() board.Player
	CurrentAction() types.PhaseType
	TurnState() types.TurnState
	GetTargetPlayer(playerName string, targetPlayerName string) (board.Player, error)
	AddHistory(msg string, cat types.Category)
	Status(viewer board.Player, newCards ...cards.Card) gamestatus.GameStatus
	StatusWithModal(viewer board.Player, modalCards []cards.Card) gamestatus.GameStatus
	OnCardMovedToPile(card cards.Card)
	DrawCards(p board.Player, count int) (cards []cards.Card, err error)
	SwitchTurn()
	SetHasMovedWarrior(value bool)
	SetHasTraded(value bool)
	SetCanMoveWarrior(value bool)
	SetCanTrade(value bool)
	GetPlayer(name string) board.Player
	PlayerIndex(name string) int
	SameTeam(i, j int) bool
	Allies(playerIdx int) []board.Player
	Enemies(playerIdx int) []board.Player
	Board() board.Board
}

type GameAction interface {
	PlayerName() string
	Validate(g Game) error
	Execute(g Game) (*Result, func() gamestatus.GameStatus, error)
	NextPhase() types.PhaseType
}

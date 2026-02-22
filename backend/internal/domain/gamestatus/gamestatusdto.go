package gamestatus

import (
	"time"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type GameStatusDTO struct {
	Viewer                 board.Player
	NewCards               []cards.Card
	ModalCards             []cards.Card
	PlayerIndex            int
	PlayersNames           []string
	Players                []board.Player
	NextTurnPlayer         string
	TurnPlayer             string
	CurrentAction          types.PhaseType
	LastAction             types.LastActionType
	GameMode               string
	IsEliminated           bool
	IsDisconnected         bool
	CanTrade               bool
	CemeteryCount          int
	CemeteryLastDead       cards.Warrior
	DiscardPileCount       int
	DiscardPileLastCard    cards.Card
	DeckCount              int
	GameStartedAt          time.Time
	TurnStartedAt          time.Time
	History                []types.HistoryLine
	LastMovedWarriorID     string
	LastAttackWeaponID     string
	LastAttackTargetID     string
	LastAttackTargetPlayer string
	StolenFrom             string
	StolenCard             cards.Card
	SpyTarget              types.SpyTarget
	SpyTargetPlayer        string
	CurrentPlayerName      string
	IsGameOver             bool
	Winner                 string
	IsPlayerWinner         bool
	SameTeamFn             func(i, j int) bool
	EliminatedPlayers      map[int]bool
	DisconnectedPlayers    map[int]bool
	CanMoveWarrior         bool
	EnemiesFn              func(playerIdx int) []board.Player
	AlliesFn               func(playerIdx int) []board.Player
}

package gamestatus

import (
	"time"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type GameStatusDTO struct {
	Viewer                   ViewerInput
	PlayersNames             []string
	Opponents                []OpponentInput
	EnemyFields              []FieldInput
	AllyFields               []FieldInput
	AnyEnemyCastleAttackable  bool
	AnyEnemyHasCards          bool
	AnyEnemyHasWeakWarriors   bool
	AllyHasCastleConstructed  bool
	NewCards                 []cards.Card
	ModalCards               []cards.Card
	NextTurnPlayer           string
	TurnPlayer               string
	CurrentAction            types.PhaseType
	LastAction               types.LastActionType
	GameMode                 string
	IsEliminated             bool
	IsDisconnected           bool
	CanTrade                 bool
	CemeteryCount            int
	CemeteryLastDead         cards.Warrior
	DiscardPileCount         int
	DiscardPileLastCard      cards.Card
	DeckCount                int
	GameStartedAt            time.Time
	TurnStartedAt            time.Time
	History                  []types.HistoryLine
	LastMovedWarriorID       string
	LastAttackWeaponID       string
	LastAttackTargetID       string
	LastAttackTargetPlayer   string
	StolenFrom               string
	StolenCard               cards.Card
	SabotagedFrom            string
	SabotagedCard            cards.Card
	SpyTarget                types.SpyTarget
	SpyTargetPlayer          string
	CurrentPlayerName        string
	IsGameOver               bool
	Winner                   string
	IsPlayerWinner           bool
	CanMoveWarrior           bool
	AmbushEffect             types.AmbushEffect
	AmbushAttackerName       string
	DeserterFromPlayer       string
	DeserterWarrior          cards.Warrior
}

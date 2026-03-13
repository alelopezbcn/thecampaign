// Package gamestatus contains the input struct for building the game status.
package gamestatus

import (
	"time"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type BuildInput struct {
	Viewer                    ViewerInput
	PlayersNames              []string
	Opponents                 []OpponentInput
	EnemyFields               []FieldInput
	AllyFields                []FieldInput
	AnyEnemyCastleAttackable  bool
	AnyEnemyHasCards          bool
	AnyEnemyHasWeakWarriors   bool
	AllyHasCastleConstructed  bool
	NewCards                  []cards.Card
	ModalCards                []cards.Card
	NextTurnPlayer            string
	TurnPlayer                string
	CurrentAction             types.PhaseType
	LastAction                types.LastActionType
	GameMode                  string
	IsEliminated              bool
	IsDisconnected            bool
	CanTrade                  bool
	CanForge                  bool
	CemeteryCount             int
	CemeteryLastDead          cards.Warrior
	DiscardPileCount          int
	DiscardPileLastCard       cards.Card
	DeckCount                 int
	GameStartedAt             time.Time
	TurnStartedAt             time.Time
	History                   []types.HistoryLine
	LastMovedWarriorID        string
	LastAttackWeaponID        string
	LastAttackTargetID        string
	LastAttackTargetPlayer    string
	StolenFrom                string
	StolenCard                cards.Card
	SabotagedFrom             string
	SabotagedCard             cards.Card
	SpyTarget                 types.SpyTarget
	SpyTargetPlayer           string
	CurrentPlayerName         string
	IsGameOver                bool
	Winner                    string
	IsPlayerWinner            bool
	CanMoveWarrior            bool
	AmbushEffect              types.AmbushEffect
	AmbushAttackerName        string
	AmbushAttackerWarriorType string
	AmbushAttackerHPBefore    int
	AmbushAttackerHPAfter     int
	AmbushAttackerDied        bool
	AmbushTargetWarriorType   string
	AmbushTargetHPBefore      int
	AmbushTargetHPAfter       int
	AmbushWeaponType          string
	AmbushDamageAmount        int
	ChampionsBountyEarner     string
	ChampionsBountyCards      int
	TraitorFromPlayer         string
	TraitorWarrior            cards.Warrior
	CurrentEvent              types.ActiveEvent
	ResurrectionWarrior       cards.Warrior
	ResurrectionTargetPlayer  string
	ResurrectionPlayerName    string
	AmbushPlacedOn            string // player whose field received the ambush (place_ambush action)
	CatapultAttacker          string
	CatapultTarget            string
	CatapultGoldStolen        int
	CatapultBlocked           bool
	PlayerStats               []PlayerStatInput
}

// PlayerStatInput holds end-of-game statistics for a single player.
type PlayerStatInput struct {
	Name        string
	Kills       int
	Damage      int
	CastleValue int
	IsWinner    bool
	IsMVP       bool
}

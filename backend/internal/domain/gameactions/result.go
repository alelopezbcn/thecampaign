package gameactions

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type Result struct {
	Action         types.LastActionType
	MovedWarriorID string               // "" = no warrior moved
	Spy            *types.SpyInfo       // nil = no spy action
	Attack         *AttackDetails       // nil = no attack
	Steal          *StealDetails        // nil = no steal
	Sabotage       *SabotageDetails     // nil = no sabotage
	Treason        *TreasonDetails      // nil = no treason
	Resurrection   *ResurrectionDetails // nil = no resurrection
	PlaceAmbush    *PlaceAmbushDetails  // nil = not a place ambush action
}

type PlaceAmbushDetails struct {
	TargetPlayer string // player whose field received the ambush
}

type AttackDetails struct {
	WeaponID              string
	TargetID              string
	TargetPlayer          string
	AmbushEffect          types.AmbushEffect // zero value = no ambush triggered
	AmbushAttackerName    string
	ChampionsBountyEarner string // "" = bounty not triggered
	ChampionsBountyCards  int
	ChampionsBountyHeal   int // > 0 = hit-heal triggered (HP healed to a random warrior)
	KillsGranted          int // 1 if a kill was earned this action, 0 otherwise
	DamageDealt           int // HP damage actually inflicted this action
}

type StealDetails struct {
	From string
	Card cards.Card
}

type SabotageDetails struct {
	From string
	Card cards.Card
}

type TreasonDetails struct {
	FromPlayer string
	Warrior    cards.Warrior
}

type ResurrectionDetails struct {
	Warrior      cards.Warrior
	TargetPlayer string // player whose field received the warrior
	PlayerName   string // player who played the card
}

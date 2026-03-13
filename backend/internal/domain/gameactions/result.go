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
	Catapult       *CatapultDetails     // nil = no catapult action
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

	// Ambush detail fields — populated only when AmbushEffect != 0.
	AmbushAttackerWarriorType string // e.g. "Knight"
	AmbushAttackerHPBefore    int
	AmbushAttackerHPAfter     int
	AmbushAttackerDied        bool
	AmbushTargetWarriorType   string // e.g. "Archer"
	AmbushTargetHPBefore      int
	AmbushTargetHPAfter       int
	AmbushWeaponType          string // e.g. "Sword"
	AmbushDamageAmount        int    // effective damage/heal (after event modifier + multiplier)
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

type CatapultDetails struct {
	AttackerName string
	TargetPlayer string
	GoldStolen   int  // 0 when Blocked is true
	Blocked      bool // fortress absorbed the hit
}

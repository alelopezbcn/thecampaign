package game

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type GameActionResult struct {
	Action             types.LastActionType
	MovedWarriorID     string
	StolenFrom         string
	StolenCard         cards.Card
	Spy                types.SpyInfo
	AttackWeaponID     string
	AttackTargetID     string
	AttackTargetPlayer string
}

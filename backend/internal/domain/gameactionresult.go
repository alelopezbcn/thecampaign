package domain

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type GameActionResult struct {
	Action             types.LastActionType
	MovedWarriorID     string
	StolenFrom         string
	StolenCard         ports.Card
	Spy                types.SpyInfo
	AttackWeaponID     string
	AttackTargetID     string
	AttackTargetPlayer string
}

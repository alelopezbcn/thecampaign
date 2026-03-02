package types

import "time"

type TurnState struct {
	CanMoveWarrior  bool
	HasMovedWarrior bool
	CanTrade        bool
	HasTraded       bool
	CanForge        bool
	HasForged       bool
	StartedAt       time.Time
}

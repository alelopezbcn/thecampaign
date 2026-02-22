package types

import "time"

type TurnState struct {
	CanMoveWarrior  bool
	HasMovedWarrior bool
	CanTrade        bool
	HasTraded       bool
	StartedAt       time.Time
}

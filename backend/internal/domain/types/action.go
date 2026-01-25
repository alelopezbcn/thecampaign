package types

type ActionType string

var (
	ActionTypeDrawCard  ActionType = "draw"
	ActionTypeAttack    ActionType = "attack"
	ActionTypeSpySteal  ActionType = "spy/steal"
	ActionTypeBuy       ActionType = "buy"
	ActionTypeConstruct ActionType = "construct"
	ActionTypeEndTurn   ActionType = "endturn"
)

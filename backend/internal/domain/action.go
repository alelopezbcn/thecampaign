package domain

type Action struct {
	PlayerID string
	Type     ActionType
}

type ActionType string

var (
	ActionTypeTakeCard  = "take_card"
	ActionTypeAttack    = "attack"
	ActionTypeSteal     = "steal"
	ActionTypeSpy       = "spy"
	ActionTypeBuy       = "buy"
	ActionTypeConstruct = "construct"
	ActionTypeSwap      = "swap"
)

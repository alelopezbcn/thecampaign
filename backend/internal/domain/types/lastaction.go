package types

type LastActionType string

const (
	LastActionDraw         LastActionType = "draw"
	LastActionAttack       LastActionType = "attack"
	LastActionMoveWarrior  LastActionType = "move_warrior"
	LastActionTrade        LastActionType = "trade"
	LastActionSpecialPower LastActionType = "special_power"
	LastActionCatapult     LastActionType = "catapult"
	LastActionSpy          LastActionType = "spy"
	LastActionSteal        LastActionType = "steal"
	LastActionBuy          LastActionType = "buy"
	LastActionConstruct    LastActionType = "construct"
	LastActionSkip         LastActionType = "skip"
	LastActionEndTurn      LastActionType = "end_turn"
)

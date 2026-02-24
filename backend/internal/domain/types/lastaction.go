package types

type LastActionType string

const (
	LastActionDraw               LastActionType = "draw"
	LastActionAttack             LastActionType = "attack"
	LastActionMoveWarrior        LastActionType = "move_warrior"
	LastActionTrade              LastActionType = "trade"
	LastActionSpecialPower       LastActionType = "special_power"
	LastActionHarpoon            LastActionType = "harpoon"
	LastActionBloodRain          LastActionType = "blood_rain"
	LastActionCatapult           LastActionType = "catapult"
	LastActionCatapultBlocked    LastActionType = "catapult_blocked"
	LastActionSpy                LastActionType = "spy"
	LastActionSteal              LastActionType = "steal"
	LastActionBuy                LastActionType = "buy"
	LastActionConstruct          LastActionType = "construct"
	LastActionFortress           LastActionType = "fortress"
	LastActionSkip               LastActionType = "skip"
	LastActionEndTurn            LastActionType = "end_turn"
)

package domain

type GameState struct {
	order int
	name  string
}

var (
	StateSettingInitialWarriors = GameState{0, "SETTING_INITIAL_WARRIORS"}
	StateWaitingDraw            = GameState{1, "WAITING_DRAW"}
	StateMainPhase              = GameState{2, "MAIN_PHASE"}
	StateEndTurn                = GameState{3, "END_TURN"}
	StateGameEnded              = GameState{4, "GAME_ENDED"}
)

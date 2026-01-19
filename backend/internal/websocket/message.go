package websocket

// MessageType defines the type of WebSocket message
type MessageType string

const (
	// Client to Server messages
	MsgJoinGame         MessageType = "join_game"
	MsgSetInitialWarriors MessageType = "set_initial_warriors"
	MsgDrawCard         MessageType = "draw_card"
	MsgAttack           MessageType = "attack"
	MsgSpecialPower     MessageType = "special_power"
	MsgMoveWarrior      MessageType = "move_warrior"
	MsgTrade            MessageType = "trade"
	MsgBuy              MessageType = "buy"
	MsgConstruct        MessageType = "construct"
	MsgSpy              MessageType = "spy"
	MsgSteal            MessageType = "steal"
	MsgCatapult         MessageType = "catapult"
	MsgEndTurn          MessageType = "end_turn"

	// Server to Client messages
	MsgGameState        MessageType = "game_state"
	MsgError            MessageType = "error"
	MsgGameStarted      MessageType = "game_started"
	MsgWaitingForPlayer MessageType = "waiting_for_player"
	MsgPlayerJoined     MessageType = "player_joined"
	MsgGameEnded        MessageType = "game_ended"
)

// Message is the base WebSocket message structure
type Message struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

// JoinGamePayload is sent when a player wants to join a game
type JoinGamePayload struct {
	GameID     string `json:"game_id"`
	PlayerName string `json:"player_name"`
}

// SetInitialWarriorsPayload for setting initial warriors
type SetInitialWarriorsPayload struct {
	WarriorIDs []string `json:"warrior_ids"`
}

// AttackPayload for attack action
type AttackPayload struct {
	WarriorID string `json:"warrior_id"`
	TargetID  string `json:"target_id"`
	WeaponID  string `json:"weapon_id"`
}

// SpecialPowerPayload for special power action
type SpecialPowerPayload struct {
	UserID   string `json:"user_id"`
	TargetID string `json:"target_id"`
	WeaponID string `json:"weapon_id"`
}

// MoveWarriorPayload for moving a warrior to field
type MoveWarriorPayload struct {
	WarriorID string `json:"warrior_id"`
}

// TradePayload for trading cards
type TradePayload struct {
	CardIDs []string `json:"card_ids"`
}

// BuyPayload for buying cards
type BuyPayload struct {
	CardID string `json:"card_id"`
}

// ConstructPayload for constructing castle
type ConstructPayload struct {
	CardID string `json:"card_id"`
}

// SpyPayload for spy action
type SpyPayload struct {
	Option int `json:"option"`
}

// StealPayload for steal action
type StealPayload struct {
	CardPosition int `json:"card_position"`
}

// CatapultPayload for catapult action
type CatapultPayload struct {
	CardPosition int `json:"card_position"`
}

// GameStatePayload is sent to update clients with game state
type GameStatePayload struct {
	GameStatus    GameStatusDTO `json:"game_status"`
	IsYourTurn    bool          `json:"is_your_turn"`
	GameEnded     bool          `json:"game_ended"`
	History       []string      `json:"history,omitempty"`
	NewlyDrawnCard string       `json:"newly_drawn_card,omitempty"` // ID of card just drawn
}

// GameStatusDTO is the JSON-friendly game status
type GameStatusDTO struct {
	CurrentPlayer              string              `json:"current_player"`
	WarriorsInHandIDs          []string            `json:"warriors_in_hand_ids"`
	UsableWeaponIDs            []string            `json:"usable_weapon_ids"`
	SpyID                      string              `json:"spy_id"`
	ThiefID                    string              `json:"thief_id"`
	ResourceIDs                []string            `json:"resource_ids"`
	SpecialPowerStatus         SpecialPowerStatusDTO `json:"special_power_status"`
	ConstructionIDs            []string            `json:"construction_ids"`
	CatapultID                 string              `json:"catapult_id"`
	CurrentPlayerHand          []CardDTO           `json:"current_player_hand"`
	CurrentPlayerField         []CardDTO           `json:"current_player_field"`
	CurrentPlayerCastle        CastleDTO           `json:"current_player_castle"`
	EnemyField                 []CardDTO           `json:"enemy_field"`
	EnemyCastle                CastleDTO           `json:"enemy_castle"`
	CardsInEnemyHand           int                 `json:"cards_in_enemy_hand"`
	ResourceCardsInEnemyCastle int                 `json:"resource_cards_in_enemy_castle"`
}

// SpecialPowerStatusDTO for special powers
type SpecialPowerStatusDTO struct {
	SpecialPowerIDs   []string `json:"special_power_ids"`
	CanHealIDs        []string `json:"can_heal_ids"`
	CanInstantKillIDs []string `json:"can_instant_kill_ids"`
	CanProtectIDs     []string `json:"can_protect_ids"`
}

// ErrorPayload for error messages
type ErrorPayload struct {
	Message string `json:"message"`
}

// GameStartedPayload when game starts
type GameStartedPayload struct {
	GameID   string   `json:"game_id"`
	Players  []string `json:"players"`
	YourName string   `json:"your_name"`
}

// PlayerJoinedPayload when a player joins
type PlayerJoinedPayload struct {
	PlayerName string `json:"player_name"`
}

package websocket

// MessageType defines the type of WebSocket message
type MessageType string

const (
	// Client to Server messages
	MsgJoinGame           MessageType = "join_game"
	MsgSetInitialWarriors MessageType = "set_initial_warriors"
	MsgDrawCard           MessageType = "draw_card"
	MsgAttack             MessageType = "attack"
	MsgSpecialPower       MessageType = "special_power"
	MsgMoveWarrior        MessageType = "move_warrior"
	MsgTrade              MessageType = "trade"
	MsgBuy                MessageType = "buy"
	MsgConstruct          MessageType = "construct"
	MsgSpy                MessageType = "spy"
	MsgSteal              MessageType = "steal"
	MsgCatapult           MessageType = "catapult"
	MsgEndTurn            MessageType = "end_turn"
	MsgSkipPhase          MessageType = "skip_phase"

	// Server to Client messages
	MsgGameState        MessageType = "game_state"
	MsgError            MessageType = "error"
	MsgGameStarted      MessageType = "game_started"
	MsgWaitingForPlayer MessageType = "waiting_for_player"
	MsgPlayerJoined     MessageType = "player_joined"
	MsgGameEnded        MessageType = "game_ended"
	MsgInitialWarriors  MessageType = "initial_warriors"
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
	TargetID string `json:"target_id"`
	WeaponID string `json:"weapon_id"`
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
	GameStatus GameStatusDTO `json:"game_status"`
	IsYourTurn bool          `json:"is_your_turn"`
	History    []string      `json:"history,omitempty"`
}

// GameStatusDTO is the JSON-friendly game status
type GameStatusDTO struct {
	CurrentPlayer  string   `json:"current_player"`
	CurrentAction  string   `json:"current_action"`
	NewCards       []string `json:"new_cards"`
	CanMoveWarrior bool     `json:"can_move_warrior"`
	CanTrade       bool     `json:"can_trade"`

	CurrentPlayerHand   []HandCardDTO  `json:"current_player_hand"`
	CurrentPlayerField  []FieldCardDTO `json:"current_player_field"`
	CurrentPlayerCastle CastleDTO      `json:"current_player_castle"`
	EnemyField          []FieldCardDTO `json:"enemy_field"`
	EnemyCastle         CastleDTO      `json:"enemy_castle"`
	CardsInEnemyHand    int            `json:"cards_in_enemy_hand"`
	Cemetery            CemeteryDTO    `json:"cemetery"`
	DiscardPile         DiscardPileDTO `json:"discard_pile"`
	CardsInDeck         int            `json:"cards_in_deck"`
	ModalCards          []CardDTO      `json:"modal_cards,omitempty"`
	History             []string       `json:"history"`
	GameOverMsg         string         `json:"game_over_msg,omitempty"`
	ErrorMsg            string         `json:"error_msg,omitempty"`
}

// HandCardDTO represents a card in the player's hand
type HandCardDTO struct {
	CardDTO
	CanBeUsedOnIDs []string       `json:"use_on"`
	CanBeUsed      bool           `json:"can_be_used"`
	DmgMultiplier  map[string]int `json:"dmg_mult,omitempty"`
}

// FieldCardDTO represents a card on the battlefield
type FieldCardDTO struct {
	CardDTO
	AttackedBy  []CardDTO `json:"attacked_by,omitempty"`
	ProtectedBy *CardDTO  `json:"protected_by,omitempty"`
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

// InitialWarriorsPayload sent to players to choose their initial warriors
type InitialWarriorsPayload struct {
	Warriors   []CardDTO `json:"warriors"`
	IsYourTurn bool      `json:"is_your_turn"`
}

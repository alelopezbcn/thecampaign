package websocket

// MessageType defines the type of WebSocket message
type MessageType string

const (
	// Client to Server messages
	MsgJoinGame     MessageType = "join_game"
	MsgDrawCard     MessageType = "draw_card"
	MsgAttack       MessageType = "attack"
	MsgSpecialPower MessageType = "special_power"
	MsgMoveWarrior  MessageType = "move_warrior"
	MsgTrade        MessageType = "trade"
	MsgBuy          MessageType = "buy"
	MsgConstruct    MessageType = "construct"
	MsgSpy          MessageType = "spy"
	MsgSteal        MessageType = "steal"
	MsgCatapult     MessageType = "catapult"
	MsgEndTurn      MessageType = "end_turn"
	MsgSkipPhase    MessageType = "skip_phase"
	MsgSwapTeam     MessageType = "swap_team"
	MsgStartGame    MessageType = "start_game"

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
	GameMode   string `json:"game_mode"`
}

// AttackPayload for attack action
type AttackPayload struct {
	TargetPlayer string `json:"target_player"`
	TargetID     string `json:"target_id"`
	WeaponID     string `json:"weapon_id"`
}

// SpecialPowerPayload for special power action
type SpecialPowerPayload struct {
	UserID   string `json:"user_id"`
	TargetID string `json:"target_id"`
	WeaponID string `json:"weapon_id"`
}

// MoveWarriorPayload for moving a warrior to field
type MoveWarriorPayload struct {
	WarriorID    string `json:"warrior_id"`
	TargetPlayer string `json:"target_player,omitempty"`
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
	CardID       string `json:"card_id"`
	TargetPlayer string `json:"target_player,omitempty"`
}

// SpyPayload for spy action
type SpyPayload struct {
	TargetPlayer string `json:"target_player"`
	Option       int    `json:"option"`
}

// StealPayload for steal action
type StealPayload struct {
	TargetPlayer string `json:"target_player"`
	CardPosition int    `json:"card_position"`
}

// CatapultPayload for catapult action
type CatapultPayload struct {
	TargetPlayer string `json:"target_player"`
	CardPosition int    `json:"card_position"`
}

// GameStatePayload is sent to update clients with game state
type GameStatePayload struct {
	GameStatus GameStatusDTO `json:"game_status"`
	IsYourTurn bool          `json:"is_your_turn"`
}

// GameStatusDTO is the JSON-friendly game status
type GameStatusDTO struct {
	CurrentPlayer  string   `json:"current_player"`
	TurnPlayer     string   `json:"turn_player"`
	CurrentAction  string   `json:"current_action"`
	LastAction     string   `json:"last_action,omitempty"`
	NewCards       []string `json:"new_cards"`
	CanMoveWarrior bool     `json:"can_move_warrior"`
	CanTrade       bool     `json:"can_trade"`

	CurrentPlayerHand   []HandCardDTO       `json:"current_player_hand"`
	CurrentPlayerField  []FieldCardDTO      `json:"current_player_field"`
	CurrentPlayerCastle CastleDTO           `json:"current_player_castle"`
	IsEliminated        bool                `json:"is_eliminated"`
	IsDisconnected      bool                `json:"is_disconnected"`
	Opponents           []OpponentStatusDTO `json:"opponents"`
	GameMode            string              `json:"game_mode"`
	Cemetery            CemeteryDTO         `json:"cemetery"`
	DiscardPile         DiscardPileDTO      `json:"discard_pile"`
	CardsInDeck         int                 `json:"cards_in_deck"`
	ModalCards          []CardDTO           `json:"modal_cards,omitempty"`
	LastMovedWarriorID    string              `json:"last_moved_warrior_id,omitempty"`
	LastAttackWeaponID    string              `json:"last_attack_weapon_id,omitempty"`
	LastAttackTargetID    string              `json:"last_attack_target_id,omitempty"`
	LastAttackTargetPlayer string             `json:"last_attack_target_player,omitempty"`
	StolenFromYouCard   []CardDTO           `json:"stolen_from_you_card,omitempty"`
	SpyNotification     string              `json:"spy_notification,omitempty"`
	History             []HistoryLineDTO    `json:"history"`
	PlayersOrder        []string            `json:"players_order"`
	NextTurnPlayer      string              `json:"next_turn_player,omitempty"`
	GameOverMsg         string              `json:"game_over_msg,omitempty"`
	IsWinner            bool                `json:"is_winner"`
	GameStartedAt       string              `json:"game_started_at"`
	TurnStartedAt       string              `json:"turn_started_at"`
	TurnTimeLimitSecs   int                 `json:"turn_time_limit_secs"`
}

type OpponentStatusDTO struct {
	PlayerName   string         `json:"player_name"`
	Field        []FieldCardDTO `json:"field"`
	Castle       CastleDTO      `json:"castle"`
	CardsInHand  int            `json:"cards_in_hand"`
	IsAlly         bool           `json:"is_ally"`
	IsEliminated   bool           `json:"is_eliminated"`
	IsDisconnected bool           `json:"is_disconnected"`
}

// HandCardDTO represents a card in the player's hand
type HandCardDTO struct {
	CardDTO
	CanBeUsedOnIDs []string       `json:"use_on"`
	CanBeUsed      bool           `json:"can_be_used"`
	DmgMultiplier  map[string]int `json:"dmg_mult,omitempty"`
	CanBeTraded    bool           `json:"can_be_traded"`
}

// FieldCardDTO represents a card on the battlefield
type FieldCardDTO struct {
	CardDTO
	AttackedBy  []CardDTO `json:"attacked_by,omitempty"`
	ProtectedBy *CardDTO  `json:"protected_by,omitempty"`
}

// HistoryLineDTO represents a line in the game history with color for UI display
type HistoryLineDTO struct {
	Msg   string `json:"msg"`
	Color string `json:"color"`
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
	GameID     string         `json:"game_id"`
	GameMode   string         `json:"game_mode"`
	MaxPlayers int            `json:"max_players"`
	PlayerName string         `json:"player_name"`
	Players    []string       `json:"players"`
	Teams      map[string]int `json:"teams,omitempty"`
}

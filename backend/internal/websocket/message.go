package websocket

import "github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"

// MessageType defines the type of WebSocket message
type MessageType string

const (
	// Client to Server messages
	MsgJoinGame     MessageType = "join_game"
	MsgDrawCard     MessageType = "draw_card"
	MsgAttack       MessageType = "attack"
	MsgSpecialPower MessageType = "special_power"
	MsgHarpoon      MessageType = "harpoon"
	MsgBloodRain    MessageType = "blood_rain"
	MsgMoveWarrior  MessageType = "move_warrior"
	MsgTrade        MessageType = "trade"
	MsgBuy          MessageType = "buy"
	MsgForge        MessageType = "forge"
	MsgBuyMercenary MessageType = "buy_mercenary"
	MsgConstruct    MessageType = "construct"
	MsgSpy          MessageType = "spy"
	MsgSteal        MessageType = "steal"
	MsgDesertion    MessageType = "desertion"
	MsgCatapult     MessageType = "catapult"
	MsgFortress     MessageType = "fortress"
	MsgResurrection MessageType = "resurrection"
	MsgSabotage     MessageType = "sabotage"
	MsgPlaceAmbush  MessageType = "place_ambush"
	MsgEndTurn      MessageType = "end_turn"
	MsgSkipPhase    MessageType = "skip_phase"
	MsgSwapTeam     MessageType = "swap_team"
	MsgStartGame    MessageType = "start_game"
	MsgRestartGame  MessageType = "restart_game"

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

const defaultCastleGoal = 25

// GameConfig holds the card counts and victory condition chosen by the room creator.
type GameConfig struct {
	Warriors          int `json:"warriors"`
	Dragons           int `json:"dragons"`
	Harpoons          int `json:"harpoons"`
	SpecialPowers     int `json:"special_powers"`
	Spies             int `json:"spies"`
	Thieves           int `json:"thieves"`
	Sabotages         int `json:"sabotages"`
	Catapults         int `json:"catapults"`
	Fortresses        int `json:"fortresses"`
	Ambushes          int `json:"ambushes"`
	BloodRains        int `json:"blood_rains"`
	Resurrections     int `json:"resurrections"`
	Desertions        int `json:"desertions"`
	ConstructionCards int `json:"construction_cards"` // copies per value 1-9 for gold/sword/arrow/poison
	CastleGoal        int `json:"castle_goal"`
}

// defaultGameConfig returns the default configuration when none is provided.
func defaultGameConfig() GameConfig {
	return GameConfig{
		Warriors:          5,
		Dragons:           1,
		Harpoons:          1,
		SpecialPowers:     4,
		Spies:             1,
		Thieves:           1,
		Sabotages:         1,
		Catapults:         1,
		Fortresses:        1,
		Ambushes:          1,
		BloodRains:        2,
		Resurrections:     1,
		Desertions:        1,
		ConstructionCards: 1,
		CastleGoal:        defaultCastleGoal,
	}
}

// JoinGamePayload is sent when a player wants to join a game
type JoinGamePayload struct {
	GameID     string     `json:"game_id"`
	PlayerName string     `json:"player_name"`
	GameMode   string     `json:"game_mode"`
	GameConfig GameConfig `json:"game_config,omitempty"`
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

// WeaponPayload is the shared payload for weapon card actions.
// TargetID is optional — AoE weapons (e.g. blood rain) leave it empty.
type WeaponPayload struct {
	TargetPlayer string `json:"target_player"`
	TargetID     string `json:"target_id,omitempty"`
	WeaponID     string `json:"weapon_id"`
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

// ForgePayload for forging two weapons into one
type ForgePayload struct {
	CardID1 string `json:"card_id_1"`
	CardID2 string `json:"card_id_2"`
}

// BuyMercenaryPayload for hiring a mercenary directly
type BuyMercenaryPayload struct {
	CardID string `json:"card_id"`
}

// ConstructPayload for constructing castle
type ConstructPayload struct {
	CardID       string `json:"card_id"`
	TargetPlayer string `json:"target_player,omitempty"`
}

// SpyPayload for spy action
type SpyPayload struct {
	CardID       string `json:"card_id"`
	TargetPlayer string `json:"target_player"`
	Option       int    `json:"option"`
}

// StealPayload for steal action
type StealPayload struct {
	CardID       string `json:"card_id"`
	TargetPlayer string `json:"target_player"`
	CardPosition int    `json:"card_position"`
}

// DesertionPayload for desertion action
type DesertionPayload struct {
	CardID       string `json:"card_id"`
	TargetPlayer string `json:"target_player"`
	WarriorID    string `json:"warrior_id"`
}

// CatapultPayload for catapult action
type CatapultPayload struct {
	CardID       string `json:"card_id"`
	TargetPlayer string `json:"target_player"`
	CardPosition int    `json:"card_position"`
}

// FortressPayload for placing a fortress on a castle
type FortressPayload struct {
	CardID       string `json:"card_id"`
	TargetPlayer string `json:"target_player,omitempty"`
}

// ResurrectionPayload for using a resurrection card
type ResurrectionPayload struct {
	CardID       string `json:"card_id"`
	TargetPlayer string `json:"target_player,omitempty"`
}

// SabotagePayload for sabotage action
type SabotagePayload struct {
	CardID       string `json:"card_id"`
	TargetPlayer string `json:"target_player"`
}

// PlaceAmbushPayload for placing an ambush card in the field
type PlaceAmbushPayload struct {
	CardID string `json:"card_id"`
}

// GameStatePayload is sent to update clients with game state
type GameStatePayload struct {
	GameStatus gamestatus.GameStatus `json:"game_status"`
	IsYourTurn bool                  `json:"is_your_turn"`
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

package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/alelopezbcn/thecampaign/internal/domain"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// ClientMessage represents a message from a client
type ClientMessage struct {
	client  *Client
	message *Message
}

// GameRoom represents a game room with two players
type GameRoom struct {
	ID         string
	GameMode   string
	MaxPlayers int
	Game       *domain.Game
	Players    map[string]*Client // playerName -> client
	mutex      sync.RWMutex
}

// Hub maintains active clients and game rooms
type Hub struct {
	clients       map[*Client]bool
	gameRooms     map[string]*GameRoom // gameID -> GameRoom
	register      chan *Client
	unregister    chan *Client
	handleMessage chan *ClientMessage
	mutex         sync.RWMutex
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients:       make(map[*Client]bool),
		gameRooms:     make(map[string]*GameRoom),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		handleMessage: make(chan *ClientMessage),
	}
}

// Register registers a client with the hub
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			log.Printf("Client registered, total clients: %d", len(h.clients))

		case client := <-h.unregister:
			var disconnectedGameID string
			var disconnectedPlayerName string

			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client unregistered, total clients: %d", len(h.clients))

				// Remove client from game room
				if client.GameID != "" {
					disconnectedGameID = client.GameID
					disconnectedPlayerName = client.PlayerName
					if room, exists := h.gameRooms[client.GameID]; exists {
						room.mutex.Lock()
						delete(room.Players, client.PlayerName)
						if len(room.Players) == 0 {
							delete(h.gameRooms, client.GameID)
							log.Printf("Game room %s removed (empty)", client.GameID)
						}
						room.mutex.Unlock()
					}
				}
			}
			h.mutex.Unlock()

			// Notify other players AFTER releasing the lock to avoid deadlock
			if disconnectedGameID != "" {
				h.broadcastToGame(disconnectedGameID, MsgError, ErrorPayload{
					Message: disconnectedPlayerName + " disconnected",
				})
			}

		case clientMsg := <-h.handleMessage:
			h.processMessage(clientMsg.client, clientMsg.message)
		}
	}
}

// processMessage processes incoming messages from clients
func (h *Hub) processMessage(client *Client, msg *Message) {
	log.Printf("Processing message type: %s from player: %s", msg.Type, client.PlayerName)

	switch msg.Type {
	case MsgJoinGame:
		h.handleJoinGame(client, msg.Payload)
	case MsgSetInitialWarriors:
		h.handleSetInitialWarriors(client, msg.Payload)
	case MsgAttack:
		h.handleAttack(client, msg.Payload)
	case MsgSpecialPower:
		h.handleSpecialPower(client, msg.Payload)
	case MsgMoveWarrior:
		h.handleMoveWarrior(client, msg.Payload)
	case MsgTrade:
		h.handleTrade(client, msg.Payload)
	case MsgBuy:
		h.handleBuy(client, msg.Payload)
	case MsgConstruct:
		h.handleConstruct(client, msg.Payload)
	case MsgSpy:
		h.handleSpy(client, msg.Payload)
	case MsgSteal:
		h.handleSteal(client, msg.Payload)
	case MsgCatapult:
		h.handleCatapult(client, msg.Payload)
	case MsgEndTurn:
		h.handleEndTurn(client)
	case MsgSkipPhase:
		h.handleSkipPhase(client)
	default:
		client.SendError("Unknown message type")
	}
}

// handleJoinGame handles a player joining a game
func (h *Hub) handleJoinGame(client *Client, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		client.SendError("Invalid payload")
		return
	}

	var joinPayload JoinGamePayload
	if err := json.Unmarshal(data, &joinPayload); err != nil {
		client.SendError("Invalid join game payload")
		return
	}

	h.mutex.Lock()
	gameID := joinPayload.GameID
	playerName := joinPayload.PlayerName

	client.PlayerName = playerName
	client.GameID = gameID

	room, exists := h.gameRooms[gameID]
	if !exists {
		// Create new game room
		room = &GameRoom{
			ID:      gameID,
			Players: make(map[string]*Client),
		}
		h.gameRooms[gameID] = room
		log.Printf("Created new game room: %s", gameID)
	}
	h.mutex.Unlock()

	room.mutex.Lock()

	// Check if this is a reconnection (player name already exists in this game)
	if oldClient, exists := room.Players[playerName]; exists {
		// Replace old client with new one
		room.Players[playerName] = client
		gameInProgress := room.Game != nil
		room.mutex.Unlock()

		// Close old client connection (outside lock)
		h.mutex.Lock()
		if _, ok := h.clients[oldClient]; ok {
			delete(h.clients, oldClient)
			close(oldClient.send)
		}
		h.mutex.Unlock()

		log.Printf("Player %s reconnected to game %s", playerName, gameID)

		if gameInProgress {
			h.sendReconnectState(gameID, playerName)
		} else {
			client.SendMessage(MsgWaitingForPlayer, nil)
		}
		return
	}

	// Check if game is full
	if room.GameMode == string(types.GameMode1v1) && len(room.Players) >= 2 ||
		room.GameMode == string(types.GameMode2v2) && len(room.Players) >= 4 ||
		room.GameMode == string(types.GameModeFFA3) && len(room.Players) >= 3 ||
		room.GameMode == string(types.GameModeFFA5) && len(room.Players) >= 5 {
		room.mutex.Unlock()
		client.SendError("Game is full")
		return
	}

	room.Players[playerName] = client

	// If a game already exists in this room, this is a reconnection
	// (player was removed by unregister before the new client joined)
	if room.Game != nil {
		room.mutex.Unlock()
		log.Printf("Player %s rejoined existing game %s", playerName, gameID)
		h.sendReconnectState(gameID, playerName)
		return
	}

	// Notify player they joined
	client.SendMessage(MsgPlayerJoined, PlayerJoinedPayload{
		GameMode:   room.GameMode,
		MaxPlayers: room.MaxPlayers,
		PlayerName: playerName,
	})

	log.Printf("Player %s joined game %s (%d/%d players)", playerName,
		gameID, len(room.Players), room.MaxPlayers)

	// Check if we should start the game
	shouldStartGame := false
	switch room.GameMode {
	case string(types.GameMode1v1):
		shouldStartGame = len(room.Players) == 2
	case string(types.GameMode2v2):
		shouldStartGame = len(room.Players) == 4
	case string(types.GameModeFFA3):
		shouldStartGame = len(room.Players) == 3
	case string(types.GameModeFFA5):
		shouldStartGame = len(room.Players) == 5
	}

	var playerNames []string

	// If we have 2 players, start the game
	if shouldStartGame {
		playerNames = make([]string, 0, 2)
		for name := range room.Players {
			playerNames = append(playerNames, name)
		}

		// Create the game
		room.Game = domain.NewGame(playerNames[0], playerNames[1], cards.NewDealer(),
			domain.NewGameStatusProvider())
		log.Printf("Game started: %s with players %v", gameID, playerNames)

		// Notify both players
		for name, c := range room.Players {
			c.SendMessage(MsgGameStarted, GameStartedPayload{
				GameID:   gameID,
				Players:  playerNames,
				YourName: name,
			})
		}
	} else {
		client.SendMessage(MsgWaitingForPlayer, nil)
	}

	room.mutex.Unlock()

	// Send initial warriors for selection AFTER releasing locks
	if shouldStartGame {
		h.sendInitialWarriors(gameID)
	}
}

// sendInitialWarriors sends each player their initial warriors to choose from
func (h *Hub) sendInitialWarriors(gameID string) {
	h.mutex.RLock()
	room, exists := h.gameRooms[gameID]
	h.mutex.RUnlock()

	if !exists || room.Game == nil {
		return
	}

	room.mutex.RLock()
	defer room.mutex.RUnlock()

	currentPlayer, _ := room.Game.WhoIsCurrent()
	currentPlayerName := currentPlayer.Name()

	// Get player objects to check who has already set warriors
	player1, player2 := room.Game.WhoIsCurrent()
	playersWithWarriors := make(map[string]bool)
	if len(player1.Field().Warriors()) > 0 {
		playersWithWarriors[player1.Name()] = true
	}
	if len(player2.Field().Warriors()) > 0 {
		playersWithWarriors[player2.Name()] = true
	}

	for playerName, client := range room.Players {
		// Skip players who have already set their warriors
		if playersWithWarriors[playerName] {
			log.Printf("Skipping initial warriors for %s (already set)", playerName)
			continue
		}

		warriors := room.Game.GetInitialWarriors(playerName)

		// Convert warriors to CardDTO, filtering out empty cards
		warriorDTOs := []CardDTO{}
		for _, w := range warriors {
			if w.CardID != "" {
				warriorDTOs = append(warriorDTOs, CardDTO{
					ID:      w.CardID,
					Type:    w.CardType.Name,
					SubType: w.CardType.SubName,
					Color:   w.CardType.Color,
					Value:   w.Value,
				})
			}
		}

		isYourTurn := playerName == currentPlayerName

		client.SendMessage(MsgInitialWarriors, InitialWarriorsPayload{
			Warriors:   warriorDTOs,
			IsYourTurn: isYourTurn,
		})

		log.Printf("Sent initial warriors to %s: isYourTurn=%v, warriors=%d",
			playerName, isYourTurn, len(warriorDTOs))
	}
}

// sendReconnectState sends the current game state to a reconnected player
func (h *Hub) sendReconnectState(gameID, playerName string) {
	h.mutex.RLock()
	room, exists := h.gameRooms[gameID]
	h.mutex.RUnlock()

	if !exists || room.Game == nil {
		return
	}

	room.mutex.RLock()
	defer room.mutex.RUnlock()

	client, ok := room.Players[playerName]
	if !ok {
		return
	}

	currentPlayer, enemyPlayer := room.Game.WhoIsCurrent()
	isCurrentPlayer := playerName == currentPlayer.Name()

	var status domain.GameStatus
	if isCurrentPlayer {
		status = room.Game.GameStatusProvider.Get(currentPlayer, enemyPlayer, room.Game)
	} else {
		status = room.Game.GameStatusProvider.Get(enemyPlayer, currentPlayer, room.Game)
	}

	// Send game_started so frontend transitions to the game screen
	playerNames := make([]string, 0, 2)
	for name := range room.Players {
		playerNames = append(playerNames, name)
	}
	client.SendMessage(MsgGameStarted, GameStartedPayload{
		GameID:   gameID,
		Players:  playerNames,
		YourName: playerName,
	})

	// Send current game state
	client.SendMessage(MsgGameState, GameStatePayload{
		GameStatus: ConvertGameStatus(status),
		IsYourTurn: isCurrentPlayer,
	})

	log.Printf("Sent reconnect state to %s: isYourTurn=%v, currentAction=%s",
		playerName, isCurrentPlayer, status.CurrentAction)
}

// sendGameStateWithStatus sends the game state using a pre-computed status for the current player
func (h *Hub) sendGameStateWithStatus(gameID string, currentPlayerStatus domain.GameStatus) {
	h.mutex.RLock()
	room, exists := h.gameRooms[gameID]
	h.mutex.RUnlock()

	if !exists || room.Game == nil {
		return
	}

	room.mutex.RLock()
	defer room.mutex.RUnlock()

	currentPlayer, enemyPlayer := room.Game.WhoIsCurrent()
	currentPlayerName := currentPlayer.Name()

	for playerName, client := range room.Players {
		var status domain.GameStatus
		isCurrentPlayer := playerName == currentPlayerName

		if isCurrentPlayer {
			// Use the pre-computed status for current player
			status = currentPlayerStatus
		} else {
			// Enemy sees their own hand and the current player's field
			status = room.Game.GameStatusProvider.Get(enemyPlayer, currentPlayer, room.Game)
			// Copy history from current player status so both players see the same history
			status.History = currentPlayerStatus.History
		}

		payload := GameStatePayload{
			GameStatus: ConvertGameStatus(status),
			IsYourTurn: isCurrentPlayer,
		}

		log.Printf("sendGameStateWithStatus to %s: isYourTurn=%v, currentAction=%s, newCards=%v",
			playerName, isCurrentPlayer, status.CurrentAction, status.NewCards)

		client.SendMessage(MsgGameState, payload)
	}
}

// broadcastToGame broadcasts a message to all players in a game
func (h *Hub) broadcastToGame(gameID string, msgType MessageType, payload interface{}) {
	h.mutex.RLock()
	room, exists := h.gameRooms[gameID]
	h.mutex.RUnlock()

	if !exists {
		return
	}

	room.mutex.RLock()
	defer room.mutex.RUnlock()

	for _, client := range room.Players {
		client.SendMessage(msgType, payload)
	}
}

// getGameRoom gets the game room for a client
func (h *Hub) getGameRoom(client *Client) (*GameRoom, bool) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	room, exists := h.gameRooms[client.GameID]
	return room, exists
}

// Helper function to execute game action and send state
func (h *Hub) executeGameAction(client *Client, action func(*domain.Game) (domain.GameStatus, error)) {
	room, exists := h.getGameRoom(client)
	if !exists || room.Game == nil {
		client.SendError("Game not found")
		return
	}

	room.mutex.Lock()
	status, err := action(room.Game)
	room.mutex.Unlock()

	if err != nil {
		client.SendError(err.Error())
		return
	}

	// Send updated game state to all players using the returned status
	h.sendGameStateWithStatus(client.GameID, status)
}

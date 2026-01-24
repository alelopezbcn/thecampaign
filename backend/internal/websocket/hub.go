package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/alelopezbcn/thecampaign/internal/domain"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
)

// ClientMessage represents a message from a client
type ClientMessage struct {
	client  *Client
	message *Message
}

// GameRoom represents a game room with two players
type GameRoom struct {
	ID      string
	Game    *domain.Game
	Players map[string]*Client // playerName -> client
	mutex   sync.RWMutex
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
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client unregistered, total clients: %d", len(h.clients))

				// Remove client from game room
				if client.GameID != "" {
					if room, exists := h.gameRooms[client.GameID]; exists {
						room.mutex.Lock()
						delete(room.Players, client.PlayerName)
						room.mutex.Unlock()

						// Notify other players
						h.broadcastToGame(client.GameID, MsgError, ErrorPayload{
							Message: client.PlayerName + " disconnected",
						})
					}
				}
			}
			h.mutex.Unlock()

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
	// Check if game is full
	if len(room.Players) >= 2 {
		room.mutex.Unlock()
		client.SendError("Game is full")
		return
	}

	// Check if player name already exists in this game
	if _, exists := room.Players[playerName]; exists {
		room.mutex.Unlock()
		client.SendError("Player name already taken in this game")
		return
	}

	room.Players[playerName] = client

	// Notify player they joined
	client.SendMessage(MsgPlayerJoined, PlayerJoinedPayload{
		PlayerName: playerName,
	})

	log.Printf("Player %s joined game %s (%d/2 players)", playerName, gameID, len(room.Players))

	// Check if we should start the game
	shouldStartGame := len(room.Players) == 2
	var playerNames []string

	// If we have 2 players, start the game
	if shouldStartGame {
		playerNames = make([]string, 0, 2)
		for name := range room.Players {
			playerNames = append(playerNames, name)
		}

		// Create the game
		room.Game = domain.NewGame(playerNames[0], playerNames[1], cards.NewDealer())
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

	// Send initial game state AFTER releasing locks
	if shouldStartGame {
		h.sendGameState(gameID)
	}
}

// sendGameState sends the current game state to all players in a game
// newlyDrawnCardID is optional - pass empty string if no card was just drawn
func (h *Hub) sendGameState(gameID string, newlyDrawnCardID ...string) {
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

	// Get the newly drawn card ID if provided
	drawnCardID := ""
	if len(newlyDrawnCardID) > 0 {
		drawnCardID = newlyDrawnCardID[0]
	}

	for playerName, client := range room.Players {
		var status gamestatus.GameStatus
		isCurrentPlayer := playerName == currentPlayerName

		if isCurrentPlayer {
			// Current player's turn - show their perspective
			status = room.Game.GetStatusForNextPlayer()
		} else {
			// Not their turn - show enemy player's perspective
			// Enemy sees their own hand and the current player's field
			status = gamestatus.NewGameStatus(enemyPlayer, currentPlayer)
		}

		// Only send newly drawn card ID to the player who received it
		cardIDForPlayer := ""
		if isCurrentPlayer {
			cardIDForPlayer = drawnCardID
		}

		payload := GameStatePayload{
			GameStatus:     ConvertGameStatus(status),
			IsYourTurn:     isCurrentPlayer,
			GameEnded:      room.Game.IsGameEnded(),
			NewlyDrawnCard: cardIDForPlayer,
		}

		log.Printf("sendGameState to %s: isYourTurn=%v, newlyDrawnCard='%s', handSize=%d, drawnCardID='%s'",
			playerName, isCurrentPlayer, cardIDForPlayer, len(status.CurrentPlayerHand), drawnCardID)

		// Debug: serialize and log the payload
		if debugData, err := json.Marshal(payload); err == nil {
			log.Printf("Payload JSON: %s", string(debugData))
		}

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
func (h *Hub) executeGameAction(client *Client, action func(*domain.Game) error) {
	room, exists := h.getGameRoom(client)
	if !exists || room.Game == nil {
		client.SendError("Game not found")
		return
	}

	room.mutex.Lock()
	err := action(room.Game)
	room.mutex.Unlock()

	if err != nil {
		client.SendError(err.Error())
		return
	}

	// Send updated game state to all players
	h.sendGameState(client.GameID)

	// Check if game ended
	if room.Game.IsGameEnded() {
		h.broadcastToGame(client.GameID, MsgGameEnded, nil)
	}
}

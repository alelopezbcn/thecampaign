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

// GameRoom represents a game room with players
type GameRoom struct {
	ID         string
	GameMode   types.GameMode
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
	case MsgDrawCard:
		h.handleDrawCard(client)
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

func maxPlayersForMode(mode types.GameMode) int {
	switch mode {
	case types.GameMode1v1:
		return 2
	case types.GameMode2v2:
		return 4
	case types.GameModeFFA3:
		return 3
	case types.GameModeFFA5:
		return 5
	default:
		return 2
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

	gameID := joinPayload.GameID
	playerName := joinPayload.PlayerName
	gameMode := types.GameMode(joinPayload.GameMode)

	client.PlayerName = playerName
	client.GameID = gameID

	h.mutex.Lock()
	room, exists := h.gameRooms[gameID]
	if !exists {
		room = &GameRoom{
			ID:         gameID,
			GameMode:   gameMode,
			MaxPlayers: maxPlayersForMode(gameMode),
			Players:    make(map[string]*Client),
		}
		h.gameRooms[gameID] = room
		log.Printf("Created new game room: %s (mode: %s, max: %d)",
			gameID, gameMode, room.MaxPlayers)
	}
	h.mutex.Unlock()

	room.mutex.Lock()

	// Reconnection: player name already exists in this game
	if oldClient, exists := room.Players[playerName]; exists {
		room.Players[playerName] = client
		gameInProgress := room.Game != nil
		room.mutex.Unlock()

		// Close old client connection
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
	if len(room.Players) >= room.MaxPlayers {
		room.mutex.Unlock()
		client.SendError("Game is full")
		return
	}

	room.Players[playerName] = client

	// If a game already exists (player was removed before the new client joined)
	if room.Game != nil {
		room.mutex.Unlock()
		log.Printf("Player %s rejoined existing game %s", playerName, gameID)
		h.sendReconnectState(gameID, playerName)
		return
	}

	log.Printf("Player %s joined game %s (%d/%d players)",
		playerName, gameID, len(room.Players), room.MaxPlayers)

	// Not enough players yet
	if len(room.Players) < room.MaxPlayers {
		allNames := make([]string, 0, len(room.Players))
		for name := range room.Players {
			allNames = append(allNames, name)
		}
		for _, c := range room.Players {
			c.SendMessage(MsgPlayerJoined, PlayerJoinedPayload{
				GameMode:   string(room.GameMode),
				MaxPlayers: room.MaxPlayers,
				PlayerName: playerName,
				Players:    allNames,
			})
		}
		room.mutex.Unlock()
		return
	}

	// All players are in - start the game
	playerNames := make([]string, 0, room.MaxPlayers)
	for name := range room.Players {
		playerNames = append(playerNames, name)
	}

	game, err := domain.NewGame(playerNames, room.GameMode, cards.NewDealer(),
		domain.NewGameStatusProvider())
	if err != nil {
		room.mutex.Unlock()
		log.Printf("Error creating game: %v", err)
		for _, c := range room.Players {
			c.SendError("Failed to create game: " + err.Error())
		}
		return
	}
	room.Game = game

	// Auto-move initial 3 warriors to field for each player
	for _, name := range playerNames {
		warriors := game.GetInitialWarriors(name)
		for _, w := range warriors {
			if w.CardID != "" {
				if err := game.AutoMoveWarriorToField(name, w.CardID); err != nil {
					log.Printf("Error auto-moving warrior %s for %s: %v",
						w.CardID, name, err)
				}
			}
		}
	}

	log.Printf("Game started: %s with players %v (mode: %s)",
		gameID, playerNames, room.GameMode)

	// Notify all players that the game started
	for name, c := range room.Players {
		c.SendMessage(MsgGameStarted, GameStartedPayload{
			GameID:   gameID,
			Players:  playerNames,
			YourName: name,
		})
	}

	room.mutex.Unlock()

	// Auto draw card for the first player and send game state
	h.autoDrawAndBroadcast(gameID)
}

// autoDrawAndBroadcast draws a card for the current player and sends state to all
func (h *Hub) autoDrawAndBroadcast(gameID string) {
	h.mutex.RLock()
	room, exists := h.gameRooms[gameID]
	h.mutex.RUnlock()

	if !exists || room.Game == nil {
		return
	}

	room.mutex.Lock()
	currentPlayer := room.Game.CurrentPlayer()
	status, err := room.Game.DrawCard(currentPlayer.Name())
	room.mutex.Unlock()

	if err != nil {
		log.Printf("Error auto-drawing card for %s: %v", currentPlayer.Name(), err)
		return
	}

	h.sendGameStateToAll(gameID, status)
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

	player := room.Game.GetPlayer(playerName)
	if player == nil {
		return
	}

	currentPlayerName := room.Game.CurrentPlayer().Name()
	isCurrentPlayer := playerName == currentPlayerName

	status := room.Game.GameStatusProvider.Get(player, room.Game)

	// Send game_started so frontend transitions to the game screen
	playerNames := make([]string, 0, len(room.Players))
	for name := range room.Players {
		playerNames = append(playerNames, name)
	}
	client.SendMessage(MsgGameStarted, GameStartedPayload{
		GameID:   gameID,
		Players:  playerNames,
		YourName: playerName,
	})

	client.SendMessage(MsgGameState, GameStatePayload{
		GameStatus: ConvertGameStatus(status),
		IsYourTurn: isCurrentPlayer,
	})

	log.Printf("Sent reconnect state to %s: isYourTurn=%v, currentAction=%s",
		playerName, isCurrentPlayer, status.CurrentAction)
}

// sendGameStateToAll sends personalized game state to every player in the room
func (h *Hub) sendGameStateToAll(gameID string, currentPlayerStatus domain.GameStatus) {
	h.mutex.RLock()
	room, exists := h.gameRooms[gameID]
	h.mutex.RUnlock()

	if !exists || room.Game == nil {
		return
	}

	room.mutex.RLock()
	defer room.mutex.RUnlock()

	currentPlayerName := room.Game.CurrentPlayer().Name()

	for playerName, client := range room.Players {
		isCurrentPlayer := playerName == currentPlayerName

		var status domain.GameStatus
		if isCurrentPlayer {
			status = currentPlayerStatus
		} else {
			player := room.Game.GetPlayer(playerName)
			if player == nil {
				continue
			}
			status = room.Game.GameStatusProvider.Get(player, room.Game)
			status.History = currentPlayerStatus.History
		}

		client.SendMessage(MsgGameState, GameStatePayload{
			GameStatus: ConvertGameStatus(status),
			IsYourTurn: isCurrentPlayer,
		})
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

// executeGameAction executes a game action and sends state to all players
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

	h.sendGameStateToAll(client.GameID, status)
}

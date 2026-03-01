// Package websocket implements the WebSocket hub and message handling for "The Campaign" game.
package websocket

import (
	"encoding/json"
	"log"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/alelopezbcn/thecampaign/internal/domain"
	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gameactions"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

const (
	turnTimeLimit           = 120 * time.Second
	waitingRoomCleanupDelay = 60 * time.Second
)

type HubGame interface {
	ExecuteAction(action gameactions.GameAction) (gamestatus.GameStatus, error)
	CurrentPlayer() board.Player
	GetPlayer(name string) board.Player
	ReconnectPlayer(name string)
	DisconnectPlayer(playerName string) error
	IsGameOver() (bool, string)
	Status(viewer board.Player, newCards ...cards.Card) gamestatus.GameStatus
}

// ClientMessage represents a message from a client
type ClientMessage struct {
	client  *Client
	message *Message
}

// GameRoom represents a game room with players
type GameRoom struct {
	ID              string
	GameMode        types.GameMode
	MaxPlayers      int
	Game            HubGame
	Players         map[string]*Client // playerName -> client
	TeamAssignments map[string]int     // playerName -> teamNumber (1 or 2), 2v2 only
	mutex           sync.RWMutex
	turnTimer       *time.Timer
	turnTimerStop   chan struct{}
	cleanupTimer    *time.Timer
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
			var gameInProgress bool

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
						if room.Game != nil {
							gameInProgress = true
							room.Players[client.PlayerName] = nil
						} else {
							// Waiting room: remove normally
							delete(room.Players, client.PlayerName)
							delete(room.TeamAssignments, client.PlayerName)
							if len(room.Players) == 0 {
								gameIDToCleanup := client.GameID
								room.cleanupTimer = time.AfterFunc(waitingRoomCleanupDelay, func() {
									h.cleanupEmptyRoom(gameIDToCleanup)
								})
								log.Printf("Game room %s empty, scheduled cleanup in %v", client.GameID, waitingRoomCleanupDelay)
							}
						}
						room.mutex.Unlock()
					}
				}
			}
			h.mutex.Unlock()

			if disconnectedGameID != "" {
				if gameInProgress {
					h.handlePlayerDisconnection(disconnectedGameID, disconnectedPlayerName)
				} else {
					h.broadcastToGame(disconnectedGameID, MsgError, ErrorPayload{
						Message: disconnectedPlayerName + " disconnected",
					})
				}
			}

		case clientMsg := <-h.handleMessage:
			h.processMessage(clientMsg.client, clientMsg.message)
		}
	}
}

// messageHandlers maps each MessageType to its handler function.
// Adding a new message type requires one entry here — no switch needed.
var messageHandlers = map[MessageType]func(*Hub, *Client, interface{}){
	MsgJoinGame:     func(h *Hub, c *Client, p interface{}) { h.handleJoinGame(c, p) },
	MsgDrawCard:     func(h *Hub, c *Client, _ interface{}) { h.handleDrawCard(c) },
	MsgAttack:       func(h *Hub, c *Client, p interface{}) { h.handleAttack(c, p) },
	MsgSpecialPower: func(h *Hub, c *Client, p interface{}) { h.handleSpecialPower(c, p) },
	MsgHarpoon:      func(h *Hub, c *Client, p interface{}) { h.handleHarpoon(c, p) },
	MsgBloodRain:    func(h *Hub, c *Client, p interface{}) { h.handleBloodRain(c, p) },
	MsgMoveWarrior:  func(h *Hub, c *Client, p interface{}) { h.handleMoveWarrior(c, p) },
	MsgTrade:        func(h *Hub, c *Client, p interface{}) { h.handleTrade(c, p) },
	MsgBuy:          func(h *Hub, c *Client, p interface{}) { h.handleBuy(c, p) },
	MsgBuyMercenary: func(h *Hub, c *Client, p interface{}) { h.handleBuyMercenary(c, p) },
	MsgConstruct:    func(h *Hub, c *Client, p interface{}) { h.handleConstruct(c, p) },
	MsgSpy:          func(h *Hub, c *Client, p interface{}) { h.handleSpy(c, p) },
	MsgSteal:        func(h *Hub, c *Client, p interface{}) { h.handleSteal(c, p) },
	MsgDesertion:    func(h *Hub, c *Client, p interface{}) { h.handleDesertion(c, p) },
	MsgCatapult:     func(h *Hub, c *Client, p interface{}) { h.handleCatapult(c, p) },
	MsgFortress:     func(h *Hub, c *Client, p interface{}) { h.handleFortress(c, p) },
	MsgResurrection: func(h *Hub, c *Client, p interface{}) { h.handleResurrection(c, p) },
	MsgSabotage:     func(h *Hub, c *Client, p interface{}) { h.handleSabotage(c, p) },
	MsgPlaceAmbush:  func(h *Hub, c *Client, p interface{}) { h.handlePlaceAmbush(c, p) },
	MsgEndTurn:      func(h *Hub, c *Client, _ interface{}) { h.handleEndTurn(c) },
	MsgSkipPhase:    func(h *Hub, c *Client, _ interface{}) { h.handleSkipPhase(c) },
	MsgSwapTeam:     func(h *Hub, c *Client, _ interface{}) { h.handleSwapTeam(c) },
	MsgStartGame:    func(h *Hub, c *Client, _ interface{}) { h.handleStartGame(c) },
	MsgRestartGame:  func(h *Hub, c *Client, _ interface{}) { h.handleRestartGame(c) },
}

// processMessage processes incoming messages from clients
func (h *Hub) processMessage(client *Client, msg *Message) {
	log.Printf("Processing message type: %s from player: %s", msg.Type, client.PlayerName)

	if handler, ok := messageHandlers[msg.Type]; ok {
		handler(h, client, msg.Payload)
	} else {
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

// generateGameID creates a short 6-character uppercase alphanumeric game ID.
// Avoids ambiguous characters (I, O, 0, 1).
func generateGameID() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	id := make([]byte, 6)
	for i := range id {
		id[i] = chars[rand.IntN(len(chars))]
	}
	return string(id)
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

	var room *GameRoom

	h.mutex.Lock()
	if gameID == "" {
		// Create mode: generate unique game ID
		for {
			gameID = generateGameID()
			if _, exists := h.gameRooms[gameID]; !exists {
				break
			}
		}
		room = &GameRoom{
			ID:              gameID,
			GameMode:        gameMode,
			MaxPlayers:      maxPlayersForMode(gameMode),
			Players:         make(map[string]*Client),
			TeamAssignments: make(map[string]int),
		}
		h.gameRooms[gameID] = room
		client.GameID = gameID
		log.Printf("Created new game room: %s (mode: %s, max: %d)",
			gameID, gameMode, room.MaxPlayers)
	} else {
		var exists bool
		room, exists = h.gameRooms[gameID]
		if !exists {
			h.mutex.Unlock()
			client.SendError("Game not found")
			return
		}
	}
	h.mutex.Unlock()

	room.mutex.Lock()

	// Cancel any pending cleanup timer now that someone is joining
	if room.cleanupTimer != nil {
		room.cleanupTimer.Stop()
		room.cleanupTimer = nil
	}

	// Reconnection: player name already exists in this game (includes nil = disconnected)
	if oldClient, exists := room.Players[playerName]; exists {
		room.Players[playerName] = client
		gameInProgress := room.Game != nil

		// Close old client connection if still active
		if oldClient != nil {
			h.mutex.Lock()
			if _, ok := h.clients[oldClient]; ok {
				delete(h.clients, oldClient)
				close(oldClient.send)
			}
			h.mutex.Unlock()
		}

		log.Printf("Player %s reconnected to game %s", playerName, gameID)

		if gameInProgress {
			room.Game.ReconnectPlayer(playerName)
			room.mutex.Unlock()
			h.broadcastToGame(gameID, MsgError, ErrorPayload{
				Message: playerName + " reconnected",
			})
			h.sendReconnectState(gameID, playerName)
			// Broadcast updated state to all so opponents see player is back
			h.broadcastGameStateToAll(gameID)
		} else {
			// Send full waiting room state including teams
			allNames := make([]string, 0, len(room.Players))
			for name := range room.Players {
				allNames = append(allNames, name)
			}
			payload := PlayerJoinedPayload{
				GameID:     room.ID,
				GameMode:   string(room.GameMode),
				MaxPlayers: room.MaxPlayers,
				PlayerName: playerName,
				Players:    allNames,
				Teams:      copyTeams(room.TeamAssignments),
			}
			room.mutex.Unlock()
			client.SendMessage(MsgPlayerJoined, payload)
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

	// Auto-assign team for 2v2
	if room.GameMode == types.GameMode2v2 {
		team1Count, team2Count := 0, 0
		for _, t := range room.TeamAssignments {
			if t == 1 {
				team1Count++
			}
			if t == 2 {
				team2Count++
			}
		}
		if team1Count <= team2Count {
			room.TeamAssignments[playerName] = 1
		} else {
			room.TeamAssignments[playerName] = 2
		}
	}

	// If a game already exists (player was removed before the new client joined)
	if room.Game != nil {
		room.mutex.Unlock()
		log.Printf("Player %s rejoined existing game %s", playerName, gameID)
		h.sendReconnectState(gameID, playerName)
		return
	}

	log.Printf("Player %s joined game %s (%d/%d players)",
		playerName, gameID, len(room.Players), room.MaxPlayers)

	// Broadcast updated player list to all players in the room
	allNames := make([]string, 0, len(room.Players))
	for name := range room.Players {
		allNames = append(allNames, name)
	}
	teams := copyTeams(room.TeamAssignments)
	for _, c := range room.Players {
		if c == nil {
			continue
		}
		c.SendMessage(MsgPlayerJoined, PlayerJoinedPayload{
			GameID:     room.ID,
			GameMode:   string(room.GameMode),
			MaxPlayers: room.MaxPlayers,
			PlayerName: playerName,
			Players:    allNames,
			Teams:      teams,
		})
	}
	room.mutex.Unlock()
}

// handleStartGame handles a player requesting to start the game
func (h *Hub) handleStartGame(client *Client) {
	room, exists := h.getGameRoom(client)
	if !exists {
		client.SendError("Game not found")
		return
	}

	room.mutex.Lock()

	if room.Game != nil {
		room.mutex.Unlock()
		client.SendError("Game already started")
		return
	}

	if len(room.Players) < room.MaxPlayers {
		room.mutex.Unlock()
		client.SendError("Not all players have joined yet")
		return
	}

	gameID := room.ID

	// For 2v2, interleave teams: T1, T2, T1, T2
	var playerNames []string
	if room.GameMode == types.GameMode2v2 {
		var t1Names, t2Names []string
		for name, team := range room.TeamAssignments {
			if team == 1 {
				t1Names = append(t1Names, name)
			} else {
				t2Names = append(t2Names, name)
			}
		}
		playerNames = []string{t1Names[0], t2Names[0], t1Names[1], t2Names[1]}
	} else {
		playerNames = make([]string, 0, room.MaxPlayers)
		for name := range room.Players {
			playerNames = append(playerNames, name)
		}
	}

	game, err := domain.NewGame(playerNames, room.GameMode, cards.NewDealer())
	if err != nil {
		room.mutex.Unlock()
		log.Printf("Error creating game: %v", err)
		for _, c := range room.Players {
			if c == nil {
				continue
			}
			c.SendError("Failed to create game: " + err.Error())
		}
		return
	}
	room.Game = game

	// Auto-move initial warriors to field for each player
	for _, name := range playerNames {
		if err := game.AutoMoveWarriorsToField(name); err != nil {
			log.Printf("Error auto-moving warriors for %s: %v", name, err)
		}
	}

	log.Printf("Game started: %s with players %v (mode: %s)",
		gameID, playerNames, room.GameMode)

	// Notify all players that the game started
	for name, c := range room.Players {
		if c == nil {
			continue
		}
		c.SendMessage(MsgGameStarted, GameStartedPayload{
			GameID:   gameID,
			Players:  playerNames,
			YourName: name,
		})
	}

	room.mutex.Unlock()

	// Auto draw card for the first player and send game state
	h.autoDrawAndBroadcast(gameID)
	h.startTurnTimer(gameID)
}

// handleRestartGame resets the game for all players in the room with the same players and mode.
func (h *Hub) handleRestartGame(client *Client) {
	room, exists := h.getGameRoom(client)
	if !exists {
		client.SendError("Game not found")
		return
	}

	room.mutex.Lock()

	if room.Game == nil {
		room.mutex.Unlock()
		client.SendError("Game not started yet")
		return
	}

	over, _ := room.Game.IsGameOver()
	if !over {
		room.mutex.Unlock()
		client.SendError("Game is not over yet")
		return
	}

	gameID := room.ID

	var playerNames []string
	if room.GameMode == types.GameMode2v2 {
		var t1Names, t2Names []string
		for name, team := range room.TeamAssignments {
			if team == 1 {
				t1Names = append(t1Names, name)
			} else {
				t2Names = append(t2Names, name)
			}
		}
		playerNames = []string{t1Names[0], t2Names[0], t1Names[1], t2Names[1]}
	} else {
		for name := range room.Players {
			playerNames = append(playerNames, name)
		}
	}

	game, err := domain.NewGame(playerNames, room.GameMode, cards.NewDealer())
	if err != nil {
		room.mutex.Unlock()
		log.Printf("Error restarting game: %v", err)
		client.SendError("Failed to restart game: " + err.Error())
		return
	}
	room.Game = game

	for _, name := range playerNames {
		if err := game.AutoMoveWarriorsToField(name); err != nil {
			log.Printf("Error auto-moving warriors for %s: %v", name, err)
		}
	}

	log.Printf("Game restarted: %s with players %v (mode: %s)", gameID, playerNames, room.GameMode)

	for name, c := range room.Players {
		if c == nil {
			continue
		}
		c.SendMessage(MsgGameStarted, GameStartedPayload{
			GameID:   gameID,
			Players:  playerNames,
			YourName: name,
		})
	}

	room.mutex.Unlock()

	h.stopTurnTimer(gameID)
	h.autoDrawAndBroadcast(gameID)
	h.startTurnTimer(gameID)
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
	status, err := room.Game.ExecuteAction(gameactions.NewDrawCardAction(currentPlayer.Name()))
	room.mutex.Unlock()

	if err != nil {
		log.Printf("Error auto-drawing card for %s: %v", currentPlayer.Name(), err)
		return
	}

	h.sendGameStateToAll(gameID, status)
}

// startTurnTimer starts (or restarts) the turn timer for a game room.
// When the timer expires, it auto-ends the current player's turn.
func (h *Hub) startTurnTimer(gameID string) {
	h.mutex.RLock()
	room, exists := h.gameRooms[gameID]
	h.mutex.RUnlock()

	if !exists || room.Game == nil {
		return
	}

	// Stop any existing timer
	if room.turnTimerStop != nil {
		close(room.turnTimerStop)
	}
	if room.turnTimer != nil {
		room.turnTimer.Stop()
	}

	room.turnTimerStop = make(chan struct{})
	room.turnTimer = time.NewTimer(turnTimeLimit)
	stop := room.turnTimerStop

	go func() {
		select {
		case <-room.turnTimer.C:
			room.mutex.Lock()
			if room.Game == nil {
				room.mutex.Unlock()
				return
			}

			if over, _ := room.Game.IsGameOver(); over {
				room.mutex.Unlock()
				return
			}

			currentPlayer := room.Game.CurrentPlayer().Name()
			log.Printf("Turn timer expired for %s in game %s", currentPlayer, gameID)

			status, err := room.Game.ExecuteAction(gameactions.NewEndTurnPhaseAction(currentPlayer, true)) // Auto-end turn due to timer expiration
			room.mutex.Unlock()

			if err != nil {
				log.Printf("Error auto-ending turn for %s: %v", currentPlayer, err)
				return
			}

			h.sendGameStateToAll(gameID, status)
			h.autoDrawAndBroadcast(gameID)
			h.startTurnTimer(gameID)
		case <-stop:
			return
		}
	}()
}

// stopTurnTimer stops the turn timer for a game room.
func (h *Hub) stopTurnTimer(gameID string) {
	h.mutex.RLock()
	room, exists := h.gameRooms[gameID]
	h.mutex.RUnlock()

	if !exists {
		return
	}

	if room.turnTimerStop != nil {
		close(room.turnTimerStop)
		room.turnTimerStop = nil
	}
	if room.turnTimer != nil {
		room.turnTimer.Stop()
		room.turnTimer = nil
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

	player := room.Game.GetPlayer(playerName)
	if player == nil {
		return
	}

	currentPlayerName := room.Game.CurrentPlayer().Name()
	isCurrentPlayer := playerName == currentPlayerName

	status := room.Game.Status(player)

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
func (h *Hub) sendGameStateToAll(gameID string, currentPlayerStatus gamestatus.GameStatus) {
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
		if client == nil {
			continue
		}
		isCurrentPlayer := playerName == currentPlayerName

		var status gamestatus.GameStatus
		if isCurrentPlayer {
			status = currentPlayerStatus
		} else {
			player := room.Game.GetPlayer(playerName)
			if player == nil {
				continue
			}
			status = room.Game.Status(player)
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
		if client != nil {
			client.SendMessage(msgType, payload)
		}
	}
}

// handlePlayerDisconnection handles a player disconnecting from an active game.
// Marks them as disconnected, advances turn if needed, and broadcasts state.
func (h *Hub) handlePlayerDisconnection(gameID, playerName string) {
	h.mutex.RLock()
	room, exists := h.gameRooms[gameID]
	h.mutex.RUnlock()

	if !exists || room.Game == nil {
		return
	}

	room.mutex.Lock()
	if over, _ := room.Game.IsGameOver(); over {
		room.mutex.Unlock()
		return
	}

	wasTheirTurn := room.Game.CurrentPlayer().Name() == playerName

	err := room.Game.DisconnectPlayer(playerName)
	gameOver, _ := room.Game.IsGameOver()
	room.mutex.Unlock()

	if err != nil {
		log.Printf("Error disconnecting player %s: %v", playerName, err)
		h.broadcastToGame(gameID, MsgError, ErrorPayload{
			Message: playerName + " disconnected",
		})
		return
	}

	h.broadcastToGame(gameID, MsgError, ErrorPayload{
		Message: playerName + " disconnected",
	})

	if gameOver {
		h.broadcastGameStateToAll(gameID)
		h.stopTurnTimer(gameID)
		return
	}

	if wasTheirTurn {
		h.autoDrawAndBroadcast(gameID)
		h.startTurnTimer(gameID)
	} else {
		h.broadcastGameStateToAll(gameID)
	}
}

// broadcastGameStateToAll gets the current player's status and sends state to all players.
func (h *Hub) broadcastGameStateToAll(gameID string) {
	h.mutex.RLock()
	room, exists := h.gameRooms[gameID]
	h.mutex.RUnlock()

	if !exists || room.Game == nil {
		return
	}

	room.mutex.RLock()
	currentPlayer := room.Game.CurrentPlayer()
	status := room.Game.Status(currentPlayer)
	room.mutex.RUnlock()

	h.sendGameStateToAll(gameID, status)
}

// getGameRoom gets the game room for a client
func (h *Hub) getGameRoom(client *Client) (*GameRoom, bool) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	room, exists := h.gameRooms[client.GameID]
	return room, exists
}

func copyTeams(src map[string]int) map[string]int {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]int, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func (h *Hub) handleSwapTeam(client *Client) {
	room, exists := h.getGameRoom(client)
	if !exists {
		client.SendError("Game not found")
		return
	}

	room.mutex.Lock()
	defer room.mutex.Unlock()

	if room.GameMode != types.GameMode2v2 {
		client.SendError("Team swap only available in 2v2 mode")
		return
	}
	if room.Game != nil {
		client.SendError("Cannot swap teams after game has started")
		return
	}

	currentTeam := room.TeamAssignments[client.PlayerName]
	targetTeam := 1
	if currentTeam == 1 {
		targetTeam = 2
	}

	// Validate target team has room
	count := 0
	for _, t := range room.TeamAssignments {
		if t == targetTeam {
			count++
		}
	}
	if count >= 2 {
		client.SendError("Target team is full")
		return
	}

	room.TeamAssignments[client.PlayerName] = targetTeam

	allNames := make([]string, 0, len(room.Players))
	for name := range room.Players {
		allNames = append(allNames, name)
	}
	teams := copyTeams(room.TeamAssignments)

	for _, c := range room.Players {
		if c == nil {
			continue
		}
		c.SendMessage(MsgPlayerJoined, PlayerJoinedPayload{
			GameID:     room.ID,
			GameMode:   string(room.GameMode),
			MaxPlayers: room.MaxPlayers,
			PlayerName: client.PlayerName,
			Players:    allNames,
			Teams:      teams,
		})
	}
}

// executeGameAction executes a game action and sends state to all players
func (h *Hub) executeGameAction(client *Client, action func(HubGame) (gamestatus.GameStatus, error)) {
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

	if status.GameOverMgs != "" {
		h.stopTurnTimer(client.GameID)
	}
}

// cleanupEmptyRoom deletes a waiting room if it is still empty when the cleanup timer fires.
// Runs in a separate goroutine via time.AfterFunc.
func (h *Hub) cleanupEmptyRoom(gameID string) {
	h.mutex.Lock()
	room, exists := h.gameRooms[gameID]
	if !exists {
		h.mutex.Unlock()
		return
	}
	h.mutex.Unlock()

	room.mutex.RLock()
	empty := len(room.Players) == 0 && room.Game == nil
	room.mutex.RUnlock()

	if empty {
		h.mutex.Lock()
		// Re-check under h.mutex to avoid a race with a concurrent join
		if r, ok := h.gameRooms[gameID]; ok {
			r.mutex.RLock()
			stillEmpty := len(r.Players) == 0 && r.Game == nil
			r.mutex.RUnlock()
			if stillEmpty {
				delete(h.gameRooms, gameID)
				log.Printf("Game room %s removed (empty after cleanup delay)", gameID)
			}
		}
		h.mutex.Unlock()
	}
}

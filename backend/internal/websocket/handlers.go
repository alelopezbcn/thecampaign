package websocket

import (
	"encoding/json"
	"log"

	"github.com/alelopezbcn/thecampaign/internal/domain"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
)

func (h *Hub) handleSetInitialWarriors(client *Client, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		client.SendError("Invalid payload")
		return
	}

	var p SetInitialWarriorsPayload
	if err := json.Unmarshal(data, &p); err != nil {
		client.SendError("Invalid set initial warriors payload")
		return
	}

	room, exists := h.getGameRoom(client)
	if !exists || room.Game == nil {
		client.SendError("Game not found")
		return
	}

	room.mutex.Lock()

	// Set initial warriors
	if err := room.Game.SetInitialWarriors(client.PlayerName, p.WarriorIDs); err != nil {
		room.mutex.Unlock()
		client.SendError(err.Error())
		return
	}

	// Check if both players have warriors on field (setup complete)
	currentPlayer, enemyPlayer := room.Game.WhoIsCurrent()
	bothHaveWarriors := len(currentPlayer.Field().Warriors()) > 0 && len(enemyPlayer.Field().Warriors()) > 0

	log.Printf("SetInitialWarriors: currentPlayer=%s, currentPlayerField=%d, enemyField=%d, bothHaveWarriors=%v",
		currentPlayer.Name(), len(currentPlayer.Field().Warriors()), len(enemyPlayer.Field().Warriors()), bothHaveWarriors)

	var status gamestatus.GameStatus
	if bothHaveWarriors {
		// Setup complete - auto draw for the current player
		log.Printf("Setup complete! Drawing card for %s", currentPlayer.Name())

		status, err = room.Game.DrawCard(currentPlayer.Name())
		if err != nil {
			log.Printf("Error drawing card: %v", err)
			room.mutex.Unlock()
			client.SendError(err.Error())
			return
		}
		log.Printf("Drew card, new cards: %v", status.NewCards)
	}

	room.mutex.Unlock()

	// Send updated game state
	if bothHaveWarriors {
		h.sendGameStateWithStatus(client.GameID, status)
	} else {
		h.sendGameState(client.GameID)
	}

	// Check if game ended
	if room.Game.IsGameEnded() {
		h.broadcastToGame(client.GameID, MsgGameEnded, nil)
	}
}

func (h *Hub) handleAttack(client *Client, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		client.SendError("Invalid payload")
		return
	}

	var p AttackPayload
	if err := json.Unmarshal(data, &p); err != nil {
		client.SendError("Invalid attack payload")
		return
	}

	h.executeGameAction(client, func(g *domain.Game) (gamestatus.GameStatus, error) {
		return g.Attack(client.PlayerName, p.TargetID, p.WeaponID)
	})
}

func (h *Hub) handleSpecialPower(client *Client, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		client.SendError("Invalid payload")
		return
	}

	var p SpecialPowerPayload
	if err := json.Unmarshal(data, &p); err != nil {
		client.SendError("Invalid special power payload")
		return
	}

	h.executeGameAction(client, func(g *domain.Game) (gamestatus.GameStatus, error) {
		return g.SpecialPower(client.PlayerName, p.UserID, p.TargetID, p.WeaponID)
	})
}

func (h *Hub) handleMoveWarrior(client *Client, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		client.SendError("Invalid payload")
		return
	}

	var p MoveWarriorPayload
	if err := json.Unmarshal(data, &p); err != nil {
		client.SendError("Invalid move warrior payload")
		return
	}

	h.executeGameAction(client, func(g *domain.Game) (gamestatus.GameStatus, error) {
		return g.MoveWarriorToField(client.PlayerName, p.WarriorID)
	})
}

func (h *Hub) handleTrade(client *Client, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		client.SendError("Invalid payload")
		return
	}

	var p TradePayload
	if err := json.Unmarshal(data, &p); err != nil {
		client.SendError("Invalid trade payload")
		return
	}

	h.executeGameAction(client, func(g *domain.Game) (gamestatus.GameStatus, error) {
		return g.Trade(client.PlayerName, p.CardIDs)
	})
}

func (h *Hub) handleBuy(client *Client, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		client.SendError("Invalid payload")
		return
	}

	var p BuyPayload
	if err := json.Unmarshal(data, &p); err != nil {
		client.SendError("Invalid buy payload")
		return
	}

	h.executeGameAction(client, func(g *domain.Game) (gamestatus.GameStatus, error) {
		return g.Buy(client.PlayerName, p.CardID)
	})
}

func (h *Hub) handleConstruct(client *Client, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		client.SendError("Invalid payload")
		return
	}

	var p ConstructPayload
	if err := json.Unmarshal(data, &p); err != nil {
		client.SendError("Invalid construct payload")
		return
	}

	h.executeGameAction(client, func(g *domain.Game) (gamestatus.GameStatus, error) {
		return g.Construct(client.PlayerName, p.CardID)
	})
}

func (h *Hub) handleSpy(client *Client, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		client.SendError("Invalid payload")
		return
	}

	var p SpyPayload
	if err := json.Unmarshal(data, &p); err != nil {
		client.SendError("Invalid spy payload")
		return
	}

	room, exists := h.getGameRoom(client)
	if !exists || room.Game == nil {
		client.SendError("Game not found")
		return
	}

	room.mutex.Lock()
	cards, status, err := room.Game.Spy(client.PlayerName, p.Option)
	room.mutex.Unlock()

	if err != nil {
		client.SendError(err.Error())
		return
	}

	// Send the revealed cards only to the client who used spy
	client.SendMessage("spy_result", map[string]interface{}{
		"cards": cards,
	})

	// Send updated game state with the returned status
	h.sendGameStateWithStatus(client.GameID, status)

	// Check if game ended
	if room.Game.IsGameEnded() {
		h.broadcastToGame(client.GameID, MsgGameEnded, nil)
	}
}

func (h *Hub) handleSteal(client *Client, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		client.SendError("Invalid payload")
		return
	}

	var p StealPayload
	if err := json.Unmarshal(data, &p); err != nil {
		client.SendError("Invalid steal payload")
		return
	}

	room, exists := h.getGameRoom(client)
	if !exists || room.Game == nil {
		client.SendError("Game not found")
		return
	}

	room.mutex.Lock()
	stolenCard, status, err := room.Game.Steal(client.PlayerName, p.CardPosition)
	room.mutex.Unlock()

	if err != nil {
		client.SendError(err.Error())
		return
	}

	// Send the stolen card only to the client who used steal
	client.SendMessage("steal_result", map[string]interface{}{
		"card": stolenCard,
	})

	// Send updated game state with the returned status
	h.sendGameStateWithStatus(client.GameID, status)

	// Check if game ended
	if room.Game.IsGameEnded() {
		h.broadcastToGame(client.GameID, MsgGameEnded, nil)
	}
}

func (h *Hub) handleCatapult(client *Client, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		client.SendError("Invalid payload")
		return
	}

	var p CatapultPayload
	if err := json.Unmarshal(data, &p); err != nil {
		client.SendError("Invalid catapult payload")
		return
	}

	h.executeGameAction(client, func(g *domain.Game) (gamestatus.GameStatus, error) {
		return g.Catapult(client.PlayerName, p.CardPosition)
	})
}

func (h *Hub) handleEndTurn(client *Client) {
	log.Printf("handleEndTurn called by %s", client.PlayerName)

	room, exists := h.getGameRoom(client)
	if !exists || room.Game == nil {
		client.SendError("Game not found")
		return
	}

	room.mutex.Lock()

	// End the current player's turn
	_, err := room.Game.EndTurn(client.PlayerName)
	if err != nil {
		log.Printf("EndTurn error: %v", err)
		room.mutex.Unlock()
		client.SendError(err.Error())
		return
	}

	// Get the new current player
	nextPlayer, _ := room.Game.WhoIsCurrent()
	log.Printf("EndTurn: nextPlayer=%s", nextPlayer.Name())

	// Automatically draw a card for the new current player
	status, err := room.Game.DrawCard(nextPlayer.Name())
	if err != nil {
		log.Printf("DrawCard error: %v", err)
		room.mutex.Unlock()
		client.SendError(err.Error())
		return
	}

	log.Printf("EndTurn: newCards=%v", status.NewCards)

	room.mutex.Unlock()

	// Send updated game state with the status from DrawCard
	h.sendGameStateWithStatus(client.GameID, status)

	// Check if game ended
	if room.Game.IsGameEnded() {
		h.broadcastToGame(client.GameID, MsgGameEnded, nil)
	}
}

func (h *Hub) handleSkipPhase(client *Client) {
	log.Printf("handleSkipPhase called by %s", client.PlayerName)

	h.executeGameAction(client, func(g *domain.Game) (gamestatus.GameStatus, error) {
		return g.SkipPhase(client.PlayerName)
	})
}

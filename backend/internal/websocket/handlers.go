package websocket

import (
	"encoding/json"
	"log"

	"github.com/alelopezbcn/thecampaign/internal/domain"
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
	// After SetInitialWarriors, the turn has already switched, so check both players' fields directly
	currentPlayer, enemyPlayer := room.Game.WhoIsCurrent()
	bothHaveWarriors := len(currentPlayer.Field().Warriors()) > 0 && len(enemyPlayer.Field().Warriors()) > 0

	log.Printf("SetInitialWarriors: currentPlayer=%s, currentPlayerField=%d, enemyField=%d, bothHaveWarriors=%v",
		currentPlayer.Name(), len(currentPlayer.Field().Warriors()), len(enemyPlayer.Field().Warriors()), bothHaveWarriors)

	var newCardID string
	if bothHaveWarriors {
		// Setup complete - auto draw for the current player
		// Get the current player's hand before drawing
		handBefore := make(map[string]bool)
		handCardsBefore := currentPlayer.Hand().ShowCards()
		for _, card := range handCardsBefore {
			handBefore[card.GetID()] = true
		}

		log.Printf("Setup complete! Drawing card for %s (hand size before: %d)", currentPlayer.Name(), len(handCardsBefore))

		if err := room.Game.DrawCards(currentPlayer.Name(), 1); err != nil {
			log.Printf("Error drawing card: %v", err)
			room.mutex.Unlock()
			client.SendError(err.Error())
			return
		}

		// Get hand after drawing to identify the new card
		handCardsAfter := currentPlayer.Hand().ShowCards()
		for _, card := range handCardsAfter {
			if !handBefore[card.GetID()] {
				newCardID = card.GetID()
				break
			}
		}
		log.Printf("Newly drawn card ID: %s (hand size after: %d)", newCardID, len(handCardsAfter))
	}

	room.mutex.Unlock()

	// Send updated game state
	log.Printf("Sending game state with newCardID: %s", newCardID)
	h.sendGameState(client.GameID, newCardID)

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

	h.executeGameAction(client, func(g *domain.Game) error {
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

	h.executeGameAction(client, func(g *domain.Game) error {
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

	h.executeGameAction(client, func(g *domain.Game) error {
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

	h.executeGameAction(client, func(g *domain.Game) error {
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

	h.executeGameAction(client, func(g *domain.Game) error {
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

	h.executeGameAction(client, func(g *domain.Game) error {
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
	cards, err := room.Game.Spy(client.PlayerName, p.Option)
	room.mutex.Unlock()

	if err != nil {
		client.SendError(err.Error())
		return
	}

	// Send the revealed cards only to the client who used spy
	client.SendMessage("spy_result", map[string]interface{}{
		"cards": cards,
	})

	// Send updated game state
	h.sendGameState(client.GameID)
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

	h.executeGameAction(client, func(g *domain.Game) error {
		return g.Steal(client.PlayerName, p.CardPosition)
	})
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

	h.executeGameAction(client, func(g *domain.Game) error {
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
	if err := room.Game.EndTurn(client.PlayerName); err != nil {
		log.Printf("EndTurn error: %v", err)
		room.mutex.Unlock()
		client.SendError(err.Error())
		return
	}

	// Get the new current player and their hand before drawing
	nextPlayer, _ := room.Game.WhoIsCurrent()
	statusBefore := room.Game.GetStatusForNextPlayer()
	handBefore := make(map[string]bool)
	for _, card := range statusBefore.CurrentPlayerHand {
		handBefore[card.Card.CardID] = true
	}

	log.Printf("EndTurn: nextPlayer=%s, hand size before draw=%d", nextPlayer.Name(), len(statusBefore.CurrentPlayerHand))

	// Automatically draw a card for the new current player
	if err := room.Game.DrawCards(nextPlayer.Name(), 1); err != nil {
		log.Printf("DrawCards error: %v", err)
		room.mutex.Unlock()
		client.SendError(err.Error())
		return
	}

	// Get hand after drawing to identify the new card
	statusAfter := room.Game.GetStatusForNextPlayer()
	var newCardID string
	for _, card := range statusAfter.CurrentPlayerHand {
		if !handBefore[card.Card.CardID] {
			newCardID = card.Card.CardID
			break
		}
	}

	log.Printf("EndTurn: newCardID=%s, hand size after draw=%d", newCardID, len(statusAfter.CurrentPlayerHand))

	room.mutex.Unlock()

	// Send updated game state with newly drawn card ID
	log.Printf("Sending game state with newCardID: %s", newCardID)
	h.sendGameState(client.GameID, newCardID)

	// Check if game ended
	if room.Game.IsGameEnded() {
		h.broadcastToGame(client.GameID, MsgGameEnded, nil)
	}
}

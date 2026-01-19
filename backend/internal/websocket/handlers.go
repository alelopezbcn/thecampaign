package websocket

import (
	"encoding/json"

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

	h.executeGameAction(client, func(g *domain.Game) error {
		return g.SetInitialWarriors(client.PlayerName, p.WarriorIDs)
	})
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
		return g.Attack(client.PlayerName, p.WarriorID, p.TargetID, p.WeaponID)
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
	room, exists := h.getGameRoom(client)
	if !exists || room.Game == nil {
		client.SendError("Game not found")
		return
	}

	room.mutex.Lock()

	// End the current player's turn
	if err := room.Game.EndTurn(client.PlayerName); err != nil {
		room.mutex.Unlock()
		client.SendError(err.Error())
		return
	}

	// Get the new current player and their hand before drawing
	nextPlayer, _ := room.Game.WhoIsCurrent()
	statusBefore := room.Game.GetStatusForNextPlayer()
	handBefore := make(map[string]bool)
	for _, card := range statusBefore.CurrentPlayerHand {
		handBefore[card.GetID()] = true
	}

	// Automatically draw a card for the new current player
	if err := room.Game.DrawCards(nextPlayer.Name(), 1); err != nil {
		room.mutex.Unlock()
		client.SendError(err.Error())
		return
	}

	// Get hand after drawing to identify the new card
	statusAfter := room.Game.GetStatusForNextPlayer()
	var newCardID string
	for _, card := range statusAfter.CurrentPlayerHand {
		if !handBefore[card.GetID()] {
			newCardID = card.GetID()
			break
		}
	}

	room.mutex.Unlock()

	// Send updated game state with newly drawn card ID
	h.sendGameState(client.GameID, newCardID)

	// Check if game ended
	if room.Game.IsGameEnded() {
		h.broadcastToGame(client.GameID, MsgGameEnded, nil)
	}
}

package websocket

import (
	"encoding/json"
	"log"

	"github.com/alelopezbcn/thecampaign/internal/domain/gameactions"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
)

func (h *Hub) handleDrawCard(client *Client) {
	log.Printf("handleDrawCard called by %s", client.PlayerName)

	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewDrawCardAction(client.PlayerName))
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

	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewAttackAction(client.PlayerName, p.TargetPlayer, p.TargetID, p.WeaponID))
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

	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewSpecialPowerAction(client.PlayerName, p.UserID, p.TargetID, p.WeaponID))
	})
}

func (h *Hub) handleHarpoon(client *Client, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		client.SendError("Invalid payload")
		return
	}

	var p HarpoonPayload
	if err := json.Unmarshal(data, &p); err != nil {
		client.SendError("Invalid harpoon payload")
		return
	}

	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewHarpoonAction(client.PlayerName, p.TargetPlayer, p.TargetID, p.WeaponID))
	})
}

func (h *Hub) handleBloodRain(client *Client, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		client.SendError("Invalid payload")
		return
	}

	var p BloodRainPayload
	if err := json.Unmarshal(data, &p); err != nil {
		client.SendError("Invalid blood rain payload")
		return
	}

	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewBloodRainAction(client.PlayerName, p.TargetPlayer, p.WeaponID))
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

	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewMoveWarriorAction(client.PlayerName, p.WarriorID, p.TargetPlayer))
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

	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewTradeAction(client.PlayerName, p.CardIDs))
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

	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewBuyAction(client.PlayerName, p.CardID))
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

	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewConstructAction(client.PlayerName, p.CardID, p.TargetPlayer))
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

	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewSpyAction(client.PlayerName, p.TargetPlayer, p.Option))
	})
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

	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewStealAction(client.PlayerName, p.TargetPlayer, p.CardPosition))
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

	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewCatapultAction(client.PlayerName, p.TargetPlayer, p.CardPosition))
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
	status, err := room.Game.ExecuteAction(gameactions.NewEndTurnPhaseAction(client.PlayerName, false))
	room.mutex.Unlock()

	if err != nil {
		client.SendError(err.Error())
		return
	}

	h.sendGameStateToAll(client.GameID, status)

	// Auto draw card for the next player and broadcast state
	h.autoDrawAndBroadcast(client.GameID)
	h.startTurnTimer(client.GameID)
}

func (h *Hub) handleSkipPhase(client *Client) {
	log.Printf("handleSkipPhase called by %s", client.PlayerName)

	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewSkipPhaseAction(client.PlayerName))
	})
}

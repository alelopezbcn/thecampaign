package websocket

import (
	"encoding/json"
	"log"

	"github.com/alelopezbcn/thecampaign/internal/domain/gameactions"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
)

// parsePayload unmarshals an arbitrary WebSocket payload into T.
// On error it sends an error message to the client and returns false.
func parsePayload[T any](client *Client, payload interface{}, errMsg string) (T, bool) {
	var zero T
	data, err := json.Marshal(payload)
	if err != nil {
		client.SendError("Invalid payload")
		return zero, false
	}
	var p T
	if err := json.Unmarshal(data, &p); err != nil {
		client.SendError(errMsg)
		return zero, false
	}
	return p, true
}

func (h *Hub) handleDrawCard(client *Client) {
	log.Printf("handleDrawCard called by %s", client.PlayerName)
	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewDrawCardAction(client.PlayerName))
	})
}

func (h *Hub) handleAttack(client *Client, payload interface{}) {
	p, ok := parsePayload[AttackPayload](client, payload, "Invalid attack payload")
	if !ok {
		return
	}
	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewAttackAction(client.PlayerName, p.WarriorID, p.TargetPlayer, p.TargetID, p.WeaponID))
	})
}

func (h *Hub) handleSpecialPower(client *Client, payload interface{}) {
	p, ok := parsePayload[SpecialPowerPayload](client, payload, "Invalid special power payload")
	if !ok {
		return
	}
	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewSpecialPowerAction(client.PlayerName, p.UserID, p.TargetID, p.WeaponID))
	})
}

func (h *Hub) handleHarpoon(client *Client, payload interface{}) {
	p, ok := parsePayload[WeaponPayload](client, payload, "Invalid weapon payload")
	if !ok {
		return
	}
	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewHarpoonAction(client.PlayerName, p.TargetPlayer, p.TargetID, p.WeaponID))
	})
}

func (h *Hub) handleBloodRain(client *Client, payload interface{}) {
	p, ok := parsePayload[WeaponPayload](client, payload, "Invalid weapon payload")
	if !ok {
		return
	}
	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewBloodRainAction(client.PlayerName, p.TargetPlayer, p.WeaponID))
	})
}

func (h *Hub) handleMoveWarrior(client *Client, payload interface{}) {
	p, ok := parsePayload[MoveWarriorPayload](client, payload, "Invalid move warrior payload")
	if !ok {
		return
	}
	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewMoveWarriorAction(client.PlayerName, p.WarriorID, p.TargetPlayer))
	})
}

func (h *Hub) handleTrade(client *Client, payload interface{}) {
	p, ok := parsePayload[TradePayload](client, payload, "Invalid trade payload")
	if !ok {
		return
	}
	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewTradeAction(client.PlayerName, p.CardIDs))
	})
}

func (h *Hub) handleBuy(client *Client, payload interface{}) {
	p, ok := parsePayload[BuyPayload](client, payload, "Invalid buy payload")
	if !ok {
		return
	}
	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewBuyAction(client.PlayerName, p.CardID))
	})
}

func (h *Hub) handleForge(client *Client, payload interface{}) {
	p, ok := parsePayload[ForgePayload](client, payload, "Invalid forge payload")
	if !ok {
		return
	}
	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewForgeAction(client.PlayerName, p.CardID1, p.CardID2))
	})
}

func (h *Hub) handleBuyMercenary(client *Client, payload interface{}) {
	p, ok := parsePayload[BuyMercenaryPayload](client, payload, "Invalid buy_mercenary payload")
	if !ok {
		return
	}
	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewBuyMercenaryAction(client.PlayerName, p.CardID))
	})
}

func (h *Hub) handleConstruct(client *Client, payload interface{}) {
	p, ok := parsePayload[ConstructPayload](client, payload, "Invalid construct payload")
	if !ok {
		return
	}
	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewConstructAction(client.PlayerName, p.CardID, p.TargetPlayer))
	})
}

func (h *Hub) handleSpy(client *Client, payload interface{}) {
	p, ok := parsePayload[SpyPayload](client, payload, "Invalid spy payload")
	if !ok {
		return
	}
	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewSpyAction(client.PlayerName, p.TargetPlayer, p.Option, p.CardID))
	})
}

func (h *Hub) handleSteal(client *Client, payload interface{}) {
	p, ok := parsePayload[StealPayload](client, payload, "Invalid steal payload")
	if !ok {
		return
	}
	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewStealAction(client.PlayerName, p.TargetPlayer, p.CardPosition, p.CardID))
	})
}

func (h *Hub) handleTreason(client *Client, payload interface{}) {
	p, ok := parsePayload[TreasonPayload](client, payload, "Invalid treason payload")
	if !ok {
		return
	}
	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewTreasonAction(client.PlayerName, p.TargetPlayer, p.WarriorID, p.CardID))
	})
}

func (h *Hub) handleCatapult(client *Client, payload interface{}) {
	p, ok := parsePayload[CatapultPayload](client, payload, "Invalid catapult payload")
	if !ok {
		return
	}
	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewCatapultAction(client.PlayerName, p.TargetPlayer, p.CardPosition, p.CardID))
	})
}

func (h *Hub) handleFortress(client *Client, payload interface{}) {
	p, ok := parsePayload[FortressPayload](client, payload, "Invalid fortress payload")
	if !ok {
		return
	}
	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewFortressAction(client.PlayerName, p.TargetPlayer, p.CardID))
	})
}

func (h *Hub) handleResurrection(client *Client, payload interface{}) {
	p, ok := parsePayload[ResurrectionPayload](client, payload, "Invalid resurrection payload")
	if !ok {
		return
	}
	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewResurrectionAction(client.PlayerName, p.TargetPlayer, p.CardID))
	})
}

func (h *Hub) handleSabotage(client *Client, payload interface{}) {
	p, ok := parsePayload[SabotagePayload](client, payload, "Invalid sabotage payload")
	if !ok {
		return
	}
	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewSabotageAction(client.PlayerName, p.TargetPlayer, p.CardID))
	})
}

func (h *Hub) handlePlaceAmbush(client *Client, payload interface{}) {
	p, ok := parsePayload[PlaceAmbushPayload](client, payload, "Invalid place_ambush payload")
	if !ok {
		return
	}
	h.executeGameAction(client, func(g HubGame) (gamestatus.GameStatus, error) {
		return g.ExecuteAction(gameactions.NewPlaceAmbushAction(client.PlayerName, p.CardID))
	})
}

func (h *Hub) handleEndTurn(client *Client) {
	log.Printf("handleEndTurn called by %s", client.PlayerName)

	room, exists := h.getGameRoom(client)
	if !exists || room.Game == nil {
		client.SendError("Game not found")
		return
	}

	status, err := func() (gamestatus.GameStatus, error) {
		room.mutex.Lock()
		defer room.mutex.Unlock()
		return room.Game.ExecuteAction(gameactions.NewEndTurnPhaseAction(client.PlayerName, false))
	}()
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

package gameactions

import (
	"errors"
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// placeAmbushGame declares the minimum Game surface needed by placeAmbushAction
type placeAmbushGame interface {
	GamePlayers
	GameTurn
	GameHistory
	GameStatusProvider
}

type placeAmbushAction struct {
	playerName       string
	targetPlayerName string // "" = own field
	cardID           string

	ambushCard  cards.Ambush
	targetField board.Field
}

func NewPlaceAmbushAction(playerName, targetPlayerName, cardID string) *placeAmbushAction {
	return &placeAmbushAction{
		playerName:       playerName,
		targetPlayerName: targetPlayerName,
		cardID:           cardID,
	}
}

func (a *placeAmbushAction) PlayerName() string { return a.playerName }

func (a *placeAmbushAction) Validate(g Game) error {
	if g.CurrentAction() != types.PhaseTypeAttack {
		return fmt.Errorf("cannot place ambush in the %s phase", g.CurrentAction())
	}

	p := g.CurrentPlayer()

	c, ok := p.GetCardFromHand(a.cardID)
	if !ok {
		return fmt.Errorf("ambush card %s not found in hand", a.cardID)
	}

	ambush, ok := c.(cards.Ambush)
	if !ok {
		return errors.New("card is not an ambush card")
	}

	// Resolve target field: own field by default, or an ally's field in 2v2.
	var targetField board.Field
	if a.targetPlayerName == "" || a.targetPlayerName == p.Name() {
		targetField = p.Field()
	} else {
		targetPlayer := g.GetPlayer(a.targetPlayerName)
		if targetPlayer == nil {
			return fmt.Errorf("target player %s not found", a.targetPlayerName)
		}
		pIdx := g.PlayerIndex(a.playerName)
		tIdx := g.PlayerIndex(a.targetPlayerName)
		if !g.SameTeam(pIdx, tIdx) {
			return fmt.Errorf("cannot place ambush on %s's field: not an ally", a.targetPlayerName)
		}
		targetField = targetPlayer.Field()
	}

	if board.HasFieldSlotCard[cards.Ambush](targetField) {
		return errors.New("field already has an ambush card")
	}

	a.ambushCard = ambush
	a.targetField = targetField
	return nil
}

func (a *placeAmbushAction) Execute(g Game) (*Result, func() gamestatus.GameStatus, error) {
	return a.execute(g)
}

func (a *placeAmbushAction) execute(g placeAmbushGame) (*Result, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()

	p.RemoveFromHand(a.ambushCard.GetID())
	a.targetField.SetSlotCard(a.ambushCard)

	placerName := p.Name()
	targetPlayer := a.targetPlayerName
	if targetPlayer == "" {
		targetPlayer = placerName
	}

	if a.targetPlayerName == "" {
		g.AddHistory(fmt.Sprintf("%s placed an ambush", placerName), types.CategoryAction)
	} else {
		g.AddHistory(fmt.Sprintf("%s placed an ambush on %s's field", placerName, a.targetPlayerName), types.CategoryAction)
	}
	result := &Result{
		Action:      types.LastActionPlaceAmbush,
		PlaceAmbush: &PlaceAmbushDetails{TargetPlayer: targetPlayer},
	}
	statusFn := func() gamestatus.GameStatus {
		return g.Status(p)
	}

	return result, statusFn, nil
}

// NextPhase stays in buy phase — the player can still buy cards after placing ambush.
func (a *placeAmbushAction) NextPhase() types.PhaseType {
	return types.PhaseTypeBuy
}

// placeAmbushPlayer is a helper interface used in tests.
type placeAmbushPlayer interface {
	board.PlayerIdentity
	board.PlayerHand
	board.PlayerField
}

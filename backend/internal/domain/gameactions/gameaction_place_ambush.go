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
	playerName string
	cardID     string

	ambushCard cards.Ambush
}

func NewPlaceAmbushAction(playerName, cardID string) *placeAmbushAction {
	return &placeAmbushAction{
		playerName: playerName,
		cardID:     cardID,
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

	if board.HasFieldSlotCard[cards.Ambush](p.Field()) {
		return errors.New("field already has an ambush card")
	}

	a.ambushCard = ambush
	return nil
}

func (a *placeAmbushAction) Execute(g Game) (*Result, func() gamestatus.GameStatus, error) {
	return a.execute(g)
}

func (a *placeAmbushAction) execute(g placeAmbushGame) (*Result, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()

	p.RemoveFromHand(a.ambushCard.GetID())
	p.Field().SetSlotCard(a.ambushCard)

	g.AddHistory(fmt.Sprintf("%s placed an ambush", p.Name()), types.CategoryAction)

	result := &Result{Action: types.LastActionPlaceAmbush}
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

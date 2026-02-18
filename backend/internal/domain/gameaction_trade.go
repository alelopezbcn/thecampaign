package domain

import (
	"errors"
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type TradeAction struct {
	playerName string
	cardIDs    []string

	currentPhase types.PhaseType
}

func NewTradeAction(playerName string, cardIDs []string) *TradeAction {
	return &TradeAction{
		playerName: playerName,
		cardIDs:    cardIDs,
	}
}

func (a *TradeAction) PlayerName() string { return a.playerName }

func (a *TradeAction) Validate(g *Game) error {
	if g.turnState.HasTraded {
		return errors.New("already traded this turn")
	}

	if len(a.cardIDs) != 3 {
		return errors.New("must trade exactly 3 cards")
	}

	return nil
}

func (a *TradeAction) Execute(g *Game) (*GameActionResult, func() GameStatus, error) {
	p := g.CurrentPlayer()
	result := &GameActionResult{}

	tradedCards, err := p.GiveCards(a.cardIDs...)
	if err != nil {
		return result, nil, fmt.Errorf("giving cards for trading failed: %w", err)
	}
	for _, c := range tradedCards {
		g.OnCardMovedToPile(c)
	}

	cards, err := g.drawCards(p, 1)
	if err != nil {
		return result, nil, fmt.Errorf("drawing card for trading failed: %w", err)
	}

	p.TakeCards(cards...)

	g.addToHistory(fmt.Sprintf("%s traded 3 cards", p.Name()), types.CategoryAction)

	g.turnState.HasTraded = true
	g.turnState.CanTrade = false
	result.Action = types.LastActionTrade
	a.currentPhase = g.currentAction

	statusFn := func() GameStatus {
		return g.GameStatusProvider.Get(p, g, cards...)
	}

	return result, statusFn, nil
}

func (a *TradeAction) NextPhase() types.PhaseType {
	return a.currentPhase
}

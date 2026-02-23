package gameactions

import (
	"errors"
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// tradeGame declares the minimum Game surface needed by tradeAction
type tradeGame interface {
	GamePlayers
	GameTurn
	GameTurnFlags
	GameCards
	GameHistory
	GameStatusProvider
}

type tradeAction struct {
	playerName string
	cardIDs    []string

	currentPhase types.PhaseType
}

func NewTradeAction(playerName string, cardIDs []string) *tradeAction {
	return &tradeAction{
		playerName: playerName,
		cardIDs:    cardIDs,
	}
}

func (a *tradeAction) PlayerName() string { return a.playerName }

func (a *tradeAction) Validate(g Game) error {
	if g.TurnState().HasTraded {
		return errors.New("already traded this turn")
	}

	if len(a.cardIDs) != 3 {
		return errors.New("must trade exactly 3 cards")
	}

	return nil
}

func (a *tradeAction) Execute(g Game) (*Result, func() gamestatus.GameStatus, error) {
	return a.execute(g)
}

func (a *tradeAction) execute(g tradeGame) (*Result, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()
	result := &Result{}

	tradedCards, err := p.RemoveFromHand(a.cardIDs...)
	if err != nil {
		return result, nil, fmt.Errorf("giving cards for trading failed: %w", err)
	}
	for _, c := range tradedCards {
		g.OnCardMovedToPile(c)
	}

	cards, err := g.DrawCards(p, 1)
	if err != nil {
		return result, nil, fmt.Errorf("drawing card for trading failed: %w", err)
	}

	p.TakeCards(cards...)

	g.AddHistory(fmt.Sprintf("%s traded 3 cards", p.Name()), types.CategoryAction)

	g.SetHasTraded(true)
	g.SetCanTrade(false)
	result.Action = types.LastActionTrade
	a.currentPhase = g.CurrentAction()

	statusFn := func() gamestatus.GameStatus {
		return g.Status(p, cards...)
	}

	return result, statusFn, nil
}

func (a *tradeAction) NextPhase() types.PhaseType {
	return a.currentPhase
}

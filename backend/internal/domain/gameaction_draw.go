package domain

import (
	"errors"
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type DrawCardAction struct {
	playerName string
}

func NewDrawCardAction(playerName string) *DrawCardAction {
	return &DrawCardAction{playerName: playerName}
}

func (a *DrawCardAction) PlayerName() string { return a.playerName }

func (a *DrawCardAction) Validate(g *Game) error {
	return nil
}

func (a *DrawCardAction) Execute(g *Game) (*GameActionResult, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()

	cards, err := g.drawCards(p, 1)
	if err != nil {
		if errors.Is(err, ErrHandLimitExceeded) {
			g.addToHistory(fmt.Sprintf("%s can't take more cards (hand limit reached)", p.Name()),
				types.CategoryError)

			result := &GameActionResult{}
			statusFn := func() gamestatus.GameStatus {
				return g.GameStatusProvider.Get(p, g)
			}
			return result, statusFn, nil
		}

		return nil, nil, fmt.Errorf("drawing card failed: %w", err)
	}

	p.TakeCards(cards...)

	g.addToHistory(fmt.Sprintf("%s drew a card", p.Name()), types.CategoryAction)

	result := &GameActionResult{Action: types.LastActionDraw}
	statusFn := func() gamestatus.GameStatus {
		return g.GameStatusProvider.Get(p, g, cards...)
	}

	return result, statusFn, nil
}

func (a *DrawCardAction) NextPhase() types.PhaseType {
	return types.PhaseTypeAttack
}

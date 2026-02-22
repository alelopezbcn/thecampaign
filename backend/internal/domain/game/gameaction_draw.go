package game

import (
	"errors"
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type drawCardAction struct {
	playerName string
}

func NewDrawCardAction(playerName string) *drawCardAction {
	return &drawCardAction{playerName: playerName}
}

func (a *drawCardAction) PlayerName() string { return a.playerName }

func (a *drawCardAction) Validate(g *game) error {
	return nil
}

func (a *drawCardAction) Execute(g *game) (*GameActionResult, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()

	cards, err := g.drawCards(p, 1)
	if err != nil {
		if errors.Is(err, board.ErrHandLimitExceeded) {
			g.AddHistory(fmt.Sprintf("%s can't take more cards (hand limit reached)", p.Name()),
				types.CategoryError)

			result := &GameActionResult{}
			statusFn := func() gamestatus.GameStatus {
				return g.gameStatusProvider.Get(p, g)
			}
			return result, statusFn, nil
		}

		return nil, nil, fmt.Errorf("drawing card failed: %w", err)
	}

	p.TakeCards(cards...)

	g.AddHistory(fmt.Sprintf("%s drew a card", p.Name()), types.CategoryAction)

	result := &GameActionResult{Action: types.LastActionDraw}
	statusFn := func() gamestatus.GameStatus {
		return g.gameStatusProvider.Get(p, g, cards...)
	}

	return result, statusFn, nil
}

func (a *drawCardAction) NextPhase() types.PhaseType {
	return types.PhaseTypeAttack
}

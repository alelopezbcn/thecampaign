package gameactions

import (
	"errors"
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// drawGame declares the minimum Game surface needed by drawCardAction
type drawGame interface {
	GamePlayers
	GameCards
	GameHistory
	GameStatusProvider
}

type drawCardAction struct {
	playerName string
}

func NewDrawCardAction(playerName string) *drawCardAction {
	return &drawCardAction{playerName: playerName}
}

func (a *drawCardAction) PlayerName() string { return a.playerName }

func (a *drawCardAction) Validate(g Game) error {
	return nil
}

func (a *drawCardAction) Execute(g Game) (*Result, func() gamestatus.GameStatus, error) {
	return a.execute(g)
}

func (a *drawCardAction) execute(g drawGame) (*Result, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()

	cards, err := g.DrawCards(p, 1)
	if err != nil {
		if errors.Is(err, board.ErrHandLimitExceeded) {
			g.AddHistory(fmt.Sprintf("%s can't take more cards (hand limit reached)", p.Name()),
				types.CategoryError)

			result := &Result{}
			statusFn := func() gamestatus.GameStatus {
				return g.Status(p)
			}
			return result, statusFn, nil
		}

		return nil, nil, fmt.Errorf("drawing card failed: %w", err)
	}

	p.TakeCards(cards...)

	g.AddHistory(fmt.Sprintf("%s drew a card", p.Name()), types.CategoryAction)

	result := &Result{Action: types.LastActionDraw}
	statusFn := func() gamestatus.GameStatus {
		return g.Status(p, cards...)
	}

	return result, statusFn, nil
}

func (a *drawCardAction) NextPhase() types.PhaseType {
	return types.PhaseTypeAttack
}

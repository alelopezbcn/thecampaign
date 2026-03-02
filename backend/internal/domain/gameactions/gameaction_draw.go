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
	GameTurn
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
	handler := g.EventHandler()

	// Apply Plague: HP modifier to active player's warriors at turn start (never kills).
	if mod := handler.TurnStartWarriorHPModifier(); mod != 0 {
		for _, warrior := range p.Field().Warriors() {
			safeAmount := mod
			if mod < 0 {
				h := warrior.Health()
				if h+mod < 1 {
					safeAmount = 1 - h
				}
			}
			if safeAmount != 0 {
				warrior.HealBy(safeAmount)
			}
		}
		_, desc := handler.Display()
		g.AddHistory(desc, types.CategoryInfo)
	}

	drawCount := 1 + handler.ExtraDrawCards()

	drawnCards, err := g.DrawCards(p, drawCount)
	if err != nil {
		if errors.Is(err, board.ErrHandLimitExceeded) {
			// If the bonus card doesn't fit, fall back to the standard 1 card
			if drawCount > 1 {
				drawnCards, err = g.DrawCards(p, 1)
			}
			if err != nil {
				g.AddHistory(fmt.Sprintf("%s can't take more cards (hand limit reached)", p.Name()),
					types.CategoryError)
				result := &Result{}
				statusFn := func() gamestatus.GameStatus {
					return g.Status(p)
				}
				return result, statusFn, nil
			}
		} else {
			return nil, nil, fmt.Errorf("drawing card failed: %w", err)
		}
	}

	p.TakeCards(drawnCards...)

	if len(drawnCards) > 1 {
		g.AddHistory(fmt.Sprintf("%s drew %d cards (Abundance)", p.Name(), len(drawnCards)), types.CategoryAction)
	} else {
		g.AddHistory(fmt.Sprintf("%s drew a card", p.Name()), types.CategoryAction)
	}

	result := &Result{Action: types.LastActionDraw}
	statusFn := func() gamestatus.GameStatus {
		return g.Status(p, drawnCards...)
	}

	return result, statusFn, nil
}

func (a *drawCardAction) NextPhase() types.PhaseType {
	return types.PhaseTypeAttack
}

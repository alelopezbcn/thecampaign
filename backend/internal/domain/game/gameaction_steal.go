package game

import (
	"errors"
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type StealAction struct {
	playerName       string
	targetPlayerName string
	cardPosition     int

	targetPlayer board.Player
}

func NewStealAction(playerName, targetPlayerName string, cardPosition int) *StealAction {
	return &StealAction{
		playerName:       playerName,
		targetPlayerName: targetPlayerName,
		cardPosition:     cardPosition,
	}
}

func (a *StealAction) PlayerName() string { return a.playerName }

func (a *StealAction) Validate(g *Game) error {
	if g.currentAction != types.PhaseTypeSpySteal {
		return fmt.Errorf("cannot steal in the %s phase", g.currentAction)
	}

	p := g.CurrentPlayer()
	if !p.HasThief() {
		return fmt.Errorf("player does not have a thief to use")
	}

	var err error
	a.targetPlayer, err = g.getTargetPlayer(a.playerName, a.targetPlayerName)
	if err != nil {
		return err
	}

	return nil
}

func (a *StealAction) Execute(g *Game) (*GameActionResult, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()

	result := &GameActionResult{}

	stolenCard, err := a.targetPlayer.CardStolenFromHand(a.cardPosition)
	if err != nil {
		return result, nil, fmt.Errorf("stealing card failed: %w", err)
	}

	t := p.Thief()
	if t == nil {
		return result, nil, errors.New("failed to retrieve thief card")
	}

	g.OnCardMovedToPile(t)
	p.TakeCards(stolenCard)

	result.StolenFrom = a.targetPlayer.Name()
	result.StolenCard = stolenCard
	result.Action = types.LastActionSteal

	g.AddHistory(fmt.Sprintf("%s stole a card from %s",
		p.Name(), a.targetPlayer.Name()), types.CategoryAction)

	statusFn := func() gamestatus.GameStatus {
		return g.gameStatusProvider.GetWithModal(p, g, []cards.Card{stolenCard})
	}

	return result, statusFn, nil
}

func (a *StealAction) NextPhase() types.PhaseType {
	return types.PhaseTypeBuy
}

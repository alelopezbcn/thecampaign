package game

import (
	"errors"
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type spyAction struct {
	playerName       string
	targetPlayerName string
	option           int

	spy cards.Spy
}

func NewSpyAction(playerName, targetPlayerName string, option int) *spyAction {
	return &spyAction{
		playerName:       playerName,
		targetPlayerName: targetPlayerName,
		option:           option,
	}
}

func (a *spyAction) PlayerName() string { return a.playerName }

func (a *spyAction) Validate(g Game) error {
	if g.CurrentAction() != types.PhaseTypeSpySteal {
		return fmt.Errorf("cannot use spy in the %s phase", g.CurrentAction())
	}

	p := g.CurrentPlayer()
	spy, ok := board.HasCardTypeInHand[cards.Spy](p)
	if !ok {
		return fmt.Errorf("player does not have a spy to use")
	}

	a.spy = spy

	return nil
}

func (a *spyAction) Execute(g Game) (*GameActionResult, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()

	var spiedCards []cards.Card
	result := &GameActionResult{}

	switch a.option {
	case 1:
		// Reveal top 5 cards from deck
		g.AddHistory(fmt.Sprintf("%s spied top 5 cards from deck", p.Name()),
			types.CategoryAction)

		result.Spy = types.SpyInfo{Target: types.SpyTargetDeck}
		spiedCards = g.Board().Deck().Reveal(5)
	case 2:
		// Reveal target's cards
		targetPlayer, err := g.GetTargetPlayer(p.Name(), a.targetPlayerName)
		if err != nil {
			return result, nil, err
		}

		g.AddHistory(fmt.Sprintf("%s spied on %s's hand",
			p.Name(), targetPlayer.Name()), types.CategoryAction)

		result.Spy = types.SpyInfo{Target: types.SpyTargetPlayer, TargetPlayer: targetPlayer.Name()}
		spiedCards = targetPlayer.Hand().ShowCards()
	default:
		return result, nil, errors.New("invalid Spy option")
	}

	spy, err := p.RemoveFromHand(a.spy.GetID())
	if err != nil {
		return result, nil, fmt.Errorf("removing spy from hand failed: %w", err)
	}

	g.OnCardMovedToPile(spy[0])

	result.Action = types.LastActionSpy
	statusFn := func() gamestatus.GameStatus {
		return g.GameStatusProvider().GetWithModal(p, g, spiedCards)
	}

	return result, statusFn, nil
}

func (a *spyAction) NextPhase() types.PhaseType {
	return types.PhaseTypeBuy
}

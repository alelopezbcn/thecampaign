package game

import (
	"errors"
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type SpyAction struct {
	playerName       string
	targetPlayerName string
	option           int
}

func NewSpyAction(playerName, targetPlayerName string, option int) *SpyAction {
	return &SpyAction{
		playerName:       playerName,
		targetPlayerName: targetPlayerName,
		option:           option,
	}
}

func (a *SpyAction) PlayerName() string { return a.playerName }

func (a *SpyAction) Validate(g *Game) error {
	if g.currentAction != types.PhaseTypeSpySteal {
		return fmt.Errorf("cannot use spy in the %s phase", g.currentAction)
	}

	p := g.CurrentPlayer()
	if !p.HasSpy() {
		return fmt.Errorf("player does not have a spy to use")
	}

	return nil
}

func (a *SpyAction) Execute(g *Game) (*GameActionResult, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()

	var spiedCards []ports.Card
	result := &GameActionResult{}

	switch a.option {
	case 1:
		// Reveal top 5 cards from deck
		g.AddHistory(fmt.Sprintf("%s spied top 5 cards from deck", p.Name()),
			types.CategoryAction)

		result.Spy = types.SpyInfo{Target: types.SpyTargetDeck}
		spiedCards = g.board.Deck().Reveal(5)
	case 2:
		// Reveal target's cards
		targetPlayer, err := g.getTargetPlayer(p.Name(), a.targetPlayerName)
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

	s := p.Spy()
	if s == nil {
		return &GameActionResult{}, nil, errors.New("failed to retrieve spy card")
	}

	g.OnCardMovedToPile(s)

	result.Action = types.LastActionSpy
	statusFn := func() gamestatus.GameStatus {
		return g.gameStatusProvider.GetWithModal(p, g, spiedCards)
	}

	return result, statusFn, nil
}

func (a *SpyAction) NextPhase() types.PhaseType {
	return types.PhaseTypeBuy
}

package game

import (
	"fmt"
	"math/rand"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type stealAction struct {
	playerName       string
	targetPlayerName string
	cardPosition     int

	targetPlayer          board.Player
	thief                 cards.Thief
	targetPlayerHandCards []cards.Card
}

func NewStealAction(playerName, targetPlayerName string, cardPosition int) *stealAction {
	return &stealAction{
		playerName:       playerName,
		targetPlayerName: targetPlayerName,
		cardPosition:     cardPosition,
	}
}

func (a *stealAction) PlayerName() string { return a.playerName }

func (a *stealAction) Validate(g Game) error {
	if g.CurrentAction() != types.PhaseTypeSpySteal {
		return fmt.Errorf("cannot steal in the %s phase", g.CurrentAction())
	}

	p := g.CurrentPlayer()
	thief, ok := board.HasCardTypeInHand[cards.Thief](p)
	if !ok {
		return fmt.Errorf("player does not have a thief to use")
	}

	var err error
	a.targetPlayer, err = g.GetTargetPlayer(a.playerName, a.targetPlayerName)
	if err != nil {
		return err
	}

	a.targetPlayerHandCards = a.targetPlayer.Hand().ShowCards()
	if a.cardPosition < 1 || a.cardPosition > len(a.targetPlayerHandCards) {
		return fmt.Errorf("invalid position %d for stealing cardBase", a.cardPosition)
	}

	a.thief = thief

	return nil
}

func (a *stealAction) Execute(g Game) (*GameActionResult, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()

	result := &GameActionResult{}
	stolenCard := a.steal()

	thief, err := p.RemoveFromHand(a.thief.GetID())
	if err != nil {
		return result, nil, fmt.Errorf("removing thief from hand failed: %w", err)
	}

	g.OnCardMovedToPile(thief[0])
	p.TakeCards(stolenCard)

	result.StolenFrom = a.targetPlayer.Name()
	result.StolenCard = stolenCard
	result.Action = types.LastActionSteal

	g.AddHistory(fmt.Sprintf("%s stole a card from %s",
		p.Name(), a.targetPlayer.Name()), types.CategoryAction)

	statusFn := func() gamestatus.GameStatus {
		return g.GameStatusProvider().GetWithModal(p, g, []cards.Card{stolenCard})
	}

	return result, statusFn, nil
}

func (a *stealAction) NextPhase() types.PhaseType {
	return types.PhaseTypeBuy
}

func (a *stealAction) steal() cards.Card {
	copied := make([]cards.Card, len(a.targetPlayerHandCards))
	copy(copied, a.targetPlayerHandCards)
	// Shuffle copied slice
	for i := range copied {
		j := i + rand.Intn(len(copied)-i)
		copied[i], copied[j] = copied[j], copied[i]
	}

	c := copied[a.cardPosition-1]
	a.targetPlayer.Hand().RemoveCard(c)

	return c
}

package gameactions

import (
	"fmt"
	"math/rand"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// sabotageGame declares the minimum Game surface needed by sabotageAction
type sabotageGame interface {
	GamePlayers
	GameTurn
	GameCards
	GameHistory
	GameStatusProvider
}

// sabotageTargetPlayer declares the minimum Player surface needed by sabotageAction
type sabotageTargetPlayer interface {
	board.PlayerIdentity
	board.PlayerHand
}

type sabotageAction struct {
	playerName       string
	targetPlayerName string
	cardID           string

	targetPlayer          sabotageTargetPlayer
	sabotageCard          cards.Sabotage
	targetPlayerHandCards []cards.Card
}

func NewSabotageAction(playerName, targetPlayerName, cardID string) *sabotageAction {
	return &sabotageAction{
		playerName:       playerName,
		targetPlayerName: targetPlayerName,
		cardID:           cardID,
	}
}

func (a *sabotageAction) PlayerName() string { return a.playerName }

func (a *sabotageAction) Validate(g Game) error {
	if g.CurrentAction() != types.PhaseTypeSpySteal {
		return fmt.Errorf("cannot use sabotage in the %s phase", g.CurrentAction())
	}

	p := g.CurrentPlayer()
	raw, ok := p.GetCardFromHand(a.cardID)
	if !ok {
		return fmt.Errorf("card %s not found in hand", a.cardID)
	}
	sabCard, ok := raw.(cards.Sabotage)
	if !ok {
		return fmt.Errorf("card %s is not a sabotage card", a.cardID)
	}

	var err error
	a.targetPlayer, err = g.GetTargetPlayer(a.playerName, a.targetPlayerName)
	if err != nil {
		return err
	}

	a.targetPlayerHandCards = a.targetPlayer.Hand().ShowCards()
	if len(a.targetPlayerHandCards) == 0 {
		return fmt.Errorf("target player has no cards to destroy")
	}

	a.sabotageCard = sabCard

	return nil
}

func (a *sabotageAction) Execute(g Game) (*Result, func() gamestatus.GameStatus, error) {
	return a.execute(g)
}

func (a *sabotageAction) execute(g sabotageGame) (*Result, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()

	result := &Result{}
	destroyedCard := a.destroyRandomCard()

	g.OnCardMovedToPile(destroyedCard)

	sabCard, err := p.RemoveFromHand(a.sabotageCard.GetID())
	if err != nil {
		return result, nil, fmt.Errorf("removing sabotage card from hand failed: %w", err)
	}

	g.OnCardMovedToPile(sabCard[0])

	targetName := a.targetPlayer.Name()
	result.SabotagedFrom = targetName
	result.SabotagedCard = destroyedCard
	result.Action = types.LastActionSabotage

	g.AddHistory(fmt.Sprintf("%s destroyed a card from %s's hand",
		p.Name(), targetName), types.CategoryAction)

	statusFn := func() gamestatus.GameStatus {
		return g.StatusWithModal(p, []cards.Card{destroyedCard})
	}

	return result, statusFn, nil
}

func (a *sabotageAction) NextPhase() types.PhaseType {
	return types.PhaseTypeBuy
}

func (a *sabotageAction) destroyRandomCard() cards.Card {
	copied := make([]cards.Card, len(a.targetPlayerHandCards))
	copy(copied, a.targetPlayerHandCards)
	for i := range copied {
		j := i + rand.Intn(len(copied)-i)
		copied[i], copied[j] = copied[j], copied[i]
	}

	c := copied[0]
	a.targetPlayer.Hand().RemoveCard(c)

	return c
}

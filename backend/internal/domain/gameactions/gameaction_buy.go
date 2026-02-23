package gameactions

import (
	"errors"
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// buyGame declares the minimum Game surface needed by buyAction
type buyGame interface {
	GamePlayers
	GameTurn
	GameCards
	GameHistory
	GameStatusProvider
}

type buyAction struct {
	playerName string
	cardID     string

	resource cards.Resource
}

func NewBuyAction(playerName, cardID string) *buyAction {
	return &buyAction{
		playerName: playerName,
		cardID:     cardID,
	}
}

func (a *buyAction) PlayerName() string { return a.playerName }

func (a *buyAction) Validate(g Game) error {
	if g.CurrentAction() != types.PhaseTypeBuy {
		return fmt.Errorf("cannot buy in the %s phase", g.CurrentAction())
	}

	p := g.CurrentPlayer()
	resourceCard, ok := p.GetCardFromHand(a.cardID)
	if !ok {
		return fmt.Errorf("Resource card not in hand: %s", a.cardID)
	}

	a.resource, ok = resourceCard.(cards.Resource)
	if !ok {
		return errors.New("only gold cards can be used to buy")
	}

	return nil
}

func (a *buyAction) Execute(g Game) (*Result, func() gamestatus.GameStatus, error) {
	return a.execute(g)
}

func (a *buyAction) execute(g buyGame) (*Result, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()

	result := &Result{}

	if _, err := p.RemoveFromHand(a.resource.GetID()); err != nil {
		return result, nil, fmt.Errorf("giving card for buying failed: %w", err)
	}

	cardsToBuy := a.resource.Value() / 2
	cards, err := g.DrawCards(p, cardsToBuy)
	if err != nil {
		p.TakeCards(a.resource) // Return the resource card to player's hand if drawing fails
		if errors.Is(err, board.ErrHandLimitExceeded) {
			return result, nil, fmt.Errorf("cards in hand limit exceeded")
		}
		return result, nil, fmt.Errorf("drawing card for buying failed: %w", err)
	}

	p.TakeCards(cards...)

	g.OnCardMovedToPile(a.resource)

	g.AddHistory(fmt.Sprintf("%s bought %d card(s)", p.Name(), cardsToBuy),
		types.CategoryAction)

	result.Action = types.LastActionBuy

	statusFn := func() gamestatus.GameStatus {
		return g.Status(p, cards...)
	}

	return result, statusFn, nil
}

func (a *buyAction) NextPhase() types.PhaseType {
	return types.PhaseTypeConstruct
}

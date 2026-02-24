package gameactions

import (
	"errors"
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/google/uuid"
)

type buyMercenaryAction struct {
	playerName string
	cardID     string

	resource cards.Resource
}

func NewBuyMercenaryAction(playerName, cardID string) *buyMercenaryAction {
	return &buyMercenaryAction{
		playerName: playerName,
		cardID:     cardID,
	}
}

func (a *buyMercenaryAction) PlayerName() string { return a.playerName }

func (a *buyMercenaryAction) Validate(g Game) error {
	if g.CurrentAction() != types.PhaseTypeBuy {
		return fmt.Errorf("cannot buy mercenary in the %s phase", g.CurrentAction())
	}

	p := g.CurrentPlayer()
	resourceCard, ok := p.GetCardFromHand(a.cardID)
	if !ok {
		return fmt.Errorf("resource card not in hand: %s", a.cardID)
	}

	a.resource, ok = resourceCard.(cards.Resource)
	if !ok {
		return errors.New("only gold cards can be used to hire a mercenary")
	}

	if a.resource.Value() < 6 {
		return errors.New("need at least 6 gold to hire a mercenary")
	}

	return nil
}

func (a *buyMercenaryAction) Execute(g Game) (*Result, func() gamestatus.GameStatus, error) {
	return a.execute(g)
}

func (a *buyMercenaryAction) execute(g buyGame) (*Result, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()
	result := &Result{}

	if _, err := p.RemoveFromHand(a.resource.GetID()); err != nil {
		return result, nil, fmt.Errorf("removing gold for mercenary hire failed: %w", err)
	}

	merc := cards.NewMercenary(uuid.NewString())
	p.TakeCards(merc)

	if err := p.MoveCardToField(merc.GetID()); err != nil {
		p.TakeCards(a.resource) // Return resource on failure
		return result, nil, fmt.Errorf("placing mercenary on field failed: %w", err)
	}

	g.OnCardMovedToPile(a.resource)
	g.AddHistory(fmt.Sprintf("%s hired a Mercenary", p.Name()), types.CategoryAction)

	result.Action = types.LastActionBuyMercenary

	statusFn := func() gamestatus.GameStatus {
		return g.Status(p)
	}

	return result, statusFn, nil
}

func (a *buyMercenaryAction) NextPhase() types.PhaseType {
	return types.PhaseTypeConstruct
}

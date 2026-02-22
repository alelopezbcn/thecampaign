package game

import (
	"errors"
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type constructAction struct {
	playerName       string
	cardID           string
	targetPlayerName string

	targetPlayer board.Player
	resourceCard cards.Card
}

func NewConstructAction(playerName, cardID string, targetPlayerName string) *constructAction {
	return &constructAction{
		playerName:       playerName,
		cardID:           cardID,
		targetPlayerName: targetPlayerName,
	}
}

func (a *constructAction) PlayerName() string { return a.playerName }

func (a *constructAction) Validate(g Game) error {
	if g.CurrentAction() != types.PhaseTypeConstruct {
		return fmt.Errorf("cannot construct in the %s phase", g.CurrentAction())
	}

	// Ally castle construction (2v2 mode)
	if a.targetPlayerName != "" && a.targetPlayerName != a.playerName {
		targetPlayer := g.GetPlayer(a.targetPlayerName)
		if targetPlayer == nil {
			return fmt.Errorf("target player %s not found", a.targetPlayerName)
		}
		a.targetPlayer = targetPlayer

		pIdx := g.PlayerIndex(a.playerName)
		tIdx := g.PlayerIndex(a.targetPlayerName)
		if !g.SameTeam(pIdx, tIdx) {
			return errors.New("can only construct on ally's castle")
		}
	}

	p := g.CurrentPlayer()
	resourceCard, ok := p.GetCardFromHand(a.cardID)
	if !ok {
		return fmt.Errorf("resource card not in hand: %s", a.cardID)
	}

	a.resourceCard = resourceCard

	return nil
}

func (a *constructAction) Execute(g Game) (*GameActionResult, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()
	result := &GameActionResult{}

	if a.targetPlayer != nil {
		// Ally castle construction
		if err := a.targetPlayer.Castle().Construct(a.resourceCard); err != nil {
			return result, nil, fmt.Errorf("constructing on ally castle failed: %w", err)
		}

		g.AddHistory(fmt.Sprintf("%s added gold to %s's castle", p.Name(),
			a.targetPlayer.Name()), types.CategoryAction)
	} else {
		// Own castle construction
		if err := p.Castle().Construct(a.resourceCard); err != nil {
			return result, nil, fmt.Errorf("constructing castle failed: %w", err)
		}

		g.AddHistory(fmt.Sprintf("%s constructed his castle", p.Name()), types.CategoryAction)
	}

	p.RemoveFromHand(a.resourceCard.GetID())

	result.Action = types.LastActionConstruct
	statusFn := func() gamestatus.GameStatus {
		return g.GameStatusProvider().Get(p, g)
	}

	return result, statusFn, nil
}

func (a *constructAction) NextPhase() types.PhaseType {
	return types.PhaseTypeEndTurn
}

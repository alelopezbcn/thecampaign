package game

import (
	"errors"
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type ConstructAction struct {
	playerName       string
	cardID           string
	targetPlayerName string

	targetPlayer ports.Player
}

func NewConstructAction(playerName, cardID string, targetPlayerName string) *ConstructAction {
	return &ConstructAction{
		playerName:       playerName,
		cardID:           cardID,
		targetPlayerName: targetPlayerName,
	}
}

func (a *ConstructAction) PlayerName() string { return a.playerName }

func (a *ConstructAction) Validate(g *Game) error {
	if g.currentAction != types.PhaseTypeConstruct {
		return fmt.Errorf("cannot construct in the %s phase", g.currentAction)
	}

	// Ally castle construction (2v2 mode)
	if a.targetPlayerName != "" && a.targetPlayerName != a.playerName {
		targetPlayer := g.GetPlayer(a.targetPlayerName)
		if targetPlayer == nil {
			return fmt.Errorf("target player %s not found", a.targetPlayerName)
		}

		pIdx := g.PlayerIndex(a.playerName)
		tIdx := g.PlayerIndex(a.targetPlayerName)
		if !g.SameTeam(pIdx, tIdx) {
			return errors.New("can only construct on ally's castle")
		}

		p := g.CurrentPlayer()
		_, ok := p.GetCardFromHand(a.cardID)
		if !ok {
			return errors.New("card not in hand: " + a.cardID)
		}

		a.targetPlayer = targetPlayer
	}

	return nil
}

func (a *ConstructAction) Execute(g *Game) (*GameActionResult, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()
	result := &GameActionResult{}

	if a.targetPlayer != nil {
		// Ally castle construction
		resourceCard, _ := p.GetCardFromHand(a.cardID)

		if err := a.targetPlayer.Castle().Construct(resourceCard); err != nil {
			return result, nil, fmt.Errorf("constructing on ally castle failed: %w", err)
		}

		p.Hand().RemoveCard(resourceCard)

		g.addToHistory(fmt.Sprintf("%s added gold to %s's castle", p.Name(),
			a.targetPlayer.Name()), types.CategoryAction)
	} else {
		// Own castle construction
		if err := p.Construct(a.cardID); err != nil {
			return result, nil, fmt.Errorf("constructing card failed: %w", err)
		}

		g.addToHistory(fmt.Sprintf("%s constructed his castle", p.Name()), types.CategoryAction)
	}

	result.Action = types.LastActionConstruct
	statusFn := func() gamestatus.GameStatus {
		return g.GameStatusProvider.Get(p, g)
	}

	return result, statusFn, nil
}

func (a *ConstructAction) NextPhase() types.PhaseType {
	return types.PhaseTypeEndTurn
}

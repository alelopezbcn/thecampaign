package gameactions

import (
	"errors"
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// constructGame declares the minimum Game surface needed by constructAction
type constructGame interface {
	GamePlayers
	GameTurn
	GameHistory
	GameStatusProvider
}

// constructTargetPlayer declares the minimum Player surface needed for ally castle construction
type constructTargetPlayer interface {
	board.PlayerIdentity
	board.PlayerCastle
}

// harvestModifiedResource wraps a Resource and overrides Value with the Harvest-adjusted amount.
type harvestModifiedResource struct {
	cards.Resource
	effectiveValue int
}

func (r *harvestModifiedResource) Value() int { return r.effectiveValue }

// applyHarvestModifier returns card unchanged when mod is 0 or the castle is not yet built
// (initial construction validates the raw card value). Otherwise wraps the resource so that
// its effective contribution to the castle total reflects the Harvest modifier (minimum 1).
func applyHarvestModifier(card cards.Card, mod int, castle board.CastleReader) cards.Card {
	if mod == 0 {
		return card
	}
	res, ok := card.(cards.Resource)
	if !ok || !castle.IsConstructed() {
		return card
	}
	effective := res.Value() + mod
	if effective < 1 {
		effective = 1
	}
	return &harvestModifiedResource{Resource: res, effectiveValue: effective}
}

type constructAction struct {
	playerName       string
	cardID           string
	targetPlayerName string

	targetPlayer constructTargetPlayer
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

func (a *constructAction) Execute(g Game) (*Result, func() gamestatus.GameStatus, error) {
	return a.execute(g)
}

func (a *constructAction) execute(g constructGame) (*Result, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()
	result := &Result{}

	mod := g.EventHandler().ConstructionValueModifier()

	if a.targetPlayer != nil {
		// Ally castle construction
		targetCastle := a.targetPlayer.Castle()
		card := applyHarvestModifier(a.resourceCard, mod, targetCastle)
		if err := targetCastle.Construct(card); err != nil {
			return result, nil, fmt.Errorf("constructing on ally castle failed: %w", err)
		}

		g.AddHistory(fmt.Sprintf("%s added gold to %s's castle", p.Name(),
			a.targetPlayer.Name()), types.CategoryAction)
	} else {
		// Own castle construction
		myCastle := p.Castle()
		card := applyHarvestModifier(a.resourceCard, mod, myCastle)
		if err := myCastle.Construct(card); err != nil {
			return result, nil, fmt.Errorf("constructing castle failed: %w", err)
		}

		g.AddHistory(fmt.Sprintf("%s constructed his castle", p.Name()), types.CategoryAction)
	}

	p.RemoveFromHand(a.resourceCard.GetID())

	result.Action = types.LastActionConstruct
	statusFn := func() gamestatus.GameStatus {
		return g.Status(p)
	}

	return result, statusFn, nil
}

func (a *constructAction) NextPhase() types.PhaseType {
	return types.PhaseTypeEndTurn
}

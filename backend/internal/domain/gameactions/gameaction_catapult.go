package gameactions

import (
	"errors"
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// catapultGame declares the minimum Game surface needed by catapultAction
type catapultGame interface {
	GamePlayers
	GameTurn
	GameCards
	GameHistory
	GameStatusProvider
}

// catapultTargetPlayer declares the minimum Player surface needed by catapultAction
type catapultTargetPlayer interface {
	board.PlayerIdentity
	board.PlayerCastle
}

type catapultAction struct {
	playerName       string
	targetPlayerName string
	cardPosition     int
	cardID           string

	catapult     cards.Catapult
	targetPlayer catapultTargetPlayer
	weapon       cards.Weapon
}

func NewCatapultAction(playerName, targetPlayerName string, cardPosition int, cardID string) *catapultAction {
	return &catapultAction{
		playerName:       playerName,
		targetPlayerName: targetPlayerName,
		cardPosition:     cardPosition,
		cardID:           cardID,
	}
}

func (a *catapultAction) PlayerName() string { return a.playerName }

func (a *catapultAction) Validate(g Game) error {
	if g.CurrentAction() != types.PhaseTypeAttack {
		return fmt.Errorf("cannot use catapult in the %s phase",
			g.CurrentAction())
	}

	p := g.CurrentPlayer()
	raw, ok := p.GetCardFromHand(a.cardID)
	if !ok {
		return fmt.Errorf("card %s not found in hand", a.cardID)
	}
	catapult, ok := raw.(cards.Catapult)
	if !ok {
		return errors.New("card is not a catapult")
	}

	var err error
	a.targetPlayer, err = g.GetTargetPlayer(a.playerName, a.targetPlayerName)
	if err != nil {
		return err
	}

	a.catapult = catapult

	return nil
}

func (a *catapultAction) Execute(g Game) (*Result, func() gamestatus.GameStatus, error) {
	return a.execute(g)
}

func (a *catapultAction) execute(g catapultGame) (*Result, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()

	// Check if castle is protected by a fortress — destroy it instead of removing gold
	if a.targetPlayer.Castle().IsProtected() {
		fortressCard := a.targetPlayer.Castle().ConsumeProtection()
		g.OnCardMovedToPile(fortressCard)

		if _, err := p.RemoveFromHand(a.catapult.GetID()); err != nil {
			return &Result{}, nil, fmt.Errorf("removing catapult from hand failed: %w", err)
		}
		g.OnCardMovedToPile(a.catapult)

		g.AddHistory(fmt.Sprintf("%s's castle wall blocked %s's catapult attack",
			a.targetPlayer.Name(), p.Name()),
			types.CategoryAction)
		result := &Result{
			Action: types.LastActionCatapultBlocked,
			Catapult: &CatapultDetails{
				AttackerName: p.Name(),
				TargetPlayer: a.targetPlayer.Name(),
				Blocked:      true,
			},
		}
		statusFn := func() gamestatus.GameStatus {
			return g.Status(p)
		}
		return result, statusFn, nil
	}

	stolenGold, err := a.catapult.Attack(a.targetPlayer.Castle(), a.cardPosition)
	if err != nil {
		result := &Result{}
		return result, nil, fmt.Errorf("attacking castle failed: %w", err)
	}

	g.OnCardMovedToPile(stolenGold)

	if _, err := p.RemoveFromHand(a.catapult.GetID()); err != nil {
		return &Result{}, nil, fmt.Errorf("removing catapult from hand failed: %w", err)
	}
	g.OnCardMovedToPile(a.catapult)

	g.AddHistory(fmt.Sprintf("%s removed %d gold from %s's castle",
		p.Name(), stolenGold.Value(), a.targetPlayer.Name()),
		types.CategoryAction)

	result := &Result{
		Action: types.LastActionCatapult,
		Catapult: &CatapultDetails{
			AttackerName: p.Name(),
			TargetPlayer: a.targetPlayer.Name(),
			GoldStolen:   stolenGold.Value(),
		},
	}
	statusFn := func() gamestatus.GameStatus {
		return g.Status(p)
	}

	return result, statusFn, nil
}

func (a *catapultAction) NextPhase() types.PhaseType {
	return types.PhaseTypeSpySteal
}

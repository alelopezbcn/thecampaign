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

type catapultAction struct {
	playerName       string
	targetPlayerName string
	cardPosition     int

	catapult     cards.Catapult
	targetPlayer board.Player
	weapon       cards.Weapon
}

func NewCatapultAction(playerName, targetPlayerName string, cardPosition int) *catapultAction {
	return &catapultAction{
		playerName:       playerName,
		targetPlayerName: targetPlayerName,
		cardPosition:     cardPosition,
	}
}

func (a *catapultAction) PlayerName() string { return a.playerName }

func (a *catapultAction) Validate(g Game) error {
	if g.CurrentAction() != types.PhaseTypeAttack {
		return fmt.Errorf("cannot use catapult in the %s phase",
			g.CurrentAction())
	}

	p := g.CurrentPlayer()
	catapult, ok := board.HasCardTypeInHand[cards.Catapult](p)
	if !ok {
		return errors.New("player does not have a catapult to use")
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

	stolenGold, err := a.catapult.Attack(a.targetPlayer.Castle(), a.cardPosition)
	if err != nil {
		result := &Result{}
		return result, nil, fmt.Errorf("attacking castle failed: %w", err)
	}

	g.OnCardMovedToPile(stolenGold)

	g.AddHistory(fmt.Sprintf("%s removed %d gold from %s's castle",
		p.Name(), stolenGold.Value(), a.targetPlayer.Name()),
		types.CategoryAction)

	result := &Result{
		Action: types.LastActionCatapult,
	}
	statusFn := func() gamestatus.GameStatus {
		return g.Status(p)
	}

	return result, statusFn, nil
}

func (a *catapultAction) NextPhase() types.PhaseType {
	return types.PhaseTypeSpySteal
}

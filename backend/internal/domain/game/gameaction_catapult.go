package game

import (
	"errors"
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type CatapultAction struct {
	playerName       string
	targetPlayerName string
	cardPosition     int

	catapult     ports.Catapult
	targetPlayer ports.Player
	weapon       ports.Weapon
}

func NewCatapultAction(playerName, targetPlayerName string, cardPosition int) *CatapultAction {
	return &CatapultAction{
		playerName:       playerName,
		targetPlayerName: targetPlayerName,
		cardPosition:     cardPosition,
	}
}

func (a *CatapultAction) PlayerName() string { return a.playerName }

func (a *CatapultAction) Validate(g *Game) error {
	if g.currentAction != types.PhaseTypeAttack {
		return fmt.Errorf("cannot use catapult in the %s phase",
			g.currentAction)
	}

	p := g.CurrentPlayer()
	if !p.HasCatapult() {
		return errors.New("player does not have a catapult to use")
	}

	var err error
	a.targetPlayer, err = g.getTargetPlayer(a.playerName, a.targetPlayerName)
	if err != nil {
		return err
	}

	a.catapult = p.Catapult()
	if a.catapult == nil {
		return errors.New("player does not have a catapult to attack")
	}

	return nil
}

func (a *CatapultAction) Execute(g *Game) (*GameActionResult, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()

	stolenGold, err := a.catapult.Attack(a.targetPlayer.Castle(), a.cardPosition)
	if err != nil {
		result := &GameActionResult{}
		return result, nil, fmt.Errorf("attacking castle failed: %w", err)
	}

	g.OnCardMovedToPile(stolenGold)

	g.addToHistory(fmt.Sprintf("%s removed %d gold from %s's castle",
		p.Name(), stolenGold.Value(), a.targetPlayer.Name()),
		types.CategoryAction)

	result := &GameActionResult{
		Action: types.LastActionCatapult,
	}
	statusFn := func() gamestatus.GameStatus {
		return g.GameStatusProvider.Get(p, g)
	}

	return result, statusFn, nil
}

func (a *CatapultAction) NextPhase() types.PhaseType {
	return types.PhaseTypeSpySteal
}

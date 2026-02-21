package game

import (
	"errors"
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type MoveWarriorAction struct {
	playerName       string
	warriorID        string
	targetPlayerName string

	targetPlayer ports.Player
	warrior      ports.Warrior
	currentPhase types.PhaseType
}

func NewMoveWarriorAction(playerName, warriorID string, targetPlayerName string) *MoveWarriorAction {
	return &MoveWarriorAction{
		playerName:       playerName,
		warriorID:        warriorID,
		targetPlayerName: targetPlayerName,
	}
}

func (a *MoveWarriorAction) PlayerName() string { return a.playerName }

func (a *MoveWarriorAction) Validate(g *Game) error {
	if g.turnState.HasMovedWarrior {
		return errors.New("already moved a warrior this turn")
	}

	p := g.CurrentPlayer()

	// Ally field move (2v2 mode)
	if a.targetPlayerName != "" && a.targetPlayerName != a.playerName {
		targetPlayer := g.GetPlayer(a.targetPlayerName)
		if targetPlayer == nil {
			return fmt.Errorf("target player %s not found", a.targetPlayerName)
		}

		pIdx := g.PlayerIndex(a.playerName)
		tIdx := g.PlayerIndex(a.targetPlayerName)
		if !g.SameTeam(pIdx, tIdx) {
			return errors.New("can only move warriors to ally's field")
		}

		c, ok := p.GetCardFromHand(a.warriorID)
		if !ok {
			return fmt.Errorf("card with ID %s not found in hand", a.warriorID)
		}

		w, ok := c.(ports.Warrior)
		if !ok {
			return fmt.Errorf("only warrior cards can be moved to field")
		}

		a.targetPlayer = targetPlayer
		a.warrior = w
	}

	return nil
}

func (a *MoveWarriorAction) Execute(g *Game) (*GameActionResult, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()
	result := &GameActionResult{}

	if a.targetPlayer != nil {
		// Ally field move
		a.targetPlayer.Field().AddWarriors(a.warrior)
		p.Hand().RemoveCard(a.warrior)

		g.addToHistory(fmt.Sprintf("%s moved warrior to %s's field", p.Name(),
			a.targetPlayer.Name()), types.CategoryAction)
	} else {
		// Own field move
		if err := p.MoveCardToField(a.warriorID); err != nil {
			return result, nil, fmt.Errorf("moving warrior to field failed: %w", err)
		}

		g.addToHistory(fmt.Sprintf("%s moved warrior to field", p.Name()),
			types.CategoryAction)
	}

	g.turnState.HasMovedWarrior = true
	g.turnState.CanMoveWarrior = false
	result.MovedWarriorID = a.warriorID
	result.Action = types.LastActionMoveWarrior
	a.currentPhase = g.currentAction

	statusFn := func() gamestatus.GameStatus {
		return g.gameStatusProvider.Get(p, g)
	}

	return result, statusFn, nil
}

func (a *MoveWarriorAction) NextPhase() types.PhaseType {
	return a.currentPhase
}

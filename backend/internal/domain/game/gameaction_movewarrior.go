package game

import (
	"errors"
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type moveWarriorAction struct {
	playerName       string
	warriorID        string
	targetPlayerName string

	targetPlayer board.Player
	warrior      cards.Warrior
	currentPhase types.PhaseType
}

func NewMoveWarriorAction(playerName, warriorID string, targetPlayerName string) *moveWarriorAction {
	return &moveWarriorAction{
		playerName:       playerName,
		warriorID:        warriorID,
		targetPlayerName: targetPlayerName,
	}
}

func (a *moveWarriorAction) PlayerName() string { return a.playerName }

func (a *moveWarriorAction) Validate(g Game) error {
	if g.TurnState().HasMovedWarrior {
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

		w, ok := c.(cards.Warrior)
		if !ok {
			return fmt.Errorf("only warrior cards can be moved to field")
		}

		a.targetPlayer = targetPlayer
		a.warrior = w
	}

	return nil
}

func (a *moveWarriorAction) Execute(g Game) (*GameActionResult, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()
	result := &GameActionResult{}

	if a.targetPlayer != nil {
		// Ally field move
		a.targetPlayer.Field().AddWarriors(a.warrior)
		p.Hand().RemoveCard(a.warrior)

		g.AddHistory(fmt.Sprintf("%s moved warrior to %s's field", p.Name(),
			a.targetPlayer.Name()), types.CategoryAction)
	} else {
		// Own field move
		if err := p.MoveCardToField(a.warriorID); err != nil {
			return result, nil, fmt.Errorf("moving warrior to field failed: %w", err)
		}

		g.AddHistory(fmt.Sprintf("%s moved warrior to field", p.Name()),
			types.CategoryAction)
	}

	g.SetHasMovedWarrior(true)
	g.SetCanMoveWarrior(false)
	result.MovedWarriorID = a.warriorID
	result.Action = types.LastActionMoveWarrior
	a.currentPhase = g.CurrentAction()

	statusFn := func() gamestatus.GameStatus {
		return g.GameStatusProvider().Get(p, g)
	}

	return result, statusFn, nil
}

func (a *moveWarriorAction) NextPhase() types.PhaseType {
	return a.currentPhase
}

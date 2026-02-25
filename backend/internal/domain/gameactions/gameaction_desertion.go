package gameactions

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// desertionGame declares the minimum Game surface needed by desertionAction.
type desertionGame interface {
	GamePlayers
	GameTurn
	GameCards
	GameHistory
	GameStatusProvider
}

// desertionTargetPlayer declares the minimum Player surface needed by desertionAction.
type desertionTargetPlayer interface {
	board.PlayerIdentity
	board.PlayerField
}

type desertionAction struct {
	playerName       string
	targetPlayerName string
	warriorID        string

	targetPlayer desertionTargetPlayer
	desertion    cards.Desertion
	warrior      cards.Warrior
}

func NewDesertionAction(playerName, targetPlayerName, warriorID string) *desertionAction {
	return &desertionAction{
		playerName:       playerName,
		targetPlayerName: targetPlayerName,
		warriorID:        warriorID,
	}
}

func (a *desertionAction) PlayerName() string { return a.playerName }

func (a *desertionAction) Validate(g Game) error {
	if g.CurrentAction() != types.PhaseTypeSpySteal {
		return fmt.Errorf("cannot use desertion in the %s phase", g.CurrentAction())
	}

	p := g.CurrentPlayer()
	desertion, ok := board.HasCardTypeInHand[cards.Desertion](p)
	if !ok {
		return fmt.Errorf("player does not have a desertion card to use")
	}

	var err error
	a.targetPlayer, err = g.GetTargetPlayer(a.playerName, a.targetPlayerName)
	if err != nil {
		return err
	}

	warrior, ok := a.targetPlayer.Field().GetWarrior(a.warriorID)
	if !ok {
		return fmt.Errorf("warrior %s not found in %s's field", a.warriorID, a.targetPlayerName)
	}

	if warrior.Health() > cards.DesertionMaxHP {
		return fmt.Errorf("warrior has %d HP — only warriors with %d or fewer HP can be deserved",
			warrior.Health(), cards.DesertionMaxHP)
	}

	a.desertion = desertion
	a.warrior = warrior

	return nil
}

func (a *desertionAction) Execute(g Game) (*Result, func() gamestatus.GameStatus, error) {
	return a.execute(g)
}

func (a *desertionAction) execute(g desertionGame) (*Result, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()
	result := &Result{}

	// Remove warrior from enemy field and add to player's field.
	if !a.targetPlayer.Field().RemoveWarrior(a.warrior) {
		return result, nil, fmt.Errorf("failed to remove warrior %s from %s's field",
			a.warriorID, a.targetPlayerName)
	}
	p.Field().AddWarriors(a.warrior)

	// Discard the Desertion card.
	discarded, err := p.RemoveFromHand(a.desertion.GetID())
	if err != nil {
		return result, nil, fmt.Errorf("removing desertion card from hand failed: %w", err)
	}
	g.OnCardMovedToPile(discarded[0])

	result.Action = types.LastActionDesertion
	result.DeserterWarriorID = a.warrior.GetID()
	result.DeserterFromPlayer = a.targetPlayer.Name()
	result.DeserterWarrior = a.warrior

	g.AddHistory(fmt.Sprintf("%s's warrior deserted to %s's ranks",
		a.targetPlayer.Name(), p.Name()), types.CategoryAction)

	statusFn := func() gamestatus.GameStatus {
		return g.Status(p)
	}

	return result, statusFn, nil
}

func (a *desertionAction) NextPhase() types.PhaseType {
	return types.PhaseTypeBuy
}

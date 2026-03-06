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
	cardID           string

	targetPlayer desertionTargetPlayer
	desertion    cards.Desertion
	warrior      cards.Warrior
}

func NewDesertionAction(playerName, targetPlayerName, warriorID, cardID string) *desertionAction {
	return &desertionAction{
		playerName:       playerName,
		targetPlayerName: targetPlayerName,
		warriorID:        warriorID,
		cardID:           cardID,
	}
}

func (a *desertionAction) PlayerName() string { return a.playerName }

func (a *desertionAction) Validate(g Game) error {
	if g.CurrentAction() != types.PhaseTypeAttack {
		return fmt.Errorf("cannot use desertion in the %s phase", g.CurrentAction())
	}

	p := g.CurrentPlayer()
	raw, ok := p.GetCardFromHand(a.cardID)
	if !ok {
		return fmt.Errorf("card %s not found in hand", a.cardID)
	}
	desertion, ok := raw.(cards.Desertion)
	if !ok {
		return fmt.Errorf("card %s is not a desertion card", a.cardID)
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

	hp := warrior.Health()
	if hp > cards.DesertionMaxHP {
		return fmt.Errorf("warrior has %d HP — only warriors with %d or fewer HP can be deserved",
			hp, cards.DesertionMaxHP)
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
	p.PlaceWarriorOnField(a.warrior)

	// Discard the Desertion card.
	discarded, err := p.RemoveFromHand(a.desertion.GetID())
	if err != nil {
		return result, nil, fmt.Errorf("removing desertion card from hand failed: %w", err)
	}
	g.OnCardMovedToPile(discarded[0])

	result.Action = types.LastActionDesertion
	result.Desertion = &DesertionDetails{
		FromPlayer: a.targetPlayer.Name(),
		Warrior:    a.warrior,
	}

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

package gameactions

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// treasonGame declares the minimum Game surface needed by treasonAction.
type treasonGame interface {
	GamePlayers
	GameTurn
	GameCards
	GameHistory
	GameStatusProvider
}

// treasonTargetPlayer declares the minimum Player surface needed by treasonAction.
type treasonTargetPlayer interface {
	board.PlayerIdentity
	board.PlayerField
}

type treasonAction struct {
	playerName       string
	targetPlayerName string
	warriorID        string
	cardID           string

	targetPlayer treasonTargetPlayer
	treason      cards.Treason
	warrior      cards.Warrior
}

func NewTreasonAction(playerName, targetPlayerName, warriorID, cardID string) *treasonAction {
	return &treasonAction{
		playerName:       playerName,
		targetPlayerName: targetPlayerName,
		warriorID:        warriorID,
		cardID:           cardID,
	}
}

func (a *treasonAction) PlayerName() string { return a.playerName }

func (a *treasonAction) Validate(g Game) error {
	if g.CurrentAction() != types.PhaseTypeAttack {
		return fmt.Errorf("cannot use treason in the %s phase", g.CurrentAction())
	}

	p := g.CurrentPlayer()
	raw, ok := p.GetCardFromHand(a.cardID)
	if !ok {
		return fmt.Errorf("card %s not found in hand", a.cardID)
	}
	treason, ok := raw.(cards.Treason)
	if !ok {
		return fmt.Errorf("card %s is not a treason card", a.cardID)
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
	if hp > cards.TreasonMaxHP {
		return fmt.Errorf("warrior has %d HP — only warriors with %d or fewer HP can be traitors",
			hp, cards.TreasonMaxHP)
	}

	a.treason = treason
	a.warrior = warrior

	return nil
}

func (a *treasonAction) Execute(g Game) (*Result, func() gamestatus.GameStatus, error) {
	return a.execute(g)
}

func (a *treasonAction) execute(g treasonGame) (*Result, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()
	result := &Result{}

	// Remove warrior from enemy field and add to player's field.
	if !a.targetPlayer.Field().RemoveWarrior(a.warrior) {
		return result, nil, fmt.Errorf("failed to remove warrior %s from %s's field",
			a.warriorID, a.targetPlayerName)
	}
	p.PlaceWarriorOnField(a.warrior)

	// Discard the Treason card.
	discarded, err := p.RemoveFromHand(a.treason.GetID())
	if err != nil {
		return result, nil, fmt.Errorf("removing treason card from hand failed: %w", err)
	}
	g.OnCardMovedToPile(discarded[0])

	result.Action = types.LastActionTreason
	result.Treason = &TreasonDetails{
		FromPlayer: a.targetPlayer.Name(),
		Warrior:    a.warrior,
	}

	g.AddHistory(fmt.Sprintf("%s's warrior moved to %s's ranks",
		a.targetPlayer.Name(), p.Name()), types.CategoryAction)

	statusFn := func() gamestatus.GameStatus {
		return g.Status(p)
	}

	return result, statusFn, nil
}

func (a *treasonAction) NextPhase() types.PhaseType {
	return types.PhaseTypeBuy
}

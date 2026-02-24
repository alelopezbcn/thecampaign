package gameactions

import (
	"errors"
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// resurrectionGame declares the minimum Game surface needed by resurrectionAction
type resurrectionGame interface {
	GamePlayers
	GameTurn
	GameHistory
	GameStatusProvider
	GameBoard
	GameCards
}

// resurrectionTargetPlayer declares the minimum Player surface needed by resurrectionAction
type resurrectionTargetPlayer interface {
	board.PlayerIdentity
	board.PlayerField
}

type resurrectionAction struct {
	playerName       string
	targetPlayerName string // empty = own field

	resurrectionCard cards.Resurrection
	targetPlayer     resurrectionTargetPlayer
	currentPhase     types.PhaseType
}

func NewResurrectionAction(playerName, targetPlayerName string) *resurrectionAction {
	return &resurrectionAction{
		playerName:       playerName,
		targetPlayerName: targetPlayerName,
	}
}

func (a *resurrectionAction) PlayerName() string { return a.playerName }

func (a *resurrectionAction) Validate(g Game) error {
	if g.CurrentAction() != types.PhaseTypeAttack {
		return fmt.Errorf("cannot use resurrection in the %s phase", g.CurrentAction())
	}

	if g.Board().Cemetery().Count() == 0 {
		return errors.New("no warriors in the cemetery to resurrect")
	}

	p := g.CurrentPlayer()
	resCard, ok := board.HasCardTypeInHand[cards.Resurrection](p)
	if !ok {
		return errors.New("player does not have a resurrection card")
	}

	targetName := a.targetPlayerName
	if targetName == "" {
		targetName = a.playerName
	}

	target := g.GetPlayer(targetName)
	if target == nil {
		return fmt.Errorf("target player %s not found", targetName)
	}

	if targetName != a.playerName {
		pIdx := g.PlayerIndex(a.playerName)
		tIdx := g.PlayerIndex(targetName)
		if !g.SameTeam(pIdx, tIdx) {
			return fmt.Errorf("cannot resurrect to %s's field: not an ally", targetName)
		}
	}

	a.resurrectionCard = resCard
	a.targetPlayer = target
	return nil
}

func (a *resurrectionAction) Execute(g Game) (*Result, func() gamestatus.GameStatus, error) {
	return a.execute(g)
}

func (a *resurrectionAction) execute(g resurrectionGame) (*Result, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()

	warrior := g.Board().Cemetery().RemoveRandom()
	if warrior == nil {
		return nil, nil, errors.New("no warriors in the cemetery")
	}

	warrior.Resurrect()
	a.targetPlayer.Field().AddWarriors(warrior)

	removed, err := p.RemoveFromHand(a.resurrectionCard.GetID())
	if err != nil {
		return nil, nil, fmt.Errorf("removing resurrection card from hand: %w", err)
	}
	g.OnCardMovedToPile(removed[0])

	playerName := p.Name()
	targetName := a.targetPlayer.Name()
	if targetName == playerName {
		g.AddHistory(fmt.Sprintf("%s resurrected a warrior from the cemetery", playerName),
			types.CategoryAction)
	} else {
		g.AddHistory(fmt.Sprintf("%s resurrected a warrior to %s's field", playerName, targetName),
			types.CategoryAction)
	}

	a.currentPhase = g.CurrentAction()
	result := &Result{Action: types.LastActionResurrection}
	statusFn := func() gamestatus.GameStatus {
		return g.Status(p)
	}

	return result, statusFn, nil
}

func (a *resurrectionAction) NextPhase() types.PhaseType {
	return a.currentPhase
}

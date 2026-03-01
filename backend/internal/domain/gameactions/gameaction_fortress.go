package gameactions

import (
	"errors"
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// fortressGame declares the minimum Game surface needed by fortressAction
type fortressGame interface {
	GamePlayers
	GameTurn
	GameHistory
	GameStatusProvider
}

// fortressTargetPlayer declares the minimum Player surface needed by fortressAction
type fortressTargetPlayer interface {
	board.PlayerIdentity
	board.PlayerCastle
}

type fortressAction struct {
	playerName       string
	targetPlayerName string // empty = own castle
	cardID           string

	targetPlayer fortressTargetPlayer
	fortressCard cards.Fortress
}

func NewFortressAction(playerName, targetPlayerName, cardID string) *fortressAction {
	return &fortressAction{
		playerName:       playerName,
		targetPlayerName: targetPlayerName,
		cardID:           cardID,
	}
}

func (a *fortressAction) PlayerName() string { return a.playerName }

func (a *fortressAction) Validate(g Game) error {
	if g.CurrentAction() != types.PhaseTypeConstruct {
		return fmt.Errorf("cannot use fortress in the %s phase", g.CurrentAction())
	}

	p := g.CurrentPlayer()
	raw, ok := p.GetCardFromHand(a.cardID)
	if !ok {
		return fmt.Errorf("card %s not found in hand", a.cardID)
	}
	fortress, ok := raw.(cards.Fortress)
	if !ok {
		return errors.New("card is not a fortress")
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
			return fmt.Errorf("cannot fortify %s's castle: not an ally", targetName)
		}
	}

	if !target.Castle().IsConstructed() {
		return errors.New("cannot fortify a castle that has not been constructed yet")
	}

	if target.Castle().IsProtected() {
		return errors.New("castle is already protected by a fortress")
	}

	a.fortressCard = fortress
	a.targetPlayer = target
	return nil
}

func (a *fortressAction) Execute(g Game) (*Result, func() gamestatus.GameStatus, error) {
	return a.execute(g)
}

func (a *fortressAction) execute(g fortressGame) (*Result, func() gamestatus.GameStatus, error) {
	p := g.CurrentPlayer()

	a.targetPlayer.Castle().SetProtection(a.fortressCard)
	p.RemoveFromHand(a.fortressCard.GetID())

	playerName := p.Name()
	targetName := a.targetPlayer.Name()
	if targetName == playerName {
		g.AddHistory(fmt.Sprintf("%s fortified their castle", playerName), types.CategoryAction)
	} else {
		g.AddHistory(fmt.Sprintf("%s fortified %s's castle", playerName, targetName), types.CategoryAction)
	}

	result := &Result{Action: types.LastActionFortress}
	statusFn := func() gamestatus.GameStatus {
		return g.Status(p)
	}

	return result, statusFn, nil
}

func (a *fortressAction) NextPhase() types.PhaseType {
	return types.PhaseTypeEndTurn
}

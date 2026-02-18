package domain

import (
	"errors"
	"fmt"
	"log"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// AutoMoveWarriorToField moves a warrior to the field during game setup (no turn validation)
func (g *Game) AutoMoveWarriorToField(playerName, warriorID string) error {
	p := g.GetPlayer(playerName)
	if p == nil {
		return fmt.Errorf("player %s not found", playerName)
	}
	return p.MoveCardToField(warriorID)
}

func (g *Game) MoveWarriorToField(playerName, warriorID string, targetPlayerName ...string) (
	status GameStatus, err error) {

	p := g.CurrentPlayer()
	if p.Name() != playerName {
		return status, fmt.Errorf("%s not your turn", playerName)
	}

	if g.hasMovedWarrior {
		return status, errors.New("already moved a warrior this turn")
	}

	// Check if moving to an ally's field (2v2 mode)
	if len(targetPlayerName) > 0 && targetPlayerName[0] != "" && targetPlayerName[0] != playerName {
		targetPlayer := g.GetPlayer(targetPlayerName[0])
		if targetPlayer == nil {
			return status, fmt.Errorf("target player %s not found", targetPlayerName[0])
		}

		pIdx := g.PlayerIndex(playerName)
		tIdx := g.PlayerIndex(targetPlayerName[0])
		if !g.SameTeam(pIdx, tIdx) {
			return status, errors.New("can only move warriors to ally's field")
		}

		c, ok := p.GetCardFromHand(warriorID)
		if !ok {
			return status, fmt.Errorf("card with ID %s not found in hand", warriorID)
		}

		w, ok := c.(ports.Warrior)
		if !ok {
			return status, fmt.Errorf("only warrior cards can be moved to field")
		}

		targetPlayer.Field().AddWarriors(w)
		p.Hand().RemoveCard(c)

		g.addToHistory(fmt.Sprintf("%s moved warrior to %s's field", p.Name(),
			targetPlayer.Name()), types.CategoryAction)
	} else {
		err = p.MoveCardToField(warriorID)
		if err != nil {
			return status, fmt.Errorf("moving warrior to field failed: %w", err)
		}
		g.addToHistory(fmt.Sprintf("%s moved warrior to field", p.Name()),
			types.CategoryAction)
	}

	g.hasMovedWarrior = true
	g.CanMoveWarrior = false
	g.lastResult.MovedWarriorID = warriorID
	g.lastResult.Action = types.LastActionMoveWarrior
	status = g.GameStatusProvider.Get(p, g)

	return status, nil
}

func (g *Game) Trade(playerName string, cardIDs []string) (
	status GameStatus, err error) {

	var cards []ports.Card
	p := g.CurrentPlayer()
	if p.Name() != playerName {
		return status, fmt.Errorf("%s not your turn", playerName)
	}

	if g.hasTraded {
		return status, errors.New("already traded this turn")
	}

	if len(cardIDs) != 3 {
		return status, errors.New("must trade exactly 3 cards")
	}

	tradedCards, err := p.GiveCards(cardIDs...)
	if err != nil {
		return status, fmt.Errorf("giving cards for trading failed: %w", err)
	}
	for _, c := range tradedCards {
		g.OnCardMovedToPile(c)
	}

	cards, err = g.drawCards(p, 1)
	if err != nil {
		return status, fmt.Errorf("drawing card for trading failed: %w", err)
	}

	p.TakeCards(cards...)

	g.addToHistory(fmt.Sprintf("%s traded 3 cards", p.Name()), types.CategoryAction)

	g.lastResult.Action = types.LastActionTrade
	g.hasTraded = true
	g.CanTrade = false
	status = g.GameStatusProvider.Get(p, g, cards...)

	return status, nil
}

func (g *Game) Buy(playerName, cardID string) (
	status GameStatus, err error) {
	p := g.CurrentPlayer()
	if p.Name() != playerName {
		return status, fmt.Errorf("%s not your turn", playerName)
	}

	if g.currentAction != types.PhaseTypeBuy {
		return status, fmt.Errorf("cannot buy in the %s phase",
			g.currentAction)
	}

	resourceCard, ok := p.GetCardFromHand(cardID)
	if !ok {
		return status, errors.New("Resource card not in hand: " + cardID)
	}

	r, ok := resourceCard.(ports.Resource)
	if !ok {
		return status, errors.New("only gold cards can be used to buy")
	}

	val := r.Value()
	if _, err = p.GiveCards(resourceCard.GetID()); err != nil {
		return status, fmt.Errorf("giving card for buying failed: %w", err)
	}

	cardsToBuy := val / 2
	cards, err := g.drawCards(p, cardsToBuy)
	if err != nil {
		p.TakeCards(resourceCard)
		if errors.Is(err, ErrHandLimitExceeded) {
			return status, fmt.Errorf("cards in hand limit exceeded")
		}
		return status, fmt.Errorf("drawing card for buying failed: %w", err)
	}

	p.TakeCards(cards...)

	g.OnCardMovedToPile(resourceCard)

	g.addToHistory(fmt.Sprintf("%s bought %d card(s)", p.Name(), cardsToBuy),
		types.CategoryAction)

	g.lastResult.Action = types.LastActionBuy
	status = g.nextAction(types.PhaseTypeConstruct,
		func() GameStatus {
			return g.GameStatusProvider.Get(p, g, cards...)
		})

	return status, nil
}

func (g *Game) Construct(playerName, cardID string, targetPlayerName ...string) (
	status GameStatus, err error) {

	p := g.CurrentPlayer()
	if p.Name() != playerName {
		return status, fmt.Errorf("%s not your turn", playerName)
	}

	if g.currentAction != types.PhaseTypeConstruct {
		return status, fmt.Errorf("cannot construct in the %s phase",
			g.currentAction)
	}

	// Check if constructing on ally's castle (2v2 mode)
	if len(targetPlayerName) > 0 && targetPlayerName[0] != "" && targetPlayerName[0] != playerName {
		targetPlayer := g.GetPlayer(targetPlayerName[0])
		if targetPlayer == nil {
			return status, fmt.Errorf("target player %s not found", targetPlayerName[0])
		}

		pIdx := g.PlayerIndex(playerName)
		tIdx := g.PlayerIndex(targetPlayerName[0])
		if !g.SameTeam(pIdx, tIdx) {
			return status, errors.New("can only construct on ally's castle")
		}

		resourceCard, ok := p.GetCardFromHand(cardID)
		if !ok {
			return status, errors.New("card not in hand: " + cardID)
		}

		_, isResource := resourceCard.(ports.Resource)
		log.Printf("DEBUG Construct ally: player=%s, target=%s, cardID=%s, isResource=%v, cardType=%T, castleConstructed=%v",
			playerName, targetPlayerName[0], cardID, isResource, resourceCard, targetPlayer.Castle().IsConstructed())

		if err = targetPlayer.Castle().Construct(resourceCard); err != nil {
			return status, fmt.Errorf("constructing on ally castle failed: %w", err)
		}

		p.Hand().RemoveCard(resourceCard)

		g.addToHistory(fmt.Sprintf("%s added gold to %s's castle", p.Name(),
			targetPlayer.Name()), types.CategoryAction)
	} else {

		if err = p.Construct(cardID); err != nil {
			return status, fmt.Errorf("constructing card failed: %w", err)
		}

		g.addToHistory(fmt.Sprintf("%s constructed", p.Name()), types.CategoryAction)
	}

	g.lastResult.Action = types.LastActionConstruct
	status = g.nextAction(types.PhaseTypeEndTurn,
		func() GameStatus {
			return g.GameStatusProvider.Get(p, g)
		})

	return status, nil
}

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
	g.lastMovedWarriorID = warriorID
	g.lastAction = "move_warrior"
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

	g.lastAction = "trade"
	g.hasTraded = true
	g.CanTrade = false
	status = g.GameStatusProvider.Get(p, g, cards...)

	return status, nil
}

func (g *Game) DrawCard(playerName string) (status GameStatus, err error) {
	var cards []ports.Card

	p := g.CurrentPlayer()
	if p.Name() != playerName {
		return status, fmt.Errorf("%s not your turn", playerName)
	}

	cards, err = g.drawCards(p, 1)
	if err != nil {
		if errors.Is(err, ErrHandLimitExceeded) {
			// Player has max cards, skip drawing but continue to attack phase
			g.addToHistory(fmt.Sprintf("%s can't take more cards (hand limit reached)", p.Name()),
				types.CategoryError)

			status = g.nextAction(types.ActionTypeAttack,
				func() GameStatus {
					return g.GameStatusProvider.Get(p, g)
				})

			return status, nil
		}

		return status, fmt.Errorf("drawing card failed: %w", err)
	}

	p.TakeCards(cards...)

	g.addToHistory(fmt.Sprintf("%s drew a card", p.Name()), types.CategoryAction)

	g.lastAction = "draw"
	status = g.nextAction(types.ActionTypeAttack,
		func() GameStatus {
			return g.GameStatusProvider.Get(p, g, cards...)
		})

	return status, nil
}

func (g *Game) Attack(playerName, targetPlayerName, targetID, weaponID string) (
	status GameStatus, err error) {

	p := g.CurrentPlayer()
	if p.Name() != playerName {
		return status, fmt.Errorf("%s not your turn", playerName)
	}

	if g.currentAction != types.ActionTypeAttack {
		return status, fmt.Errorf("cannot attack in the %s phase",
			g.currentAction)
	}

	targetPlayer, err := g.getTargetPlayer(playerName, targetPlayerName)
	if err != nil {
		return status, err
	}

	targetCard, ok := targetPlayer.GetCardFromField(targetID)
	if !ok {
		return status, errors.New("target card not in enemy field: " + targetID)
	}

	weaponCard, ok := p.GetCardFromHand(weaponID)
	if !ok {
		return status, errors.New("weapon card not in hand: " + weaponID)
	}

	t, ok := targetCard.(ports.Attackable)
	if !ok {
		return status, fmt.Errorf("the target cardBase cannot be attacked")
	}

	w, ok := weaponCard.(ports.Weapon)
	if !ok {
		return status, fmt.Errorf("the card is not a weapon")
	}

	if err = p.Attack(t, w); err != nil {
		return status, fmt.Errorf("attack action failed: %w", err)
	}

	g.addToHistory(fmt.Sprintf("%s attacked %s with %s",
		playerName, t.String(), w.String()),
		types.CategoryAction)

	g.lastAction = "attack"
	status = g.nextAction(types.ActionTypeSpySteal,
		func() GameStatus {
			return g.GameStatusProvider.Get(p, g)
		})

	return status, nil
}

func (g *Game) SpecialPower(playerName, userID, targetID, weaponID string) (
	status GameStatus, err error) {

	p := g.CurrentPlayer()
	if p.Name() != playerName {
		return status, fmt.Errorf("%s not your turn", playerName)
	}

	if g.currentAction != types.ActionTypeAttack {
		return status, fmt.Errorf("cannot use special power in the %s phase",
			g.currentAction)
	}

	userCard, ok := p.GetCardFromField(userID)
	if !ok {
		return status, errors.New("warrior card not in field: " + userID)
	}

	// Determine user warrior type for validation
	userWarrior, ok := userCard.(ports.Warrior)
	if !ok {
		return status, fmt.Errorf("the attacking card is not a warrior")
	}
	userType := userWarrior.Type()

	var targetCard ports.Card
	targetIsAllyOrSelf := false

	// Search own field
	targetCard, ok = p.GetCardFromField(targetID)
	if ok {
		targetIsAllyOrSelf = true
	}
	if !ok {
		// Search ally fields (2v2)
		for _, ally := range g.Allies(g.PlayerIndex(playerName)) {
			targetCard, ok = ally.GetCardFromField(targetID)
			if ok {
				targetIsAllyOrSelf = true
				break
			}
		}
	}
	if !ok {
		// Search enemy fields
		for _, enemy := range g.Enemies(g.PlayerIndex(playerName)) {
			targetCard, ok = enemy.GetCardFromField(targetID)
			if ok {
				break
			}
		}
	}
	if !ok {
		return status, errors.New("target card not valid: " + targetID)
	}

	// Validate target side based on warrior type
	if userType == types.ArcherWarriorType && targetIsAllyOrSelf {
		return status, errors.New("archer instant kill can only target enemies")
	}
	if (userType == types.KnightWarriorType || userType == types.MageWarriorType) && !targetIsAllyOrSelf {
		return status, errors.New("knight/mage special power can only target allies")
	}

	weaponCard, ok := p.GetCardFromHand(weaponID)
	if !ok {
		return status, errors.New("weapon card not in hand: " + weaponID)
	}

	s, ok := weaponCard.(ports.SpecialPower)
	if !ok {
		return status, fmt.Errorf("the card is not a special power")
	}

	t, ok := targetCard.(ports.Warrior)
	if !ok {
		return status, fmt.Errorf("the target card is not a warrior")
	}

	if err = p.UseSpecialPower(userWarrior, t, s); err != nil {
		return status, fmt.Errorf("special power action failed: %w", err)
	}

	g.addToHistory(fmt.Sprintf("%s used special power on %s",
		playerName, t.String()), types.CategoryAction)

	g.lastAction = "special_power"
	status = g.nextAction(types.ActionTypeSpySteal,
		func() GameStatus {
			return g.GameStatusProvider.Get(p, g)
		})

	return status, nil
}

func (g *Game) Catapult(playerName, targetPlayerName string, cardPosition int) (
	status GameStatus, err error) {

	p := g.CurrentPlayer()
	if p.Name() != playerName {
		return status, fmt.Errorf("%s not your turn", playerName)
	}

	if g.currentAction != types.ActionTypeAttack {
		return status, fmt.Errorf("cannot use catapult in the %s phase",
			g.currentAction)
	}

	if !p.HasCatapult() {
		return status, errors.New("player does not have a catapult to use")
	}

	targetPlayer, err := g.getTargetPlayer(playerName, targetPlayerName)
	if err != nil {
		return status, err
	}

	t := p.Catapult()
	if t == nil {
		return status, errors.New("player does not have a catapult to attack")
	}

	stolenGold, err := t.Attack(targetPlayer.Castle(), cardPosition)
	if err != nil {
		return status, fmt.Errorf("attacking castle failed: %w", err)
	}

	g.OnCardMovedToPile(stolenGold)

	g.addToHistory(fmt.Sprintf("%s removed %d gold from %s's castle",
		p.Name(), stolenGold.Value(), targetPlayer.Name()),
		types.CategoryAction)

	g.lastAction = "catapult"
	status = g.nextAction(types.ActionTypeSpySteal,
		func() GameStatus {
			return g.GameStatusProvider.Get(p, g)
		})

	return status, nil
}

func (g *Game) Spy(playerName, targetPlayerName string, option int) (
	status GameStatus, err error) {

	p := g.CurrentPlayer()
	if p.Name() != playerName {
		return status, fmt.Errorf("%s not your turn", playerName)
	}

	if g.currentAction != types.ActionTypeSpySteal {
		return status, fmt.Errorf("cannot use Spy in the %s phase",
			g.currentAction)
	}

	if !p.HasSpy() {
		return status, errors.New("player does not have a spy to use")
	}

	var spiedCards []ports.Card

	switch option {
	case 1:
		// Reveal top 5 cards from deck
		g.addToHistory(fmt.Sprintf("%s spied top 5 cards from deck", p.Name()),
			types.CategoryAction)

		g.lastSpyInfo = "deck"
		spiedCards = g.deck.Reveal(5)
	case 2:
		// Reveal target's cards
		targetPlayer, err := g.getTargetPlayer(playerName, targetPlayerName)
		if err != nil {
			return status, err
		}

		g.addToHistory(fmt.Sprintf("%s spied on %s's hand",
			p.Name(), targetPlayer.Name()), types.CategoryAction)

		g.lastSpyInfo = targetPlayer.Name()
		spiedCards = targetPlayer.Hand().ShowCards()
	default:
		return status, errors.New("invalid Spy option")
	}

	s := p.Spy()
	if s == nil {
		return status, errors.New("failed to retrieve spy card")
	}

	g.OnCardMovedToPile(s)

	g.lastAction = "spy"
	status = g.nextAction(types.ActionTypeBuy,
		func() GameStatus {
			return g.GameStatusProvider.GetWithModal(p, g, spiedCards)
		})

	return status, nil
}

func (g *Game) Steal(playerName, targetPlayerName string, cardPosition int) (
	status GameStatus, err error) {

	p := g.CurrentPlayer()
	if p.Name() != playerName {
		return status, fmt.Errorf("%s not your turn", playerName)
	}

	if g.currentAction != types.ActionTypeSpySteal {
		return status, fmt.Errorf("cannot use Steal in the %s phase",
			g.currentAction)
	}

	if !p.HasThief() {
		return status, errors.New("player does not have a thief to steal with")
	}

	targetPlayer, err := g.getTargetPlayer(playerName, targetPlayerName)
	if err != nil {
		return status, err
	}

	stolenCard, err := targetPlayer.CardStolenFromHand(cardPosition)
	if err != nil {
		return status, fmt.Errorf("stealing card failed: %w", err)
	}

	t := p.Thief()
	if t == nil {
		return status, errors.New("failed to retrieve thief card")
	}

	g.OnCardMovedToPile(t)
	p.TakeCards(stolenCard)

	g.lastStolenFrom = targetPlayer.Name()
	g.lastStolenCard = stolenCard

	g.addToHistory(fmt.Sprintf("%s stole a card from %s",
		p.Name(), targetPlayer.Name()), types.CategoryAction)

	g.lastAction = "steal"
	status = g.nextAction(types.ActionTypeBuy,
		func() GameStatus {
			return g.GameStatusProvider.GetWithModal(p, g, []ports.Card{stolenCard})
		})

	return status, nil
}

func (g *Game) Buy(playerName, cardID string) (
	status GameStatus, err error) {
	p := g.CurrentPlayer()
	if p.Name() != playerName {
		return status, fmt.Errorf("%s not your turn", playerName)
	}

	if g.currentAction != types.ActionTypeBuy {
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

	g.lastAction = "buy"
	status = g.nextAction(types.ActionTypeConstruct,
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

	if g.currentAction != types.ActionTypeConstruct {
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

	g.lastAction = "construct"
	status = g.nextAction(types.ActionTypeEndTurn,
		func() GameStatus {
			return g.GameStatusProvider.Get(p, g)
		})

	return status, nil
}

func (g *Game) SkipPhase(playerName string) (status GameStatus, err error) {
	var nextAction types.ActionType

	p := g.CurrentPlayer()
	if p.Name() != playerName {
		return status, fmt.Errorf("%s not your turn", playerName)
	}

	skippedPhase := g.currentAction

	switch g.currentAction {
	case types.ActionTypeAttack:
		nextAction = types.ActionTypeSpySteal
	case types.ActionTypeSpySteal:
		nextAction = types.ActionTypeBuy
	case types.ActionTypeBuy:
		nextAction = types.ActionTypeConstruct
	case types.ActionTypeConstruct:
		nextAction = types.ActionTypeEndTurn
	default:
		return status, errors.New("cannot skip this phase")
	}

	g.addToHistory(fmt.Sprintf("%s skipped phase %s", p.Name(), skippedPhase),
		types.CategorySkip)

	g.lastAction = "skip"
	status = g.nextAction(nextAction,
		func() GameStatus {
			return g.GameStatusProvider.Get(p, g)
		})

	return status, nil
}

func (g *Game) EndTurn(player string, expired bool) (status GameStatus, err error) {
	p := g.CurrentPlayer()
	if p.Name() != player {
		return status, fmt.Errorf("%s not your turn", player)
	}

	if expired {
		g.addToHistory(fmt.Sprintf("%s's turn expired", p.Name()),
			types.CategoryTurnExpired)
	} else {
		g.addToHistory(fmt.Sprintf("%s ended their turn", p.Name()),
			types.CategoryEndTurn)
	}

	g.lastAction = "end_turn"
	g.switchTurn()
	status = g.nextAction(types.ActionTypeDrawCard,
		func() GameStatus {
			return g.GameStatusProvider.Get(g.CurrentPlayer(), g)
		})

	return status, nil
}

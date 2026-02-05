package domain

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/google/uuid"
)

type Games []Game

type Game struct {
	id                 string
	Players            []ports.Player
	CurrentTurn        int
	currentAction      types.ActionType
	CanMoveWarrior     bool
	hasMovedWarrior    bool
	CanTrade           bool
	hasTraded          bool
	deck               ports.Deck
	discardPile        ports.DiscardPile
	cemetery           ports.Cemetery
	history            []string
	historyTracker     int
	dealer             ports.Dealer
	GameStatusProvider GameStatusProvider
	gameOver           bool
	winner             string
}

func NewGame(player1, player2 string,
	dealer ports.Dealer, gameStatusProvider GameStatusProvider) *Game {

	playersArr := []string{player1, player2}
	rand.Shuffle(len(playersArr), func(i, j int) {
		playersArr[i], playersArr[j] = playersArr[j], playersArr[i]
	})

	g := &Game{
		id:                 uuid.NewString(),
		CurrentTurn:        0,
		discardPile:        newDiscardPile(),
		cemetery:           newCemetery(),
		history:            []string{},
		dealer:             dealer,
		GameStatusProvider: gameStatusProvider,
	}

	p1 := NewPlayer(playersArr[0], g, g, g, g)
	p2 := NewPlayer(playersArr[1], g, g, g, g)
	g.Players = []ports.Player{p1, p2}

	g.addToHistory(fmt.Sprintf("Game created between %s and %s",
		p1.Name(), p2.Name()))

	g.deal()

	return g
}

func (g *Game) deal() {
	warriorCards := shuffle(g.dealer.WarriorsCards())

	// Each player gets 3 Warrior cards
	warriorsIdx := 0
	for _, p := range g.Players {
		p.TakeCards(warriorCards[warriorsIdx : warriorsIdx+3]...)
		warriorsIdx += 3
	}

	deckCards := append(warriorCards[warriorsIdx:], g.dealer.OtherCards()...)
	deckCards = shuffle(deckCards)
	otherIdx := 0
	for _, p := range g.Players {
		p.TakeCards(deckCards[otherIdx : otherIdx+4]...)
		otherIdx += 4
	}

	deckCards = deckCards[otherIdx:]
	g.deck = NewDeck(deckCards)
}

func (g *Game) GetInitialWarriors(playerName string) (warriors [3]gamestatus.Card) {
	i := 0
	for _, p := range g.Players {
		if p.Name() == playerName {
			for _, c := range p.Hand().ShowCards() {
				if w, ok := c.(ports.Warrior); ok {
					warriors[i] = gamestatus.FromDomainCard(w)
					i++
					if i == 3 {
						return warriors
					}
				}
			}
		}
	}

	return warriors
}

func (g *Game) SetInitialWarriors(playerName string, warriorIDs []string) error {
	p, _ := g.WhoIsCurrent()
	if p.Name() != playerName {
		return fmt.Errorf("%s not your turn", playerName)
	}

	if len(warriorIDs) < 1 {
		return errors.New("must place at least 1 warrior")
	}
	if len(warriorIDs) > 3 {
		return errors.New("cannot place more than 3 warriors")
	}

	for _, id := range warriorIDs {
		if err := p.MoveCardToField(strings.TrimSpace(id)); err != nil {
			return err
		}
	}
	g.switchTurn()

	// Check if both players have set their warriors
	allSet := true
	for _, p := range g.Players {
		if len(p.Field().Warriors()) == 0 {
			allSet = false
			break
		}
	}
	if allSet {
		return nil
	}

	return nil
}

func (g *Game) WhoIsCurrent() (current ports.Player, enemy ports.Player) {
	return g.Players[g.CurrentTurn],
		g.Players[(g.CurrentTurn+1)%len(g.Players)]
}

func (g *Game) MoveWarriorToField(playerName, warriorID string) (
	status GameStatus, err error) {

	p, e := g.WhoIsCurrent()
	if p.Name() != playerName {
		return status, fmt.Errorf("%s not your turn", playerName)
	}

	err = p.MoveCardToField(warriorID)
	if err != nil {
		return status, fmt.Errorf("moving warrior to field failed: %w", err)
	}

	g.addToHistory(fmt.Sprintf("%s moved warrior to field", p.Name()))
	g.hasMovedWarrior = true
	status = g.GameStatusProvider.Get(p, e, g)

	return status, nil
}

func (g *Game) Trade(playerName string, cardIDs []string) (
	status GameStatus, err error) {

	var cards []ports.Card
	p, e := g.WhoIsCurrent()
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

	g.addToHistory(fmt.Sprintf("%s traded 3 cards", p.Name()))
	g.hasTraded = true
	g.CanTrade = false
	status = g.GameStatusProvider.Get(p, e, g, cards...)

	return status, nil
}

func (g *Game) DrawCard(playerName string) (status GameStatus, err error) {
	var cards []ports.Card

	p, e := g.WhoIsCurrent()
	if p.Name() != playerName {
		return status, fmt.Errorf("%s not your turn", playerName)
	}

	cards, err = g.drawCards(p, 1)
	if err != nil {
		if errors.Is(err, ErrHandLimitExceeded) {
			// Player has max cards, skip drawing but continue to attack phase
			g.addToHistory(fmt.Sprintf("%s can't take more cards (hand limit reached)", p.Name()))
			status = g.nextAction(types.ActionTypeAttack,
				func() GameStatus {
					return g.GameStatusProvider.Get(p, e, g)
				})
			return status, nil
		}
		return status, fmt.Errorf("drawing card failed: %w", err)
	}

	p.TakeCards(cards...)
	g.addToHistory(fmt.Sprintf("%s drew a card", p.Name()))
	status = g.nextAction(types.ActionTypeAttack,
		func() GameStatus {
			return g.GameStatusProvider.Get(p, e, g, cards...)
		})

	return status, nil
}

func (g *Game) Attack(playerName, targetID, weaponID string) (
	status GameStatus, err error) {

	p, e := g.WhoIsCurrent()
	if p.Name() != playerName {
		return status, fmt.Errorf("%s not your turn", playerName)
	}

	if g.currentAction != types.ActionTypeAttack {
		return status, fmt.Errorf("cannot attack in the %s phase",
			g.currentAction)
	}

	targetCard, ok := e.GetCardFromField(targetID)
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
		playerName, t.String(), w.String()))
	status = g.nextAction(types.ActionTypeSpySteal,
		func() GameStatus {
			return g.GameStatusProvider.Get(p, e, g)
		})

	return status, nil
}

func (g *Game) SpecialPower(playerName, userID, targetID, weaponID string) (
	status GameStatus, err error) {

	p, e := g.WhoIsCurrent()
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

	var targetCard ports.Card
	targetCard, ok = p.GetCardFromField(targetID)
	if !ok {
		targetCard, ok = e.GetCardFromField(targetID)
		if !ok {
			return status, errors.New("target card not valid: " + targetID)
		}
	}

	weaponCard, ok := p.GetCardFromHand(weaponID)
	if !ok {
		return status, errors.New("weapon card not in hand: " + weaponID)
	}

	s, ok := weaponCard.(ports.SpecialPower)
	if !ok {
		return status, fmt.Errorf("the card is not a special power")
	}

	w, ok := userCard.(ports.Warrior)
	if !ok {
		return status, fmt.Errorf("the attacking card is not a warrior")
	}

	t, ok := targetCard.(ports.Warrior)
	if !ok {
		return status, fmt.Errorf("the target card is not a warrior")
	}

	if err = p.UseSpecialPower(w, t, s); err != nil {
		return status, fmt.Errorf("special power action failed: %w", err)
	}

	g.addToHistory(fmt.Sprintf("%s used special power on %s",
		playerName, t.String()))
	status = g.nextAction(types.ActionTypeSpySteal,
		func() GameStatus {
			return g.GameStatusProvider.Get(p, e, g)
		})

	return status, nil
}

func (g *Game) Catapult(playerName string, cardPosition int) (
	status GameStatus, err error) {

	p, e := g.WhoIsCurrent()
	if p.Name() != playerName {
		return status, fmt.Errorf("%s not your turn", playerName)
	}

	if g.currentAction != types.ActionTypeAttack {
		return status, fmt.Errorf("cannot use catapult in the %s phase",
			g.currentAction)
	}

	t := p.Catapult()
	if t == nil {
		return status, errors.New("player does not have a catapult to attack")
	}

	stolenGold, err := t.Attack(e.Castle(), cardPosition)
	if err != nil {
		return status, fmt.Errorf("attacking castle failed: %w", err)
	}

	g.OnCardMovedToPile(stolenGold)

	g.addToHistory(fmt.Sprintf("%s removed %d gold from %s's castle",
		p.Name(), stolenGold.Value(), e.Name()))
	status = g.nextAction(types.ActionTypeSpySteal,
		func() GameStatus {
			return g.GameStatusProvider.Get(p, e, g)
		})

	return status, nil
}

func (g *Game) Spy(playerName string, option int) (status GameStatus, err error) {
	p, e := g.WhoIsCurrent()
	if p.Name() != playerName {
		return status, fmt.Errorf("%s not your turn", playerName)
	}

	if g.currentAction != types.ActionTypeSpySteal {
		return status, fmt.Errorf("cannot use Spy in the %s phase",
			g.currentAction)
	}

	s := p.Spy()
	if s == nil {
		return status, errors.New("player does not have a Spy to use")
	}

	g.OnCardMovedToPile(s)
	var spiedCards []ports.Card

	switch option {
	case 1:
		// Reveal top 5 cards from deck
		g.addToHistory(fmt.Sprintf("%s spied top 5 cards from deck", p.Name()))
		spiedCards = g.deck.Reveal(5)
	case 2:
		// Reveal enemy's cards
		g.addToHistory(fmt.Sprintf("%s spied on %s's hand",
			p.Name(), e.Name()))
		spiedCards = e.Hand().ShowCards()
	default:
		return status, errors.New("invalid Spy option")
	}

	status = g.nextAction(types.ActionTypeBuy,
		func() GameStatus {
			return g.GameStatusProvider.GetWithModal(p, e, g, spiedCards)
		})

	return status, nil
}

func (g *Game) Steal(playerName string, cardPosition int) (
	status GameStatus, err error) {
	p, e := g.WhoIsCurrent()
	if p.Name() != playerName {
		return status, fmt.Errorf("%s not your turn", playerName)
	}

	if g.currentAction != types.ActionTypeSpySteal {
		return status, fmt.Errorf("cannot use Steal in the %s phase",
			g.currentAction)
	}

	t := p.Thief()
	if t == nil {
		return status, errors.New("player does not have a thief to steal with")
	}

	stolenCard, err := e.CardStolenFromHand(cardPosition)
	if err != nil {
		return status, fmt.Errorf("stealing card failed: %w", err)
	}

	g.OnCardMovedToPile(t)
	p.TakeCards(stolenCard)

	g.addToHistory(fmt.Sprintf("%s stole a card from %s",
		p.Name(), e.Name()))
	status = g.nextAction(types.ActionTypeBuy,
		func() GameStatus {
			return g.GameStatusProvider.GetWithModal(p, e, g, []ports.Card{stolenCard})
		})

	return status, nil
}

func (g *Game) Buy(playerName, cardID string) (
	status GameStatus, err error) {
	p, e := g.WhoIsCurrent()
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
	g.addToHistory(fmt.Sprintf("%s bought %d card(s)", p.Name(), cardsToBuy))
	status = g.nextAction(types.ActionTypeConstruct,
		func() GameStatus {
			return g.GameStatusProvider.Get(p, e, g, cards...)
		})

	return status, nil
}

func (g *Game) Construct(playerName, cardID string) (
	status GameStatus, err error) {

	p, e := g.WhoIsCurrent()
	if p.Name() != playerName {
		return status, fmt.Errorf("%s not your turn", playerName)
	}

	if g.currentAction != types.ActionTypeConstruct {
		return status, fmt.Errorf("cannot construct in the %s phase",
			g.currentAction)
	}

	if err = p.Construct(cardID); err != nil {
		return status, fmt.Errorf("constructing card failed: %w", err)
	}

	g.addToHistory(fmt.Sprintf("%s constructed", p.Name()))
	status = g.nextAction(types.ActionTypeEndTurn,
		func() GameStatus {
			return g.GameStatusProvider.Get(p, e, g)
		})

	return status, nil
}

func (g *Game) SkipPhase(playerName string) (status GameStatus, err error) {
	var nextAction types.ActionType

	p, e := g.WhoIsCurrent()
	if p.Name() != playerName {
		return status, fmt.Errorf("%s not your turn", playerName)
	}

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

	g.addToHistory(fmt.Sprintf("%s skipped phase", p.Name()))
	status = g.nextAction(nextAction,
		func() GameStatus {
			return g.GameStatusProvider.Get(p, e, g)
		})

	return status, nil
}

func (g *Game) EndTurn(player string) (status GameStatus, err error) {
	p, _ := g.WhoIsCurrent()
	if p.Name() != player {
		return status, fmt.Errorf("%s not your turn", player)
	}

	g.switchTurn()
	p, e := g.WhoIsCurrent()
	status = g.nextAction(types.ActionTypeDrawCard,
		func() GameStatus {
			return g.GameStatusProvider.Get(p, e, g)
		})

	g.addToHistory(fmt.Sprintf("is %s's turn", p.Name()))

	return status, nil
}

func (g *Game) IsGameOver() (bool, string) {
	return g.gameOver, g.winner
}

func (g *Game) drawCards(p ports.Player, count int) (cards []ports.Card, err error) {

	if !p.CanTakeCards(count) {
		return nil, ErrHandLimitExceeded
	}

	cards = make([]ports.Card, 0, count)
	for i := 0; i < count; i++ {
		c, ok := g.deck.DrawCard()
		if !ok {
			g.shuffleDiscardPileIntoDeck()

			c, ok = g.deck.DrawCard()
			if !ok {
				return nil, errors.New("no cards left to draw")
			}
		}

		cards = append(cards, c)
	}

	return cards, nil
}

func (g *Game) shuffleDiscardPileIntoDeck() {
	g.deck.Replenish(g.discardPile.Empty())
}

func (g *Game) addToHistory(msg string) {
	if len(msg) == 0 {
		return
	}

	g.history = append(g.history, msg)
}

func (g *Game) GetHistory() []string {
	if g.historyTracker == 0 {
		g.historyTracker = len(g.history)
		return g.history
	}
	newMessages := g.history[g.historyTracker:]
	g.historyTracker = len(g.history)
	return newMessages
}

func (g *Game) OnCardMovedToPile(card ports.Card) {
	g.discardPile.Discard(card)
}

func (g *Game) OnWarriorMovedToCemetery(warrior ports.Warrior) {
	g.cemetery.AddCorp(warrior)
	g.addToHistory("warrior buried in cemetery")
}

func (g *Game) OnCastleCompletion(p ports.Player) {
	g.gameOver = true
	g.winner = p.Name()
}

func (g *Game) OnFieldWithoutWarriors() {
	g.gameOver = true
	p, _ := g.WhoIsCurrent()
	g.winner = p.Name()
}

func (g *Game) switchTurn() {
	g.hasMovedWarrior = false
	g.hasTraded = false
	g.currentAction = types.ActionTypeDrawCard
	g.CurrentTurn = (g.CurrentTurn + 1) % len(g.Players)
}

func (g *Game) CurrentAction() types.ActionType {
	return g.currentAction
}

func (g *Game) nextAction(expectedAction types.ActionType,
	gameStatusFn func() GameStatus) GameStatus {

	p, enemy := g.WhoIsCurrent()
	g.CanMoveWarrior = !g.hasMovedWarrior && p.HasWarriorsInHand()
	g.CanTrade = !g.hasTraded && p.CanTradeCards()

	if expectedAction == types.ActionTypeAttack {
		// Check if player can attack with weapons OR catapult
		canAttackWithCatapult := p.HasCatapult() && enemy.Castle().CanBeAttacked()

		if p.CanAttack() || canAttackWithCatapult || g.CanMoveWarrior {
			g.currentAction = types.ActionTypeAttack
			return gameStatusFn()
		}
		expectedAction = types.ActionTypeSpySteal
	}
	if expectedAction == types.ActionTypeSpySteal {
		if p.HasSpy() || p.HasThief() {
			g.currentAction = types.ActionTypeSpySteal
			return gameStatusFn()
		}
		expectedAction = types.ActionTypeBuy
	}
	if expectedAction == types.ActionTypeBuy {
		if p.CanBuy() {
			g.currentAction = types.ActionTypeBuy
			return gameStatusFn()
		}
		expectedAction = types.ActionTypeConstruct
	}
	if expectedAction == types.ActionTypeConstruct {
		if p.CanConstruct() {
			g.currentAction = types.ActionTypeConstruct
			return gameStatusFn()
		}
		expectedAction = types.ActionTypeEndTurn
	}

	if expectedAction == types.ActionTypeEndTurn {
		g.currentAction = types.ActionTypeEndTurn
		return gameStatusFn()
	}

	g.currentAction = types.ActionTypeDrawCard
	return gameStatusFn()
}

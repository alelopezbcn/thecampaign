package domain

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

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
	CanTrade           bool
	state              GameState
	deck               ports.Deck
	discardPile        []ports.Card
	cemetery           []ports.Warrior
	history            []string
	dealer             ports.Dealer
	GameStatusProvider GameStatusProvider
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
		discardPile:        []ports.Card{},
		cemetery:           []ports.Warrior{},
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
	g.addToHistory("Dealing Cards")

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

	g.state = StateSettingInitialWarriors
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
		return errors.New(fmt.Sprintf("%s not your turn", playerName))
	}

	if g.state != StateSettingInitialWarriors {
		return errors.New("not in initial warrior setting phase")
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
	g.addToHistory(p.Name() + " has set their initial warriors.")
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
		g.state = StateWaitingDraw
		g.addToHistory("Both players have set their initial warriors.")
		return nil
	}

	return nil
}

func (g *Game) WhoIsCurrent() (current ports.Player, enemy ports.Player) {
	return g.Players[g.CurrentTurn],
		g.Players[(g.CurrentTurn+1)%len(g.Players)]
}

func (g *Game) DrawCard(playerName string) (status gamestatus.GameStatus, err error) {
	p, e := g.WhoIsCurrent()
	if p.Name() != playerName {
		return status, fmt.Errorf("%s not your turn", playerName)
	}

	cards, err := g.drawCards(p, 1)
	if err != nil && !errors.Is(err, ErrHandLimitExceeded) {
		return status, err
	}

	if err == nil {
		p.TakeCards(cards...)
		g.addToHistory(fmt.Sprintf("%s drew %d card(s).", p.Name(), 1))
	} else {
		g.addToHistory(fmt.Sprintf("%s can't take more cards.", p.Name()))
	}

	g.currentAction = types.ActionTypeAttack

	status = g.GameStatusProvider.Get(p, e, g.currentAction, g.CanMoveWarrior, g.CanTrade, cards...)

	return status, nil
}

func (g *Game) MoveWarriorToField(playerName, warriorID string) (
	status gamestatus.GameStatus, err error) {

	p, e := g.WhoIsCurrent()
	if p.Name() != playerName {
		return status, errors.New(fmt.Sprintf("%s not your turn", playerName))
	}

	err = p.MoveCardToField(warriorID)
	if err != nil {
		return status, fmt.Errorf("moving warrior to field failed: %w", err)
	}

	g.addToHistory(fmt.Sprintf("%s moved warrior %s to field",
		p.Name(), warriorID))
	g.CanMoveWarrior = false
	status = g.GameStatusProvider.Get(p, e, g.currentAction, g.CanMoveWarrior, g.CanTrade)

	return status, nil
}

func (g *Game) Trade(playerName string, cardIDs []string) (
	status gamestatus.GameStatus, err error) {

	p, e := g.WhoIsCurrent()
	if p.Name() != playerName {
		return status, errors.New(fmt.Sprintf("%s not your turn", playerName))
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

	cards, err := g.drawCards(p, 1)
	if err != nil {
		return status, fmt.Errorf("drawing card for trading failed: %w", err)
	}

	p.TakeCards(cards...)

	g.addToHistory(fmt.Sprintf("%s traded cards", p.Name()))
	g.CanTrade = false
	status = g.GameStatusProvider.Get(p, e, g.currentAction, g.CanMoveWarrior, g.CanTrade, cards...)

	return status, nil
}

func (g *Game) Attack(playerName, targetID, weaponID string) (
	status gamestatus.GameStatus, err error) {

	p, e := g.WhoIsCurrent()
	if p.Name() != playerName {
		return status, errors.New(fmt.Sprintf("%s not your turn", playerName))
	}

	targetCard, ok := e.GetCardFromField(targetID)
	if !ok {
		return status, errors.New("target card not in enemy field: " + targetID)
	}

	weaponCard, ok := p.GetCardFromHand(weaponID)
	if !ok {
		return status, errors.New("weapon card not in hand: " + weaponID)
	}

	if err := p.Attack(targetCard, weaponCard); err != nil {
		return status, fmt.Errorf("attack action failed: %w", err)
	}

	g.addToHistory(fmt.Sprintf("%s\nwas attacked with \n%s",
		targetCard.String(), targetCard.String()))

	g.currentAction = types.ActionTypeSpySteal
	status = g.GameStatusProvider.Get(p, e, g.currentAction, g.CanMoveWarrior, g.CanTrade)
	return status, nil
}

func (g *Game) SpecialPower(playerName, userID, targetID, weaponID string) (
	status gamestatus.GameStatus, err error) {

	p, e := g.WhoIsCurrent()
	if p.Name() != playerName {
		return status, errors.New(fmt.Sprintf("%s not your turn", playerName))
	}

	warriorCard, ok := p.GetCardFromField(userID)
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

	if err := p.UseSpecialPower(warriorCard, targetCard, weaponCard); err != nil {
		return status, fmt.Errorf("special power action failed: %w", err)
	}

	g.addToHistory(fmt.Sprintf("%s\nattacked\n%s",
		warriorCard.String(), targetCard.String()))

	g.currentAction = types.ActionTypeSpySteal
	status = g.GameStatusProvider.Get(p, e, g.currentAction, g.CanMoveWarrior, g.CanTrade)

	return status, nil
}

func (g *Game) Catapult(playerName string, cardPosition int) (
	status gamestatus.GameStatus, err error) {

	p, e := g.WhoIsCurrent()
	if p.Name() != playerName {
		return status, errors.New(fmt.Sprintf("%s not your turn", playerName))
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

	g.addToHistory(fmt.Sprintf("%s used catapult to steal gold from %s's castle",
		p.Name(), e.Name()))

	g.currentAction = types.ActionTypeSpySteal
	status = g.GameStatusProvider.Get(p, e, g.currentAction, g.CanMoveWarrior, g.CanTrade)

	return status, nil
}

// TODO: Spy debe devolver UI cards
func (g *Game) Spy(playerName string, option int) (spiedCards []gamestatus.Card,
	status gamestatus.GameStatus, err error) {

	p, e := g.WhoIsCurrent()
	if p.Name() != playerName {
		return nil, status, errors.New(fmt.Sprintf("%s not your turn", playerName))
	}

	s := p.Spy()
	if s == nil {
		return nil, status, errors.New("player does not have a Spy to use")
	}

	g.OnCardMovedToPile(s)

	g.currentAction = types.ActionTypeBuy
	status = g.GameStatusProvider.Get(p, e, g.currentAction, g.CanMoveWarrior, g.CanTrade)

	switch option {
	case 1:
		// Reveal top 5 cards from deck
		g.addToHistory(fmt.Sprintf("%s spied top 5 cards from deck", p.Name()))

		return gamestatus.FromDomainCards(g.deck.Reveal(5)), status, nil
	case 2:
		// Reveal enemy's cards
		g.addToHistory(fmt.Sprintf("%s spied on %s's hand",
			p.Name(), e.Name()))

		return gamestatus.FromDomainCards(e.Hand().ShowCards()), status, nil
	default:
		return nil, status, errors.New("invalid Spy option")
	}
}

func (g *Game) Steal(playerName string, cardPosition int) (
	status gamestatus.GameStatus, err error) {

	p, e := g.WhoIsCurrent()
	if p.Name() != playerName {
		return status, errors.New(fmt.Sprintf("%s not your turn", playerName))
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
	g.currentAction = types.ActionTypeBuy

	g.addToHistory(fmt.Sprintf("%s stole a card from %s",
		p.Name(), e.Name()))

	status = g.GameStatusProvider.Get(p, e, g.currentAction, g.CanMoveWarrior, g.CanTrade, stolenCard)

	return status, nil

}

func (g *Game) Buy(playerName, cardID string) (
	status gamestatus.GameStatus, err error) {

	p, e := g.WhoIsCurrent()
	if p.Name() != playerName {
		return status, errors.New(fmt.Sprintf("%s not your turn", playerName))
	}

	resourceCard, ok := p.GetCardFromHand(cardID)
	if !ok {
		return status, errors.New("Resource card not in hand: " + cardID)
	}

	r, ok := resourceCard.(ports.Resource)
	if !ok {
		return status, errors.New("only gold cards can be used to buy")
	}

	if !r.CanBuy() {
		return status, errors.New("cannot buy with gold card")
	}

	val := r.Value()
	p.GiveCards(resourceCard.GetID())
	g.OnCardMovedToPile(resourceCard)

	cardsToBuy := val / 2
	cards, err := g.drawCards(p, cardsToBuy)
	if err != nil {
		return status, fmt.Errorf("drawing card for buying failed: %w", err)
	}

	p.TakeCards(cards...)

	g.addToHistory(fmt.Sprintf("%s bought %d card(s) using %s",
		p.Name(), cardsToBuy, resourceCard.String()))

	g.currentAction = types.ActionTypeConstruct

	status = g.GameStatusProvider.Get(p, e, g.currentAction, g.CanMoveWarrior, g.CanTrade, cards...)

	return status, nil
}

func (g *Game) Construct(playerName, cardID string) (
	status gamestatus.GameStatus, err error) {

	p, e := g.WhoIsCurrent()
	if p.Name() != playerName {
		return status, errors.New(fmt.Sprintf("%s not your turn", playerName))
	}

	if err := p.Construct(cardID); err != nil {
		return status, fmt.Errorf("constructing card failed: %w", err)
	}

	g.addToHistory(fmt.Sprintf("%s constructed castle with card %s",
		p.Name(), cardID))

	g.currentAction = types.ActionTypeEndTurn
	status = g.GameStatusProvider.Get(p, e, g.currentAction, g.CanMoveWarrior, g.CanTrade)
	return status, nil

}

func (g *Game) EndTurn(player string) (status gamestatus.GameStatus, err error) {
	p, _ := g.WhoIsCurrent()
	if p.Name() != player {
		return status, errors.New(fmt.Sprintf("%s not your turn", player))
	}

	g.switchTurn()
	p, e := g.WhoIsCurrent()
	g.currentAction = types.ActionTypeDrawCard
	status = g.GameStatusProvider.Get(p, e, g.currentAction, g.CanMoveWarrior, g.CanTrade)

	return status, nil
}

func (g *Game) IsGameEnded() bool {
	return g.state == StateGameEnded
}

func (g *Game) drawCards(p ports.Player, count int) (cards []ports.Card, err error) {

	if !p.CanTakeCards(count) {
		g.addToHistory(p.Name() + " exceeded max number of cards in hand.")

		return nil, ErrHandLimitExceeded
	}

	cards = make([]ports.Card, 0, count)
	for i := 0; i < count; i++ {
		c, ok := g.deck.DrawCard()
		if !ok {
			g.addToHistory("deck is empty, shuffling discard pile into deck")
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
	g.deck.Replenish(g.discardPile)
	g.discardPile = []ports.Card{}
	g.addToHistory("Shuffled discard pile into deck")
}

func (g *Game) addToHistory(msg string) {
	if len(msg) == 0 {
		return
	}

	g.history = append(g.history, msg)
	log.Print(fmt.Sprintf("***********: %s %s", time.Now().Format("2006-01-02 15:04:05"), msg))
}

func (g *Game) OnCardMovedToPile(card ports.Card) {
	g.discardPile = append(g.discardPile, card)
	g.addToHistory(fmt.Sprintf("card moved to discard pile (%d): %s",
		len(g.discardPile), card.String()))
}

func (g *Game) OnWarriorMovedToCemetery(warrior ports.Warrior) {
	g.cemetery = append(g.cemetery, warrior)
	g.addToHistory(fmt.Sprintf("warrior died and moved to cemetery (%d): %s",
		len(g.cemetery), warrior.String()))
}

func (g *Game) OnCastleCompletion(p ports.Player) {
	g.state = StateGameEnded
	g.addToHistory(fmt.Sprintf("%s wins: Castle completed", p.Name()))
}

func (g *Game) OnFieldWithoutWarriors(p ports.Player) {
	g.state = StateGameEnded
	g.addToHistory(fmt.Sprintf("%s loses: No more warriors in field", p.Name()))
}

func (g *Game) OnMessage(msg string) {
	g.addToHistory(msg)
}

func (g *Game) switchTurn() {
	g.CanMoveWarrior = true
	g.CanTrade = true
	g.CurrentTurn = (g.CurrentTurn + 1) % len(g.Players)
}

func (g *Game) CurrentAction() types.ActionType {
	return g.currentAction
}

func (g *Game) SkipPhase(playerName string) (status gamestatus.GameStatus, err error) {
	p, e := g.WhoIsCurrent()
	if p.Name() != playerName {
		return status, errors.New(fmt.Sprintf("%s not your turn", playerName))
	}

	switch g.currentAction {
	case types.ActionTypeAttack:
		g.currentAction = types.ActionTypeSpySteal
	case types.ActionTypeSpySteal:
		g.currentAction = types.ActionTypeBuy
	case types.ActionTypeBuy:
		g.currentAction = types.ActionTypeConstruct
	case types.ActionTypeConstruct:
		g.currentAction = types.ActionTypeEndTurn
	default:
		return status, errors.New("cannot skip this phase")
	}

	g.addToHistory(fmt.Sprintf("%s skipped phase", p.Name()))
	status = g.GameStatusProvider.Get(p, e, g.currentAction, g.CanMoveWarrior, g.CanTrade)
	return status, nil
}

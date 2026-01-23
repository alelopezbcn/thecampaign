package domain

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/google/uuid"
)

type Games []Game

type Game struct {
	id          string
	Players     []ports.Player
	CurrentTurn int
	state       GameState
	deck        ports.Deck
	discardPile []ports.Card
	cemetery    []ports.Warrior
	history     []string
	dealer      ports.Dealer
}

func NewGame(player1, player2 string,
	dealer ports.Dealer) *Game {
	playersArr := []string{player1, player2}
	rand.Shuffle(len(playersArr), func(i, j int) {
		playersArr[i], playersArr[j] = playersArr[j], playersArr[i]
	})

	g := &Game{
		id:          uuid.NewString(),
		CurrentTurn: 0,
		discardPile: []ports.Card{},
		cemetery:    []ports.Warrior{},
		history:     []string{},
		dealer:      dealer,
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

func (g *Game) DrawCards(playerName string, count int) (err error) {
	p, _ := g.WhoIsCurrent()
	if p.Name() != playerName {
		return errors.New(fmt.Sprintf("%s not your turn", playerName))
	}

	if !p.CanTakeCards(count) {
		g.addToHistory(p.Name() + " exceeded max number of cards in hand.")

		return ErrHandLimitExceeded
	}

	cards := make([]ports.Card, 0, count)
	for i := 0; i < count; i++ {
		c, ok := g.deck.DrawCard()
		if !ok {
			g.addToHistory("deck is empty, shuffling discard pile into deck")
			g.shuffleDiscardPileIntoDeck()

			c, ok = g.deck.DrawCard()
			if !ok {
				return errors.New("no cards left to draw")
			}
		}

		cards = append(cards, c)
	}

	p.TakeCards(cards...)

	g.addToHistory(fmt.Sprintf("%s drew %d card(s).", p.Name(), count))

	return nil
}

func (g *Game) GetStatusForNextPlayer() (status GameStatus) {
	player, enemy := g.WhoIsCurrent()
	return NewGameStatus(player, enemy)
}

func (g *Game) Attack(playerName, targetID, weaponID string) error {
	current, enemy := g.WhoIsCurrent()
	if current.Name() != playerName {
		return errors.New(fmt.Sprintf("%s not your turn", playerName))
	}

	targetCard, ok := enemy.GetCardFromField(targetID)
	if !ok {
		return errors.New("target card not in enemy field: " + targetID)
	}

	weaponCard, ok := current.GetCardFromHand(weaponID)
	if !ok {
		return errors.New("weapon card not in hand: " + weaponID)
	}

	if err := current.Attack(targetCard, weaponCard); err != nil {
		return fmt.Errorf("attack action failed: %w", err)
	}

	g.addToHistory(fmt.Sprintf("%s\nwas attacked with \n%s",
		targetCard.String(), targetCard.String()))

	return nil
}

func (g *Game) MoveWarriorToField(playerName, warriorID string) error {
	current, _ := g.WhoIsCurrent()
	if current.Name() != playerName {
		return errors.New(fmt.Sprintf("%s not your turn", playerName))
	}

	err := current.MoveCardToField(warriorID)
	if err != nil {
		return fmt.Errorf("moving warrior to field failed: %w", err)
	}

	g.addToHistory(fmt.Sprintf("%s moved warrior %s to field",
		current.Name(), warriorID))

	return nil
}

func (g *Game) Trade(playerName string, cardIDs []string) error {
	current, _ := g.WhoIsCurrent()
	if current.Name() != playerName {
		return errors.New(fmt.Sprintf("%s not your turn", playerName))
	}

	if len(cardIDs) != 3 {
		return errors.New("must trade exactly 3 cards")
	}

	cards, err := current.GiveCards(cardIDs...)
	if err != nil {
		return fmt.Errorf("giving cards for trade failed: %w", err)
	}
	for _, c := range cards {
		g.discardPile = append(g.discardPile, c)
	}

	if err := g.DrawCards(playerName, 1); err != nil {
		return fmt.Errorf("drawing card for trading failed: %w", err)
	}

	g.addToHistory(fmt.Sprintf("%s traded cards %v",
		current.Name(), cardIDs))

	return nil
}

func (g *Game) Buy(playerName, cardID string) error {
	current, _ := g.WhoIsCurrent()
	if current.Name() != playerName {
		return errors.New(fmt.Sprintf("%s not your turn", playerName))
	}

	resourceCard, ok := current.GetCardFromHand(cardID)
	if !ok {
		return errors.New("Resource card not in hand: " + cardID)
	}

	r, ok := resourceCard.(ports.Resource)
	if !ok {
		return errors.New("only gold cards can be used to buy")
	}

	val := r.Value()
	if val == 1 {
		return errors.New("cannot buy with gold card of value 1")
	}

	g.OnCardMovedToPile(resourceCard)

	cardsToBuy := val / 2
	if err := g.DrawCards(playerName, cardsToBuy); err != nil {
		return fmt.Errorf("drawing card for buying failed: %w", err)
	}

	g.addToHistory(fmt.Sprintf("%s bought %d card(s) using %s",
		current.Name(), cardsToBuy, resourceCard.String()))

	return nil

}

func (g *Game) SpecialPower(playerName, userID, targetID, weaponID string) error {
	current, enemy := g.WhoIsCurrent()
	if current.Name() != playerName {
		return errors.New(fmt.Sprintf("%s not your turn", playerName))
	}

	warriorCard, ok := current.GetCardFromField(userID)
	if !ok {
		return errors.New("warrior card not in field: " + userID)
	}

	var targetCard ports.Card
	targetCard, ok = current.GetCardFromField(targetID)
	if !ok {
		targetCard, ok = enemy.GetCardFromField(targetID)
		if !ok {
			return errors.New("target card not valid: " + targetID)
		}
	}

	weaponCard, ok := current.GetCardFromHand(weaponID)
	if !ok {
		return errors.New("weapon card not in hand: " + weaponID)
	}

	if err := current.UseSpecialPower(warriorCard, targetCard, weaponCard); err != nil {
		return fmt.Errorf("special power action failed: %w", err)
	}

	g.addToHistory(fmt.Sprintf("%s\nattacked\n%s",
		warriorCard.String(), targetCard.String()))

	return nil
}

func (g *Game) Construct(playerName, cardID string) error {
	current, _ := g.WhoIsCurrent()
	if current.Name() != playerName {
		return errors.New(fmt.Sprintf("%s not your turn", playerName))
	}

	if err := current.Construct(cardID); err != nil {
		return fmt.Errorf("constructing card failed: %w", err)
	}

	g.addToHistory(fmt.Sprintf("%s constructed castle with card %s",
		current.Name(), cardID))

	return nil

}

func (g *Game) Spy(playerName string, option int) ([]ports.Card, error) {
	current, enemy := g.WhoIsCurrent()
	if current.Name() != playerName {
		return nil, errors.New(fmt.Sprintf("%s not your turn", playerName))
	}

	s := current.Spy()
	if s == nil {
		return nil, errors.New("player does not have a Spy to use")
	}

	g.OnCardMovedToPile(s)

	switch option {
	case 1:
		// Reveal top 5 cards from deck
		g.addToHistory(fmt.Sprintf("%s spied top 5 cards from deck", current.Name()))

		return g.deck.Reveal(5), nil
	case 2:
		// Reveal enemy's cards
		g.addToHistory(fmt.Sprintf("%s spied on %s's hand",
			current.Name(), enemy.Name()))

		return enemy.Hand().ShowCards(), nil
	default:
		return nil, errors.New("invalid Spy option")
	}
}

func (g *Game) Steal(playerName string, cardPosition int) error {
	current, enemy := g.WhoIsCurrent()
	if current.Name() != playerName {
		return errors.New(fmt.Sprintf("%s not your turn", playerName))
	}

	t := current.Thief()
	if t == nil {
		return errors.New("player does not have a thief to steal with")
	}

	stolenCard, err := enemy.CardStolenFromHand(cardPosition)
	if err != nil {
		return fmt.Errorf("stealing card failed: %w", err)
	}

	g.OnCardMovedToPile(t)
	current.TakeCards(stolenCard)

	g.addToHistory(fmt.Sprintf("%s stole a card from %s",
		current.Name(), enemy.Name()))

	return nil

}

func (g *Game) Catapult(playerName string, cardPosition int) error {
	current, enemy := g.WhoIsCurrent()
	if current.Name() != playerName {
		return errors.New(fmt.Sprintf("%s not your turn", playerName))
	}

	t := current.Catapult()
	if t == nil {
		return errors.New("player does not have a catapult to attack")
	}

	stolenGold, err := t.Attack(enemy.Castle(), cardPosition)
	if err != nil {
		return fmt.Errorf("attacking castle failed: %w", err)
	}

	g.OnCardMovedToPile(stolenGold)

	g.addToHistory(fmt.Sprintf("%s used catapult to steal gold from %s's castle",
		current.Name(), enemy.Name()))

	return nil

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

func (g *Game) EndTurn(player string) error {
	next, _ := g.WhoIsCurrent()
	if next.Name() != player {
		return errors.New(fmt.Sprintf("%s not your turn", player))
	}

	g.addToHistory(player + " ended their turn")
	g.switchTurn()

	return nil
}

func (g *Game) IsGameEnded() bool {
	return g.state == StateGameEnded
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
	g.CurrentTurn = (g.CurrentTurn + 1) % len(g.Players)
}

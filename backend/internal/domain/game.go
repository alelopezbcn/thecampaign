package domain

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Games []Game

type Game struct {
	id          string
	Players     []*Player
	CurrentTurn int
	state       GameState
	deck        Deck
	discardPile []iCard
	cemetery    []iCard
	history     []string
}

func NewGame(player1, player2 string) *Game {
	playersArr := []string{player1, player2}
	rand.Shuffle(len(playersArr), func(i, j int) {
		playersArr[i], playersArr[j] = playersArr[j], playersArr[i]
	})

	g := &Game{
		id:          uuid.NewString(),
		CurrentTurn: 0,
		discardPile: []iCard{},
		cemetery:    []iCard{},
		history:     []string{},
	}

	p1 := NewPlayer(playersArr[0], g, g)
	p2 := NewPlayer(playersArr[1], g, g)
	g.Players = []*Player{p1, p2}

	g.addToHistory(fmt.Sprintf("Game created between %s and %s", p1.Name, p2.Name))

	g.deal()

	return g
}

func (g *Game) deal() {
	g.addToHistory("Dealing Cards")

	warriorCards := shuffle(warriorsCards(g))

	// Each player gets 3 warrior cards
	warriorsIdx := 0
	for _, player := range g.Players {
		player.takeCards(warriorCards[warriorsIdx : warriorsIdx+3]...)
		warriorsIdx += 3
	}

	deckCards := append(warriorCards[warriorsIdx:], otherButWarriorsCards(g)...)
	deckCards = shuffle(deckCards)
	otherIdx := 0
	for _, player := range g.Players {
		player.takeCards(deckCards[otherIdx : otherIdx+3]...)
		otherIdx += 4
	}

	deckCards = deckCards[otherIdx:]
	g.deck = NewDeck(deckCards)

	g.state = StateSettingInitialWarriors
}

func (g *Game) SetInitialWarriors(playerName string, warriorIDs []string) error {
	player, _ := g.WhoIsCurrent()
	if player.Name != playerName {
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
		if err := player.moveCardToField(strings.TrimSpace(id)); err != nil {
			return err
		}
	}
	g.addToHistory(player.Name + " has set their initial warriors.")
	g.switchTurn()

	// Check if both players have set their warriors
	allSet := true
	for _, p := range g.Players {
		if len(p.field.showCards()) == 0 {
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

func (g *Game) switchTurn() {
	g.CurrentTurn = (g.CurrentTurn + 1) % len(g.Players)
}

func (g *Game) WhoIsCurrent() (current *Player, enemy *Player) {
	return g.Players[g.CurrentTurn],
		g.Players[(g.CurrentTurn+1)%len(g.Players)]
}

func (g *Game) WhoIsEnemy() *Player {
	return g.Players[(g.CurrentTurn+1)%len(g.Players)]
}

func (g *Game) DrawCard(playerName string) (card iCard, err error) {
	player, _ := g.WhoIsCurrent()
	if player.Name != playerName {
		return card, errors.New(fmt.Sprintf("%s not your turn", playerName))
	}

	if !player.hand.canAddCards(1) {
		g.addToHistory(player.Name + " exceeded max number of cards in hand.")

		return nil, ErrHandLimitExceeded
	}

	card, ok := g.deck.DrawCard()
	if !ok {
		g.addToHistory("Deck is empty, shuffling discard pile into deck")
		g.shuffleDiscardPileIntoDeck()

		card, ok = g.deck.DrawCard()
		if !ok {
			return card, errors.New("no cards left to draw")
		}
	}

	player.takeCards(card)

	g.addToHistory(player.Name + " drew a Card")

	return card, nil
}

func (g *Game) GetStatusForNextPlayer() (status BoardStatus) {
	player, enemy := g.WhoIsCurrent()
	status.Player = player.Name
	status.Hand = player.ShowHand()
	status.OwnField = player.ShowField()
	status.OwnCastle = player.ShowCastle()

	status.EnemyField = enemy.ShowField()
	status.EnemyCastle = enemy.ShowCastle()

	return status
}

func (g *Game) Attack(playerName, warriorID, targetID, weaponID string) error {
	next, enemy := g.WhoIsCurrent()
	if next.Name != playerName {
		return errors.New(fmt.Sprintf("%s not your turn", playerName))
	}

	warriorCard, ok := next.GetCardFromField(warriorID)
	if !ok {
		return errors.New("warrior card not in field: " + warriorID)
	}

	targetCard, ok := enemy.GetCardFromField(targetID)
	if !ok {
		return errors.New("target card not in enemy field: " + targetID)
	}

	weaponCard, ok := next.GetCardFromHand(weaponID)
	if !ok {
		return errors.New("weapon card not in hand: " + weaponID)
	}

	if err := next.Attack(warriorCard, targetCard, weaponCard); err != nil {
		return fmt.Errorf("attack action failed: %w", err)
	}

	g.addToHistory(fmt.Sprintf("%s attacked %s using %s",
		warriorCard.String(), targetCard.String(), weaponCard.String()))

	return nil

}

func (g *Game) shuffleDiscardPileIntoDeck() {
	g.deck.Replenish(g.discardPile)
	g.discardPile = []iCard{}
	g.addToHistory("Shuffled discard pile into deck")
}

func (g *Game) addToHistory(msg string) {
	if len(msg) == 0 {
		return
	}

	g.history = append(g.history, msg)
	println(fmt.Sprintf("*********: %s %s", time.Now().Format("2006-01-02 15:04:05"), msg))
}

func (g *Game) EndTurn(player string) error {
	next, _ := g.WhoIsCurrent()
	if next.Name != player {
		return errors.New(fmt.Sprintf("%s not your turn", player))
	}

	g.addToHistory(player + " ended their turn")
	g.switchTurn()

	return nil
}

func (g *Game) IsGameEnded() bool {
	return g.state == StateGameEnded
}

func (g *Game) OnCardUsed(player *Player, card iCard) {
	player.removeCardFromHand(card)
	g.discardPile = append(g.discardPile, card)
	g.addToHistory("Card moved to discard pile: " + card.String())
}

func (g *Game) OnWarriorDead(player *Player, card iCard) {
	player.removeCardFromField(card)
	g.cemetery = append(g.cemetery, card)
	g.addToHistory("Warrior died and moved to cemetery: " + card.String())
}

func (g *Game) OnGameEnded(reason string) {
	g.state = StateGameEnded
	current, _ := g.WhoIsCurrent()
	g.addToHistory(fmt.Sprintf("%s wins: %s", current.Name, reason))
}

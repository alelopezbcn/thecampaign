package domain

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
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
	discardPile []card
	cemetery    []card
	history     []string
}

func NewGame(player1, player2 string) *Game {
	playersArr := []string{player1, player2}
	rand.Shuffle(len(playersArr), func(i, j int) {
		playersArr[i], playersArr[j] = playersArr[j], playersArr[i]
	})

	p1 := NewPlayer(playersArr[0])
	p2 := NewPlayer(playersArr[1])

	g := &Game{
		id:          uuid.NewString(),
		Players:     []*Player{p1, p2},
		CurrentTurn: 0,
		discardPile: []card{},
		cemetery:    []card{},
		history:     []string{},
	}

	g.deal()

	return g
}

func (g *Game) deal() {
	g.addToHistory("Dealing Cards")

	warriorCards := shuffle(WarriorsCards())

	// Each player gets 3 warrior cards
	warriorsIdx := 0
	for _, player := range g.Players {
		player.Hand = append(player.Hand, warriorCards[warriorsIdx:warriorsIdx+3]...)
		warriorsIdx += 3
	}

	deckCards := append(warriorCards[warriorsIdx:], OtherButWarriorsCards()...)
	deckCards = shuffle(deckCards)
	otherIdx := 0
	for _, player := range g.Players {
		player.Hand = append(player.Hand, deckCards[otherIdx:otherIdx+4]...)
		otherIdx += 4
	}

	deckCards = deckCards[otherIdx:]
	g.deck = Deck{Cards: deckCards}

	g.state = StateSettingInitialWarriors
	g.addToHistory("Set initial warriors: " + g.Players[0].Name + " goes first.")
}

func (g *Game) SetInitialWarriors(playerID string, warriorIDs []string) error {
	if g.Players[g.CurrentTurn].ID != playerID {
		return errors.New("not your turn")
	}

	player := g.Players[g.CurrentTurn]
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
		if !player.MoveWarriorToField(id) {
			return errors.New("failed to move warrior to field: " + id)
		}
	}
	g.addToHistory(player.Name + " has set their initial warriors.")
	g.switchTurn()

	// Check if both players have set their warriors
	allSet := true
	for _, p := range g.Players {
		if len(p.Field) == 0 {
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
	g.addToHistory("It's now " + g.Players[g.CurrentTurn].Name + "'s turn.")
}

func (g *Game) WhoIsNext() string {
	return g.Players[g.CurrentTurn].Name
}

func (g *Game) HandleAction(playerID string, action string,
	source Card, destination Card) error {

	if g.Players[g.CurrentTurn].ID != playerID {
		return errors.New("not your turn")
	}

	switch action {
	case "take_card":
	}

	return nil
}

func (g *Game) drawCard() error {
	player := g.Players[g.CurrentTurn]
	card, ok := g.deck.DrawCard()
	if !ok {
		g.addToHistory("Deck is empty, shuffling discard pile into deck")
		g.shuffleDiscardIntoDeck()

		card, ok = g.deck.DrawCard()
		if !ok {
			return errors.New("no cards left to draw")
		}
	}

	player.Hand = append(player.Hand, card)
	log.Println(player.Name + " drew a card: " + card.Name)
	return nil
}

func (g *Game) shuffleDiscardIntoDeck() {
	g.deck.Replenish(g.discardPile)
	g.discardPile = []card{}
	g.addToHistory("Shuffled discard pile into deck")
}

func (g *Game) addToHistory(msg string) {
	g.history = append(g.history, msg)
	println(fmt.Sprintf("%s %s", time.Now(), msg))
}

package domain

import (
	"errors"
	"fmt"
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
	discardPile []Card
	cemetery    []Card
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
		discardPile: []Card{},
		cemetery:    []Card{},
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
		player.takeCards(warriorCards[warriorsIdx : warriorsIdx+3]...)
		warriorsIdx += 3
	}

	deckCards := append(warriorCards[warriorsIdx:], OtherButWarriorsCards()...)
	deckCards = shuffle(deckCards)
	otherIdx := 0
	for _, player := range g.Players {
		player.takeCards(deckCards[otherIdx : otherIdx+3]...)
		otherIdx += 4
	}

	deckCards = deckCards[otherIdx:]
	g.deck = NewDeck(deckCards)

	g.state = StateSettingInitialWarriors
	g.addToHistory("Set initial warriors: " + g.Players[0].Name + " goes first.")
}

func (g *Game) SetInitialWarriors(playerName string, warriorIDs []string) error {
	player := g.WhoIsNext()
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
		if !player.moveWarriorToField(id) {
			return errors.New("failed to move warrior to field: " + id)
		}
	}
	g.addToHistory(player.Name + " has set their initial warriors.")
	g.switchTurn()

	// Check if both players have set their warriors
	allSet := true
	for _, p := range g.Players {
		if len(p.field) == 0 {
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

func (g *Game) WhoIsNext() *Player {
	return g.Players[g.CurrentTurn]
}

func (g *Game) WhoIsEnemy() *Player {
	return g.Players[(g.CurrentTurn+1)%len(g.Players)]
}

// func (g *Game) HandleAction(playerName string, action string,
// 	source Card, destination Card) error {
// 	player := g.WhoIsNext()
// 	if player.Name != playerName {
// 		return errors.New(fmt.Sprintf("%s not your turn", playerName))
// 	}
//
//
// 	switch action {
// 	case "take_card":
// 	}
//
// 	return nil
// }

func (g *Game) DrawCard(playerName string) (status ActionReadyStatus, err error) {
	player := g.WhoIsNext()
	if player.Name != playerName {
		return status, errors.New(fmt.Sprintf("%s not your turn", playerName))
	}

	card, ok := g.deck.DrawCard()
	if !ok {
		g.addToHistory("Deck is empty, shuffling discard pile into deck")
		g.shuffleDiscardPileIntoDeck()

		card, ok = g.deck.DrawCard()
		if !ok {
			return status, errors.New("no cards left to draw")
		}
	}

	player.takeCards(card)
	status.Player = playerName
	status.CardTaken = card
	status.Hand = player.ShowHand()
	status.OwnField = player.ShowField()
	status.OwnCastle = player.ShowCastle()

	enemy := g.WhoIsEnemy()
	status.EnemyField = enemy.ShowField()
	status.EnemyCastle = enemy.ShowCastle()

	g.addToHistory(player.Name + " drew a Card: " + card.String())

	return status, nil
}

func (g *Game) shuffleDiscardPileIntoDeck() {
	g.deck.Replenish(g.discardPile)
	g.discardPile = []Card{}
	g.addToHistory("Shuffled discard pile into deck")
}

func (g *Game) addToHistory(msg string) {
	g.history = append(g.history, msg)
	println(fmt.Sprintf("*********: %s %s", time.Now().Format("2006-01-02 15:04:05"), msg))
}

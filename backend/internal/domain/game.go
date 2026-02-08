package domain

import (
	"errors"
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/google/uuid"
)

type Games []Game

type Game struct {
	id                 string
	Mode               types.GameMode
	Players            []ports.Player
	Teams              map[int][]int // teamID -> player indices (2v2 only)
	EliminatedPlayers  map[int]bool  // player index -> eliminated (FFA only)
	CurrentTurn        int
	currentAction      types.ActionType
	CanMoveWarrior     bool
	hasMovedWarrior    bool
	CanTrade           bool
	hasTraded          bool
	deck               ports.Deck
	discardPile        ports.DiscardPile
	cemetery           ports.Cemetery
	dealer             ports.Dealer
	GameStatusProvider GameStatusProvider
	history            []string
	historyTracker     int
	gameOver           bool
	winner             string
}

func NewGame(playerNames []string, mode types.GameMode, dealer ports.Dealer,
	gameStatusProvider GameStatusProvider) (*Game, error) {

	if err := validatePlayers(playerNames, mode); err != nil {
		return nil, err
	}

	g := &Game{
		id:                 uuid.NewString(),
		CurrentTurn:        0,
		discardPile:        newDiscardPile(),
		cemetery:           newCemetery(),
		history:            []string{},
		dealer:             dealer,
		GameStatusProvider: gameStatusProvider,
		Players:            make([]ports.Player, len(playerNames)),
		Mode:               mode,
		EliminatedPlayers:  make(map[int]bool),
	}

	if mode == types.GameMode2v2 {
		g.Teams = map[int][]int{
			1: {0, 2}, // Team 1: Player 1 and Player 3
			2: {1, 3}, // Team 2: Player 2 and Player 4
		}
	}

	for i, name := range playerNames {
		p := NewPlayer(name, i, g, g, g, g)
		g.Players[i] = p
	}

	g.deal()

	return g, nil
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

func validatePlayers(playerNames []string, mode types.GameMode) error {
	switch mode {
	case types.GameMode1v1:
		if len(playerNames) != 2 {
			return errors.New("1v1 mode requires 2 players")
		}
	case types.GameMode2v2:
		if len(playerNames) != 4 {
			return errors.New("2v2 mode requires 4 players")
		}
	case types.GameModeFFA3:
		if len(playerNames) != 3 {
			return errors.New("FFA3 mode requires 3 players")
		}
	case types.GameModeFFA5:
		if len(playerNames) != 5 {
			return errors.New("FFA mode requires 5 players")
		}
	default:
		return errors.New("invalid game mode")
	}

	return nil
}

func (g *Game) deal() {
	warriorCards := shuffle(g.dealer.WarriorsCards(len(g.Players)))

	// Each player gets 3 Warrior cards
	warriorsIdx := 0
	for _, p := range g.Players {
		p.TakeCards(warriorCards[warriorsIdx : warriorsIdx+3]...)
		warriorsIdx += 3
	}

	deckCards := append(warriorCards[warriorsIdx:],
		g.dealer.OtherCards(len(g.Players))...)

	deckCards = shuffle(deckCards)
	otherIdx := 0
	for _, p := range g.Players {
		p.TakeCards(deckCards[otherIdx : otherIdx+4]...)
		otherIdx += 4
	}

	deckCards = deckCards[otherIdx:]
	g.deck = NewDeck(deckCards)
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
	if g.Mode == types.GameMode2v2 {
		g.winner = p.Name() + "'s team"
	} else {
		g.winner = p.Name()
	}
}

func (g *Game) OnFieldWithoutWarriors(playerName string) {
	eliminatedIdx := g.PlayerIndex(playerName)

	switch g.Mode {
	case types.GameMode1v1:
		g.gameOver = true
		g.winner = g.CurrentPlayer().Name()
		return

	case types.GameModeFFA3, types.GameModeFFA5:
		g.EliminatedPlayers[eliminatedIdx] = true
		active := 0
		var lastActive string
		for i, p := range g.Players {
			if !g.EliminatedPlayers[i] {
				active++
				lastActive = p.Name()
			}
		}
		if active == 1 {
			g.gameOver = true
			g.winner = lastActive
		}

	case types.GameMode2v2:
		g.EliminatedPlayers[eliminatedIdx] = true
		// Check if all enemies of the eliminated player's team are also eliminated
		// (i.e., the opposing team is fully eliminated)
		attackerIdx := g.CurrentTurn
		allEnemiesEliminated := true
		for _, enemy := range g.Enemies(attackerIdx) {
			enemyIdx := g.PlayerIndex(enemy.Name())
			if !g.EliminatedPlayers[enemyIdx] {
				allEnemiesEliminated = false
				break
			}
		}
		if allEnemiesEliminated {
			g.gameOver = true
			g.winner = g.CurrentPlayer().Name() + "'s team"
		}
	}

	g.addToHistory(playerName + " has been eliminated!")
	eliminatedPlayer := g.Players[eliminatedIdx]
	// Move all cards from the eliminated player's hand to the discard pile
	for _, c := range eliminatedPlayer.Hand().ShowCards() {
		g.discardPile.Discard(c)
	}
	// Move all castled cards to the discard pile
	for _, c := range eliminatedPlayer.Castle().ResourceCards() {
		g.discardPile.Discard(c)
	}
}

func (g *Game) switchTurn() {
	g.hasMovedWarrior = false
	g.hasTraded = false
	g.currentAction = types.ActionTypeDrawCard

	for {
		g.CurrentTurn = (g.CurrentTurn + 1) % len(g.Players)
		if !g.EliminatedPlayers[g.CurrentTurn] {
			break
		}
	}
}

func (g *Game) nextAction(expectedAction types.ActionType,
	gameStatusFn func() GameStatus) GameStatus {

	p := g.CurrentPlayer()
	g.CanMoveWarrior = !g.hasMovedWarrior && p.HasWarriorsInHand()
	g.CanTrade = !g.hasTraded && p.CanTradeCards()

	if expectedAction == types.ActionTypeAttack {
		// Check if player can attack with weapons OR catapult
		canAttackWithCatapult := false
		if p.HasCatapult() {
			for _, e := range g.Enemies(g.CurrentTurn) {
				if e.Castle().CanBeAttacked() {
					canAttackWithCatapult = true
					break
				}
			}
		}

		if p.CanAttack() || canAttackWithCatapult || g.CanMoveWarrior {
			g.currentAction = types.ActionTypeAttack
			return gameStatusFn()
		}
		expectedAction = types.ActionTypeSpySteal
		g.addToHistory(fmt.Sprintf("%s has no cards to attack, skipping phase.", p.Name()))
	}
	if expectedAction == types.ActionTypeSpySteal {
		if p.HasSpy() || p.HasThief() {
			g.currentAction = types.ActionTypeSpySteal
			return gameStatusFn()
		}
		expectedAction = types.ActionTypeBuy
		g.addToHistory(fmt.Sprintf("%s has no Spy or Thief to steal, skipping phase.", p.Name()))
	}
	if expectedAction == types.ActionTypeBuy {
		if p.CanBuy() {
			g.currentAction = types.ActionTypeBuy
			return gameStatusFn()
		}
		expectedAction = types.ActionTypeConstruct
		g.addToHistory(fmt.Sprintf("%s has no cards to buy or exceeds hand limit, skipping phase.", p.Name()))
	}
	if expectedAction == types.ActionTypeConstruct {
		canConstruct := p.CanConstruct()
		if !canConstruct {
			// In 2v2, check if player has resources and any ally has a constructed castle
			for _, ally := range g.Allies(g.PlayerIndex(p.Name())) {
				if ally.Castle().IsConstructed() {
					for _, c := range p.Hand().ShowCards() {
						if _, ok := c.(ports.Resource); ok {
							canConstruct = true
							break
						}
					}
					break
				}
			}
		}
		if canConstruct {
			g.currentAction = types.ActionTypeConstruct
			return gameStatusFn()
		}
		expectedAction = types.ActionTypeEndTurn
		g.addToHistory(fmt.Sprintf("%s has no cards to construct, skipping phase.", p.Name()))
	}

	if expectedAction == types.ActionTypeEndTurn {
		g.currentAction = types.ActionTypeEndTurn
		return gameStatusFn()
	}

	g.currentAction = types.ActionTypeDrawCard
	return gameStatusFn()
}

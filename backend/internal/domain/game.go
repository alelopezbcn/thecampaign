package domain

import "errors"

type Games []Game

type Game struct {
	ID          string
	Players     [2]Player
	CurrentTurn int
	State       GameState
	Deck        []Card
	DiscardPile []Card
	Cemetery    []Card
	IsStarted   bool
	History     []string
}

func (g Game) HandleAction(playerID string, action string) error {

	if g.Players[g.CurrentTurn].ID != playerID {
		return errors.New("not your turn")
	}

	switch action {
	case "take_card":
	}

	return nil
}

package domain

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type GameStatusProvider interface {
	Get(p ports.Player, e ports.Player, game *Game, newCards ...ports.Card) GameStatus
	GetWithModal(p ports.Player, e ports.Player, game *Game, modalCards []ports.Card) GameStatus
}

type gameStatusProvider struct{}

func (gsp *gameStatusProvider) Get(p ports.Player, e ports.Player,
	game *Game, newCards ...ports.Card) GameStatus {

	return newGameStatus(p, e, game, newCards...)

}

func (gsp *gameStatusProvider) GetWithModal(p ports.Player, e ports.Player,
	game *Game, modalCards []ports.Card) GameStatus {

	return newGameStatusWithModalCards(p, e, game, modalCards)

}

func NewGameStatusProvider() *gameStatusProvider {
	return &gameStatusProvider{}
}

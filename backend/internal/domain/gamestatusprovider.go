package domain

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type GameStatusProvider interface {
	Get(viewer ports.Player, game *Game, newCards ...ports.Card) GameStatus
	GetWithModal(viewer ports.Player, game *Game, modalCards []ports.Card) GameStatus
}

type gameStatusProvider struct{}

func (gsp *gameStatusProvider) Get(viewer ports.Player, game *Game,
	newCards ...ports.Card) GameStatus {

	return newGameStatus(viewer, game, newCards...)

}

func (gsp *gameStatusProvider) GetWithModal(viewer ports.Player, game *Game,
	modalCards []ports.Card) GameStatus {

	return newGameStatusWithModalCards(viewer, game, modalCards)

}

func NewGameStatusProvider() *gameStatusProvider {
	return &gameStatusProvider{}
}

package domain

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type GameStatusProvider interface {
	Get(p ports.Player, e ports.Player, game *Game, newCards ...ports.Card) GameStatus
}

type gameStatusProvider struct{}

func (gsp *gameStatusProvider) Get(p ports.Player, e ports.Player,
	game *Game, newCards ...ports.Card) GameStatus {

	return newGameStatus(p, e, game, newCards...)

}

func NewGameStatusProvider() *gameStatusProvider {
	return &gameStatusProvider{}
}

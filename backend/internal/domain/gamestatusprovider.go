package domain

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type GameStatusProvider interface {
	Get(currentPlayer ports.Player,
		enemy ports.Player, action types.ActionType,
		canMove bool, canTrade bool,
		cemetery ports.Cemetery, newCards ...ports.Card) GameStatus
}

type gameStatusProvider struct{}

func (gsp *gameStatusProvider) Get(currentPlayer ports.Player,
	enemy ports.Player, action types.ActionType,
	canMove bool, canTrade bool,
	cemetery ports.Cemetery, newCards ...ports.Card) GameStatus {

	return newGameStatus(currentPlayer, enemy,
		action, canMove, canTrade, cemetery, newCards...)

}

func NewGameStatusProvider() *gameStatusProvider {
	return &gameStatusProvider{}
}

package game

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
)

type GameStatusProvider interface {
	Get(viewer board.Player, game Game, newCards ...cards.Card) gamestatus.GameStatus
	GetWithModal(viewer board.Player, game Game, modalCards []cards.Card) gamestatus.GameStatus
}

type gameStatusProvider struct{}

func (gsp *gameStatusProvider) Get(viewer board.Player, game Game,
	newCards ...cards.Card,
) gamestatus.GameStatus {
	return gsp.getGameStatus(viewer, game, newCards, nil)
}

func (gsp *gameStatusProvider) GetWithModal(viewer board.Player, game Game,
	modalCards []cards.Card,
) gamestatus.GameStatus {
	return gsp.getGameStatus(viewer, game, nil, modalCards)
}

func (gsp *gameStatusProvider) getGameStatus(viewer board.Player,
	game Game, newCards []cards.Card, modalCards []cards.Card,
) gamestatus.GameStatus {
	viewerIdx := game.PlayerIndex(viewer.Name())
	gameStatusDTO := gamestatus.GameStatusDTO{
		Viewer:                 viewer,
		NewCards:               newCards,
		ModalCards:             modalCards,
		PlayerIndex:            viewerIdx,
		Players:                game.Board().Players(),
		NextTurnPlayer:         game.NextActiveTurnPlayer(),
		TurnPlayer:             game.CurrentPlayer().Name(),
		CurrentAction:          game.CurrentAction(),
		LastAction:             game.LastResult().Action,
		GameMode:               string(game.Mode()),
		IsEliminated:           game.EliminatedPlayers()[viewerIdx],
		IsDisconnected:         game.DisconnectedPlayers()[viewerIdx],
		CanTrade:               game.TurnState().CanTrade,
		CemeteryCount:          game.Board().Cemetery().Count(),
		CemeteryLastDead:       game.Board().Cemetery().GetLast(),
		DiscardPileCount:       game.Board().DiscardPile().Count(),
		DiscardPileLastCard:    game.Board().DiscardPile().GetLast(),
		DeckCount:              game.Board().Deck().Count(),
		GameStartedAt:          game.GameStartedAt(),
		TurnStartedAt:          game.TurnState().StartedAt,
		History:                game.GetHistory(),
		LastMovedWarriorID:     game.LastResult().MovedWarriorID,
		LastAttackWeaponID:     game.LastResult().AttackWeaponID,
		LastAttackTargetID:     game.LastResult().AttackTargetID,
		LastAttackTargetPlayer: game.LastResult().AttackTargetPlayer,
		StolenFrom:             game.LastResult().StolenFrom,
		StolenCard:             game.LastResult().StolenCard,
		SpyTarget:              game.LastResult().Spy.Target,
		SpyTargetPlayer:        game.LastResult().Spy.TargetPlayer,
		CurrentPlayerName:      game.CurrentPlayer().Name(),
		IsPlayerWinner:         game.IsPlayerWinner(viewerIdx),
		SameTeamFn:             game.SameTeam,
		EliminatedPlayers:      game.EliminatedPlayers(),
		DisconnectedPlayers:    game.DisconnectedPlayers(),
		CanMoveWarrior:         game.TurnState().CanMoveWarrior,
		EnemiesFn:              game.Enemies,
		AlliesFn:               game.Allies,
	}

	gameStatusDTO.IsGameOver, gameStatusDTO.Winner = game.IsGameOver()

	return gamestatus.NewGameStatus(gameStatusDTO)
}

func NewGameStatusProvider() *gameStatusProvider {
	return &gameStatusProvider{}
}

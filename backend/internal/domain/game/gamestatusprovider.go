package game

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type GameStatusProvider interface {
	Get(viewer ports.Player, game *Game, newCards ...ports.Card) gamestatus.GameStatus
	GetWithModal(viewer ports.Player, game *Game, modalCards []ports.Card) gamestatus.GameStatus
}

type gameStatusProvider struct{}

func (gsp *gameStatusProvider) Get(viewer ports.Player, game *Game,
	newCards ...ports.Card,
) gamestatus.GameStatus {
	return gsp.getGameStatus(viewer, game, newCards, nil)
}

func (gsp *gameStatusProvider) GetWithModal(viewer ports.Player, game *Game,
	modalCards []ports.Card,
) gamestatus.GameStatus {
	return gsp.getGameStatus(viewer, game, nil, modalCards)
}

func (gsp *gameStatusProvider) getGameStatus(viewer ports.Player,
	game *Game, newCards []ports.Card, modalCards []ports.Card,
) gamestatus.GameStatus {
	viewerIdx := game.PlayerIndex(viewer.Name())
	gameStatusDTO := gamestatus.GameStatusDTO{
		Viewer:                 viewer,
		NewCards:               newCards,
		ModalCards:             modalCards,
		PlayerIndex:            viewerIdx,
		Players:                game.Players,
		NextTurnPlayer:         game.nextActiveTurnPlayer(),
		TurnPlayer:             game.CurrentPlayer().Name(),
		CurrentAction:          game.currentAction,
		LastAction:             game.lastResult.Action,
		GameMode:               string(game.Mode),
		IsEliminated:           game.EliminatedPlayers[viewerIdx],
		IsDisconnected:         game.DisconnectedPlayers[viewerIdx],
		CanTrade:               game.turnState.CanTrade,
		CemeteryCount:          game.cemetery.Count(),
		CemeteryLastDead:       game.cemetery.GetLast(),
		DiscardPileCount:       game.discardPile.Count(),
		DiscardPileLastCard:    game.discardPile.GetLast(),
		DeckCount:              game.deck.Count(),
		GameStartedAt:          game.GameStartedAt,
		TurnStartedAt:          game.turnState.StartedAt,
		History:                game.GetHistory(),
		LastMovedWarriorID:     game.lastResult.MovedWarriorID,
		LastAttackWeaponID:     game.lastResult.AttackWeaponID,
		LastAttackTargetID:     game.lastResult.AttackTargetID,
		LastAttackTargetPlayer: game.lastResult.AttackTargetPlayer,
		StolenFrom:             game.lastResult.StolenFrom,
		StolenCard:             game.lastResult.StolenCard,
		SpyTarget:              game.lastResult.Spy.Target,
		SpyTargetPlayer:        game.lastResult.Spy.TargetPlayer,
		CurrentPlayerName:      game.CurrentPlayer().Name(),
		IsPlayerWinner:         game.isPlayerWinner(viewerIdx),
		SameTeamFn:             game.SameTeam,
		EliminatedPlayers:      game.EliminatedPlayers,
		DisconnectedPlayers:    game.DisconnectedPlayers,
		CanMoveWarrior:         game.turnState.CanMoveWarrior,
		EnemiesFn:              game.Enemies,
		AlliesFn:               game.Allies,
	}

	gameStatusDTO.IsGameOver, gameStatusDTO.Winner = game.IsGameOver()

	return gamestatus.NewGameStatus(gameStatusDTO)
}

func NewGameStatusProvider() *gameStatusProvider {
	return &gameStatusProvider{}
}

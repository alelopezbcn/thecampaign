package domain

import (
	"time"

	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type GameStatus struct {
	CurrentPlayer  string   `json:"current_player"`
	TurnPlayer     string   `json:"turn_player"`
	CurrentAction  string   `json:"current_action"`
	NewCards       []string `json:"new_cards"`
	CanMoveWarrior bool     `json:"can_move_warrior"`
	CanTrade       bool     `json:"can_trade"`

	CurrentPlayerHand   []gamestatus.HandCard  `json:"current_player_hand"`
	CurrentPlayerField  []gamestatus.FieldCard `json:"current_player_field"`
	CurrentPlayerCastle gamestatus.Castle      `json:"current_player_castle"`
	IsEliminated        bool                   `json:"is_eliminated"`
	Opponents           []OpponentStatus       `json:"opponents"`
	GameMode            string                 `json:"game_mode"`
	Cemetery            gamestatus.Cemetery    `json:"cemetery"`
	DiscardPile         gamestatus.DiscardPile `json:"discard_pile"`
	CardsInDeck         int                    `json:"deck"`
	ModalCards          []gamestatus.Card      `json:"modal_cards"`
	History             []string               `json:"history"`
	GameOverMgs         string                 `json:"game_over_msg"`
	IsWinner            bool                   `json:"is_winner"`
	GameStartedAt       time.Time              `json:"game_started_at"`
	TurnStartedAt       time.Time              `json:"turn_started_at"`
	TurnTimeLimitSecs   int                    `json:"turn_time_limit_secs"`
}

type OpponentStatus struct {
	PlayerName   string
	Field        []gamestatus.FieldCard
	Castle       gamestatus.Castle
	CardsInHand  int
	IsAlly       bool
	IsEliminated bool
}

func newGameStatusWithModalCards(viewer ports.Player, game *Game,
	modalCards []ports.Card) GameStatus {
	gs := newGameStatus(viewer, game)

	gs.ModalCards = gamestatus.FromDomainCards(modalCards)

	return gs
}

func newGameStatus(viewer ports.Player, game *Game, newCards ...ports.Card,
) GameStatus {

	viewerIdx := game.PlayerIndex(viewer.Name())
	gs := GameStatus{
		CurrentPlayer:       viewer.Name(),
		TurnPlayer:          game.CurrentPlayer().Name(),
		CurrentAction:       string(game.currentAction),
		GameMode:            string(game.Mode),
		NewCards:            []string{},
		CurrentPlayerHand:   []gamestatus.HandCard{},
		CurrentPlayerField:  []gamestatus.FieldCard{},
		CurrentPlayerCastle: gamestatus.NewCastle(viewer.Castle()),
		IsEliminated:        game.EliminatedPlayers[viewerIdx],
		CanTrade:            game.CanTrade,
		Cemetery:            gamestatus.NewCemetery(game.cemetery),
		DiscardPile:         gamestatus.NewDiscardPile(game.discardPile),
		CardsInDeck:         game.deck.Count(),
		History:             game.GetHistory(),
		GameStartedAt:       game.GameStartedAt,
		TurnStartedAt:       game.TurnStartedAt,
		TurnTimeLimitSecs:   60,
	}

	if len(newCards) > 0 {
		for _, c := range newCards {
			gs.NewCards = append(gs.NewCards, c.GetID())
		}
	}

	processHandCards(viewer, game, &gs)

	for _, warrior := range viewer.Field().Warriors() {
		gs.CurrentPlayerField = append(gs.CurrentPlayerField, gamestatus.NewFieldCard(warrior))
	}

	processOpponents(viewer, game, &gs)

	if over, winner := game.IsGameOver(); over {
		gs.GameOverMgs = "Game over! The winner is " + winner
		gs.IsWinner = game.isPlayerWinner(viewerIdx)
	}

	return gs
}

func processHandCards(viewer ports.Player, game *Game, gs *GameStatus) {
	action := game.currentAction
	canMove := game.CanMoveWarrior

	for _, card := range viewer.Hand().ShowCards() {
		switch ct := card.(type) {
		case ports.Warrior:
			gs.CanMoveWarrior = canMove
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand, gamestatus.NewWarriorHandCard(ct))

		case ports.Weapon:
			var enemyFields []ports.Field
			for _, enemy := range game.Enemies(viewer.Idx()) {
				enemyFields = append(enemyFields, enemy.Field())
			}

			if ct.Type() == types.SpecialPowerWeaponType {
				var allyFields []ports.Field
				for _, ally := range game.Allies(viewer.Idx()) {
					allyFields = append(allyFields, ally.Field())
				}

				gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
					gamestatus.NewSpecialPowerHandCard(ct.(ports.SpecialPower), viewer.Field(),
						allyFields, enemyFields, action))

				continue
			}

			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				gamestatus.NewWeaponHandCard(ct, viewer.Field(),
					enemyFields, viewer.Castle().IsConstructed(), action))

		case ports.Catapult:
			canBeAttacked := false
			for _, enemy := range game.Enemies(viewer.Idx()) {
				if enemy.Castle().CanBeAttacked() {
					canBeAttacked = true
					break
				}
			}
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				gamestatus.NewCatapultHandCard(ct.GetID(), canBeAttacked,
					action))

		case ports.Spy:
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				gamestatus.NewSpyHandCard(ct.GetID(), action))

		case ports.Thief:
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				gamestatus.NewThiefHandCard(ct.GetID(), action))

		case ports.Resource:
			allyCastleConstructed := false
			for _, ally := range game.Allies(game.PlayerIndex(viewer.Name())) {
				if ally.Castle().IsConstructed() {
					allyCastleConstructed = true
					break
				}
			}
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				gamestatus.NewResourceHandCard(ct, viewer.Castle().IsConstructed(),
					allyCastleConstructed, viewer.CanBuyWith(ct), action))
		}
	}
}

func processOpponents(viewer ports.Player, game *Game, gs *GameStatus) {
	viewerIdx := game.PlayerIndex(viewer.Name())

	for i, p := range game.Players {
		if i == viewerIdx {
			continue
		}
		opp := OpponentStatus{
			PlayerName:   p.Name(),
			CardsInHand:  p.CardsInHand(),
			Castle:       gamestatus.NewCastle(p.Castle()),
			IsAlly:       game.SameTeam(viewerIdx, i),
			IsEliminated: game.EliminatedPlayers[i],
		}
		for _, warrior := range p.Field().Warriors() {
			opp.Field = append(opp.Field, gamestatus.NewFieldCard(warrior))
		}
		gs.Opponents = append(gs.Opponents, opp)
	}
}

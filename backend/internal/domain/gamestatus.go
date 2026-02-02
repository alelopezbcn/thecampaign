package domain

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type GameStatus struct {
	CurrentPlayer  string   `json:"current_player"`
	CurrentAction  string   `json:"current_action"`
	NewCards       []string `json:"new_cards"`
	CanMoveWarrior bool     `json:"can_move_warrior"`
	CanTrade       bool     `json:"can_trade"`

	CurrentPlayerHand   []gamestatus.HandCard  `json:"current_player_hand"`
	CurrentPlayerField  []gamestatus.FieldCard `json:"current_player_field"`
	CurrentPlayerCastle gamestatus.Castle      `json:"current_player_castle"`
	EnemyField          []gamestatus.FieldCard `json:"enemy_field"`
	EnemyCastle         gamestatus.Castle      `json:"enemy_castle"`
	CardsInEnemyHand    int                    `json:"cards_in_enemy_hand"`
	Cemetery            gamestatus.Cemetery    `json:"cemetery"`
	DiscardPile         gamestatus.DiscardPile `json:"discard_pile"`
	CardsInDeck         int                    `json:"deck"`
	ModalCards          []gamestatus.Card      `json:"modal_cards"`
	History             []string               `json:"history"`
	GameOverMgs         string                 `json:"game_over_msg"`
	ErrorMsg            string                 `json:"error_msg,omitempty"`
}

func newGameStatusWithModalCards(currentPlayer ports.Player, enemy ports.Player,
	game *Game, modalCards []ports.Card) GameStatus {
	gs := newGameStatus(currentPlayer, enemy, game)

	gs.ModalCards = gamestatus.FromDomainCards(modalCards)

	return gs
}

func newGameStatus(currentPlayer ports.Player, enemy ports.Player, game *Game,
	newCards ...ports.Card) GameStatus {

	action := game.currentAction
	canMove := game.CanMoveWarrior
	canTrade := game.CanTrade

	gs := GameStatus{
		CurrentPlayer:       currentPlayer.Name(),
		CurrentAction:       string(action),
		NewCards:            []string{},
		CurrentPlayerHand:   []gamestatus.HandCard{},
		CurrentPlayerField:  []gamestatus.FieldCard{},
		CurrentPlayerCastle: gamestatus.NewCastle(currentPlayer.Castle()),
		CardsInEnemyHand:    enemy.CardsInHand(),
		EnemyField:          []gamestatus.FieldCard{},
		EnemyCastle:         gamestatus.NewCastle(enemy.Castle()),
		CanTrade:            canTrade && len(currentPlayer.Hand().ShowCards()) >= 3,
		Cemetery:            gamestatus.NewCemetery(game.cemetery),
		DiscardPile:         gamestatus.NewDiscardPile(game.discardPile),
		CardsInDeck:         game.deck.Count(),
		History:             game.GetHistory(),
	}

	if len(newCards) > 0 {
		for _, c := range newCards {
			gs.NewCards = append(gs.NewCards, c.GetID())
		}
	}

	for _, card := range currentPlayer.Hand().ShowCards() {
		switch ct := card.(type) {
		case ports.Warrior:
			gs.CanMoveWarrior = canMove
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand, gamestatus.NewWarriorHandCard(ct))
		case ports.Weapon:
			if ct.Type() == types.SpecialPowerWeaponType {
				gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
					gamestatus.NewSpecialPowerHandCard(ct.(ports.SpecialPower), currentPlayer.Field(),
						enemy.Field(), action))
				continue
			}

			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				gamestatus.NewWeaponHandCard(ct, currentPlayer.Field(),
					enemy.Field(), currentPlayer.Castle().IsConstructed(), action))
		case ports.Catapult:
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				gamestatus.NewCatapultHandCard(ct.GetID(), enemy.Castle().CanBeAttacked(),
					action))
		case ports.Spy:
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				gamestatus.NewSpyHandCard(ct.GetID(), action))
		case ports.Thief:
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				gamestatus.NewThiefHandCard(ct.GetID(), action))
		case ports.Resource:
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				gamestatus.NewResourceHandCard(ct, currentPlayer.Castle().IsConstructed(),
					currentPlayer.CanBuyWith(ct), action))
		}
	}

	for _, warrior := range currentPlayer.Field().Warriors() {
		gs.CurrentPlayerField = append(gs.CurrentPlayerField, gamestatus.NewFieldCard(warrior))
	}
	for _, warrior := range enemy.Field().Warriors() {
		gs.EnemyField = append(gs.EnemyField, gamestatus.NewFieldCard(warrior))
	}

	if over, winner := game.IsGameOver(); over {
		gs.GameOverMgs = "Game over! The winner is " + winner
	}

	return gs
}

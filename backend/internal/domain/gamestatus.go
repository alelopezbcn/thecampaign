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

	CurrentPlayerHand          []gamestatus.HandCard  `json:"current_player_hand"`
	CurrentPlayerField         []gamestatus.FieldCard `json:"current_player_field"`
	CurrentPlayerCastle        gamestatus.Castle      `json:"current_player_castle"`
	EnemyField                 []gamestatus.FieldCard `json:"enemy_field"`
	EnemyCastle                gamestatus.Castle      `json:"enemy_castle"`
	CardsInEnemyHand           int                    `json:"cards_in_enemy_hand"`
	ResourceCardsInEnemyCastle int                    `json:"resource_cards_in_enemy_castle"`
	Cemetery                   gamestatus.Cemetery    `json:"cemetery"`
	DiscardPile                gamestatus.DiscardPile `json:"discard_pile"`
}

func newGameStatus(currentPlayer ports.Player, enemy ports.Player, game *Game, newCards ...ports.Card) GameStatus {
	action := game.currentAction
	canMove := game.CanMoveWarrior
	canTrade := game.CanTrade

	gs := GameStatus{
		CurrentPlayer:              currentPlayer.Name(),
		CurrentAction:              string(action),
		NewCards:                   []string{},
		CurrentPlayerHand:          []gamestatus.HandCard{},
		CurrentPlayerField:         []gamestatus.FieldCard{},
		CurrentPlayerCastle:        gamestatus.NewCastle(currentPlayer.Castle()),
		CardsInEnemyHand:           enemy.CardsInHand(),
		EnemyField:                 []gamestatus.FieldCard{},
		EnemyCastle:                gamestatus.NewCastle(enemy.Castle()),
		ResourceCardsInEnemyCastle: enemy.Castle().ResourceCards(),
		CanTrade:                   canTrade && len(currentPlayer.Hand().ShowCards()) >= 3,
		Cemetery:                   gamestatus.NewCemetery(game.cemetery),
		DiscardPile:                gamestatus.NewDiscardPile(game.discardPile),
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
				gamestatus.NewResourceHandCard(ct, currentPlayer.Castle().IsConstructed(), action))
		}
	}

	for _, warrior := range currentPlayer.Field().Warriors() {
		gs.CurrentPlayerField = append(gs.CurrentPlayerField, gamestatus.NewFieldCard(warrior))
	}
	for _, warrior := range enemy.Field().Warriors() {
		gs.EnemyField = append(gs.EnemyField, gamestatus.NewFieldCard(warrior))
	}

	return gs
}

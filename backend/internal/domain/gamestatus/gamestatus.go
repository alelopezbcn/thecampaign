package gamestatus

import (
	"fmt"
	"strings"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type GameStatus struct {
	CurrentPlayer  string   `json:"current_player"`
	CurrentAction  string   `json:"current_action"`
	NewCards       []string `json:"new_cards"`
	CanMoveWarrior bool     `json:"can_move_warrior"`
	CanTrade       bool     `json:"can_trade"`

	CurrentPlayerHand          []HandCard  `json:"current_player_hand"`
	CurrentPlayerField         []FieldCard `json:"current_player_field"`
	CurrentPlayerCastle        Castle      `json:"current_player_castle"`
	EnemyField                 []FieldCard `json:"enemy_field"`
	EnemyCastle                Castle      `json:"enemy_castle"`
	CardsInEnemyHand           int         `json:"cards_in_enemy_hand"`
	ResourceCardsInEnemyCastle int         `json:"resource_cards_in_enemy_castle"`
}

func (g *GameStatus) ShowBoard() string {
	sb := strings.Builder{}

	sb.WriteString(fmt.Sprintf("%s \n", g.EnemyCastle.String()))
	sb.WriteString("Enemy's field: \n")
	for _, c := range g.EnemyField {
		sb.WriteString("  - " + c.String() + "\n")
	}
	sb.WriteString("--------\n")

	sb.WriteString("Your field: \n")
	for _, c := range g.CurrentPlayerField {
		sb.WriteString("  - " + c.String() + "\n")
	}

	sb.WriteString(fmt.Sprintf("%s \n", g.CurrentPlayerCastle.String()))
	sb.WriteString("--------\n")

	sb.WriteString("Your hand: \n")
	for _, c := range g.CurrentPlayerHand {
		sb.WriteString("  - " + c.String() + "\n")
	}
	sb.WriteString("\n--------")
	return sb.String()
}

func NewGameStatus(currentPlayer ports.Player, enemy ports.Player,
	action types.ActionType, canMove bool, canTrade bool, newCards ...ports.Card) GameStatus {

	// Cache castle info with nil checks
	gs := GameStatus{
		CurrentPlayer:              currentPlayer.Name(),
		CurrentAction:              string(action),
		NewCards:                   []string{},
		CurrentPlayerHand:          []HandCard{},
		CurrentPlayerField:         []FieldCard{},
		CurrentPlayerCastle:        newCastle(currentPlayer.Castle()),
		CardsInEnemyHand:           enemy.CardsInHand(),
		EnemyField:                 []FieldCard{},
		EnemyCastle:                newCastle(enemy.Castle()),
		ResourceCardsInEnemyCastle: enemy.Castle().ResourceCards(),
		CanTrade:                   canTrade && len(currentPlayer.Hand().ShowCards()) >= 3,
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
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand, newWarriorHandCard(ct))
		case ports.Weapon:
			if ct.Type() == types.SpecialPowerWeaponType {
				gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
					newSpecialPowerHandCard(ct.(ports.SpecialPower), currentPlayer.Field(),
						enemy.Field(), action))
				continue
			}

			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				newWeaponHandCard(ct, currentPlayer.Field(),
					enemy.Field(), currentPlayer.Castle().IsConstructed(), action))

		case ports.Catapult:
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				newCatapultHandCard(ct.GetID(), enemy.Castle().CanBeAttacked(),
					action))

		case ports.Spy:
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				newSpyHandCard(ct.GetID(), action))
		case ports.Thief:
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				newThiefHandCard(ct.GetID(), action))
		case ports.Resource:
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				newResourceHandCard(ct, currentPlayer.Castle().IsConstructed(), action))
		}
	}

	for _, warrior := range currentPlayer.Field().Warriors() {
		gs.CurrentPlayerField = append(gs.CurrentPlayerField, newFieldCard(warrior))
	}
	for _, warrior := range enemy.Field().Warriors() {
		gs.EnemyField = append(gs.EnemyField, newFieldCard(warrior))
	}

	return gs
}

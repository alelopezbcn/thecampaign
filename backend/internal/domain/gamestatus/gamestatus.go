package gamestatus

import (
	"fmt"
	"strings"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type GameStatus struct {
	CurrentPlayer     string `json:"current_player"`
	CanMoveWarrior    bool   `json:"can_move_warrior"`
	CanAttack         bool   `json:"can_attack"`
	CanCatapult       bool   `json:"can_catapult"`
	CanSpy            bool   `json:"can_spy"`
	CanSteal          bool   `json:"can_steal"`
	CanBuy            bool   `json:"can_buy"`
	CanInitiateCastle bool   `json:"can_initiate_castle"`
	CanGrowCastle     bool   `json:"can_grow_castle"`

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

func NewGameStatus(currentPlayer ports.Player, enemy ports.Player) GameStatus {
	gs := GameStatus{}
	gs.CurrentPlayer = currentPlayer.Name()
	gs.CurrentPlayerHand = []HandCard{}
	gs.CurrentPlayerField = []FieldCard{}
	gs.EnemyField = []FieldCard{}

	for _, card := range currentPlayer.Hand().ShowCards() {
		switch ct := card.(type) {
		case ports.Warrior:
			gs.CanMoveWarrior = true
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand, newWarriorHandCard(ct))
		case ports.Weapon:
			gs.CanInitiateCastle = ct.CanConstruct()

			gs.CanAttack = true
			if ct.Type() == ports.SpecialPowerWeaponType {
				gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
					newSpecialPowerHandCard(ct.(ports.SpecialPower), currentPlayer.Field(),
						enemy.Field()))
				continue
			}

			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				newWeaponHandCard(ct, currentPlayer.Field(),
					enemy.Field().AttackableIDs()))
		case ports.Catapult:
			canBeUsed := enemy.Castle().CanBeAttacked()
			gs.CanAttack = gs.CanAttack || canBeUsed
			castleID := ""
			if canBeUsed {
				castleID = enemy.Castle().GetID()
			}
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				newCatapultHandCard(ct.GetID(), castleID))

		case ports.Spy:
			gs.CanSpy = true
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				newSpyHandCard(ct.GetID()))
		case ports.Thief:
			gs.CanSteal = true
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				newThiefHandCard(ct.GetID()))
		case ports.Resource:
			gs.CanInitiateCastle = gs.CanInitiateCastle || ct.CanConstruct()
			gs.CanGrowCastle = true
			gs.CanBuy = ct.CanBuy()
			gs.CurrentPlayerHand = append(gs.CurrentPlayerHand,
				newResourceHandCard(ct))
		}
	}

	for _, warrior := range currentPlayer.Field().Warriors() {
		gs.CurrentPlayerField = append(gs.CurrentPlayerField, newFieldCard(warrior))
	}
	for _, warrior := range enemy.Field().Warriors() {
		gs.EnemyField = append(gs.EnemyField, newFieldCard(warrior))
	}

	gs.CurrentPlayerCastle = newCastle(currentPlayer.Castle())
	gs.EnemyCastle = newCastle(enemy.Castle())
	gs.CardsInEnemyHand = enemy.CardsInHand()
	gs.ResourceCardsInEnemyCastle = enemy.Castle().ResourceCards()

	return gs
}

package domain

import (
	"fmt"
	"strings"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type GameStatus struct {
	CurrentPlayer      string
	WarriorsInHandIDs  []string
	UsableWeaponIDs    []string
	SpyID              string
	ThiefID            string
	ResourceIDs        []string
	SpecialPowerStatus *SpecialPowerStatus
	ConstructionIDs    []string

	CurrentPlayerHand          []ports.Card
	CurrentPlayerField         []ports.Warrior
	CurrentPlayerCastle        ports.Castle
	EnemyField                 []ports.Warrior
	EnemyCastle                ports.Castle
	CardsInEnemyHand           int
	ResourceCardsInEnemyCastle int
}

type SpecialPowerStatus struct {
	SpecialPowerIDs   []string
	CanHealIDs        []string
	CanInstantKillIDs []string
	CanProtectIDs     []string
}

func newSpecialPowerStatus(ids []string, myField ports.Field, enemyField ports.Field) *SpecialPowerStatus {
	sp := &SpecialPowerStatus{
		SpecialPowerIDs: ids,
	}

	if myField.HasArcher() {
		for _, warrior := range enemyField.Warriors() {
			if ok, card := warrior.IsProtected(); ok {
				sp.CanInstantKillIDs = append(sp.CanInstantKillIDs, card.GetID())
			} else {
				sp.CanInstantKillIDs = append(sp.CanInstantKillIDs, warrior.GetID())
			}
		}
	}
	if myField.HasKnight() {
		for _, warrior := range myField.Warriors() {
			isProtected, _ := warrior.IsProtected()
			if warrior.Type() == ports.DragonType || isProtected {
				continue
			}
			sp.CanProtectIDs = append(sp.CanProtectIDs, warrior.GetID())
		}
	}
	if myField.HasMage() {
		for _, warrior := range myField.Warriors() {
			if warrior.Type() == ports.DragonType || !warrior.IsDamaged() {
				continue
			}
			sp.CanHealIDs = append(sp.CanHealIDs, warrior.GetID())
		}
	}

	return sp
}

func NewGameStatus(currentPlayer ports.Player, enemy ports.Player) GameStatus {
	gs := GameStatus{}
	gs.CurrentPlayer = currentPlayer.Name()
	gs.WarriorsInHandIDs = []string{}
	gs.UsableWeaponIDs = []string{}
	gs.ResourceIDs = []string{}
	gs.ConstructionIDs = []string{}
	var specialPowerIDs []string

	for _, v := range currentPlayer.Hand().ShowCards() {
		switch c := v.(type) {
		case ports.Warrior:
			gs.WarriorsInHandIDs = append(gs.WarriorsInHandIDs, c.GetID())
		case ports.Weapon:
			w := c.(ports.Weapon)
			if w.DamageAmount() == 1 {
				gs.ConstructionIDs = append(gs.ConstructionIDs, w.GetID())
			}

			switch w.Type() {
			case ports.ArrowType:
				if currentPlayer.Field().HasArcher() ||
					currentPlayer.Field().HasDragon() {
					gs.UsableWeaponIDs = append(gs.UsableWeaponIDs, w.GetID())
				}
			case ports.PoisonType:
				if currentPlayer.Field().HasMage() ||
					currentPlayer.Field().HasDragon() {
					gs.UsableWeaponIDs = append(gs.UsableWeaponIDs, w.GetID())
				}
			case ports.SwordType:
				if currentPlayer.Field().HasKnight() ||
					currentPlayer.Field().HasDragon() {
					gs.UsableWeaponIDs = append(gs.UsableWeaponIDs, w.GetID())
				}
			}
		case ports.Catapult:
			if enemy.Castle().ResourceCards() > 0 {
				gs.UsableWeaponIDs = append(gs.UsableWeaponIDs, c.GetID())
			}
		case ports.Spy:
			gs.SpyID = c.GetID()
		case ports.Thief:
			gs.ThiefID = c.GetID()
		case ports.Resource:
			if c.Value() == 1 {
				gs.ConstructionIDs = append(gs.ConstructionIDs, c.GetID())
			}

			gs.ResourceIDs = append(gs.ResourceIDs, c.GetID())
		case ports.SpecialPower:
			specialPowerIDs = append(specialPowerIDs, c.GetID())
		}
	}

	if len(specialPowerIDs) > 0 {
		gs.SpecialPowerStatus = newSpecialPowerStatus(specialPowerIDs,
			currentPlayer.Field(), enemy.Field())
	}

	gs.CurrentPlayerHand = currentPlayer.Hand().ShowCards()
	gs.CurrentPlayerField = currentPlayer.Field().Warriors()
	gs.CurrentPlayerCastle = currentPlayer.Castle()
	gs.EnemyField = enemy.Field().Warriors()
	gs.EnemyCastle = enemy.Castle()
	gs.CardsInEnemyHand = enemy.CardsInHand()
	gs.ResourceCardsInEnemyCastle = enemy.Castle().ResourceCards()

	return gs
}

func (g *GameStatus) ShowBoard() string {
	sb := strings.Builder{}

	if !g.EnemyCastle.IsConstructed() {
		sb.WriteString("Enemy's castle: not constructed \n")
	} else {
		sb.WriteString(fmt.Sprintf("Enemy's castle: %s \n", g.EnemyCastle.String()))
	}

	sb.WriteString("Enemy's field: \n")
	for _, c := range g.EnemyField {
		sb.WriteString("  - " + c.String() + "\n")
	}
	sb.WriteString("--------\n")

	sb.WriteString("Your field: \n")
	for _, c := range g.CurrentPlayerField {
		sb.WriteString("  - " + c.String() + "\n")
	}

	if !g.CurrentPlayerCastle.IsConstructed() {
		sb.WriteString("Your castle: not constructed \n")
	} else {
		sb.WriteString(fmt.Sprintf("Your castle: %s \n", g.CurrentPlayerCastle.String()))
	}
	sb.WriteString("--------\n")

	sb.WriteString("Your hand: \n")
	for _, c := range g.CurrentPlayerHand {
		sb.WriteString("  - " + c.String() + "\n")
	}
	sb.WriteString("\n--------")
	return sb.String()
}

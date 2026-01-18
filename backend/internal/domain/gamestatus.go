package domain

import "github.com/alelopezbcn/thecampaign/internal/domain/ports"

type GameStatus struct {
	CurrentPlayer      string
	WarriorsInHandIDs  []string
	UsableWeaponIDs    []string
	SpyID              string
	ThiefID            string
	ResourceIDs        []string
	SpecialPowerStatus *SpecialPowerStatus

	CurrentPlayerHand          []Card
	CurrentPlayerField         []Card
	CurrentPlayerCastle        CastleStatus
	EnemyField                 []Card
	EnemyCastle                CastleStatus
	CardsInEnemyHand           int
	ResourceCardsInEnemyCastle int
}

type Card struct {
	AffectedBy   []Card
	ID           string
	Value        int
	Type         string
	CanAttack    []Card
	CanSpy       bool
	CanSteal     bool
	CanBuy       bool
	CanConstruct bool
}

func NewCard(c ports.Card) Card {
	card := Card{
		ID: c.GetID(),
	}
	return card
}

type SpecialPowerStatus struct {
	SpecialPowerIDs   []string
	CanHealIDs        []string
	CanInstantKillIDs []string
	CanProtectIDs     []string
}

func NewSpecialPowerStatus(ids []string, myField ports.Field, enemyField ports.Field) *SpecialPowerStatus {
	sp := &SpecialPowerStatus{
		SpecialPowerIDs: ids,
	}

	if myField.HasArcher() {
		// loopear enemyField y determinar los IDs que puedo InstantKill, enemies y escudos

	}
	if myField.HasKnight() {
		// loopear myField y ver a quien puedo proteger (no dragon, no ya protegidos)
		// en game validar que no se proteja a uno ya protegido!
	}
	if myField.HasMage() {
		// loopear myField y ver a quien puedo curar (los que tengan daño recibido, no Dragon)
	}

	return sp
}

type CastleStatus struct {
	AffectedBy []Card
}

func NewGameStatus(currentPlayer ports.Player, enemy ports.Player) GameStatus {
	gs := GameStatus{}
	gs.CurrentPlayer = currentPlayer.Name()
	gs.WarriorsInHandIDs = []string{}
	gs.UsableWeaponIDs = []string{}
	gs.ResourceIDs = []string{}
	var specialPowerIDs []string
	for _, v := range currentPlayer.ShowHand() {
		switch c := v.(type) {
		case ports.Warrior:
			gs.WarriorsInHandIDs = append(gs.WarriorsInHandIDs, c.GetID())
		case ports.Weapon:
			switch w := c.(type) {
			case ports.Arrow:
				if currentPlayer.ShowField().HasArcher() ||
					currentPlayer.ShowField().HasDragon() {
					gs.UsableWeaponIDs = append(gs.UsableWeaponIDs, w.GetID())
				}
			case ports.Poison:
				if currentPlayer.ShowField().HasMage() ||
					currentPlayer.ShowField().HasDragon() {
					gs.UsableWeaponIDs = append(gs.UsableWeaponIDs, w.GetID())
				}
			case ports.Sword:
				if currentPlayer.ShowField().HasKnight() ||
					currentPlayer.ShowField().HasDragon() {
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
			gs.ResourceIDs = append(gs.ResourceIDs, c.GetID())
		case ports.SpecialPower:
			specialPowerIDs = append(specialPowerIDs, c.GetID())
		}
	}

	if len(specialPowerIDs) > 0 {
		gs.SpecialPowerStatus = NewSpecialPowerStatus(specialPowerIDs,
			currentPlayer.ShowField(), enemy.ShowField())
	}

	return gs
}

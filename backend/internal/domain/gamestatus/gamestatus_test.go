package gamestatus

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/stretchr/testify/assert"
)

func TestGameStatus_WarriorsInHand(t *testing.T) {
	k := cards.NewKnight("k1")
	a := cards.NewArcher("a1")
	m := cards.NewMage("m1")
	d := cards.NewDragon("d1")

	cardsInHand := []ports.Card{k, a, m, d}
	p1 := newPlayerWithCards("p1", cardsInHand, nil)
	p2 := newPlayerWithCards("p2", nil, nil)
	gameStatus := NewGameStatus(p1, p2)

	assert.Equal(t, 4, len(gameStatus.WarriorsInHandIDs))
	assert.Contains(t, gameStatus.WarriorsInHandIDs, "K1")
	assert.Contains(t, gameStatus.WarriorsInHandIDs, "A1")
	assert.Contains(t, gameStatus.WarriorsInHandIDs, "M1")
	assert.Contains(t, gameStatus.WarriorsInHandIDs, "D1")
}

func TestGameStatus_UsableWeapons_All(t *testing.T) {
	k := cards.NewKnight("k1")
	a := cards.NewArcher("a1")
	m := cards.NewMage("m1")

	cardsInField := []ports.Warrior{k, a, m}
	cardsInHand := []ports.Card{
		cards.NewSword("s1", 5),
		cards.NewArrow("a1", 3),
		cards.NewPoison("p1", 4),
	}
	p1 := newPlayerWithCards("p1", cardsInHand, cardsInField)
	p2 := newPlayerWithCards("p2", nil, nil)
	gameStatus := NewGameStatus(p1, p2)

	assert.Equal(t, 3, len(gameStatus.UsableWeaponIDs))
	assert.Contains(t, gameStatus.UsableWeaponIDs, "S1")
	assert.Contains(t, gameStatus.UsableWeaponIDs, "A1")
	assert.Contains(t, gameStatus.UsableWeaponIDs, "P1")
}

func TestGameStatus_UsableWeapons_Two(t *testing.T) {
	k := cards.NewKnight("k1")
	a := cards.NewArcher("a1")

	cardsInField := []ports.Warrior{k, a}
	cardsInHand := []ports.Card{
		cards.NewSword("s1", 5),
		cards.NewArrow("a1", 3),
		cards.NewPoison("p1", 4),
	}
	p1 := newPlayerWithCards("p1", cardsInHand, cardsInField)
	p2 := newPlayerWithCards("p2", nil, nil)
	gameStatus := NewGameStatus(p1, p2)

	assert.Equal(t, 2, len(gameStatus.UsableWeaponIDs))
	assert.Contains(t, gameStatus.UsableWeaponIDs, "S1")
	assert.Contains(t, gameStatus.UsableWeaponIDs, "A1")
	assert.NotContains(t, gameStatus.UsableWeaponIDs, "P1")
}

func TestGameStatus_ConstructionIDs_AsWeapons(t *testing.T) {
	cardsInHand := []ports.Card{
		cards.NewSword("s1", 1),
		cards.NewSword("s2", 5),
		cards.NewArrow("a1", 1),
		cards.NewArrow("a2", 8),
		cards.NewPoison("p1", 1),
		cards.NewPoison("p2", 9),
	}
	p1 := newPlayerWithCards("p1", cardsInHand, nil)
	p2 := newPlayerWithCards("p2", nil, nil)
	gameStatus := NewGameStatus(p1, p2)

	assert.Equal(t, 3, len(gameStatus.ConstructionIDs))
	assert.Contains(t, gameStatus.ConstructionIDs, "S1")
	assert.Contains(t, gameStatus.ConstructionIDs, "A1")
	assert.Contains(t, gameStatus.ConstructionIDs, "P1")
	assert.NotContains(t, gameStatus.ConstructionIDs, "S2")
	assert.NotContains(t, gameStatus.ConstructionIDs, "A2")
	assert.NotContains(t, gameStatus.ConstructionIDs, "P2")
}

func TestGameStatus_ConstructionIDs_AsResource(t *testing.T) {

	cardsInHand := []ports.Card{
		cards.NewGold("g1", 1),
		cards.NewGold("g2", 9),
	}

	p1 := newPlayerWithCards("p1", cardsInHand, nil)
	p2 := newPlayerWithCards("p2", nil, nil)
	gameStatus := NewGameStatus(p1, p2)

	assert.Equal(t, 1, len(gameStatus.ConstructionIDs))
	assert.Contains(t, gameStatus.ConstructionIDs, "G1")
	assert.NotContains(t, gameStatus.ConstructionIDs, "G2")
}

func TestGameStatus_ResourceIDs(t *testing.T) {

	cardsInHand := []ports.Card{
		cards.NewGold("g1", 1),
		cards.NewGold("g2", 9),
	}

	p1 := newPlayerWithCards("p1", cardsInHand, nil)
	p2 := newPlayerWithCards("p2", nil, nil)
	gameStatus := NewGameStatus(p1, p2)

	assert.Equal(t, 2, len(gameStatus.ResourceIDs))
	assert.Contains(t, gameStatus.ResourceIDs, "G1")
	assert.Contains(t, gameStatus.ResourceIDs, "G2")
}

func TestGameStatus_SpecialPower_CanProtect(t *testing.T) {
	sp := cards.NewSpecialPower("sp1")
	cardsInHand := []ports.Card{sp}

	cardsInField := []ports.Warrior{
		cards.NewKnight("m1"),
		cards.NewArcher("a1"),
		cards.NewDragon("d2"),
	}
	enemyField := []ports.Warrior{
		cards.NewKnight("ek1"),
	}

	p1 := newPlayerWithCards("p1", cardsInHand, cardsInField)
	p2 := newPlayerWithCards("p2", nil, enemyField)
	gameStatus := NewGameStatus(p1, p2)

	assert.Equal(t, 1, len(gameStatus.SpecialPowerStatus.SpecialPowerIDs))
	assert.Equal(t, 2, len(gameStatus.SpecialPowerStatus.CanProtectIDs))
	assert.Contains(t, gameStatus.SpecialPowerStatus.SpecialPowerIDs, "SP1")
	assert.Contains(t, gameStatus.SpecialPowerStatus.CanProtectIDs, "M1")
	assert.Contains(t, gameStatus.SpecialPowerStatus.CanProtectIDs, "A1")
	assert.NotContains(t, gameStatus.SpecialPowerStatus.CanProtectIDs, "EK1")
	assert.NotContains(t, gameStatus.SpecialPowerStatus.CanProtectIDs, "D2")
}

func TestGameStatus_SpecialPower_CanInstantKill(t *testing.T) {
	sp := cards.NewSpecialPower("sp1")
	cardsInHand := []ports.Card{sp}

	// Enemy field: one protected, one unprotected
	shield := cards.NewSpecialPower("shield1")
	protectedWarrior := cards.NewKnight("ek1")
	protectedWarrior.Protect(shield)

	unprotectedWarrior := cards.NewArcher("ea1")

	enemyField := []ports.Warrior{protectedWarrior, unprotectedWarrior}
	myField := []ports.Warrior{cards.NewArcher("a1")}

	p1 := newPlayerWithCards("p1", cardsInHand, myField)
	p2 := newPlayerWithCards("p2", nil, enemyField)
	gameStatus := NewGameStatus(p1, p2)

	assert.Equal(t, 1, len(gameStatus.SpecialPowerStatus.SpecialPowerIDs))
	assert.Equal(t, 2, len(gameStatus.SpecialPowerStatus.CanInstantKillIDs))
	assert.Contains(t, gameStatus.SpecialPowerStatus.SpecialPowerIDs, "SP1")
	assert.Contains(t, gameStatus.SpecialPowerStatus.CanInstantKillIDs, "EA1")
	assert.Contains(t, gameStatus.SpecialPowerStatus.CanInstantKillIDs, "SHIELD1")
	assert.NotContains(t, gameStatus.SpecialPowerStatus.CanInstantKillIDs, "EK1")
}

func TestGameStatus_SpecialPower_CanHeal(t *testing.T) {
	sp := cards.NewSpecialPower("sp1")
	cardsInHand := []ports.Card{sp}

	arrow := cards.NewArrow("a1", 4)
	damagedWarrior := cards.NewKnight("ek1")
	damagedWarrior.ReceiveDamage(arrow, 1)

	myField := []ports.Warrior{damagedWarrior,
		cards.NewMage("m1")}

	p1 := newPlayerWithCards("p1", cardsInHand, myField)
	p2 := newPlayerWithCards("p2", nil, nil)
	gameStatus := NewGameStatus(p1, p2)

	assert.Equal(t, 1, len(gameStatus.SpecialPowerStatus.SpecialPowerIDs))
	assert.Equal(t, 1, len(gameStatus.SpecialPowerStatus.CanHealIDs))
	assert.Contains(t, gameStatus.SpecialPowerStatus.SpecialPowerIDs, "SP1")
	assert.Contains(t, gameStatus.SpecialPowerStatus.CanHealIDs, "EK1")
	assert.NotContains(t, gameStatus.SpecialPowerStatus.CanHealIDs, "M1")
}

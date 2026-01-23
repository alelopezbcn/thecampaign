package cards

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// Knight tests

func TestNewKnightCard(t *testing.T) {
	knight := NewKnight("k1")

	assert.Equal(t, "K1", knight.GetID())
	assert.Equal(t, WarriorMaxHealth, knight.Health())
}

func TestKnight_BeAttacked_WeaponNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	knight := NewKnight("k1")

	err := knight.BeAttacked(nil)
	assert.ErrorContains(t, err, "weapon cannot be nil")
}

func TestKnight_BeAttacked_WithSword_NoMultiplier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dmgAmnt := 5
	target := NewKnight("k1")
	sword := NewSword("id", dmgAmnt)

	err := target.BeAttacked(sword)
	assert.NoError(t, err)
	assert.Equal(t, WarriorMaxHealth-dmgAmnt, target.Health())
}

func TestKnight_BeAttacked_WithArrow_NoMultiplier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dmgAmnt := 5
	target := NewKnight("k1")
	arrow := NewArrow("id", dmgAmnt)

	err := target.BeAttacked(arrow)
	assert.NoError(t, err)
	assert.Equal(t, WarriorMaxHealth-dmgAmnt, target.Health())
}

func TestKnight_BeAttacked_WithPoison_Multiplier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dmgAmnt := 5
	target := NewKnight("k1")
	poison := NewPoison("id", dmgAmnt)

	err := target.BeAttacked(poison)
	assert.NoError(t, err)
	assert.Equal(t, WarriorMaxHealth-dmgAmnt*2, target.Health())
}

// Mage tests

func TestNewMageCard(t *testing.T) {
	mage := NewMage("m1")

	assert.Equal(t, "M1", mage.GetID())
	assert.Equal(t, WarriorMaxHealth, mage.Health())
}

func TestMage_BeAttacked_WeaponNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mage := NewMage("m1")

	err := mage.BeAttacked(nil)
	assert.ErrorContains(t, err, "weapon cannot be nil")
}

func TestMage_BeAttacked_WithSword_NoMultiplier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dmgAmnt := 5
	target := NewMage("m1")
	sword := NewSword("id", dmgAmnt)

	err := target.BeAttacked(sword)
	assert.NoError(t, err)
	assert.Equal(t, WarriorMaxHealth-dmgAmnt, target.Health())
}

func TestMage_BeAttacked_WithPoison_NoMultiplier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dmgAmnt := 5
	target := NewMage("m1")
	poison := NewPoison("id", dmgAmnt)

	err := target.BeAttacked(poison)
	assert.NoError(t, err)
	assert.Equal(t, WarriorMaxHealth-dmgAmnt, target.Health())
}

func TestMage_BeAttacked_WithArrow_Multiplier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dmgAmnt := 5
	target := NewMage("m1")
	arrow := NewArrow("id", dmgAmnt)

	err := target.BeAttacked(arrow)
	assert.NoError(t, err)
	assert.Equal(t, WarriorMaxHealth-dmgAmnt*2, target.Health())
}

// Archer tests

func TestNewArcherCard(t *testing.T) {
	archer := NewArcher("a1")

	assert.Equal(t, "A1", archer.GetID())
	assert.Equal(t, WarriorMaxHealth, archer.Health())
}

func TestArcher_BeAttacked_WeaponNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	archer := NewArcher("a1")

	err := archer.BeAttacked(nil)
	assert.ErrorContains(t, err, "weapon cannot be nil")
}

func TestArcher_BeAttacked_WithArrow_NoMultiplier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dmgAmnt := 5
	target := NewArcher("a1")
	arrow := NewArrow("id", dmgAmnt)

	err := target.BeAttacked(arrow)
	assert.NoError(t, err)
	assert.Equal(t, WarriorMaxHealth-dmgAmnt, target.Health())
}

func TestArcher_BeAttacked_WithPoison_NoMultiplier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dmgAmnt := 5
	target := NewArcher("a1")
	poison := NewPoison("id", dmgAmnt)

	err := target.BeAttacked(poison)
	assert.NoError(t, err)
	assert.Equal(t, WarriorMaxHealth-dmgAmnt, target.Health())
}

func TestArcher_BeAttacked_WithSword_Multiplier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dmgAmnt := 5
	target := NewArcher("a1")
	sword := NewSword("id", dmgAmnt)

	err := target.BeAttacked(sword)
	assert.NoError(t, err)
	assert.Equal(t, WarriorMaxHealth-dmgAmnt*2, target.Health())
}

// Dragon tests

func TestNewDragonCard(t *testing.T) {
	dragon := NewDragon("d1")

	assert.Equal(t, "D1", dragon.GetID())
	assert.Equal(t, DragonMaxHealth, dragon.Health())
}

func TestDragon_BeAttacked_WeaponNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dragon := NewDragon("d1")

	err := dragon.BeAttacked(nil)
	assert.ErrorContains(t, err, "weapon cannot be nil")
}

func TestDragon_BeAttacked_WithSword_NoMultiplier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dmgAmnt := 5
	target := NewDragon("d1")
	sword := NewSword("id", dmgAmnt)

	err := target.BeAttacked(sword)
	assert.NoError(t, err)
	assert.Equal(t, DragonMaxHealth-dmgAmnt, target.Health())
}

func TestDragon_BeAttacked_WithArrow_NoMultiplier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dmgAmnt := 5
	target := NewDragon("d1")
	arrow := NewArrow("id", dmgAmnt)

	err := target.BeAttacked(arrow)
	assert.NoError(t, err)
	assert.Equal(t, DragonMaxHealth-dmgAmnt, target.Health())
}

func TestDragon_BeAttacked_WithPoison_NoMultiplier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dmgAmnt := 5
	target := NewDragon("d1")
	poison := NewPoison("id", dmgAmnt)

	err := target.BeAttacked(poison)
	assert.NoError(t, err)
	assert.Equal(t, DragonMaxHealth-dmgAmnt, target.Health())
}

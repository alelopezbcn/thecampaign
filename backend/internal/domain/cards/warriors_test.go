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
	assert.Equal(t, warriorMaxHealth, knight.Health())
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
	assert.Equal(t, warriorMaxHealth-dmgAmnt, target.Health())
}

func TestKnight_BeAttacked_WithArrow_NoMultiplier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dmgAmnt := 5
	target := NewKnight("k1")
	arrow := NewArrow("id", dmgAmnt)

	err := target.BeAttacked(arrow)
	assert.NoError(t, err)
	assert.Equal(t, warriorMaxHealth-dmgAmnt, target.Health())
}

func TestKnight_BeAttacked_WithPoison_Multiplier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dmgAmnt := 5
	target := NewKnight("k1")
	poison := NewPoison("id", dmgAmnt)

	err := target.BeAttacked(poison)
	assert.NoError(t, err)
	assert.Equal(t, warriorMaxHealth-dmgAmnt*2, target.Health())
}

// Mage tests

func TestNewMageCard(t *testing.T) {
	mage := NewMage("m1")

	assert.Equal(t, "M1", mage.GetID())
	assert.Equal(t, warriorMaxHealth, mage.Health())
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
	assert.Equal(t, warriorMaxHealth-dmgAmnt, target.Health())
}

func TestMage_BeAttacked_WithPoison_NoMultiplier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dmgAmnt := 5
	target := NewMage("m1")
	poison := NewPoison("id", dmgAmnt)

	err := target.BeAttacked(poison)
	assert.NoError(t, err)
	assert.Equal(t, warriorMaxHealth-dmgAmnt, target.Health())
}

func TestMage_BeAttacked_WithArrow_Multiplier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dmgAmnt := 5
	target := NewMage("m1")
	arrow := NewArrow("id", dmgAmnt)

	err := target.BeAttacked(arrow)
	assert.NoError(t, err)
	assert.Equal(t, warriorMaxHealth-dmgAmnt*2, target.Health())
}

// Archer tests

func TestNewArcherCard(t *testing.T) {
	archer := NewArcher("a1")

	assert.Equal(t, "A1", archer.GetID())
	assert.Equal(t, warriorMaxHealth, archer.Health())
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
	assert.Equal(t, warriorMaxHealth-dmgAmnt, target.Health())
}

func TestArcher_BeAttacked_WithPoison_NoMultiplier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dmgAmnt := 5
	target := NewArcher("a1")
	poison := NewPoison("id", dmgAmnt)

	err := target.BeAttacked(poison)
	assert.NoError(t, err)
	assert.Equal(t, warriorMaxHealth-dmgAmnt, target.Health())
}

func TestArcher_BeAttacked_WithSword_Multiplier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dmgAmnt := 5
	target := NewArcher("a1")
	sword := NewSword("id", dmgAmnt)

	err := target.BeAttacked(sword)
	assert.NoError(t, err)
	assert.Equal(t, warriorMaxHealth-dmgAmnt*2, target.Health())
}

// Dragon tests

func TestNewDragonCard(t *testing.T) {
	dragon := NewDragon("d1")

	assert.Equal(t, "D1", dragon.GetID())
	assert.Equal(t, dragonMaxHealth, dragon.Health())
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
	assert.Equal(t, dragonMaxHealth-dmgAmnt, target.Health())
}

func TestDragon_BeAttacked_WithArrow_NoMultiplier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dmgAmnt := 5
	target := NewDragon("d1")
	arrow := NewArrow("id", dmgAmnt)

	err := target.BeAttacked(arrow)
	assert.NoError(t, err)
	assert.Equal(t, dragonMaxHealth-dmgAmnt, target.Health())
}

func TestDragon_BeAttacked_WithPoison_NoMultiplier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dmgAmnt := 5
	target := NewDragon("d1")
	poison := NewPoison("id", dmgAmnt)

	err := target.BeAttacked(poison)
	assert.NoError(t, err)
	assert.Equal(t, dragonMaxHealth-dmgAmnt, target.Health())
}

// Mercenary tests

func TestNewMercenaryCard(t *testing.T) {
	merc := NewMercenary("mc1")

	assert.Equal(t, "MC1", merc.GetID())
	assert.Equal(t, "Mercenary", merc.Name())
	assert.Equal(t, mercenaryMaxHealth, merc.Health())
}

func TestMercenary_IsMercenaryCard(t *testing.T) {
	merc := NewMercenary("mc1")
	assert.True(t, merc.IsMercenaryCard())
}

func TestMercenary_BeAttacked_WeaponNil(t *testing.T) {
	merc := NewMercenary("mc1")
	err := merc.BeAttacked(nil)
	assert.ErrorContains(t, err, "weapon cannot be nil")
}

func TestMercenary_BeAttacked_WithSword_NoMultiplier(t *testing.T) {
	dmgAmnt := 5
	target := NewMercenary("mc1")
	sword := NewSword("sw1", dmgAmnt)

	err := target.BeAttacked(sword)
	assert.NoError(t, err)
	assert.Equal(t, mercenaryMaxHealth-dmgAmnt, target.Health())
}

func TestMercenary_BeAttacked_WithArrow_NoMultiplier(t *testing.T) {
	dmgAmnt := 5
	target := NewMercenary("mc1")
	arrow := NewArrow("ar1", dmgAmnt)

	err := target.BeAttacked(arrow)
	assert.NoError(t, err)
	assert.Equal(t, mercenaryMaxHealth-dmgAmnt, target.Health())
}

func TestMercenary_BeAttacked_WithPoison_NoMultiplier(t *testing.T) {
	dmgAmnt := 5
	target := NewMercenary("mc1")
	poison := NewPoison("po1", dmgAmnt)

	err := target.BeAttacked(poison)
	assert.NoError(t, err)
	assert.Equal(t, mercenaryMaxHealth-dmgAmnt, target.Health())
}

func TestMercenary_IsDamaged_FalseAtFullHealth(t *testing.T) {
	merc := NewMercenary("mc1")
	assert.False(t, merc.IsDamaged())
}

func TestMercenary_IsDamaged_TrueAfterDamage(t *testing.T) {
	merc := NewMercenary("mc1")
	_ = merc.BeAttacked(NewSword("sw1", 3))
	assert.True(t, merc.IsDamaged())
}

func TestMercenary_Heal_RestoresToMercenaryMaxHealth(t *testing.T) {
	weaponObs := &fakeCardObs{}
	spObs := &fakeCardObs{}
	deadObs := &fakeWarriorDeadObs{}
	weapon := &fakeWeapon{id: "w1", damage: 5, observer: weaponObs}
	sp := &fakeSP{id: "sp1", observer: spObs}

	merc := NewMercenary("mc1")
	merc.AddWarriorDeadObserver(deadObs)
	_ = merc.BeAttacked(weapon)
	assert.True(t, merc.IsDamaged())

	merc.Heal(sp)

	assert.Equal(t, mercenaryMaxHealth, merc.Health())
	assert.False(t, merc.IsDamaged())
}

func TestMercenary_Resurrect_RestoresToMercenaryMaxHealth(t *testing.T) {
	merc := NewMercenary("mc1")
	_ = merc.BeAttacked(NewSword("sw1", 5))
	assert.True(t, merc.IsDamaged())

	merc.Resurrect()

	assert.Equal(t, mercenaryMaxHealth/2, merc.Health())
}

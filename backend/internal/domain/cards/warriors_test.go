package cards

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewKnightCard(t *testing.T) {
	knight := NewKnightCard("k1")

	assert.Equal(t, "K1", knight.GetID())
	assert.Equal(t, WarriorHealth, knight.Health())
}

func TestKnight_Attack_WithInvalidWeapon(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	knight := NewKnightCard("k1")
	target := mocks.NewMockAttackable(ctrl)
	arrow := NewArrowCard("id", 0)

	err := knight.Attack(target, arrow)
	assert.ErrorContains(t, err, "knight can only attack with sword")
}

func TestKnight_Attack_TargetNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	knight := NewKnightCard("k1")
	sword := mocks.NewMockWeapon(ctrl)

	err := knight.Attack(nil, sword)
	assert.ErrorContains(t, err, "target cannot be nil")
}

func TestKnight_Attack_WeaponNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	knight := NewKnightCard("k1")
	target := mocks.NewMockAttackable(ctrl)

	err := knight.Attack(target, nil)
	assert.ErrorContains(t, err, "weapon cannot be nil")
}

func TestKnight_Attack_WithSword_NoMultiplier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dmgAmnt := 5
	knight := NewKnightCard("k1")
	target := NewMageCard("id")
	sword := NewSwordCard("id", dmgAmnt)

	err := knight.Attack(target, sword)
	assert.NoError(t, err)
	assert.Equal(t, WarriorHealth-dmgAmnt, target.Health())
}

func TestKnight_Attack_WithSword_Multiplier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dmgAmnt := 5
	knight := NewKnightCard("k1")
	target := NewArcherCard("id")
	sword := NewSwordCard("id", dmgAmnt)

	err := knight.Attack(target, sword)
	assert.NoError(t, err)
	assert.Equal(t, WarriorHealth-dmgAmnt*2, target.Health())
}

func TestNewArcherCard(t *testing.T) {
	archer := NewArcherCard("a1")

	assert.Equal(t, "A1", archer.GetID())
	assert.Equal(t, WarriorHealth, archer.Health())
}

func TestArcher_Attack_WithInvalidWeapon(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	archer := NewArcherCard("a1")
	target := mocks.NewMockAttackable(ctrl)
	sword := NewSwordCard("id", 0)

	err := archer.Attack(target, sword)
	assert.ErrorContains(t, err, "archer can only attack with arrow")
}

func TestArcher_Attack_TargetNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	archer := NewArcherCard("a1")
	arrow := mocks.NewMockWeapon(ctrl)

	err := archer.Attack(nil, arrow)
	assert.ErrorContains(t, err, "target cannot be nil")
}

func TestArcher_Attack_WeaponNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	archer := NewArcherCard("a1")
	target := mocks.NewMockAttackable(ctrl)

	err := archer.Attack(target, nil)
	assert.ErrorContains(t, err, "weapon cannot be nil")
}

func TestArcher_Attack_WithArrow_NoMultiplier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dmgAmnt := 4
	archer := NewArcherCard("a1")
	target := NewKnightCard("id")
	arw := NewArrowCard("id", dmgAmnt)

	err := archer.Attack(target, arw)
	assert.NoError(t, err)
	assert.Equal(t, WarriorHealth-dmgAmnt, target.Health())
}

func TestArcher_Attack_WithArrow_Multiplier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dmgAmnt := 4
	archer := NewArcherCard("a1")
	target := NewMageCard("id")
	arw := NewArrowCard("id", dmgAmnt)

	err := archer.Attack(target, arw)
	assert.NoError(t, err)
	assert.Equal(t, WarriorHealth-dmgAmnt*2, target.Health())
}

func TestNewMageCard(t *testing.T) {
	mage := NewMageCard("m1")

	assert.Equal(t, "M1", mage.GetID())
	assert.Equal(t, WarriorHealth, mage.Health())
}

func TestMage_Attack_WithInvalidWeapon(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mage := NewMageCard("m1")
	target := mocks.NewMockAttackable(ctrl)
	sword := NewSwordCard("id", 0)

	err := mage.Attack(target, sword)
	assert.ErrorContains(t, err, "mage can only attack with poison")
}

func TestMage_Attack_TargetNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mage := NewMageCard("m1")
	staff := mocks.NewMockWeapon(ctrl)

	err := mage.Attack(nil, staff)
	assert.ErrorContains(t, err, "target cannot be nil")
}

func TestMage_Attack_WeaponNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mage := NewMageCard("m1")
	target := mocks.NewMockAttackable(ctrl)

	err := mage.Attack(target, nil)
	assert.ErrorContains(t, err, "weapon cannot be nil")
}

func TestMage_Attack_WithPoison_NoMultiplier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dmgAmnt := 6
	mage := NewMageCard("m1")
	target := NewArcherCard("id")
	poison := NewPoisonCard("id", dmgAmnt)

	err := mage.Attack(target, poison)
	assert.NoError(t, err)
	assert.Equal(t, WarriorHealth-dmgAmnt, target.Health())
}

func TestMage_Attack_WithPoison_Multiplier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dmgAmnt := 6
	mage := NewMageCard("m1")
	target := NewKnightCard("id")
	poison := NewPoisonCard("id", dmgAmnt)

	err := mage.Attack(target, poison)
	assert.NoError(t, err)
	assert.Equal(t, WarriorHealth-dmgAmnt*2, target.Health())
}

func TestNewDragonCard(t *testing.T) {
	dragon := NewDragonCard("d1")

	assert.Equal(t, "D1", dragon.GetID())
	assert.Equal(t, WarriorHealth, dragon.Health())
}

func TestDragon_Attack_TargetNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dragon := NewDragonCard("d1")
	fireBreath := mocks.NewMockWeapon(ctrl)

	err := dragon.Attack(nil, fireBreath)
	assert.ErrorContains(t, err, "target cannot be nil")
}

func TestDragon_Attack_WeaponNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dragon := NewDragonCard("d1")
	target := mocks.NewMockAttackable(ctrl)

	err := dragon.Attack(target, nil)
	assert.ErrorContains(t, err, "weapon cannot be nil")
}

func TestDragon_Attack_Archer_Multiplier(t *testing.T) {
	dmgAmnt := 2
	dragon := NewDragonCard("d1")
	target := NewArcherCard("id")
	sword := NewSwordCard("id", dmgAmnt)

	err := dragon.Attack(target, sword)
	assert.NoError(t, err)
	assert.Equal(t, WarriorHealth-dmgAmnt*2, target.Health())
}

func TestDragon_Attack_Archer_NoMultiplier(t *testing.T) {
	dmgAmnt := 2
	dragon := NewDragonCard("d1")
	target := NewArcherCard("id")
	arrow := NewArrowCard("id", dmgAmnt)

	err := dragon.Attack(target, arrow)
	assert.NoError(t, err)
	assert.Equal(t, WarriorHealth-dmgAmnt, target.Health())
}

func TestDragon_Attack_Mage_Multiplier(t *testing.T) {
	dmgAmnt := 2
	dragon := NewDragonCard("d1")
	target := NewMageCard("id")
	arrow := NewArrowCard("id", dmgAmnt)

	err := dragon.Attack(target, arrow)
	assert.NoError(t, err)
	assert.Equal(t, WarriorHealth-dmgAmnt*2, target.Health())
}

func TestDragon_Attack_Mage_NoMultiplier(t *testing.T) {
	dmgAmnt := 2
	dragon := NewDragonCard("d1")
	target := NewMageCard("id")
	poison := NewPoisonCard("id", dmgAmnt)

	err := dragon.Attack(target, poison)
	assert.NoError(t, err)
	assert.Equal(t, WarriorHealth-dmgAmnt, target.Health())
}

func TestDragon_Attack_Knight_Multiplier(t *testing.T) {
	dmgAmnt := 2
	dragon := NewDragonCard("d1")
	target := NewKnightCard("id")
	poison := NewPoisonCard("id", dmgAmnt)

	err := dragon.Attack(target, poison)
	assert.NoError(t, err)
	assert.Equal(t, WarriorHealth-dmgAmnt*2, target.Health())
}

func TestDragon_Attack_Knight_NoMultiplier(t *testing.T) {
	dmgAmnt := 2
	dragon := NewDragonCard("d1")
	target := NewKnightCard("id")
	sword := NewSwordCard("id", dmgAmnt)

	err := dragon.Attack(target, sword)
	assert.NoError(t, err)
	assert.Equal(t, WarriorHealth-dmgAmnt, target.Health())
}

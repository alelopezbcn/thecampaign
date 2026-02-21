package cards

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewSpecialPower(t *testing.T) {
	sp := NewSpecialPower("sp1")
	assert.Equal(t, "SP1", sp.GetID())
	assert.Equal(t, specialPowerMaxHealth, sp.Health())
	assert.Equal(t, specialPowerDamage, sp.DamageAmount())
	// String() returns format: "WeaponType (DamageAmount)"
	assert.Contains(t, sp.String(), "Special Power")
	assert.Contains(t, sp.String(), "10")
}

func TestSpecialPower_Use_ByKnight(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sp := NewSpecialPower("sp1")
	knight := NewKnight("k1")
	target := mocks.NewMockWarrior(ctrl)
	target.EXPECT().Protect(sp)

	err := sp.Use(knight, target)
	assert.NoError(t, err)
}

func TestSpecialPower_Use_ByKnight_OnDragon(t *testing.T) {
	sp := NewSpecialPower("sp1")
	knight := NewKnight("k1")
	target := NewDragon("d1")

	err := sp.Use(knight, target)
	assert.EqualError(t, err, "dragon cannot be protected")
}

func TestSpecialPower_Use_ByArcher(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sp := NewSpecialPower("sp1")
	archer := NewArcher("a1")
	target := mocks.NewMockWarrior(ctrl)
	target.EXPECT().InstantKill(sp)

	err := sp.Use(archer, target)
	assert.NoError(t, err)
}

func TestSpecialPower_Use_ByArcher_OnDragon(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sp := NewSpecialPower("sp1")
	archer := NewArcher("a1")
	target := mocks.NewMockDragon(ctrl)
	target.EXPECT().InstantKill(sp)

	err := sp.Use(archer, target)
	assert.NoError(t, err)
}

func TestSpecialPower_Use_ByMage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sp := NewSpecialPower("sp1")
	mage := NewMage("m1")
	target := mocks.NewMockWarrior(ctrl)
	target.EXPECT().Heal(sp)

	err := sp.Use(mage, target)
	assert.NoError(t, err)
}

func TestSpecialPower_Use_ByMage_OnDragon(t *testing.T) {
	sp := NewSpecialPower("sp1")
	mage := NewMage("m1")
	target := NewDragon("d1")

	err := sp.Use(mage, target)
	assert.EqualError(t, err, "dragon cannot be healed")
}

func TestSpecialPower_Use_ByDragon(t *testing.T) {
	sp := NewSpecialPower("sp1")
	dragon := NewDragon("d1")
	target := NewKnight("k1")

	err := sp.Use(dragon, target)
	assert.EqualError(t, err, "special power action not allowed to be used by Dragon")
}

func TestSpecialPower_Use_ByUnknownType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sp := NewSpecialPower("sp1")
	unknown := mocks.NewMockWarrior(ctrl)
	target := NewKnight("k1")

	err := sp.Use(unknown, target)
	assert.EqualError(t, err, "special power action not allowed for this warrior type")
}

func TestSpecialPower_Destroyed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	weapon1 := mocks.NewMockWeapon(ctrl)
	weapon2 := mocks.NewMockWeapon(ctrl)
	discardObs := mocks.NewMockCardMovedToPileObserver(ctrl)

	weapon1.EXPECT().GetCardMovedToPileObserver().Return(discardObs)
	weapon2.EXPECT().GetCardMovedToPileObserver().Return(discardObs)
	discardObs.EXPECT().OnCardMovedToPile(weapon1)
	discardObs.EXPECT().OnCardMovedToPile(weapon2)

	sp := &specialPower{
		cardBase: &cardBase{
			cardMovedToPileObserver: discardObs,
		},
		attackableBase: &attackableBase{
			attackedBy: []Weapon{weapon1, weapon2},
		},
	}

	discardObs.EXPECT().OnCardMovedToPile(sp)

	sp.Destroyed()
	assert.Empty(t, sp.attackedBy)
}

func TestSpecialPower_ReceiveDamage_NotDefeated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sp := &specialPower{
		attackableBase: &attackableBase{
			attackedBy: []Weapon{},
			health:     10,
		},
	}

	weapon := mocks.NewMockWeapon(ctrl)
	weapon.EXPECT().DamageAmount().Return(4)
	defeated := sp.ReceiveDamage(weapon, 1)
	assert.False(t, defeated)
	assert.Equal(t, 6, sp.Health())
	assert.Len(t, sp.attackedBy, 1)
	assert.Equal(t, weapon, sp.attackedBy[0])
}

func TestSpecialPower_ReceiveDamage_Defeated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	weapon := mocks.NewMockWeapon(ctrl)
	weapon.EXPECT().DamageAmount().Return(15)

	discardObs := mocks.NewMockCardMovedToPileObserver(ctrl)
	weapon.EXPECT().GetCardMovedToPileObserver().Return(discardObs)
	discardObs.EXPECT().OnCardMovedToPile(weapon)

	sp := &specialPower{
		cardBase: &cardBase{
			cardMovedToPileObserver: discardObs,
		},
		attackableBase: &attackableBase{
			attackedBy: []Weapon{},
			health:     10,
		},
	}

	discardObs.EXPECT().OnCardMovedToPile(sp)

	defeated := sp.ReceiveDamage(weapon, 1)
	assert.True(t, defeated)
	assert.LessOrEqual(t, sp.Health(), 0)
	assert.Empty(t, sp.attackedBy)
}

func TestSpecialPower_String_AliveWithWeapons(t *testing.T) {
	// Create a real special power to test String() behavior
	sp := NewSpecialPower("sp1")

	str := sp.String()
	// String() returns format: "WeaponType (DamageAmount)"
	assert.Contains(t, str, "Special Power")
	assert.Contains(t, str, "10")
}

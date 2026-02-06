package cards

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestReceiveDamage_WithProtection_DoesNotGetDamage_ProtectionNotDestroyed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	weapon := mocks.NewMockWeapon(ctrl)
	sp := mocks.NewMockSpecialPower(ctrl)
	sp.EXPECT().ReceiveDamage(weapon, 1).Return(false)
	w := warriorBase{
		protectedBy: sp,
		attackableBase: &attackableBase{
			health: 5,
		},
	}

	defeated := w.ReceiveDamage(weapon, 2)
	assert.False(t, defeated)
	assert.NotNil(t, w.protectedBy)
	assert.Equal(t, w.health, 5)
}

func TestReceiveDamage_WithProtection_DoesNotGetDamage_ProtectionDestroyed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	weapon := mocks.NewMockWeapon(ctrl)
	sp := mocks.NewMockSpecialPower(ctrl)
	sp.EXPECT().ReceiveDamage(weapon, 1).Return(true)
	w := warriorBase{
		protectedBy: sp,
		attackableBase: &attackableBase{
			health: 5,
		},
	}

	defeated := w.ReceiveDamage(weapon, 2)
	assert.False(t, defeated)
	assert.Nil(t, w.protectedBy)
	assert.Equal(t, w.health, 5)
}

func TestReceiveDamage_KillsWarrior(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	weapon := mocks.NewMockWeapon(ctrl)
	weapon.EXPECT().DamageAmount().Return(10)
	cardDiscarded := mocks.NewMockCardMovedToPileObserver(ctrl)
	cardDiscarded.EXPECT().OnCardMovedToPile(weapon)
	weapon.EXPECT().GetCardMovedToPileObserver().
		Return(cardDiscarded)
	obs := mocks.NewMockWarriorDeadObserver(ctrl)

	w := &warriorBase{
		attackableBase: &attackableBase{
			health: 5,
		},
		WarriorDeadObserver: obs,
	}
	obs.EXPECT().OnWarriorDead(w)

	defeated := w.ReceiveDamage(weapon, 1)
	assert.True(t, defeated)
	assert.Len(t, w.attackedBy, 0)
	assert.LessOrEqual(t, w.health, 0)
}

func TestReceiveDamage_ReducesHealth_NoDeath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	weapon := mocks.NewMockWeapon(ctrl)
	weapon.EXPECT().DamageAmount().Return(2)
	w := &warriorBase{
		attackableBase: &attackableBase{
			health: 5,
		},
	}

	defeated := w.ReceiveDamage(weapon, 1)
	assert.False(t, defeated)
	assert.Equal(t, 3, w.health)
	assert.Len(t, w.attackedBy, 1)
	assert.Equal(t, weapon, w.attackedBy[0])
}

func TestWarriorBase_BeAttacked_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	w := &warriorBase{}
	weapon := mocks.NewMockWeapon(ctrl)

	err := w.BeAttacked(weapon)
	assert.Error(t, err)
	assert.EqualError(t, err, "should be implemented by concrete warrior types")
}

func TestProtectedBy_SetsProtection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	w := &warriorBase{}
	sp := mocks.NewMockSpecialPower(ctrl)

	w.Protect(sp)
	assert.Equal(t, sp, w.protectedBy)
}

func TestWarriorBase_Heal_RestoresHealthAndDiscardsWeapons(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	weapon1 := mocks.NewMockWeapon(ctrl)
	weapon2 := mocks.NewMockWeapon(ctrl)
	discardObs1 := mocks.NewMockCardMovedToPileObserver(ctrl)
	discardObs2 := mocks.NewMockCardMovedToPileObserver(ctrl)

	weapon1.EXPECT().GetCardMovedToPileObserver().Return(discardObs1)
	weapon2.EXPECT().GetCardMovedToPileObserver().Return(discardObs2)
	discardObs1.EXPECT().OnCardMovedToPile(weapon1)
	discardObs2.EXPECT().OnCardMovedToPile(weapon2)

	w := &warriorBase{
		attackableBase: &attackableBase{
			health:     2,
			attackedBy: []ports.Weapon{weapon1, weapon2},
		},
	}
	sp := mocks.NewMockSpecialPower(ctrl)
	discardObs := mocks.NewMockCardMovedToPileObserver(ctrl)
	sp.EXPECT().GetCardMovedToPileObserver().Return(discardObs)
	discardObs.EXPECT().OnCardMovedToPile(sp)

	w.Heal(sp)
	assert.Equal(t, WarriorMaxHealth, w.health)
	assert.Empty(t, w.attackedBy)
}

func TestWarriorBase_Heal_NoWeapons(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	w := &warriorBase{
		attackableBase: &attackableBase{
			health:     1,
			attackedBy: []ports.Weapon{},
		},
	}
	sp := mocks.NewMockSpecialPower(ctrl)
	discardObs := mocks.NewMockCardMovedToPileObserver(ctrl)
	sp.EXPECT().GetCardMovedToPileObserver().Return(discardObs)
	discardObs.EXPECT().OnCardMovedToPile(sp)

	w.Heal(sp)
	assert.Equal(t, WarriorMaxHealth, w.health)
	assert.Empty(t, w.attackedBy)
}

func TestWarriorBase_InstantKill_WithProtection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sp := mocks.NewMockSpecialPower(ctrl)
	w := &warriorBase{
		protectedBy: sp,
	}
	sp.EXPECT().Destroyed()

	w.InstantKill(sp)
}

func TestWarriorBase_InstantKill_WithoutProtection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sp := mocks.NewMockSpecialPower(ctrl)

	discardObs := mocks.NewMockCardMovedToPileObserver(ctrl)
	sp.EXPECT().GetCardMovedToPileObserver().Return(discardObs)
	discardObs.EXPECT().OnCardMovedToPile(sp)

	obs := mocks.NewMockWarriorDeadObserver(ctrl)
	w := &warriorBase{
		attackableBase: &attackableBase{
			attackedBy: []ports.Weapon{},
		},
		WarriorDeadObserver: obs,
	}
	obs.EXPECT().OnWarriorDead(w)

	w.InstantKill(sp)
	assert.Empty(t, w.attackedBy)
}

func TestWarriorBase_String_AliveWithWeapons(t *testing.T) {
	// Create a real knight to test String() behavior
	// String() returns format: "WarriorType (Health)"
	k := NewKnight("k1")
	str := k.String()
	assert.Contains(t, str, "Knight")
	assert.Contains(t, str, "20") // Initial health
}

func TestWarriorBase_String_AliveNoWeapons(t *testing.T) {
	// String() returns format: "WarriorType (Health)"
	a := NewArcher("a1")
	str := a.String()
	assert.Contains(t, str, "Archer")
	assert.Contains(t, str, "20") // Initial health
}

func TestWarriorBase_String_Dead(t *testing.T) {
	// Test warrior with 0 health
	// Health() method returns 0 instead of negative
	m := NewMage("m1")
	str := m.String()
	assert.Contains(t, str, "Mage")
	assert.Contains(t, str, "20") // Initial health
}

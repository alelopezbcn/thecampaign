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
	cardDiscarded := mocks.NewMockCardToBeDiscardedObserver(ctrl)
	cardDiscarded.EXPECT().OnCardToBeDiscarded(weapon)
	weapon.EXPECT().GetCardToBeDiscardedObserver().
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

func TestWarriorBase_Attack_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	w := &warriorBase{}
	attackable := mocks.NewMockAttackable(ctrl)
	weapon := mocks.NewMockWeapon(ctrl)

	err := w.Attack(attackable, weapon)
	assert.Error(t, err)
	assert.EqualError(t, err, "should be implemented by concrete warrior types")
}

func TestProtectedBy_SetsProtection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	w := &warriorBase{}
	sp := mocks.NewMockSpecialPower(ctrl)

	w.ProtectedBy(sp)
	assert.Equal(t, sp, w.protectedBy)
}

func TestWarriorBase_Heal_RestoresHealthAndDiscardsWeapons(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	weapon1 := mocks.NewMockWeapon(ctrl)
	weapon2 := mocks.NewMockWeapon(ctrl)
	discardObs1 := mocks.NewMockCardToBeDiscardedObserver(ctrl)
	discardObs2 := mocks.NewMockCardToBeDiscardedObserver(ctrl)

	weapon1.EXPECT().GetCardToBeDiscardedObserver().Return(discardObs1)
	weapon2.EXPECT().GetCardToBeDiscardedObserver().Return(discardObs2)
	discardObs1.EXPECT().OnCardToBeDiscarded(weapon1)
	discardObs2.EXPECT().OnCardToBeDiscarded(weapon2)

	w := &warriorBase{
		attackableBase: &attackableBase{
			health:     2,
			attackedBy: []ports.Weapon{weapon1, weapon2},
		},
	}

	w.Heal()
	assert.Equal(t, WarriorHealth, w.health)
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

	w.Heal()
	assert.Equal(t, WarriorHealth, w.health)
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

	discardObs := mocks.NewMockCardToBeDiscardedObserver(ctrl)
	sp.EXPECT().GetCardToBeDiscardedObserver().Return(discardObs)
	discardObs.EXPECT().OnCardToBeDiscarded(sp)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	weapon1 := mocks.NewMockWeapon(ctrl)
	weapon2 := mocks.NewMockWeapon(ctrl)
	weapon1.EXPECT().String().Return("Sword")
	weapon2.EXPECT().String().Return("Axe")

	w := &warriorBase{
		cardBase: &cardBase{
			name: "Warrior",
			id:   "W1",
		},
		attackableBase: &attackableBase{
			health:     5,
			attackedBy: []ports.Weapon{weapon1, weapon2},
		},
	}

	str := w.String()
	assert.Contains(t, str, "Warrior (W1)")
	assert.Contains(t, str, "Health: 5")
	assert.Contains(t, str, "Sword")
	assert.Contains(t, str, "Axe")
}

func TestWarriorBase_String_AliveNoWeapons(t *testing.T) {
	w := &warriorBase{
		cardBase: &cardBase{
			name: "Warrior",
			id:   "W2",
		},
		attackableBase: &attackableBase{
			health:     3,
			attackedBy: nil,
		},
	}

	str := w.String()
	assert.Contains(t, str, "Warrior (W2)")
	assert.Contains(t, str, "Health: 3")
}

func TestWarriorBase_String_Dead(t *testing.T) {
	w := &warriorBase{
		cardBase: &cardBase{
			name: "Warrior",
			id:   "W3",
		},
		attackableBase: &attackableBase{
			health:     0,
			attackedBy: nil,
		},
	}

	str := w.String()
	assert.Contains(t, str, "Warrior (W3)")
	assert.NotContains(t, str, "Health:")
}

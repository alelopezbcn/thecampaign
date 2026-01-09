package cards

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// Helper to create a warriorCardBase with mocks
func newTestWarriorBase(ctrl *gomock.Controller) (
	*warriorCardBase, *mocks.MockSpecialPower, *mocks.MockWarriorDeadObserver) {

	w := &warriorCardBase{}
	sp := mocks.NewMockSpecialPower(ctrl)
	obs := mocks.NewMockWarriorDeadObserver(ctrl)
	w.AddWarriorDeadObserver(obs)
	return w, sp, obs
}

func TestReceiveDamage_WithProtection_DoesNotGetDamage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	w, sp, _ := newTestWarriorBase(ctrl)
	weapon := mocks.NewMockWeapon(ctrl)
	w.ProtectedBy(sp)

	sp.EXPECT().ReceiveDamage(weapon, 1).Return(false)

	defeated := w.ReceiveDamage(weapon, 2)
	assert.False(t, defeated)

	continuar por testing
}

//
// func TestReceiveDamage_WithProtection_Destroyed(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()
//
// 	w, sp, _ := newTestWarriorBase(ctrl)
// 	weapon := mocks.NewMockWeapon(ctrl)
// 	w.ProtectedBy(sp)
//
// 	sp.EXPECT().ReceiveDamage(weapon, 1).Return(true)
//
// 	defeated := w.ReceiveDamage(weapon, 2)
// 	assert.False(t, defeated)
// 	assert.Nil(t, w.ProtectedBy)
// }
//
// func TestReceiveDamage_KillsWarrior(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()
//
// 	w, _, obs := newTestWarriorBase(ctrl)
// 	weapon := mocks.NewMockWeapon(ctrl)
// 	weapon.EXPECT().DamageAmount().Return(10)
// 	weapon.EXPECT().GetCardToBeDiscardedObserver().Return(mocks.NewMockCardToBeDiscardedObserver(ctrl))
// 	w.Health = 5
//
// 	obs.EXPECT().OnWarriorDead(w)
//
// 	defeated := w.ReceiveDamage(weapon, 1)
// 	assert.True(t, defeated)
// }
//
// func TestReceiveDamage_ReducesHealth_NoDeath(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()
//
// 	w, _, _ := newTestWarriorBase(ctrl)
// 	weapon := mocks.NewMockWeapon(ctrl)
// 	weapon.EXPECT().DamageAmount().Return(2)
// 	w.Health = 5
//
// 	defeated := w.ReceiveDamage(weapon, 1)
// 	assert.False(t, defeated)
// 	assert.Equal(t, 3, w.Health)
// }
//
// func TestHeal_ResetsHealth_AndNotifiesDiscard(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()
//
// 	w, _, _ := newTestWarriorBase(ctrl)
// 	weapon := mocks.NewMockWeapon(ctrl)
// 	obs := mocks.NewMockCardToBeDiscardedObserver(ctrl)
// 	weapon.EXPECT().GetCardToBeDiscardedObserver().Return(obs)
// 	obs.EXPECT().OnCardToBeDiscarded(weapon)
// 	w.AttackedBy = []ports.Weapon{weapon}
// 	w.Health = 1
//
// 	w.Heal()
// 	assert.Equal(t, WarriorHealth, w.Health)
// 	assert.Empty(t, w.AttackedBy)
// }
//
// func TestInstantKill_WithProtection(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()
//
// 	w, sp, _ := newTestWarriorBase(ctrl)
// 	w.ProtectedBy(sp)
// 	sp.EXPECT().Destroyed()
//
// 	w.InstantKill()
// }
//
// func TestInstantKill_WithoutProtection(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()
//
// 	w, _, obs := newTestWarriorBase(ctrl)
// 	weapon := mocks.NewMockWeapon(ctrl)
// 	obsDiscard := mocks.NewMockCardToBeDiscardedObserver(ctrl)
// 	weapon.EXPECT().GetCardToBeDiscardedObserver().Return(obsDiscard)
// 	obsDiscard.EXPECT().OnCardToBeDiscarded(weapon)
// 	w.AttackedBy = []ports.Weapon{weapon}
//
// 	obs.EXPECT().OnWarriorDead(w)
//
// 	w.InstantKill()
// 	assert.Empty(t, w.AttackedBy)
// }

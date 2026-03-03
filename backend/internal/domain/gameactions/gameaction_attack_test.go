package gameactions_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/alelopezbcn/thecampaign/internal/domain/gameactions"
)

func TestAttackAction_PlayerName(t *testing.T) {
	action := gameactions.NewAttackAction("Player1", "Player2", "t1", "w1")
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestAttackAction_NextPhase(t *testing.T) {
	action := gameactions.NewAttackAction("Player1", "Player2", "t1", "w1")
	assert.Equal(t, types.PhaseTypeSpySteal, action.NextPhase())
}

func TestAttackAction_Validate(t *testing.T) {
	t.Run("Error when not in Attack phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeBuy).Times(2)

		action := gameactions.NewAttackAction("Player1", "Player2", "targetID", "weaponID")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot attack in the")
	})

	t.Run("Error when target card not in enemy field", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().GetTargetPlayer("Player1", "Player2").Return(mockPlayer2, nil)
		mockPlayer2.EXPECT().GetCardFromField("targetID").Return(nil, false)

		action := gameactions.NewAttackAction("Player1", "Player2", "targetID", "weaponID")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target card not in enemy field")
	})

	t.Run("Error when weapon card not in hand", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().GetTargetPlayer("Player1", "Player2").Return(mockPlayer2, nil)
		mockPlayer2.EXPECT().GetCardFromField("targetID").Return(mockWarrior, true)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("weaponID").Return(nil, false)

		action := gameactions.NewAttackAction("Player1", "Player2", "targetID", "weaponID")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "weapon card not in hand")
	})

	t.Run("Error when target is not attackable", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockCard := mocks.NewMockCard(ctrl) // Not Attackable
		mockWeapon := mocks.NewMockWeapon(ctrl)
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().GetTargetPlayer("Player1", "Player2").Return(mockPlayer2, nil)
		mockPlayer2.EXPECT().GetCardFromField("targetID").Return(mockCard, true)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("weaponID").Return(mockWeapon, true)

		action := gameactions.NewAttackAction("Player1", "Player2", "targetID", "weaponID")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "the target card cannot be attacked")
	})

	t.Run("Error when card is not a weapon", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockResource := mocks.NewMockResource(ctrl) // Not a weapon
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().GetTargetPlayer("Player1", "Player2").Return(mockPlayer2, nil)
		mockPlayer2.EXPECT().GetCardFromField("targetID").Return(mockWarrior, true)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("weaponID").Return(mockResource, true)

		action := gameactions.NewAttackAction("Player1", "Player2", "targetID", "weaponID")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "the card is not a weapon")
	})

	t.Run("Error when weapon cannot be used with player's field", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockWeapon := mocks.NewMockWeapon(ctrl)
		mockField := mocks.NewMockField(ctrl)
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().GetTargetPlayer("Player1", "Player2").Return(mockPlayer2, nil)
		mockPlayer2.EXPECT().GetCardFromField("targetID").Return(mockWarrior, true)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("weaponID").Return(mockWeapon, true)
		mockPlayer1.EXPECT().Field().Return(mockField)
		mockWeapon.EXPECT().CanBeUsedWith(mockField).Return(false)
		mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)

		action := gameactions.NewAttackAction("Player1", "Player2", "targetID", "weaponID")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "weapon cannot be used")
	})

	t.Run("Success validates without error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockWeapon := mocks.NewMockWeapon(ctrl)
		mockField := mocks.NewMockField(ctrl)
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().GetTargetPlayer("Player1", "Player2").Return(mockPlayer2, nil)
		mockPlayer2.EXPECT().GetCardFromField("targetID").Return(mockWarrior, true)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("weaponID").Return(mockWeapon, true)
		mockPlayer1.EXPECT().Field().Return(mockField)
		mockWeapon.EXPECT().CanBeUsedWith(mockField).Return(true)

		action := gameactions.NewAttackAction("Player1", "Player2", "targetID", "weaponID")
		err := action.Validate(mockGame)

		assert.NoError(t, err)
	})
}

// validateForExecuteWithTargetPlayer is the full helper that returns all mocks including mockPlayer2.
func validateForExecuteWithTargetPlayer(t *testing.T, ctrl *gomock.Controller, targetID, weaponID string) (
	action gameactions.GameAction,
	mockGame *mocks.MockGame,
	mockPlayer1 *mocks.MockPlayer,
	mockPlayer2 *mocks.MockPlayer,
	mockWarrior *mocks.MockWarrior,
	mockWeapon *mocks.MockWeapon,
) {
	t.Helper()
	mockGame = mocks.NewMockGame(ctrl)
	mockPlayer1 = mocks.NewMockPlayer(ctrl)
	mockPlayer2 = mocks.NewMockPlayer(ctrl)
	mockWarrior = mocks.NewMockWarrior(ctrl)
	mockWeapon = mocks.NewMockWeapon(ctrl)
	mockField := mocks.NewMockField(ctrl)

	mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
	mockGame.EXPECT().GetTargetPlayer(gomock.Any(), gomock.Any()).Return(mockPlayer2, nil)
	mockPlayer2.EXPECT().GetCardFromField(targetID).Return(mockWarrior, true)
	mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
	mockPlayer1.EXPECT().GetCardFromHand(weaponID).Return(mockWeapon, true)
	mockPlayer1.EXPECT().Field().Return(mockField)
	mockWeapon.EXPECT().CanBeUsedWith(mockField).Return(true)

	a := gameactions.NewAttackAction("Player1", "Player2", targetID, weaponID)
	if err := a.Validate(mockGame); err != nil {
		t.Fatalf("validateForExecute: unexpected validation error: %v", err)
	}
	return a, mockGame, mockPlayer1, mockPlayer2, mockWarrior, mockWeapon
}

// validateForExecute sets up expectations for a successful Validate call and
// calls Validate on the action, populating its internal target/weapon/player.
func validateForExecute(t *testing.T, ctrl *gomock.Controller, targetID, weaponID string) (
	action gameactions.GameAction,
	mockGame *mocks.MockGame,
	mockPlayer1 *mocks.MockPlayer,
	mockWarrior *mocks.MockWarrior,
	mockWeapon *mocks.MockWeapon,
) {
	t.Helper()
	a, mockGame, mockPlayer1, _, mockWarrior, mockWeapon := validateForExecuteWithTargetPlayer(t, ctrl, targetID, weaponID)
	return a, mockGame, mockPlayer1, mockWarrior, mockWeapon
}

func TestAttackAction_Execute(t *testing.T) {
	t.Run("Error when attack fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, _, mockPlayer2, mockWarrior, mockWeapon := validateForExecuteWithTargetPlayer(t, ctrl, "targetID", "weaponID")
		mockDefField := mocks.NewMockField(ctrl)
		mockPlayer2.EXPECT().Field().Return(mockDefField)
		mockDefField.EXPECT().SlotCards().Return(nil)

		mockGame.EXPECT().EventHandler().Return(calmEvent())
		mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)
		mockWarrior.EXPECT().BeAttacked(mockWeapon).Return(errors.New("attack failed"))

		result, _, err := action.Execute(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "attack action failed")
		assert.NotNil(t, result)
	})

	t.Run("Success returns result with attack details", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		action, mockGame, mockPlayer1, mockPlayer2, mockWarrior, mockWeapon := validateForExecuteWithTargetPlayer(t, ctrl, "K1", "S1")
		mockDefField := mocks.NewMockField(ctrl)
		mockPlayer2.EXPECT().Field().Return(mockDefField)
		mockDefField.EXPECT().SlotCards().Return(nil)

		mockGame.EXPECT().EventHandler().Return(calmEvent())
		mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockWarrior.EXPECT().BeAttacked(mockWeapon).Return(nil)
		mockWarrior.EXPECT().String().Return("Knight (20)")
		mockWeapon.EXPECT().String().Return("Sword (5)")
		mockPlayer1.EXPECT().RemoveFromHand("S1").Return(nil, nil)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer1).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionAttack, result.Action)
		assert.Equal(t, "S1", result.AttackWeaponID)
		assert.Equal(t, "K1", result.AttackTargetID)
		assert.Equal(t, "Player2", result.AttackTargetPlayer)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("History is updated on successful attack", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockPlayer2, mockWarrior, mockWeapon := validateForExecuteWithTargetPlayer(t, ctrl, "K1", "S1")
		mockDefField := mocks.NewMockField(ctrl)
		mockPlayer2.EXPECT().Field().Return(mockDefField)
		mockDefField.EXPECT().SlotCards().Return(nil)

		mockGame.EXPECT().EventHandler().Return(calmEvent())
		mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockWarrior.EXPECT().BeAttacked(mockWeapon).Return(nil)
		mockWarrior.EXPECT().String().Return("Knight (20)")
		mockWeapon.EXPECT().String().Return("Sword (5)")
		mockPlayer1.EXPECT().RemoveFromHand("S1").Return(nil, nil)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Do(func(msg string, _ types.Category) {
			assert.True(t, strings.Contains(msg, "Player1") && strings.Contains(msg, "attacked"),
				"History should contain attack action")
		})

		_, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
	})
}

// setupAmbushAttack validates an attack action and sets up the defender's field to have an ambush
// with a specific effect. Returns all mocks needed for execute expectations.
func setupAmbushAttack(t *testing.T, ctrl *gomock.Controller, effect types.AmbushEffect) (
	action gameactions.GameAction,
	mockGame *mocks.MockGame,
	mockPlayer1 *mocks.MockPlayer,
	mockPlayer2 *mocks.MockPlayer,
	mockDefenderField *mocks.MockField,
	mockWeapon *mocks.MockWeapon,
	ambush *mocks.MockAmbush,
) {
	t.Helper()
	mockGame = mocks.NewMockGame(ctrl)
	mockPlayer1 = mocks.NewMockPlayer(ctrl)
	mockPlayer2 = mocks.NewMockPlayer(ctrl)
	mockWarrior := mocks.NewMockWarrior(ctrl)
	mockWeapon = mocks.NewMockWeapon(ctrl)
	mockAttackerField := mocks.NewMockField(ctrl)
	mockDefenderField = mocks.NewMockField(ctrl)
	ambush = mocks.NewMockAmbush(ctrl)

	// Validate
	mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
	mockGame.EXPECT().GetTargetPlayer(gomock.Any(), gomock.Any()).Return(mockPlayer2, nil)
	mockPlayer2.EXPECT().GetCardFromField("targetID").Return(mockWarrior, true)
	mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
	mockPlayer1.EXPECT().GetCardFromHand("weaponID").Return(mockWeapon, true)
	mockPlayer1.EXPECT().Field().Return(mockAttackerField)
	mockWeapon.EXPECT().CanBeUsedWith(mockAttackerField).Return(true)

	action = gameactions.NewAttackAction("Player1", "Player2", "targetID", "weaponID")
	if err := action.Validate(mockGame); err != nil {
		t.Fatalf("setupAmbushAttack: unexpected validation error: %v", err)
	}

	// Set up defender field to return ambush
	mockPlayer2.EXPECT().Field().Return(mockDefenderField).AnyTimes()
	mockDefenderField.EXPECT().SlotCards().Return([]cards.Card{ambush})
	mockDefenderField.EXPECT().RemoveSlotCard(ambush)

	// Ambush discards itself
	mockCardObs := mocks.NewMockCardMovedToPileObserver(ctrl)
	ambush.EXPECT().GetCardMovedToPileObserver().Return(mockCardObs)
	mockCardObs.EXPECT().OnCardMovedToPile(ambush)
	ambush.EXPECT().Effect().Return(effect)

	return action, mockGame, mockPlayer1, mockPlayer2, mockDefenderField, mockWeapon, ambush
}

func TestAttackAction_Execute_AmbushCancelAttack(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	action, mockGame, mockPlayer1, _, _, mockWeapon, _ := setupAmbushAttack(t, ctrl, types.AmbushEffectCancelAttack)

	mockGame.EXPECT().EventHandler().Return(calmEvent())
	mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)
	mockPlayer1.EXPECT().Name().Return("Player1")
	mockPlayer1.EXPECT().RemoveFromHand("weaponID").Return([]cards.Card{mockWeapon}, nil)
	mockCardObs := mocks.NewMockCardMovedToPileObserver(ctrl)
	mockWeapon.EXPECT().GetCardMovedToPileObserver().Return(mockCardObs)
	mockCardObs.EXPECT().OnCardMovedToPile(mockWeapon)
	mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())

	result, _, err := action.Execute(mockGame)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, types.LastActionAmbush, result.Action)
	assert.Equal(t, types.AmbushEffectCancelAttack, result.AmbushEffect)
	assert.Equal(t, "Player1", result.AmbushAttackerName)
}

func TestAttackAction_Execute_AmbushStealWeapon(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	action, mockGame, mockPlayer1, mockPlayer2, _, mockWeapon, _ := setupAmbushAttack(t, ctrl, types.AmbushEffectStealWeapon)

	mockGame.EXPECT().EventHandler().Return(calmEvent())
	mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)
	mockPlayer1.EXPECT().Name().Return("Player1")
	mockPlayer1.EXPECT().RemoveFromHand("weaponID").Return([]cards.Card{mockWeapon}, nil)
	mockPlayer2.EXPECT().ForceAddCard(mockWeapon)
	mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())

	result, _, err := action.Execute(mockGame)

	assert.NoError(t, err)
	assert.Equal(t, types.LastActionAmbush, result.Action)
	assert.Equal(t, types.AmbushEffectStealWeapon, result.AmbushEffect)
}

func TestAttackAction_Execute_AmbushReflectDamage_NoWarriors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	action, mockGame, mockPlayer1, _, _, mockWeapon, _ := setupAmbushAttack(t, ctrl, types.AmbushEffectReflectDamage)

	mockAttackerField2 := mocks.NewMockField(ctrl)
	mockPlayer1.EXPECT().Field().Return(mockAttackerField2)
	mockGame.EXPECT().EventHandler().Return(calmEvent())
	mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)
	mockAttackerField2.EXPECT().Warriors().Return([]cards.Warrior{})
	mockPlayer1.EXPECT().Name().Return("Player1")
	mockPlayer1.EXPECT().RemoveFromHand("weaponID").Return(nil, nil)
	mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())

	result, _, err := action.Execute(mockGame)

	assert.NoError(t, err)
	assert.Equal(t, types.LastActionAmbush, result.Action)
	assert.Equal(t, types.AmbushEffectReflectDamage, result.AmbushEffect)
}

func TestAttackAction_Execute_AmbushReflectDamage_WithWarrior_UsesTargetMultiplier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	action, mockGame, mockPlayer1, _, _, mockWeapon, _ := setupAmbushAttack(t, ctrl, types.AmbushEffectReflectDamage)

	mockAttackerField2 := mocks.NewMockField(ctrl)
	mockAttackerWarrior := mocks.NewMockWarrior(ctrl)
	mockPlayer1.EXPECT().Field().Return(mockAttackerField2)
	mockGame.EXPECT().EventHandler().Return(calmEvent())
	mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)
	mockAttackerField2.EXPECT().Warriors().Return([]cards.Warrior{mockAttackerWarrior})
	mockAttackerWarrior.EXPECT().Type().Return(types.KnightWarriorType)
	// MultiplierFactor is called with the ORIGINAL target (mockWarrior from setupAmbushAttack)
	mockWeapon.EXPECT().MultiplierFactor(gomock.Any()).Return(2)
	mockAttackerWarrior.EXPECT().ReceiveDamage(mockWeapon, 2)
	mockAttackerWarrior.EXPECT().String().Return("Knight (10)")
	mockPlayer1.EXPECT().Name().Return("Player1")
	mockPlayer1.EXPECT().RemoveFromHand("weaponID").Return(nil, nil)
	mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())

	result, _, err := action.Execute(mockGame)

	assert.NoError(t, err)
	assert.Equal(t, types.LastActionAmbush, result.Action)
	assert.Equal(t, types.AmbushEffectReflectDamage, result.AmbushEffect)
}

func TestAttackAction_Execute_AmbushInstantKill_NoWarriors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	action, mockGame, mockPlayer1, _, _, mockWeapon, _ := setupAmbushAttack(t, ctrl, types.AmbushEffectInstantKill)

	mockAttackerField2 := mocks.NewMockField(ctrl)
	mockPlayer1.EXPECT().Field().Return(mockAttackerField2)
	mockGame.EXPECT().EventHandler().Return(calmEvent())
	mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)
	mockAttackerField2.EXPECT().Warriors().Return([]cards.Warrior{})
	mockPlayer1.EXPECT().Name().Return("Player1")
	mockPlayer1.EXPECT().RemoveFromHand("weaponID").Return([]cards.Card{mockWeapon}, nil)
	mockCardObs := mocks.NewMockCardMovedToPileObserver(ctrl)
	mockWeapon.EXPECT().GetCardMovedToPileObserver().Return(mockCardObs)
	mockCardObs.EXPECT().OnCardMovedToPile(mockWeapon)
	mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())

	result, _, err := action.Execute(mockGame)

	assert.NoError(t, err)
	assert.Equal(t, types.LastActionAmbush, result.Action)
	assert.Equal(t, types.AmbushEffectInstantKill, result.AmbushEffect)
}

func TestAttackAction_Execute_NoAmbush_NormalAttack(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	action, mockGame, mockPlayer1, mockPlayer2, mockWarrior, mockWeapon := validateForExecuteWithTargetPlayer(t, ctrl, "K1", "S1")

	mockDefenderField := mocks.NewMockField(ctrl)
	mockPlayer2.EXPECT().Field().Return(mockDefenderField)
	mockDefenderField.EXPECT().SlotCards().Return(nil)

	mockGame.EXPECT().EventHandler().Return(calmEvent())
	mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)
	mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
	mockWarrior.EXPECT().BeAttacked(mockWeapon).Return(nil)
	mockWarrior.EXPECT().String().Return("Knight (20)")
	mockWeapon.EXPECT().String().Return("Sword (5)")
	mockPlayer1.EXPECT().RemoveFromHand("S1").Return(nil, nil)
	mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())

	result, _, err := action.Execute(mockGame)

	assert.NoError(t, err)
	assert.Equal(t, types.LastActionAttack, result.Action)
}

func TestAttackAction_Execute_AmbushDrainLife(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGame := mocks.NewMockGame(ctrl)
	mockPlayer1 := mocks.NewMockPlayer(ctrl)
	mockPlayer2 := mocks.NewMockPlayer(ctrl)
	mockTargetWarrior := mocks.NewMockWarrior(ctrl)
	mockWeapon := mocks.NewMockWeapon(ctrl)
	mockAttackerField := mocks.NewMockField(ctrl)
	mockDefenderField := mocks.NewMockField(ctrl)
	ambush := mocks.NewMockAmbush(ctrl)

	// Validate
	mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
	mockGame.EXPECT().GetTargetPlayer(gomock.Any(), gomock.Any()).Return(mockPlayer2, nil)
	mockPlayer2.EXPECT().GetCardFromField("targetID").Return(mockTargetWarrior, true)
	mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
	mockPlayer1.EXPECT().GetCardFromHand("weaponID").Return(mockWeapon, true)
	mockPlayer1.EXPECT().Field().Return(mockAttackerField)
	mockWeapon.EXPECT().CanBeUsedWith(mockAttackerField).Return(true)

	action := gameactions.NewAttackAction("Player1", "Player2", "targetID", "weaponID")
	if err := action.Validate(mockGame); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	// Set up ambush
	mockPlayer2.EXPECT().Field().Return(mockDefenderField).AnyTimes()
	mockDefenderField.EXPECT().SlotCards().Return([]cards.Card{ambush})
	mockDefenderField.EXPECT().RemoveSlotCard(ambush)
	ambushObs := mocks.NewMockCardMovedToPileObserver(ctrl)
	ambush.EXPECT().GetCardMovedToPileObserver().Return(ambushObs)
	ambushObs.EXPECT().OnCardMovedToPile(ambush)
	ambush.EXPECT().Effect().Return(types.AmbushEffectDrainLife)

	// EventHandler (calm — no modifier)
	mockGame.EXPECT().EventHandler().Return(calmEvent())
	mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)

	// DrainLife heals the target equal to weapon damage × multiplier
	mockWeapon.EXPECT().MultiplierFactor(mockTargetWarrior).Return(1)
	mockWeapon.EXPECT().DamageAmount().Return(5)
	mockTargetWarrior.EXPECT().HealBy(5)

	mockPlayer1.EXPECT().RemoveFromHand("weaponID").Return(nil, nil)
	weaponObs := mocks.NewMockCardMovedToPileObserver(ctrl)
	mockWeapon.EXPECT().GetCardMovedToPileObserver().Return(weaponObs)
	weaponObs.EXPECT().OnCardMovedToPile(mockWeapon)
	mockPlayer1.EXPECT().Name().Return("Player1")
	mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())

	result, _, err := action.Execute(mockGame)

	assert.NoError(t, err)
	assert.Equal(t, types.LastActionAmbush, result.Action)
	assert.Equal(t, types.AmbushEffectDrainLife, result.AmbushEffect)
}

func TestAttackAction_Execute_Curse(t *testing.T) {
	t.Run("Curse reduces weapon damage — DamageAmount called to compute effective damage", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Curse excludes Sword; Arrow and Poison are affected with -2
		event := curseEvent(types.SwordWeaponType, -2)

		action, mockGame, mockPlayer1, mockPlayer2, mockWarrior, mockWeapon :=
			validateForExecuteWithTargetPlayer(t, ctrl, "K1", "A1")
		mockDefField := mocks.NewMockField(ctrl)
		mockPlayer2.EXPECT().Field().Return(mockDefField)
		mockDefField.EXPECT().SlotCards().Return(nil)

		mockGame.EXPECT().EventHandler().Return(event)
		mockWeapon.EXPECT().Type().Return(types.ArrowWeaponType) // Arrow is affected
		mockWeapon.EXPECT().DamageAmount().Return(5)             // 5 + (-2) = 3 effective

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		// BeAttacked receives the curse-modified wrapper (not the original mockWeapon)
		mockWarrior.EXPECT().BeAttacked(gomock.Any()).Return(nil)
		mockWarrior.EXPECT().String().Return("Knight (20)")
		mockWeapon.EXPECT().String().Return("Arrow (5)")
		mockPlayer1.EXPECT().RemoveFromHand("A1").Return(nil, nil)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer1).Return(gamestatus.GameStatus{})

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.Equal(t, types.LastActionAttack, result.Action)
		assert.Equal(t, gamestatus.GameStatus{}, statusFn())
	})

	t.Run("Curse excluded weapon is unaffected — original weapon passed to BeAttacked", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Curse excludes Sword — Sword deals normal damage
		event := curseEvent(types.SwordWeaponType, -2)

		action, mockGame, mockPlayer1, mockPlayer2, mockWarrior, mockWeapon :=
			validateForExecuteWithTargetPlayer(t, ctrl, "K1", "S1")
		mockDefField := mocks.NewMockField(ctrl)
		mockPlayer2.EXPECT().Field().Return(mockDefField)
		mockDefField.EXPECT().SlotCards().Return(nil)

		mockGame.EXPECT().EventHandler().Return(event)
		mockWeapon.EXPECT().Type().Return(types.SwordWeaponType) // Sword is excluded → mod=0
		// DamageAmount NOT called (mod=0 so wrapper is not created)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockWarrior.EXPECT().BeAttacked(mockWeapon).Return(nil) // original weapon, unmodified
		mockWarrior.EXPECT().String().Return("Knight (20)")
		mockWeapon.EXPECT().String().Return("Sword (5)")
		mockPlayer1.EXPECT().RemoveFromHand("S1").Return(nil, nil)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer1).Return(gamestatus.GameStatus{})

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.Equal(t, types.LastActionAttack, result.Action)
		assert.Equal(t, gamestatus.GameStatus{}, statusFn())
	})

	t.Run("Curse affects reflected damage in ambush", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Curse excludes Sword; Arrow is affected with -2 modifier
		event := curseEvent(types.SwordWeaponType, -2)

		action, mockGame, mockPlayer1, _, _, mockWeapon, _ := setupAmbushAttack(t, ctrl, types.AmbushEffectReflectDamage)

		mockAttackerField2 := mocks.NewMockField(ctrl)
		mockAttackerWarrior := mocks.NewMockWarrior(ctrl)
		mockPlayer1.EXPECT().Field().Return(mockAttackerField2)

		mockGame.EXPECT().EventHandler().Return(event)
		mockWeapon.EXPECT().Type().Return(types.ArrowWeaponType) // Arrow is affected → mod=-2
		mockWeapon.EXPECT().DamageAmount().Return(5)             // effective = 3

		mockAttackerField2.EXPECT().Warriors().Return([]cards.Warrior{mockAttackerWarrior})
		mockAttackerWarrior.EXPECT().Type().Return(types.ArcherWarriorType)
		// MultiplierFactor called on the curse-modified wrapper (delegates to embedded weapon's method)
		mockWeapon.EXPECT().MultiplierFactor(gomock.Any()).Return(1)
		// ReceiveDamage receives the curse-modified wrapper (effectiveDamage=3)
		mockAttackerWarrior.EXPECT().ReceiveDamage(gomock.Any(), 1)
		mockAttackerWarrior.EXPECT().String().Return("Archer (15)")
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockPlayer1.EXPECT().RemoveFromHand("weaponID").Return(nil, nil)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())

		result, _, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.Equal(t, types.LastActionAmbush, result.Action)
		assert.Equal(t, types.AmbushEffectReflectDamage, result.AmbushEffect)
	})
}

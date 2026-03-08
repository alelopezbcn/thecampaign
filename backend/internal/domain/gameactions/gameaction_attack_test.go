package gameactions_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/alelopezbcn/thecampaign/internal/domain/gameactions"
)

func TestAttackAction_PlayerName(t *testing.T) {
	action := gameactions.NewAttackAction("Player1", "w1", "Player2", "t1", "w1")
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestAttackAction_NextPhase(t *testing.T) {
	action := gameactions.NewAttackAction("Player1", "w1", "Player2", "t1", "w1")
	assert.Equal(t, types.PhaseTypeSpySteal, action.NextPhase())
}

func TestAttackAction_Validate(t *testing.T) {
	t.Run("Error when not in Attack phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeBuy).Times(2)

		action := gameactions.NewAttackAction("Player1", "warriorID", "Player2", "targetID", "weaponID")
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

		action := gameactions.NewAttackAction("Player1", "warriorID", "Player2", "targetID", "weaponID")
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

		action := gameactions.NewAttackAction("Player1", "warriorID", "Player2", "targetID", "weaponID")
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

		action := gameactions.NewAttackAction("Player1", "warriorID", "Player2", "targetID", "weaponID")
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

		action := gameactions.NewAttackAction("Player1", "warriorID", "Player2", "targetID", "weaponID")
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

		action := gameactions.NewAttackAction("Player1", "warriorID", "Player2", "targetID", "weaponID")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "weapon cannot be used")
	})

	t.Run("Error when attacker warrior not in field", func(t *testing.T) {
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
		mockPlayer1.EXPECT().GetCardFromField("warriorID").Return(nil, false)

		action := gameactions.NewAttackAction("Player1", "warriorID", "Player2", "targetID", "weaponID")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found in your field")
	})

	t.Run("Error when warrior type incompatible with weapon", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockTargetWarrior := mocks.NewMockWarrior(ctrl)
		mockAttackerWarrior := mocks.NewMockWarrior(ctrl)
		mockWeapon := mocks.NewMockWeapon(ctrl)
		mockField := mocks.NewMockField(ctrl)
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().GetTargetPlayer("Player1", "Player2").Return(mockPlayer2, nil)
		mockPlayer2.EXPECT().GetCardFromField("targetID").Return(mockTargetWarrior, true)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("weaponID").Return(mockWeapon, true)
		mockPlayer1.EXPECT().Field().Return(mockField)
		mockWeapon.EXPECT().CanBeUsedWith(mockField).Return(true)
		mockPlayer1.EXPECT().GetCardFromField("warriorID").Return(mockAttackerWarrior, true)
		mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)
		mockAttackerWarrior.EXPECT().CanUseWeapon(types.SwordWeaponType).Return(false)
		mockAttackerWarrior.EXPECT().Type().Return(types.ArcherWarriorType) // for error message

		action := gameactions.NewAttackAction("Player1", "warriorID", "Player2", "targetID", "weaponID")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot use")
	})

	t.Run("Success validates without error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockTargetWarrior := mocks.NewMockWarrior(ctrl)
		mockAttackerWarrior := mocks.NewMockWarrior(ctrl)
		mockWeapon := mocks.NewMockWeapon(ctrl)
		mockField := mocks.NewMockField(ctrl)
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().GetTargetPlayer("Player1", "Player2").Return(mockPlayer2, nil)
		mockPlayer2.EXPECT().GetCardFromField("targetID").Return(mockTargetWarrior, true)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("weaponID").Return(mockWeapon, true)
		mockPlayer1.EXPECT().Field().Return(mockField)
		mockWeapon.EXPECT().CanBeUsedWith(mockField).Return(true)
		mockPlayer1.EXPECT().GetCardFromField("warriorID").Return(mockAttackerWarrior, true)
		mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)
		mockAttackerWarrior.EXPECT().CanUseWeapon(types.SwordWeaponType).Return(true)

		action := gameactions.NewAttackAction("Player1", "warriorID", "Player2", "targetID", "weaponID")
		err := action.Validate(mockGame)

		assert.NoError(t, err)
	})
}

// validateForExecuteWithTargetPlayer is the full helper that returns all mocks including mockPlayer2.
// The weapon is assumed to be SwordWeaponType and the attacker warrior KnightWarriorType.
func validateForExecuteWithTargetPlayer(t *testing.T, ctrl *gomock.Controller, targetID, weaponID string) (
	action gameactions.GameAction,
	mockGame *mocks.MockGame,
	mockPlayer1 *mocks.MockPlayer,
	mockPlayer2 *mocks.MockPlayer,
	mockTargetWarrior *mocks.MockWarrior,
	mockWeapon *mocks.MockWeapon,
	mockAttackerWarrior *mocks.MockWarrior,
) {
	t.Helper()
	mockGame = mocks.NewMockGame(ctrl)
	mockPlayer1 = mocks.NewMockPlayer(ctrl)
	mockPlayer2 = mocks.NewMockPlayer(ctrl)
	mockTargetWarrior = mocks.NewMockWarrior(ctrl)
	mockAttackerWarrior = mocks.NewMockWarrior(ctrl)
	mockWeapon = mocks.NewMockWeapon(ctrl)
	mockField := mocks.NewMockField(ctrl)

	mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
	mockGame.EXPECT().GetTargetPlayer(gomock.Any(), gomock.Any()).Return(mockPlayer2, nil)
	mockPlayer2.EXPECT().GetCardFromField(targetID).Return(mockTargetWarrior, true)
	mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
	mockPlayer1.EXPECT().GetCardFromHand(weaponID).Return(mockWeapon, true)
	mockPlayer1.EXPECT().Field().Return(mockField)
	mockWeapon.EXPECT().CanBeUsedWith(mockField).Return(true)
	mockPlayer1.EXPECT().GetCardFromField("warriorID").Return(mockAttackerWarrior, true)
	mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)
	mockAttackerWarrior.EXPECT().CanUseWeapon(types.SwordWeaponType).Return(true)

	a := gameactions.NewAttackAction("Player1", "warriorID", "Player2", targetID, weaponID)
	if err := a.Validate(mockGame); err != nil {
		t.Fatalf("validateForExecute: unexpected validation error: %v", err)
	}
	return a, mockGame, mockPlayer1, mockPlayer2, mockTargetWarrior, mockWeapon, mockAttackerWarrior
}

func TestAttackAction_Execute(t *testing.T) {
	t.Run("Error when attack fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, _, mockPlayer2, mockWarrior, mockWeapon, mockAttackerWarrior := validateForExecuteWithTargetPlayer(t, ctrl, "targetID", "weaponID")
		mockDefField := mocks.NewMockField(ctrl)
		mockPlayer2.EXPECT().Field().Return(mockDefField)
		mockDefField.EXPECT().SlotCards().Return(nil)

		mockGame.EXPECT().EventHandler().Return(calmEvent())
		mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)
		mockAttackerWarrior.EXPECT().Kills().Return(0)
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

		action, mockGame, mockPlayer1, mockPlayer2, mockWarrior, mockWeapon, mockAttackerWarrior := validateForExecuteWithTargetPlayer(t, ctrl, "K1", "S1")
		mockDefField := mocks.NewMockField(ctrl)
		mockPlayer2.EXPECT().Field().Return(mockDefField)
		mockDefField.EXPECT().SlotCards().Return(nil)

		mockGame.EXPECT().EventHandler().Return(calmEvent())
		mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)
		mockAttackerWarrior.EXPECT().Kills().Return(0)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockWarrior.EXPECT().BeAttacked(mockWeapon).Return(nil)
		mockWarrior.EXPECT().Health().Return(15) // target survives — no kill
		mockWarrior.EXPECT().String().Return("Knight (20)")
		mockWeapon.EXPECT().String().Return("Sword (5)")
		mockPlayer1.EXPECT().RemoveFromHand("S1").Return(nil, nil)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer1).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionAttack, result.Action)
		assert.NotNil(t, result.Attack)
		assert.Equal(t, "S1", result.Attack.WeaponID)
		assert.Equal(t, "K1", result.Attack.TargetID)
		assert.Equal(t, "Player2", result.Attack.TargetPlayer)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("History is updated on successful attack", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockPlayer2, mockWarrior, mockWeapon, mockAttackerWarrior := validateForExecuteWithTargetPlayer(t, ctrl, "K1", "S1")
		mockDefField := mocks.NewMockField(ctrl)
		mockPlayer2.EXPECT().Field().Return(mockDefField)
		mockDefField.EXPECT().SlotCards().Return(nil)

		mockGame.EXPECT().EventHandler().Return(calmEvent())
		mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)
		mockAttackerWarrior.EXPECT().Kills().Return(0)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockWarrior.EXPECT().BeAttacked(mockWeapon).Return(nil)
		mockWarrior.EXPECT().Health().Return(15) // target survives — no kill
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
	mockAttackerWarrior *mocks.MockWarrior,
) {
	t.Helper()
	mockGame = mocks.NewMockGame(ctrl)
	mockPlayer1 = mocks.NewMockPlayer(ctrl)
	mockPlayer2 = mocks.NewMockPlayer(ctrl)
	mockTargetWarrior := mocks.NewMockWarrior(ctrl)
	mockAttackerWarrior = mocks.NewMockWarrior(ctrl)
	mockWeapon = mocks.NewMockWeapon(ctrl)
	mockAttackerField := mocks.NewMockField(ctrl)
	mockDefenderField = mocks.NewMockField(ctrl)
	ambush = mocks.NewMockAmbush(ctrl)

	// Validate
	mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
	mockGame.EXPECT().GetTargetPlayer(gomock.Any(), gomock.Any()).Return(mockPlayer2, nil)
	mockPlayer2.EXPECT().GetCardFromField("targetID").Return(mockTargetWarrior, true)
	mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
	mockPlayer1.EXPECT().GetCardFromHand("weaponID").Return(mockWeapon, true)
	mockPlayer1.EXPECT().Field().Return(mockAttackerField)
	mockWeapon.EXPECT().CanBeUsedWith(mockAttackerField).Return(true)
	mockPlayer1.EXPECT().GetCardFromField("warriorID").Return(mockAttackerWarrior, true)
	mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)
	mockAttackerWarrior.EXPECT().CanUseWeapon(types.SwordWeaponType).Return(true)

	action = gameactions.NewAttackAction("Player1", "warriorID", "Player2", "targetID", "weaponID")
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

	return action, mockGame, mockPlayer1, mockPlayer2, mockDefenderField, mockWeapon, ambush, mockAttackerWarrior
}

func TestAttackAction_Execute_AmbushCancelAttack(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	action, mockGame, mockPlayer1, _, _, mockWeapon, _, mockAttackerWarrior := setupAmbushAttack(t, ctrl, types.AmbushEffectCancelAttack)

	mockGame.EXPECT().EventHandler().Return(calmEvent())
	mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)
	mockAttackerWarrior.EXPECT().Kills().Return(0)
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
	assert.NotNil(t, result.Attack)
	assert.Equal(t, types.AmbushEffectCancelAttack, result.Attack.AmbushEffect)
	assert.Equal(t, "Player1", result.Attack.AmbushAttackerName)
}

func TestAttackAction_Execute_AmbushStealWeapon(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	action, mockGame, mockPlayer1, mockPlayer2, _, mockWeapon, _, mockAttackerWarrior := setupAmbushAttack(t, ctrl, types.AmbushEffectStealWeapon)

	mockGame.EXPECT().EventHandler().Return(calmEvent())
	mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)
	mockAttackerWarrior.EXPECT().Kills().Return(0)
	mockPlayer1.EXPECT().Name().Return("Player1")
	mockPlayer1.EXPECT().RemoveFromHand("weaponID").Return([]cards.Card{mockWeapon}, nil)
	mockPlayer2.EXPECT().ForceAddCard(mockWeapon)
	mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())

	result, _, err := action.Execute(mockGame)

	assert.NoError(t, err)
	assert.Equal(t, types.LastActionAmbush, result.Action)
	assert.Equal(t, types.AmbushEffectStealWeapon, result.Attack.AmbushEffect)
}

func TestAttackAction_Execute_AmbushReflectDamage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	action, mockGame, mockPlayer1, _, _, mockWeapon, _, mockAttackerWarrior := setupAmbushAttack(t, ctrl, types.AmbushEffectReflectDamage)

	mockGame.EXPECT().EventHandler().Return(calmEvent())
	mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)
	mockAttackerWarrior.EXPECT().Kills().Return(0)
	// MultiplierFactor is called with the ORIGINAL target (mockTargetWarrior from setupAmbushAttack)
	mockWeapon.EXPECT().MultiplierFactor(gomock.Any()).Return(2)
	mockAttackerWarrior.EXPECT().ReceiveDamage(mockWeapon, 2)
	mockAttackerWarrior.EXPECT().String().Return("Knight (10)")
	mockPlayer1.EXPECT().Name().Return("Player1")
	mockPlayer1.EXPECT().RemoveFromHand("weaponID").Return(nil, nil)
	mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())

	result, _, err := action.Execute(mockGame)

	assert.NoError(t, err)
	assert.Equal(t, types.LastActionAmbush, result.Action)
	assert.Equal(t, types.AmbushEffectReflectDamage, result.Attack.AmbushEffect)
}

func TestAttackAction_Execute_AmbushInstantKill(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	action, mockGame, mockPlayer1, _, _, mockWeapon, _, mockAttackerWarrior := setupAmbushAttack(t, ctrl, types.AmbushEffectInstantKill)

	mockGame.EXPECT().EventHandler().Return(calmEvent())
	mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)
	mockAttackerWarrior.EXPECT().Kills().Return(0)
	// The warrior who attacked is instantly killed by the ambush
	mockAttackerWarrior.EXPECT().KillByAmbush()
	mockAttackerWarrior.EXPECT().String().Return("Knight (20)")
	mockPlayer1.EXPECT().Name().Return("Player1")
	mockPlayer1.EXPECT().RemoveFromHand("weaponID").Return([]cards.Card{mockWeapon}, nil)
	mockCardObs := mocks.NewMockCardMovedToPileObserver(ctrl)
	mockWeapon.EXPECT().GetCardMovedToPileObserver().Return(mockCardObs)
	mockCardObs.EXPECT().OnCardMovedToPile(mockWeapon)
	mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())

	result, _, err := action.Execute(mockGame)

	assert.NoError(t, err)
	assert.Equal(t, types.LastActionAmbush, result.Action)
	assert.Equal(t, types.AmbushEffectInstantKill, result.Attack.AmbushEffect)
}

func TestAttackAction_Execute_NoAmbush_NormalAttack(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	action, mockGame, mockPlayer1, mockPlayer2, mockWarrior, mockWeapon, mockAttackerWarrior := validateForExecuteWithTargetPlayer(t, ctrl, "K1", "S1")

	mockDefenderField := mocks.NewMockField(ctrl)
	mockPlayer2.EXPECT().Field().Return(mockDefenderField)
	mockDefenderField.EXPECT().SlotCards().Return(nil)

	mockGame.EXPECT().EventHandler().Return(calmEvent())
	mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)
	mockAttackerWarrior.EXPECT().Kills().Return(0)
	mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
	mockWarrior.EXPECT().BeAttacked(mockWeapon).Return(nil)
	mockWarrior.EXPECT().Health().Return(15) // target survives — no kill
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
	mockAttackerWarrior := mocks.NewMockWarrior(ctrl)
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
	mockPlayer1.EXPECT().GetCardFromField("warriorID").Return(mockAttackerWarrior, true)
	mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)
	mockAttackerWarrior.EXPECT().CanUseWeapon(types.SwordWeaponType).Return(true)

	action := gameactions.NewAttackAction("Player1", "warriorID", "Player2", "targetID", "weaponID")
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
	mockAttackerWarrior.EXPECT().Kills().Return(0)

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
	assert.Equal(t, types.AmbushEffectDrainLife, result.Attack.AmbushEffect)
}

func TestAttackAction_Execute_Curse(t *testing.T) {
	t.Run("Curse reduces weapon damage — DamageAmount called to compute effective damage", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Curse excludes Sword; Arrow and Poison are affected with -2
		event := curseEvent(types.SwordWeaponType, -2)

		action, mockGame, mockPlayer1, mockPlayer2, mockWarrior, mockWeapon, mockAttackerWarrior := validateForExecuteWithTargetPlayer(t, ctrl, "K1", "A1")
		mockDefField := mocks.NewMockField(ctrl)
		mockPlayer2.EXPECT().Field().Return(mockDefField)
		mockDefField.EXPECT().SlotCards().Return(nil)

		mockGame.EXPECT().EventHandler().Return(event)
		mockWeapon.EXPECT().Type().Return(types.ArrowWeaponType) // Arrow is affected
		mockWeapon.EXPECT().DamageAmount().Return(5)             // 5 + (-2) = 3 effective
		mockAttackerWarrior.EXPECT().Kills().Return(0)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		// BeAttacked receives the curse-modified wrapper (not the original mockWeapon)
		mockWarrior.EXPECT().BeAttacked(gomock.Any()).Return(nil)
		mockWarrior.EXPECT().Health().Return(15) // target survives — no kill
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

		action, mockGame, mockPlayer1, mockPlayer2, mockWarrior, mockWeapon, mockAttackerWarrior := validateForExecuteWithTargetPlayer(t, ctrl, "K1", "S1")
		mockDefField := mocks.NewMockField(ctrl)
		mockPlayer2.EXPECT().Field().Return(mockDefField)
		mockDefField.EXPECT().SlotCards().Return(nil)

		mockGame.EXPECT().EventHandler().Return(event)
		mockWeapon.EXPECT().Type().Return(types.SwordWeaponType) // Sword is excluded → mod=0
		// DamageAmount NOT called (mod=0 so wrapper is not created)
		mockAttackerWarrior.EXPECT().Kills().Return(0)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockWarrior.EXPECT().BeAttacked(mockWeapon).Return(nil) // original weapon, unmodified
		mockWarrior.EXPECT().Health().Return(15)                // target survives — no kill
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

		action, mockGame, mockPlayer1, _, _, mockWeapon, _, mockAttackerWarrior := setupAmbushAttack(t, ctrl, types.AmbushEffectReflectDamage)

		mockGame.EXPECT().EventHandler().Return(event)
		mockWeapon.EXPECT().Type().Return(types.ArrowWeaponType) // Arrow is affected → mod=-2
		mockWeapon.EXPECT().DamageAmount().Return(5)             // effective = 3
		mockAttackerWarrior.EXPECT().Kills().Return(0)

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
		assert.Equal(t, types.AmbushEffectReflectDamage, result.Attack.AmbushEffect)
	})
}

func TestAttackAction_Execute_Bloodlust(t *testing.T) {
	t.Run("Attacker heals when target is killed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockPlayer2, mockTargetWarrior, mockWeapon, mockAttackerWarrior := validateForExecuteWithTargetPlayer(t, ctrl, "K1", "S1")
		mockDefField := mocks.NewMockField(ctrl)
		mockPlayer2.EXPECT().Field().Return(mockDefField)
		mockDefField.EXPECT().SlotCards().Return(nil)

		mockGame.EXPECT().EventHandler().Return(bloodlustEvent())
		mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)
		mockAttackerWarrior.EXPECT().Kills().Return(0)
		mockTargetWarrior.EXPECT().BeAttacked(mockWeapon).Return(nil)
		mockTargetWarrior.EXPECT().Health().Return(0).Times(2) // target died: AddKill check + bloodlust check
		mockAttackerWarrior.EXPECT().AddKill()
		mockAttackerWarrior.EXPECT().HealBy(2)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockTargetWarrior.EXPECT().String().Return("Knight (0)")
		mockWeapon.EXPECT().String().Return("Sword (5)")
		mockPlayer1.EXPECT().RemoveFromHand("S1").Return(nil, nil)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())

		result, _, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.Equal(t, types.LastActionAttack, result.Action)
	})

	t.Run("Attacker does not heal when target survives", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockPlayer2, mockTargetWarrior, mockWeapon, mockAttackerWarrior := validateForExecuteWithTargetPlayer(t, ctrl, "K1", "S1")
		mockDefField := mocks.NewMockField(ctrl)
		mockPlayer2.EXPECT().Field().Return(mockDefField)
		mockDefField.EXPECT().SlotCards().Return(nil)

		mockGame.EXPECT().EventHandler().Return(bloodlustEvent())
		mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)
		mockAttackerWarrior.EXPECT().Kills().Return(0)
		mockTargetWarrior.EXPECT().BeAttacked(mockWeapon).Return(nil)
		mockTargetWarrior.EXPECT().Health().Return(10).Times(2) // target survived: AddKill check + bloodlust check
		// HealBy must NOT be called
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockTargetWarrior.EXPECT().String().Return("Knight (10)")
		mockWeapon.EXPECT().String().Return("Sword (5)")
		mockPlayer1.EXPECT().RemoveFromHand("S1").Return(nil, nil)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())

		result, _, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.Equal(t, types.LastActionAttack, result.Action)
	})
}

func TestAttackAction_Execute_ChampionsBounty(t *testing.T) {
	t.Run("Kill top-HP enemy warrior draws a card", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockPlayer2, mockTargetWarrior, mockWeapon, mockAttackerWarrior := validateForExecuteWithTargetPlayer(t, ctrl, "K1", "S1")
		mockDefField := mocks.NewMockField(ctrl)
		mockPlayer2.EXPECT().Field().Return(mockDefField).AnyTimes()
		mockDefField.EXPECT().SlotCards().Return(nil)
		mockDefField.EXPECT().Warriors().Return([]cards.Warrior{mockTargetWarrior})

		mockGame.EXPECT().EventHandler().Return(championsBountyEvent())
		mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)
		mockAttackerWarrior.EXPECT().Kills().Return(0)
		mockAttackerWarrior.EXPECT().AddKill()
		gomock.InOrder(
			mockTargetWarrior.EXPECT().Health().Return(10), // pre-kill snapshot
			mockTargetWarrior.EXPECT().Health().Return(0),  // AddKill check
			mockTargetWarrior.EXPECT().Health().Return(0),  // bounty kill check
		)
		mockTargetWarrior.EXPECT().BeAttacked(mockWeapon).Return(nil)

		// Only enemy is the target player — isTopEnemy trivially true
		mockGame.EXPECT().PlayerIndex("Player1").Return(0)
		mockGame.EXPECT().Enemies(0).Return([]board.Player{mockPlayer2})
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		mockBountyCard := mocks.NewMockCard(ctrl)
		mockGame.EXPECT().DrawCards(mockPlayer1, 2).Return([]cards.Card{mockBountyCard}, nil)
		mockPlayer1.EXPECT().TakeCards(mockBountyCard)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockTargetWarrior.EXPECT().String().Return("Knight (0)")
		mockWeapon.EXPECT().String().Return("Sword (5)")
		mockPlayer1.EXPECT().RemoveFromHand("S1").Return(nil, nil)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Times(2) // bounty + attack

		result, _, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.Equal(t, types.LastActionAttack, result.Action)
	})

	t.Run("Kill non-top enemy warrior draws no card", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockPlayer2, mockTargetWarrior, mockWeapon, mockAttackerWarrior := validateForExecuteWithTargetPlayer(t, ctrl, "K1", "S1")
		mockDefField := mocks.NewMockField(ctrl)
		mockPlayer2.EXPECT().Field().Return(mockDefField).AnyTimes()
		mockDefField.EXPECT().SlotCards().Return(nil)
		mockDefField.EXPECT().Warriors().Return([]cards.Warrior{mockTargetWarrior}) // total HP 5

		mockGame.EXPECT().EventHandler().Return(championsBountyEvent())
		mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)
		mockAttackerWarrior.EXPECT().Kills().Return(0)
		mockAttackerWarrior.EXPECT().AddKill()
		gomock.InOrder(
			mockTargetWarrior.EXPECT().Health().Return(5), // pre-kill snapshot
			mockTargetWarrior.EXPECT().Health().Return(0), // AddKill check
			mockTargetWarrior.EXPECT().Health().Return(0), // bounty kill check
		)
		mockTargetWarrior.EXPECT().BeAttacked(mockWeapon).Return(nil)

		// Player3 has higher total HP — target player is not top
		mockGame.EXPECT().PlayerIndex("Player1").Return(0)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer3Field := mocks.NewMockField(ctrl)
		mockGame.EXPECT().Enemies(0).Return([]board.Player{mockPlayer2, mockPlayer3})
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer3.EXPECT().Name().Return("Player3")
		mockPlayer3.EXPECT().Field().Return(mockPlayer3Field)
		strongerWarrior := mocks.NewMockWarrior(ctrl)
		mockPlayer3Field.EXPECT().Warriors().Return([]cards.Warrior{strongerWarrior})
		strongerWarrior.EXPECT().Health().Return(8)

		// DrawCards must NOT be called
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockTargetWarrior.EXPECT().String().Return("Knight (0)")
		mockWeapon.EXPECT().String().Return("Sword (5)")
		mockPlayer1.EXPECT().RemoveFromHand("S1").Return(nil, nil)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()) // attack only

		result, _, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.Equal(t, types.LastActionAttack, result.Action)
	})

	t.Run("Kill tied-top enemy warrior draws a card", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockPlayer2, mockTargetWarrior, mockWeapon, mockAttackerWarrior := validateForExecuteWithTargetPlayer(t, ctrl, "K1", "S1")
		mockDefField := mocks.NewMockField(ctrl)
		mockPlayer2.EXPECT().Field().Return(mockDefField).AnyTimes()
		mockDefField.EXPECT().SlotCards().Return(nil)
		mockDefField.EXPECT().Warriors().Return([]cards.Warrior{mockTargetWarrior}) // total HP 8

		mockGame.EXPECT().EventHandler().Return(championsBountyEvent())
		mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)
		mockAttackerWarrior.EXPECT().Kills().Return(0)
		mockAttackerWarrior.EXPECT().AddKill()
		gomock.InOrder(
			mockTargetWarrior.EXPECT().Health().Return(8), // pre-kill snapshot
			mockTargetWarrior.EXPECT().Health().Return(0), // AddKill check
			mockTargetWarrior.EXPECT().Health().Return(0), // bounty kill check
		)
		mockTargetWarrior.EXPECT().BeAttacked(mockWeapon).Return(nil)

		// Player3 ties at 8 HP — target player ties for top → bounty applies
		mockGame.EXPECT().PlayerIndex("Player1").Return(0)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer3Field := mocks.NewMockField(ctrl)
		mockGame.EXPECT().Enemies(0).Return([]board.Player{mockPlayer2, mockPlayer3})
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer3.EXPECT().Name().Return("Player3")
		mockPlayer3.EXPECT().Field().Return(mockPlayer3Field)
		equalWarrior := mocks.NewMockWarrior(ctrl)
		mockPlayer3Field.EXPECT().Warriors().Return([]cards.Warrior{equalWarrior})
		equalWarrior.EXPECT().Health().Return(8)

		mockBountyCard := mocks.NewMockCard(ctrl)
		mockGame.EXPECT().DrawCards(mockPlayer1, 2).Return([]cards.Card{mockBountyCard}, nil)
		mockPlayer1.EXPECT().TakeCards(mockBountyCard)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockTargetWarrior.EXPECT().String().Return("Knight (0)")
		mockWeapon.EXPECT().String().Return("Sword (5)")
		mockPlayer1.EXPECT().RemoveFromHand("S1").Return(nil, nil)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Times(2)

		result, _, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.Equal(t, types.LastActionAttack, result.Action)
	})

	t.Run("Target survives — no card drawn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockPlayer2, mockTargetWarrior, mockWeapon, mockAttackerWarrior := validateForExecuteWithTargetPlayer(t, ctrl, "K1", "S1")
		mockDefField := mocks.NewMockField(ctrl)
		mockPlayer2.EXPECT().Field().Return(mockDefField).AnyTimes()
		mockDefField.EXPECT().SlotCards().Return(nil)
		mockDefField.EXPECT().Warriors().Return([]cards.Warrior{mockTargetWarrior})

		mockGame.EXPECT().EventHandler().Return(championsBountyEvent())
		mockWeapon.EXPECT().Type().Return(types.SwordWeaponType)
		mockAttackerWarrior.EXPECT().Kills().Return(0)
		mockTargetWarrior.EXPECT().BeAttacked(mockWeapon).Return(nil)
		mockTargetWarrior.EXPECT().Health().Return(5).Times(3) // pre-kill + AddKill check + bounty kill check (survived)

		// PlayerIndex/Enemies/DrawCards must NOT be called
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockTargetWarrior.EXPECT().String().Return("Knight (5)")
		mockWeapon.EXPECT().String().Return("Sword (3)")
		mockPlayer1.EXPECT().RemoveFromHand("S1").Return(nil, nil)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())

		result, _, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.Equal(t, types.LastActionAttack, result.Action)
	})
}

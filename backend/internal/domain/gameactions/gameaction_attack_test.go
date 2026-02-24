package gameactions_test

import (
	"errors"
	"strings"
	"testing"

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
	mockGame = mocks.NewMockGame(ctrl)
	mockPlayer1 = mocks.NewMockPlayer(ctrl)
	mockPlayer2 := mocks.NewMockPlayer(ctrl)
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
	return a, mockGame, mockPlayer1, mockWarrior, mockWeapon
}

func TestAttackAction_Execute(t *testing.T) {
	t.Run("Error when attack fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, _, mockWarrior, mockWeapon := validateForExecute(t, ctrl, "targetID", "weaponID")
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

		action, mockGame, mockPlayer1, mockWarrior, mockWeapon := validateForExecute(t, ctrl, "K1", "S1")

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

		action, mockGame, mockPlayer1, mockWarrior, mockWeapon := validateForExecute(t, ctrl, "K1", "S1")

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

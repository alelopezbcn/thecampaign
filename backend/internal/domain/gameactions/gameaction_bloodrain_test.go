package gameactions_test

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/alelopezbcn/thecampaign/internal/domain/gameactions"
)

// validateBloodRain sets up and runs Validate for a blood rain action.
func validateBloodRain(
	t *testing.T, ctrl *gomock.Controller, targets []cards.Warrior,
) (gameactions.GameAction, *mocks.MockGame, *mocks.MockPlayer, *mocks.MockBloodRain) {
	t.Helper()

	mockGame := mocks.NewMockGame(ctrl)
	mockPlayer := mocks.NewMockPlayer(ctrl)
	mockTargetPlayer := mocks.NewMockPlayer(ctrl)
	mockField := mocks.NewMockField(ctrl)
	mockBloodRain := mocks.NewMockBloodRain(ctrl)

	mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
	mockGame.EXPECT().GetTargetPlayer("Player1", "Player2").Return(mockTargetPlayer, nil)
	mockTargetPlayer.EXPECT().Field().Return(mockField)
	mockField.EXPECT().Warriors().Return(targets)
	mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
	mockPlayer.EXPECT().GetCardFromHand("bloodrain-id").Return(mockBloodRain, true)

	action := gameactions.NewBloodRainAction("Player1", "Player2", "bloodrain-id")
	if err := action.Validate(mockGame); err != nil {
		t.Fatalf("validateBloodRain: unexpected error: %v", err)
	}
	return action, mockGame, mockPlayer, mockBloodRain
}

func TestBloodRainAction_PlayerName(t *testing.T) {
	action := gameactions.NewBloodRainAction("Player1", "Player2", "bloodrain-id")
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestBloodRainAction_NextPhase(t *testing.T) {
	action := gameactions.NewBloodRainAction("Player1", "Player2", "bloodrain-id")
	assert.Equal(t, types.PhaseTypeSpySteal, action.NextPhase())
}

func TestBloodRainAction_Execute_Bloodlust(t *testing.T) {
	t.Run("Field warriors healed for each target killed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		target1 := mocks.NewMockWarrior(ctrl)
		target2 := mocks.NewMockWarrior(ctrl)
		action, mockGame, mockPlayer, mockBloodRain := validateBloodRain(t, ctrl, []cards.Warrior{target1, target2})

		fieldWarrior := mocks.NewMockWarrior(ctrl)
		mockField := mocks.NewMockField(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockBloodRain.EXPECT().Attack([]cards.Warrior{target1, target2}).Return(nil)
		mockGame.EXPECT().EventHandler().Return(bloodlustEvent())
		target1.EXPECT().Health().Return(0) // killed
		target2.EXPECT().Health().Return(0) // killed
		// 2 kills × 2 heal = 4 total heal
		mockPlayer.EXPECT().Field().Return(mockField)
		mockField.EXPECT().Warriors().Return([]cards.Warrior{fieldWarrior})
		fieldWarrior.EXPECT().HealBy(4)
		mockBloodRain.EXPECT().GetID().Return("bloodrain-id")
		mockPlayer.EXPECT().RemoveFromHand("bloodrain-id").Return(nil, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockBloodRain)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionBloodRain, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Partial kills: only killed targets contribute to heal", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		target1 := mocks.NewMockWarrior(ctrl)
		target2 := mocks.NewMockWarrior(ctrl)
		action, mockGame, mockPlayer, mockBloodRain := validateBloodRain(t, ctrl, []cards.Warrior{target1, target2})

		fieldWarrior := mocks.NewMockWarrior(ctrl)
		mockField := mocks.NewMockField(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockBloodRain.EXPECT().Attack([]cards.Warrior{target1, target2}).Return(nil)
		mockGame.EXPECT().EventHandler().Return(bloodlustEvent())
		target1.EXPECT().Health().Return(0) // killed
		target2.EXPECT().Health().Return(3) // survived
		// 1 kill × 2 heal = 2
		mockPlayer.EXPECT().Field().Return(mockField)
		mockField.EXPECT().Warriors().Return([]cards.Warrior{fieldWarrior})
		fieldWarrior.EXPECT().HealBy(2)
		mockBloodRain.EXPECT().GetID().Return("bloodrain-id")
		mockPlayer.EXPECT().RemoveFromHand("bloodrain-id").Return(nil, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockBloodRain)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.Equal(t, types.LastActionBloodRain, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("No healing when all targets survive", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		target1 := mocks.NewMockWarrior(ctrl)
		action, mockGame, mockPlayer, mockBloodRain := validateBloodRain(t, ctrl, []cards.Warrior{target1})

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockBloodRain.EXPECT().Attack([]cards.Warrior{target1}).Return(nil)
		mockGame.EXPECT().EventHandler().Return(bloodlustEvent())
		target1.EXPECT().Health().Return(5) // survived — HealBy must NOT be called
		mockBloodRain.EXPECT().GetID().Return("bloodrain-id")
		mockPlayer.EXPECT().RemoveFromHand("bloodrain-id").Return(nil, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockBloodRain)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.Equal(t, types.LastActionBloodRain, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("No healing when no bloodlust event", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		target1 := mocks.NewMockWarrior(ctrl)
		action, mockGame, mockPlayer, mockBloodRain := validateBloodRain(t, ctrl, []cards.Warrior{target1})

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockBloodRain.EXPECT().Attack([]cards.Warrior{target1}).Return(nil)
		mockGame.EXPECT().EventHandler().Return(calmEvent())
		// Health must NOT be called; HealBy must NOT be called
		mockBloodRain.EXPECT().GetID().Return("bloodrain-id")
		mockPlayer.EXPECT().RemoveFromHand("bloodrain-id").Return(nil, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockBloodRain)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.Equal(t, types.LastActionBloodRain, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})
}

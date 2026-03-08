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

// validateHarpoon sets up and runs Validate for a harpoon action, returning the
// action and mocks needed to exercise Execute.
func validateHarpoon(
	t *testing.T, ctrl *gomock.Controller,
) (gameactions.GameAction, *mocks.MockGame, *mocks.MockPlayer, *mocks.MockWarrior, *mocks.MockHarpoon) {
	t.Helper()

	mockGame := mocks.NewMockGame(ctrl)
	mockPlayer := mocks.NewMockPlayer(ctrl)
	mockTargetPlayer := mocks.NewMockPlayer(ctrl)
	mockDragon := mocks.NewMockWarrior(ctrl) // Dragon is the same interface as Warrior
	mockHarpoon := mocks.NewMockHarpoon(ctrl)

	mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
	mockGame.EXPECT().GetTargetPlayer("Player1", "Player2").Return(mockTargetPlayer, nil)
	mockTargetPlayer.EXPECT().GetCardFromField("dragon-id").Return(mockDragon, true)
	mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
	mockPlayer.EXPECT().GetCardFromHand("harpoon-id").Return(mockHarpoon, true)

	action := gameactions.NewHarpoonAction("Player1", "Player2", "dragon-id", "harpoon-id")
	if err := action.Validate(mockGame); err != nil {
		t.Fatalf("validateHarpoon: unexpected error: %v", err)
	}
	return action, mockGame, mockPlayer, mockDragon, mockHarpoon
}

func TestHarpoonAction_PlayerName(t *testing.T) {
	action := gameactions.NewHarpoonAction("Player1", "Player2", "dragon-id", "harpoon-id")
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestHarpoonAction_NextPhase(t *testing.T) {
	action := gameactions.NewHarpoonAction("Player1", "Player2", "dragon-id", "harpoon-id")
	assert.Equal(t, types.PhaseTypeSpySteal, action.NextPhase())
}

func TestHarpoonAction_Execute_Bloodlust(t *testing.T) {
	t.Run("All field warriors are healed when dragon is killed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer, mockDragon, mockHarpoon := validateHarpoon(t, ctrl)
		mockField := mocks.NewMockField(ctrl)
		mockFieldWarrior1 := mocks.NewMockWarrior(ctrl)
		mockFieldWarrior2 := mocks.NewMockWarrior(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockHarpoon.EXPECT().Attack(mockDragon).Return(nil)
		mockGame.EXPECT().EventHandler().Return(bloodlustEvent())
		mockDragon.EXPECT().Health().Return(0) // dragon died
		mockPlayer.EXPECT().Field().Return(mockField)
		mockField.EXPECT().Warriors().Return([]cards.Warrior{mockFieldWarrior1, mockFieldWarrior2})
		mockFieldWarrior1.EXPECT().HealBy(2)
		mockFieldWarrior2.EXPECT().HealBy(2)
		mockHarpoon.EXPECT().GetID().Return("harpoon-id")
		mockPlayer.EXPECT().RemoveFromHand("harpoon-id").Return(nil, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockHarpoon)
		mockDragon.EXPECT().String().Return("Dragon (0)")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionHarpoon, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("No healing when dragon survives", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer, mockDragon, mockHarpoon := validateHarpoon(t, ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockHarpoon.EXPECT().Attack(mockDragon).Return(nil)
		mockGame.EXPECT().EventHandler().Return(bloodlustEvent())
		mockDragon.EXPECT().Health().Return(5) // dragon survived — HealBy must NOT be called
		mockHarpoon.EXPECT().GetID().Return("harpoon-id")
		mockPlayer.EXPECT().RemoveFromHand("harpoon-id").Return(nil, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockHarpoon)
		mockDragon.EXPECT().String().Return("Dragon (5)")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}
		mockGame.EXPECT().Status(mockPlayer).Return(expectedStatus)
		result, statusFn, err := action.Execute(mockGame)
		_ = statusFn()

		assert.NoError(t, err)
		assert.Equal(t, types.LastActionHarpoon, result.Action)
	})

	t.Run("No healing when no bloodlust event", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer, mockDragon, mockHarpoon := validateHarpoon(t, ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockHarpoon.EXPECT().Attack(mockDragon).Return(nil)
		mockGame.EXPECT().EventHandler().Return(calmEvent())
		// HealBy must NOT be called; Health must NOT be called
		mockHarpoon.EXPECT().GetID().Return("harpoon-id")
		mockPlayer.EXPECT().RemoveFromHand("harpoon-id").Return(nil, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockHarpoon)
		mockDragon.EXPECT().String().Return("Dragon (10)")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}
		mockGame.EXPECT().Status(mockPlayer).Return(expectedStatus)
		result, statusFn, err := action.Execute(mockGame)
		_ = statusFn()

		assert.NoError(t, err)
		assert.Equal(t, types.LastActionHarpoon, result.Action)
	})
}

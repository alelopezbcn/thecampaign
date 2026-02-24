package gameactions_test

import (
	"errors"
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/gameactions"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// setUpSPArcherEnemySearch configures mock expectations for searching an enemy field in 1v1.
func setUpSPArcherEnemySearch(
	ctrl *gomock.Controller,
	mockGame *mocks.MockGame,
	mockPlayer1 *mocks.MockPlayer,
	mockPlayer2 *mocks.MockPlayer,
	targetID string,
	targetReturn board.Player,
	targetFound bool,
) {
	// Own field miss
	mockPlayer1.EXPECT().GetCardFromField(targetID).Return(nil, false)
	// Allies search (no allies in 1v1)
	mockGame.EXPECT().PlayerIndex("Player1").Return(0)
	mockGame.EXPECT().Allies(0).Return([]board.Player{})
	// Enemies search
	mockGame.EXPECT().PlayerIndex("Player1").Return(0)
	if targetFound {
		target, _ := targetReturn.(board.Player) // use the mock directly below
		_ = target
		mockGame.EXPECT().Enemies(0).Return([]board.Player{mockPlayer2})
	} else {
		mockGame.EXPECT().Enemies(0).Return([]board.Player{mockPlayer2})
	}
}

// validateSPActionArcher runs Validate for archer targeting enemy warrior.
func validateSPActionArcher(
	t *testing.T, ctrl *gomock.Controller,
) (gameactions.GameAction, *mocks.MockGame, *mocks.MockPlayer, *mocks.MockWarrior, *mocks.MockWarrior, *mocks.MockSpecialPower) {
	t.Helper()
	mockGame := mocks.NewMockGame(ctrl)
	mockPlayer1 := mocks.NewMockPlayer(ctrl)
	mockPlayer2 := mocks.NewMockPlayer(ctrl)
	mockWarrior := mocks.NewMockWarrior(ctrl)
	mockTarget := mocks.NewMockWarrior(ctrl)
	mockSP := mocks.NewMockSpecialPower(ctrl)

	mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
	mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
	mockPlayer1.EXPECT().GetCardFromField("A1").Return(mockWarrior, true)
	mockWarrior.EXPECT().Type().Return(types.ArcherWarriorType)
	// Target search: own field miss, no allies, enemy hit
	mockPlayer1.EXPECT().GetCardFromField("EK1").Return(nil, false)
	mockGame.EXPECT().PlayerIndex("Player1").Return(0)
	mockGame.EXPECT().Allies(0).Return([]board.Player{})
	mockGame.EXPECT().PlayerIndex("Player1").Return(0)
	mockGame.EXPECT().Enemies(0).Return([]board.Player{mockPlayer2})
	mockPlayer2.EXPECT().GetCardFromField("EK1").Return(mockTarget, true)
	// Weapon
	mockPlayer1.EXPECT().GetCardFromHand("SP1").Return(mockSP, true)

	action := gameactions.NewSpecialPowerAction("Player1", "A1", "EK1", "SP1")
	if err := action.Validate(mockGame); err != nil {
		t.Fatalf("validateSPActionArcher: unexpected error: %v", err)
	}
	return action, mockGame, mockPlayer1, mockWarrior, mockTarget, mockSP
}

func TestSpecialPowerAction_PlayerName(t *testing.T) {
	action := gameactions.NewSpecialPowerAction("Player1", "K1", "EK1", "SP1")
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestSpecialPowerAction_NextPhase(t *testing.T) {
	action := gameactions.NewSpecialPowerAction("Player1", "K1", "EK1", "SP1")
	assert.Equal(t, types.PhaseTypeSpySteal, action.NextPhase())
}

func TestSpecialPowerAction_Validate(t *testing.T) {
	t.Run("Error when not in Attack phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeBuy).Times(2)

		action := gameactions.NewSpecialPowerAction("Player1", "K1", "EK1", "SP1")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot use special power in the")
	})

	t.Run("Error when warrior not in field", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromField("K1").Return(nil, false)

		action := gameactions.NewSpecialPowerAction("Player1", "K1", "EK1", "SP1")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "warrior card not in field")
	})

	t.Run("Error when user card is not a warrior", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockCard := mocks.NewMockCard(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromField("K1").Return(mockCard, true)

		action := gameactions.NewSpecialPowerAction("Player1", "K1", "EK1", "SP1")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "the attacking card is not a warrior")
	})

	t.Run("Error when target not in any field", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromField("K1").Return(mockWarrior, true)
		mockWarrior.EXPECT().Type().Return(types.ArcherWarriorType)
		mockPlayer1.EXPECT().GetCardFromField("EK1").Return(nil, false)
		mockGame.EXPECT().PlayerIndex("Player1").Return(0)
		mockGame.EXPECT().Allies(0).Return([]board.Player{})
		mockGame.EXPECT().PlayerIndex("Player1").Return(0)
		mockGame.EXPECT().Enemies(0).Return([]board.Player{mockPlayer2})
		mockPlayer2.EXPECT().GetCardFromField("EK1").Return(nil, false)

		action := gameactions.NewSpecialPowerAction("Player1", "K1", "EK1", "SP1")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target card not valid")
	})

	t.Run("Error when archer targets ally", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockTarget := mocks.NewMockWarrior(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromField("A1").Return(mockWarrior, true)
		mockWarrior.EXPECT().Type().Return(types.ArcherWarriorType)
		// Target found in own field (ally)
		mockPlayer1.EXPECT().GetCardFromField("T1").Return(mockTarget, true)

		action := gameactions.NewSpecialPowerAction("Player1", "A1", "T1", "SP1")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "archer instant kill can only target enemies")
	})

	t.Run("Error when knight targets enemy", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockTarget := mocks.NewMockWarrior(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromField("K1").Return(mockWarrior, true)
		mockWarrior.EXPECT().Type().Return(types.KnightWarriorType)
		mockPlayer1.EXPECT().GetCardFromField("EK1").Return(nil, false)
		mockGame.EXPECT().PlayerIndex("Player1").Return(0)
		mockGame.EXPECT().Allies(0).Return([]board.Player{})
		mockGame.EXPECT().PlayerIndex("Player1").Return(0)
		mockGame.EXPECT().Enemies(0).Return([]board.Player{mockPlayer2})
		mockPlayer2.EXPECT().GetCardFromField("EK1").Return(mockTarget, true)

		action := gameactions.NewSpecialPowerAction("Player1", "K1", "EK1", "SP1")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "knight/mage special power can only target allies")
	})

	t.Run("Error when weapon not in hand", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockTarget := mocks.NewMockWarrior(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromField("K1").Return(mockWarrior, true)
		mockWarrior.EXPECT().Type().Return(types.ArcherWarriorType)
		mockPlayer1.EXPECT().GetCardFromField("EK1").Return(nil, false)
		mockGame.EXPECT().PlayerIndex("Player1").Return(0)
		mockGame.EXPECT().Allies(0).Return([]board.Player{})
		mockGame.EXPECT().PlayerIndex("Player1").Return(0)
		mockGame.EXPECT().Enemies(0).Return([]board.Player{mockPlayer2})
		mockPlayer2.EXPECT().GetCardFromField("EK1").Return(mockTarget, true)
		mockPlayer1.EXPECT().GetCardFromHand("SP1").Return(nil, false)

		action := gameactions.NewSpecialPowerAction("Player1", "K1", "EK1", "SP1")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "weapon card not in hand")
	})

	t.Run("Error when card is not a special power", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockTarget := mocks.NewMockWarrior(ctrl)
		mockResource := mocks.NewMockResource(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromField("K1").Return(mockWarrior, true)
		mockWarrior.EXPECT().Type().Return(types.ArcherWarriorType)
		mockPlayer1.EXPECT().GetCardFromField("EK1").Return(nil, false)
		mockGame.EXPECT().PlayerIndex("Player1").Return(0)
		mockGame.EXPECT().Allies(0).Return([]board.Player{})
		mockGame.EXPECT().PlayerIndex("Player1").Return(0)
		mockGame.EXPECT().Enemies(0).Return([]board.Player{mockPlayer2})
		mockPlayer2.EXPECT().GetCardFromField("EK1").Return(mockTarget, true)
		mockPlayer1.EXPECT().GetCardFromHand("SP1").Return(mockResource, true)

		action := gameactions.NewSpecialPowerAction("Player1", "K1", "EK1", "SP1")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "the card is not a special power")
	})

	t.Run("Error when target card is not a warrior", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockTargetCard := mocks.NewMockCard(ctrl)
		mockSP := mocks.NewMockSpecialPower(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromField("A1").Return(mockWarrior, true)
		mockWarrior.EXPECT().Type().Return(types.ArcherWarriorType)
		mockPlayer1.EXPECT().GetCardFromField("EK1").Return(nil, false)
		mockGame.EXPECT().PlayerIndex("Player1").Return(0)
		mockGame.EXPECT().Allies(0).Return([]board.Player{})
		mockGame.EXPECT().PlayerIndex("Player1").Return(0)
		mockGame.EXPECT().Enemies(0).Return([]board.Player{mockPlayer2})
		mockPlayer2.EXPECT().GetCardFromField("EK1").Return(mockTargetCard, true)
		mockPlayer1.EXPECT().GetCardFromHand("SP1").Return(mockSP, true)

		action := gameactions.NewSpecialPowerAction("Player1", "A1", "EK1", "SP1")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "the target card is not a warrior")
	})

	t.Run("Success on enemy target stores fields", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, _, _, _, _, _ := validateSPActionArcher(t, ctrl)
		assert.NotNil(t, action)
	})

	t.Run("Success on own target (knight protect/heal)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockTarget := mocks.NewMockWarrior(ctrl)
		mockSP := mocks.NewMockSpecialPower(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromField("K1").Return(mockWarrior, true)
		mockWarrior.EXPECT().Type().Return(types.KnightWarriorType)
		// Target found in own field
		mockPlayer1.EXPECT().GetCardFromField("A1").Return(mockTarget, true)
		mockPlayer1.EXPECT().GetCardFromHand("SP1").Return(mockSP, true)

		action := gameactions.NewSpecialPowerAction("Player1", "K1", "A1", "SP1")
		err := action.Validate(mockGame)

		assert.NoError(t, err)
	})
}

func TestSpecialPowerAction_Execute(t *testing.T) {
	t.Run("Error when RemoveFromHand fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockWarrior, mockTarget, mockSP := validateSPActionArcher(t, ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockSP.EXPECT().Use(mockWarrior, mockTarget).Return(nil)
		mockSP.EXPECT().GetID().Return("SP1")
		mockPlayer1.EXPECT().RemoveFromHand("SP1").Return(nil, errors.New("card not found"))

		result, _, err := action.Execute(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "removing special power from hand failed")
		assert.NotNil(t, result)
	})

	t.Run("Error when special power fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockWarrior, mockTarget, mockSP := validateSPActionArcher(t, ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockSP.EXPECT().Use(mockWarrior, mockTarget).Return(errors.New("power failed"))

		result, _, err := action.Execute(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "special power action failed")
		assert.NotNil(t, result)
	})

	t.Run("Success returns result", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockWarrior, mockTarget, mockSP := validateSPActionArcher(t, ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockSP.EXPECT().Use(mockWarrior, mockTarget).Return(nil)
		mockSP.EXPECT().GetID().Return("SP1")
		mockPlayer1.EXPECT().RemoveFromHand("SP1").Return(nil, nil)
		mockTarget.EXPECT().String().Return("Knight (20)")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer1).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionSpecialPower, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("History is updated on success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockWarrior, mockTarget, mockSP := validateSPActionArcher(t, ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockSP.EXPECT().Use(mockWarrior, mockTarget).Return(nil)
		mockSP.EXPECT().GetID().Return("SP1")
		mockPlayer1.EXPECT().RemoveFromHand("SP1").Return(nil, nil)
		mockTarget.EXPECT().String().Return("Knight (20)")

		var capturedMsg string
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Do(func(msg string, _ types.Category) {
			capturedMsg = msg
		})
		// statusFn not called; no Status expectation

		_, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		assert.Contains(t, capturedMsg, "Player1")
		assert.Contains(t, capturedMsg, "special power")
	})
}

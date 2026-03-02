package gameactions_test

import (
	"errors"
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/gameactions"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// validateMoveOwnField runs Validate for an own-field warrior move.
func validateMoveOwnField(
	t *testing.T, ctrl *gomock.Controller, warriorID string,
) (gameactions.GameAction, *mocks.MockGame, *mocks.MockPlayer) {
	t.Helper()
	mockGame := mocks.NewMockGame(ctrl)
	mockPlayer1 := mocks.NewMockPlayer(ctrl)

	mockGame.EXPECT().TurnState().Return(types.TurnState{})
	mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)

	action := gameactions.NewMoveWarriorAction("Player1", warriorID, "")
	if err := action.Validate(mockGame); err != nil {
		t.Fatalf("validateMoveOwnField: unexpected error: %v", err)
	}
	return action, mockGame, mockPlayer1
}

// validateMoveAllyField runs Validate for a 2v2 ally-field warrior move.
func validateMoveAllyField(
	t *testing.T, ctrl *gomock.Controller,
) (gameactions.GameAction, *mocks.MockGame, *mocks.MockPlayer, *mocks.MockPlayer, *mocks.MockWarrior) {
	t.Helper()
	mockGame := mocks.NewMockGame(ctrl)
	mockPlayer1 := mocks.NewMockPlayer(ctrl)
	mockPlayer2 := mocks.NewMockPlayer(ctrl)
	mockWarrior := mocks.NewMockWarrior(ctrl)

	mockGame.EXPECT().TurnState().Return(types.TurnState{})
	mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
	mockGame.EXPECT().GetPlayer("Player2").Return(mockPlayer2)
	mockGame.EXPECT().PlayerIndex("Player1").Return(0)
	mockGame.EXPECT().PlayerIndex("Player2").Return(1)
	mockGame.EXPECT().SameTeam(0, 1).Return(true)
	mockPlayer1.EXPECT().GetCardFromHand("K1").Return(mockWarrior, true)

	action := gameactions.NewMoveWarriorAction("Player1", "K1", "Player2")
	if err := action.Validate(mockGame); err != nil {
		t.Fatalf("validateMoveAllyField: unexpected error: %v", err)
	}
	return action, mockGame, mockPlayer1, mockPlayer2, mockWarrior
}

func TestMoveWarriorAction_PlayerName(t *testing.T) {
	action := gameactions.NewMoveWarriorAction("Player1", "K1", "")
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestMoveWarriorAction_Validate(t *testing.T) {
	t.Run("Error when already moved warrior this turn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockGame.EXPECT().TurnState().Return(types.TurnState{HasMovedWarrior: true})

		action := gameactions.NewMoveWarriorAction("Player1", "K1", "")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already moved a warrior this turn")
	})

	t.Run("Success for own field move", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, _, _ := validateMoveOwnField(t, ctrl, "K1")
		assert.NotNil(t, action)
	})

	t.Run("Error when target player not found for ally move", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)

		mockGame.EXPECT().TurnState().Return(types.TurnState{})
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockGame.EXPECT().GetPlayer("Unknown").Return(nil)

		action := gameactions.NewMoveWarriorAction("Player1", "K1", "Unknown")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target player Unknown not found")
	})

	t.Run("Error when target is not ally", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockGame.EXPECT().TurnState().Return(types.TurnState{})
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockGame.EXPECT().GetPlayer("Player2").Return(mockPlayer2)
		mockGame.EXPECT().PlayerIndex("Player1").Return(0)
		mockGame.EXPECT().PlayerIndex("Player2").Return(1)
		mockGame.EXPECT().SameTeam(0, 1).Return(false)

		action := gameactions.NewMoveWarriorAction("Player1", "K1", "Player2")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can only move warriors to ally's field")
	})

	t.Run("Error when card not in hand for ally move", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockGame.EXPECT().TurnState().Return(types.TurnState{})
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockGame.EXPECT().GetPlayer("Player2").Return(mockPlayer2)
		mockGame.EXPECT().PlayerIndex("Player1").Return(0)
		mockGame.EXPECT().PlayerIndex("Player2").Return(1)
		mockGame.EXPECT().SameTeam(0, 1).Return(true)
		mockPlayer1.EXPECT().GetCardFromHand("K1").Return(nil, false)

		action := gameactions.NewMoveWarriorAction("Player1", "K1", "Player2")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "card with ID K1 not found in hand")
	})

	t.Run("Error when card is not a warrior for ally move", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockCard := mocks.NewMockCard(ctrl)

		mockGame.EXPECT().TurnState().Return(types.TurnState{})
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockGame.EXPECT().GetPlayer("Player2").Return(mockPlayer2)
		mockGame.EXPECT().PlayerIndex("Player1").Return(0)
		mockGame.EXPECT().PlayerIndex("Player2").Return(1)
		mockGame.EXPECT().SameTeam(0, 1).Return(true)
		mockPlayer1.EXPECT().GetCardFromHand("G1").Return(mockCard, true)

		action := gameactions.NewMoveWarriorAction("Player1", "G1", "Player2")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "only warrior cards can be moved to field")
	})

	t.Run("Success stores target player and warrior for ally move", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, _, _, _, _ := validateMoveAllyField(t, ctrl)
		assert.NotNil(t, action)
	})
}

func TestMoveWarriorAction_Execute(t *testing.T) {
	t.Run("Error when MoveCardToField fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1 := validateMoveOwnField(t, ctrl, "K1")

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().MoveCardToField("K1").Return(errors.New("card not found"))

		result, _, err := action.Execute(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "moving warrior to field failed")
		assert.NotNil(t, result)
	})

	t.Run("Success moving warrior to own field", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1 := validateMoveOwnField(t, ctrl, "K1")
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().MoveCardToField("K1").Return(nil)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().SetHasMovedWarrior(true)
		mockGame.EXPECT().SetCanMoveWarrior(false)
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().Status(mockPlayer1).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionMoveWarrior, result.Action)
		assert.Equal(t, "K1", result.MovedWarriorID)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Success moving warrior to ally field in 2v2", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockPlayer2, mockWarrior := validateMoveAllyField(t, ctrl)
		mockHand := mocks.NewMockHand(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer2.EXPECT().PlaceWarriorOnField(mockWarrior)
		mockPlayer1.EXPECT().Hand().Return(mockHand)
		mockHand.EXPECT().RemoveCard(mockWarrior).Return(true)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockPlayer2.EXPECT().Name().Return("Player2")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().SetHasMovedWarrior(true)
		mockGame.EXPECT().SetCanMoveWarrior(false)
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().Status(mockPlayer1).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionMoveWarrior, result.Action)
		assert.Equal(t, "K1", result.MovedWarriorID)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("NextPhase returns current game phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1 := validateMoveOwnField(t, ctrl, "K1")

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().MoveCardToField("K1").Return(nil)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().SetHasMovedWarrior(true)
		mockGame.EXPECT().SetCanMoveWarrior(false)
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeBuy)
		// statusFn not called; no Status expectation

		_, _, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.Equal(t, types.PhaseTypeBuy, action.NextPhase())
	})

	t.Run("History updated on own field move", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1 := validateMoveOwnField(t, ctrl, "K1")

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().MoveCardToField("K1").Return(nil)
		mockPlayer1.EXPECT().Name().Return("Player1")

		var capturedMsg string
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Do(func(msg string, _ types.Category) {
			capturedMsg = msg
		})
		mockGame.EXPECT().SetHasMovedWarrior(true)
		mockGame.EXPECT().SetCanMoveWarrior(false)
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)

		_, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		assert.Contains(t, capturedMsg, "Player1")
		assert.Contains(t, capturedMsg, "moved warrior")
	})

	t.Run("History updated on ally field move", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockPlayer2, mockWarrior := validateMoveAllyField(t, ctrl)
		mockHand := mocks.NewMockHand(ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer2.EXPECT().PlaceWarriorOnField(mockWarrior)
		mockPlayer1.EXPECT().Hand().Return(mockHand)
		mockHand.EXPECT().RemoveCard(mockWarrior).Return(true)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockPlayer2.EXPECT().Name().Return("Player2")

		var capturedMsg string
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Do(func(msg string, _ types.Category) {
			capturedMsg = msg
		})
		mockGame.EXPECT().SetHasMovedWarrior(true)
		mockGame.EXPECT().SetCanMoveWarrior(false)
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)

		_, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		assert.Contains(t, capturedMsg, "Player1")
		assert.Contains(t, capturedMsg, "Player2's field")
	})
}

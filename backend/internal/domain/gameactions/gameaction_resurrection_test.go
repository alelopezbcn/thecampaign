package gameactions_test

import (
	"errors"
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gameactions"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// validateResurrectionOwnField runs Validate for a resurrection targeting the player's own field.
func validateResurrectionOwnField(
	t *testing.T, ctrl *gomock.Controller,
) (gameactions.GameAction, *mocks.MockGame, *mocks.MockPlayer, *mocks.MockResurrection, *mocks.MockBoard, *mocks.MockCemetery) {
	t.Helper()
	mockGame := mocks.NewMockGame(ctrl)
	mockPlayer := mocks.NewMockPlayer(ctrl)
	mockResurrection := mocks.NewMockResurrection(ctrl)
	mockBoard := mocks.NewMockBoard(ctrl)
	mockCemetery := mocks.NewMockCemetery(ctrl)

	mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
	mockGame.EXPECT().Board().Return(mockBoard)
	mockBoard.EXPECT().Cemetery().Return(mockCemetery)
	mockCemetery.EXPECT().Count().Return(1)
	mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
	mockPlayer.EXPECT().GetCardFromHand("resurrection-id").Return(mockResurrection, true)
	mockGame.EXPECT().GetPlayer("Player1").Return(mockPlayer)

	action := gameactions.NewResurrectionAction("Player1", "", "resurrection-id")
	if err := action.Validate(mockGame); err != nil {
		t.Fatalf("validateResurrectionOwnField: unexpected error: %v", err)
	}
	return action, mockGame, mockPlayer, mockResurrection, mockBoard, mockCemetery
}

// validateResurrectionAllyField runs Validate for a resurrection targeting an ally's field (2v2).
func validateResurrectionAllyField(
	t *testing.T, ctrl *gomock.Controller,
) (gameactions.GameAction, *mocks.MockGame, *mocks.MockPlayer, *mocks.MockPlayer, *mocks.MockResurrection, *mocks.MockBoard, *mocks.MockCemetery) {
	t.Helper()
	mockGame := mocks.NewMockGame(ctrl)
	mockPlayer1 := mocks.NewMockPlayer(ctrl)
	mockPlayer2 := mocks.NewMockPlayer(ctrl)
	mockResurrection := mocks.NewMockResurrection(ctrl)
	mockBoard := mocks.NewMockBoard(ctrl)
	mockCemetery := mocks.NewMockCemetery(ctrl)

	mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
	mockGame.EXPECT().Board().Return(mockBoard)
	mockBoard.EXPECT().Cemetery().Return(mockCemetery)
	mockCemetery.EXPECT().Count().Return(1)
	mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
	mockPlayer1.EXPECT().GetCardFromHand("resurrection-id").Return(mockResurrection, true)
	mockGame.EXPECT().GetPlayer("Player2").Return(mockPlayer2)
	mockGame.EXPECT().PlayerIndex("Player1").Return(0)
	mockGame.EXPECT().PlayerIndex("Player2").Return(1)
	mockGame.EXPECT().SameTeam(0, 1).Return(true)

	action := gameactions.NewResurrectionAction("Player1", "Player2", "resurrection-id")
	if err := action.Validate(mockGame); err != nil {
		t.Fatalf("validateResurrectionAllyField: unexpected error: %v", err)
	}
	return action, mockGame, mockPlayer1, mockPlayer2, mockResurrection, mockBoard, mockCemetery
}

func TestResurrectionAction_PlayerName(t *testing.T) {
	action := gameactions.NewResurrectionAction("Player1", "", "resurrection-id")
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestResurrectionAction_Validate(t *testing.T) {
	t.Run("Error when not in Attack phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeDrawCard).Times(2)

		action := gameactions.NewResurrectionAction("Player1", "", "resurrection-id")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot use resurrection in the")
	})

	t.Run("Error when cemetery is empty", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockBoard := mocks.NewMockBoard(ctrl)
		mockCemetery := mocks.NewMockCemetery(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().Board().Return(mockBoard)
		mockBoard.EXPECT().Cemetery().Return(mockCemetery)
		mockCemetery.EXPECT().Count().Return(0)

		action := gameactions.NewResurrectionAction("Player1", "", "resurrection-id")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no warriors in the cemetery to resurrect")
	})

	t.Run("Error when card not found in hand", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer := mocks.NewMockPlayer(ctrl)
		mockBoard := mocks.NewMockBoard(ctrl)
		mockCemetery := mocks.NewMockCemetery(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().Board().Return(mockBoard)
		mockBoard.EXPECT().Cemetery().Return(mockCemetery)
		mockCemetery.EXPECT().Count().Return(1)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockPlayer.EXPECT().GetCardFromHand("resurrection-id").Return(nil, false)

		action := gameactions.NewResurrectionAction("Player1", "", "resurrection-id")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found in hand")
	})

	t.Run("Error when card found but wrong type", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer := mocks.NewMockPlayer(ctrl)
		mockBoard := mocks.NewMockBoard(ctrl)
		mockCemetery := mocks.NewMockCemetery(ctrl)
		mockCard := mocks.NewMockCard(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().Board().Return(mockBoard)
		mockBoard.EXPECT().Cemetery().Return(mockCemetery)
		mockCemetery.EXPECT().Count().Return(1)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockPlayer.EXPECT().GetCardFromHand("resurrection-id").Return(mockCard, true)

		action := gameactions.NewResurrectionAction("Player1", "", "resurrection-id")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a resurrection card")
	})

	t.Run("Error when target player not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer := mocks.NewMockPlayer(ctrl)
		mockBoard := mocks.NewMockBoard(ctrl)
		mockCemetery := mocks.NewMockCemetery(ctrl)
		mockResurrection := mocks.NewMockResurrection(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().Board().Return(mockBoard)
		mockBoard.EXPECT().Cemetery().Return(mockCemetery)
		mockCemetery.EXPECT().Count().Return(1)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockPlayer.EXPECT().GetCardFromHand("resurrection-id").Return(mockResurrection, true)
		mockGame.EXPECT().GetPlayer("Unknown").Return(nil)

		action := gameactions.NewResurrectionAction("Player1", "Unknown", "resurrection-id")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target player Unknown not found")
	})

	t.Run("Error when target is not an ally", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockBoard := mocks.NewMockBoard(ctrl)
		mockCemetery := mocks.NewMockCemetery(ctrl)
		mockResurrection := mocks.NewMockResurrection(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().Board().Return(mockBoard)
		mockBoard.EXPECT().Cemetery().Return(mockCemetery)
		mockCemetery.EXPECT().Count().Return(1)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("resurrection-id").Return(mockResurrection, true)
		mockGame.EXPECT().GetPlayer("Player2").Return(mockPlayer2)
		mockGame.EXPECT().PlayerIndex("Player1").Return(0)
		mockGame.EXPECT().PlayerIndex("Player2").Return(1)
		mockGame.EXPECT().SameTeam(0, 1).Return(false)

		action := gameactions.NewResurrectionAction("Player1", "Player2", "resurrection-id")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not an ally")
	})

	t.Run("Success for own field", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, _, _, _, _, _ := validateResurrectionOwnField(t, ctrl)
		assert.NotNil(t, action)
	})

	t.Run("Success for ally field in 2v2", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, _, _, _, _, _, _ := validateResurrectionAllyField(t, ctrl)
		assert.NotNil(t, action)
	})
}

func TestResurrectionAction_Execute(t *testing.T) {
	t.Run("Error when cemetery returns no warrior", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer, _, mockBoard, mockCemetery := validateResurrectionOwnField(t, ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockGame.EXPECT().Board().Return(mockBoard)
		mockBoard.EXPECT().Cemetery().Return(mockCemetery)
		mockCemetery.EXPECT().RemoveRandom().Return(nil)

		result, statusFn, err := action.Execute(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no warriors in the cemetery")
		assert.Nil(t, result)
		assert.Nil(t, statusFn)
	})

	t.Run("Error when RemoveFromHand fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer, mockResurrection, mockBoard, mockCemetery := validateResurrectionOwnField(t, ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockField := mocks.NewMockField(ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockGame.EXPECT().Board().Return(mockBoard)
		mockBoard.EXPECT().Cemetery().Return(mockCemetery)
		mockCemetery.EXPECT().RemoveRandom().Return(mockWarrior)
		mockWarrior.EXPECT().Resurrect()
		mockPlayer.EXPECT().Field().Return(mockField)
		mockField.EXPECT().AddWarriors(mockWarrior)
		mockResurrection.EXPECT().GetID().Return("res1")
		mockPlayer.EXPECT().RemoveFromHand("res1").Return(nil, errors.New("card not found"))

		result, statusFn, err := action.Execute(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "removing resurrection card from hand")
		assert.Nil(t, result)
		assert.Nil(t, statusFn)
	})

	t.Run("Success resurrecting to own field", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer, mockResurrection, mockBoard, mockCemetery := validateResurrectionOwnField(t, ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockField := mocks.NewMockField(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockGame.EXPECT().Board().Return(mockBoard)
		mockBoard.EXPECT().Cemetery().Return(mockCemetery)
		mockCemetery.EXPECT().RemoveRandom().Return(mockWarrior)
		mockWarrior.EXPECT().Resurrect()
		mockPlayer.EXPECT().Field().Return(mockField)
		mockField.EXPECT().AddWarriors(mockWarrior)
		mockResurrection.EXPECT().GetID().Return("res1")
		mockPlayer.EXPECT().RemoveFromHand("res1").Return([]cards.Card{mockResurrection}, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockResurrection)
		mockPlayer.EXPECT().Name().Return("Player1").Times(2)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().Status(mockPlayer).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionResurrection, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Success resurrecting to ally field in 2v2", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockPlayer2, mockResurrection, mockBoard, mockCemetery := validateResurrectionAllyField(t, ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockField := mocks.NewMockField(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockGame.EXPECT().Board().Return(mockBoard)
		mockBoard.EXPECT().Cemetery().Return(mockCemetery)
		mockCemetery.EXPECT().RemoveRandom().Return(mockWarrior)
		mockWarrior.EXPECT().Resurrect()
		mockPlayer2.EXPECT().Field().Return(mockField)
		mockField.EXPECT().AddWarriors(mockWarrior)
		mockResurrection.EXPECT().GetID().Return("res1")
		mockPlayer1.EXPECT().RemoveFromHand("res1").Return([]cards.Card{mockResurrection}, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockResurrection)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockPlayer2.EXPECT().Name().Return("Player2")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().Status(mockPlayer1).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionResurrection, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("NextPhase returns current game phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer, mockResurrection, mockBoard, mockCemetery := validateResurrectionOwnField(t, ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockField := mocks.NewMockField(ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockGame.EXPECT().Board().Return(mockBoard)
		mockBoard.EXPECT().Cemetery().Return(mockCemetery)
		mockCemetery.EXPECT().RemoveRandom().Return(mockWarrior)
		mockWarrior.EXPECT().Resurrect()
		mockPlayer.EXPECT().Field().Return(mockField)
		mockField.EXPECT().AddWarriors(mockWarrior)
		mockResurrection.EXPECT().GetID().Return("res1")
		mockPlayer.EXPECT().RemoveFromHand("res1").Return([]cards.Card{mockResurrection}, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockResurrection)
		mockPlayer.EXPECT().Name().Return("Player1").Times(2)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeBuy)
		// statusFn not called; no Status expectation

		_, _, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.Equal(t, types.PhaseTypeBuy, action.NextPhase())
	})

	t.Run("History message for own field resurrection", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer, mockResurrection, mockBoard, mockCemetery := validateResurrectionOwnField(t, ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockField := mocks.NewMockField(ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockGame.EXPECT().Board().Return(mockBoard)
		mockBoard.EXPECT().Cemetery().Return(mockCemetery)
		mockCemetery.EXPECT().RemoveRandom().Return(mockWarrior)
		mockWarrior.EXPECT().Resurrect()
		mockPlayer.EXPECT().Field().Return(mockField)
		mockField.EXPECT().AddWarriors(mockWarrior)
		mockResurrection.EXPECT().GetID().Return("res1")
		mockPlayer.EXPECT().RemoveFromHand("res1").Return([]cards.Card{mockResurrection}, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockResurrection)
		mockPlayer.EXPECT().Name().Return("Player1").Times(2)

		var capturedMsg string
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Do(func(msg string, _ types.Category) {
			capturedMsg = msg
		})
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)

		_, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		assert.Contains(t, capturedMsg, "Player1")
		assert.Contains(t, capturedMsg, "resurrected")
	})

	t.Run("History message for ally field resurrection", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockPlayer2, mockResurrection, mockBoard, mockCemetery := validateResurrectionAllyField(t, ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockField := mocks.NewMockField(ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockGame.EXPECT().Board().Return(mockBoard)
		mockBoard.EXPECT().Cemetery().Return(mockCemetery)
		mockCemetery.EXPECT().RemoveRandom().Return(mockWarrior)
		mockWarrior.EXPECT().Resurrect()
		mockPlayer2.EXPECT().Field().Return(mockField)
		mockField.EXPECT().AddWarriors(mockWarrior)
		mockResurrection.EXPECT().GetID().Return("res1")
		mockPlayer1.EXPECT().RemoveFromHand("res1").Return([]cards.Card{mockResurrection}, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockResurrection)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockPlayer2.EXPECT().Name().Return("Player2")

		var capturedMsg string
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Do(func(msg string, _ types.Category) {
			capturedMsg = msg
		})
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)

		_, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		assert.Contains(t, capturedMsg, "Player1")
		assert.Contains(t, capturedMsg, "Player2")
	})
}

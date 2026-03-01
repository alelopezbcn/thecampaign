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

// validateCatapultAction runs Validate with mockCatapult looked up by "catapult-id".
func validateCatapultAction(
	t *testing.T, ctrl *gomock.Controller,
	targetPlayerName string, cardPosition int,
) (gameactions.GameAction, *mocks.MockGame, *mocks.MockPlayer, *mocks.MockPlayer, *mocks.MockCatapult) {
	t.Helper()

	mockGame := mocks.NewMockGame(ctrl)
	mockPlayer1 := mocks.NewMockPlayer(ctrl)
	mockPlayer2 := mocks.NewMockPlayer(ctrl)
	mockCatapult := mocks.NewMockCatapult(ctrl)

	mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
	mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
	mockPlayer1.EXPECT().GetCardFromHand("catapult-id").Return(mockCatapult, true)
	mockGame.EXPECT().GetTargetPlayer("Player1", targetPlayerName).Return(mockPlayer2, nil)

	action := gameactions.NewCatapultAction("Player1", targetPlayerName, cardPosition, "catapult-id")
	if err := action.Validate(mockGame); err != nil {
		t.Fatalf("validateCatapultAction: unexpected error: %v", err)
	}
	return action, mockGame, mockPlayer1, mockPlayer2, mockCatapult
}

func TestCatapultAction_PlayerName(t *testing.T) {
	action := gameactions.NewCatapultAction("Player1", "Player2", 0, "catapult-id")
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestCatapultAction_NextPhase(t *testing.T) {
	action := gameactions.NewCatapultAction("Player1", "Player2", 0, "catapult-id")
	assert.Equal(t, types.PhaseTypeSpySteal, action.NextPhase())
}

func TestCatapultAction_Validate(t *testing.T) {
	t.Run("Error when not in Attack phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeBuy).Times(2)

		action := gameactions.NewCatapultAction("Player1", "Player2", 0, "catapult-id")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot use catapult in the")
	})

	t.Run("Error when card not found in hand", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("catapult-id").Return(nil, false)

		action := gameactions.NewCatapultAction("Player1", "Player2", 0, "catapult-id")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found in hand")
	})

	t.Run("Error when card found but wrong type", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockCard := mocks.NewMockCard(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("catapult-id").Return(mockCard, true)

		action := gameactions.NewCatapultAction("Player1", "Player2", 0, "catapult-id")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a catapult")
	})

	t.Run("Success validates and stores catapult and target player", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, _, _, _, _ := validateCatapultAction(t, ctrl, "Player2", 2)
		assert.NotNil(t, action)
	})
}

func TestCatapultAction_Execute(t *testing.T) {
	t.Run("Error when castle attack fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockPlayer2, mockCatapult := validateCatapultAction(t, ctrl, "Player2", 0)
		mockCastle := mocks.NewMockCastle(ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer2.EXPECT().Castle().Return(mockCastle).Times(2)
		mockCastle.EXPECT().IsProtected().Return(false)
		mockCatapult.EXPECT().Attack(mockCastle, 0).Return(nil, errors.New("invalid position"))

		result, _, err := action.Execute(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "attacking castle failed")
		assert.NotNil(t, result)
	})

	t.Run("Success returns result and discards stolen gold", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockPlayer2, mockCatapult := validateCatapultAction(t, ctrl, "Player2", 2)
		mockCastle := mocks.NewMockCastle(ctrl)
		mockStolenGold := mocks.NewMockResource(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer2.EXPECT().Castle().Return(mockCastle).Times(2)
		mockCastle.EXPECT().IsProtected().Return(false)
		mockCatapult.EXPECT().Attack(mockCastle, 2).Return(mockStolenGold, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockStolenGold)
		mockCatapult.EXPECT().GetID().Return("catapult-id")
		mockPlayer1.EXPECT().RemoveFromHand("catapult-id").Return([]cards.Card{mockCatapult}, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockCatapult)
		mockStolenGold.EXPECT().Value().Return(3)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockPlayer2.EXPECT().Name().Return("Player2")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer1).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionCatapult, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("History is updated on success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockPlayer2, mockCatapult := validateCatapultAction(t, ctrl, "Player2", 1)
		mockCastle := mocks.NewMockCastle(ctrl)
		mockStolenGold := mocks.NewMockResource(ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer2.EXPECT().Castle().Return(mockCastle).Times(2)
		mockCastle.EXPECT().IsProtected().Return(false)
		mockCatapult.EXPECT().Attack(mockCastle, 1).Return(mockStolenGold, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockStolenGold)
		mockCatapult.EXPECT().GetID().Return("catapult-id")
		mockPlayer1.EXPECT().RemoveFromHand("catapult-id").Return([]cards.Card{mockCatapult}, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockCatapult)
		mockStolenGold.EXPECT().Value().Return(5)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockPlayer2.EXPECT().Name().Return("Player2")

		var capturedMsg string
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Do(func(msg string, _ types.Category) {
			capturedMsg = msg
		})

		_, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		assert.Contains(t, capturedMsg, "Player1")
		assert.Contains(t, capturedMsg, "gold")
		assert.Contains(t, capturedMsg, "Player2")
	})

	t.Run("Fortress blocks catapult — wall destroyed, gold not removed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockPlayer2, mockCatapult := validateCatapultAction(t, ctrl, "Player2", 1)
		mockCastle := mocks.NewMockCastle(ctrl)
		mockFortressCard := mocks.NewMockCard(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer2.EXPECT().Castle().Return(mockCastle).Times(2)
		mockCastle.EXPECT().IsProtected().Return(true)
		mockCastle.EXPECT().ConsumeProtection().Return(mockFortressCard)
		mockGame.EXPECT().OnCardMovedToPile(mockFortressCard)
		mockCatapult.EXPECT().GetID().Return("catapult-id")
		mockPlayer1.EXPECT().RemoveFromHand("catapult-id").Return([]cards.Card{mockCatapult}, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockCatapult)
		mockPlayer2.EXPECT().Name().Return("Player2")
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer1).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionCatapultBlocked, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})
}

package gameactions_test

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/gameactions"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// validateFortressOwnCastle runs Validate for own castle fortification.
func validateFortressOwnCastle(t *testing.T, ctrl *gomock.Controller) (
	gameactions.GameAction, *mocks.MockGame, *mocks.MockPlayer, *mocks.MockFortress, *mocks.MockCastle,
) {
	t.Helper()
	mockGame := mocks.NewMockGame(ctrl)
	mockPlayer := mocks.NewMockPlayer(ctrl)
	mockFortress := mocks.NewMockFortress(ctrl)
	mockCastle := mocks.NewMockCastle(ctrl)

	mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeConstruct)
	mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
	mockPlayer.EXPECT().GetCardFromHand("fortress-id").Return(mockFortress, true)
	mockGame.EXPECT().GetPlayer("Player1").Return(mockPlayer)
	mockPlayer.EXPECT().Castle().Return(mockCastle).Times(2)
	mockCastle.EXPECT().IsConstructed().Return(true)
	mockCastle.EXPECT().IsProtected().Return(false)

	action := gameactions.NewFortressAction("Player1", "", "fortress-id")
	if err := action.Validate(mockGame); err != nil {
		t.Fatalf("validateFortressOwnCastle: unexpected error: %v", err)
	}
	return action, mockGame, mockPlayer, mockFortress, mockCastle
}

// validateFortressAllyCastle runs Validate for ally castle fortification (2v2).
func validateFortressAllyCastle(t *testing.T, ctrl *gomock.Controller) (
	gameactions.GameAction, *mocks.MockGame, *mocks.MockPlayer, *mocks.MockPlayer, *mocks.MockFortress, *mocks.MockCastle,
) {
	t.Helper()
	mockGame := mocks.NewMockGame(ctrl)
	mockPlayer1 := mocks.NewMockPlayer(ctrl)
	mockPlayer2 := mocks.NewMockPlayer(ctrl)
	mockFortress := mocks.NewMockFortress(ctrl)
	mockCastle := mocks.NewMockCastle(ctrl)

	mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeConstruct)
	mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
	mockPlayer1.EXPECT().GetCardFromHand("fortress-id").Return(mockFortress, true)
	mockGame.EXPECT().GetPlayer("Player2").Return(mockPlayer2)
	mockGame.EXPECT().PlayerIndex("Player1").Return(0)
	mockGame.EXPECT().PlayerIndex("Player2").Return(1)
	mockGame.EXPECT().SameTeam(0, 1).Return(true)
	mockPlayer2.EXPECT().Castle().Return(mockCastle).Times(2)
	mockCastle.EXPECT().IsConstructed().Return(true)
	mockCastle.EXPECT().IsProtected().Return(false)

	action := gameactions.NewFortressAction("Player1", "Player2", "fortress-id")
	if err := action.Validate(mockGame); err != nil {
		t.Fatalf("validateFortressAllyCastle: unexpected error: %v", err)
	}
	return action, mockGame, mockPlayer1, mockPlayer2, mockFortress, mockCastle
}

func TestFortressAction_PlayerName(t *testing.T) {
	action := gameactions.NewFortressAction("Player1", "", "fortress-id")
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestFortressAction_NextPhase(t *testing.T) {
	action := gameactions.NewFortressAction("Player1", "", "fortress-id")
	assert.Equal(t, types.PhaseTypeEndTurn, action.NextPhase())
}

func TestFortressAction_Validate(t *testing.T) {
	t.Run("Error when not in Construct phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack).Times(2)

		action := gameactions.NewFortressAction("Player1", "", "fortress-id")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot use fortress in the")
	})

	t.Run("Error when card not found in hand", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer := mocks.NewMockPlayer(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeConstruct)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockPlayer.EXPECT().GetCardFromHand("fortress-id").Return(nil, false)

		action := gameactions.NewFortressAction("Player1", "", "fortress-id")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found in hand")
	})

	t.Run("Error when card found but wrong type", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer := mocks.NewMockPlayer(ctrl)
		mockCard := mocks.NewMockCard(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeConstruct)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockPlayer.EXPECT().GetCardFromHand("fortress-id").Return(mockCard, true)

		action := gameactions.NewFortressAction("Player1", "", "fortress-id")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a fortress")
	})

	t.Run("Error when target castle is not constructed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer := mocks.NewMockPlayer(ctrl)
		mockFortress := mocks.NewMockFortress(ctrl)
		mockCastle := mocks.NewMockCastle(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeConstruct)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockPlayer.EXPECT().GetCardFromHand("fortress-id").Return(mockFortress, true)
		mockGame.EXPECT().GetPlayer("Player1").Return(mockPlayer)
		mockPlayer.EXPECT().Castle().Return(mockCastle)
		mockCastle.EXPECT().IsConstructed().Return(false)

		action := gameactions.NewFortressAction("Player1", "", "fortress-id")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot fortify a castle that has not been constructed yet")
	})

	t.Run("Error when castle is already protected", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer := mocks.NewMockPlayer(ctrl)
		mockFortress := mocks.NewMockFortress(ctrl)
		mockCastle := mocks.NewMockCastle(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeConstruct)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockPlayer.EXPECT().GetCardFromHand("fortress-id").Return(mockFortress, true)
		mockGame.EXPECT().GetPlayer("Player1").Return(mockPlayer)
		mockPlayer.EXPECT().Castle().Return(mockCastle).Times(2)
		mockCastle.EXPECT().IsConstructed().Return(true)
		mockCastle.EXPECT().IsProtected().Return(true)

		action := gameactions.NewFortressAction("Player1", "", "fortress-id")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "castle is already protected by a fortress")
	})

	t.Run("Error when target is not an ally in 2v2", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockFortress := mocks.NewMockFortress(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeConstruct)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("fortress-id").Return(mockFortress, true)
		mockGame.EXPECT().GetPlayer("Player2").Return(mockPlayer2)
		mockGame.EXPECT().PlayerIndex("Player1").Return(0)
		mockGame.EXPECT().PlayerIndex("Player2").Return(1)
		mockGame.EXPECT().SameTeam(0, 1).Return(false)

		action := gameactions.NewFortressAction("Player1", "Player2", "fortress-id")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not an ally")
	})

	t.Run("Success for own castle", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, _, _, _, _ := validateFortressOwnCastle(t, ctrl)
		assert.NotNil(t, action)
	})

	t.Run("Success for ally castle in 2v2", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, _, _, _, _, _ := validateFortressAllyCastle(t, ctrl)
		assert.NotNil(t, action)
	})
}

func TestFortressAction_Execute(t *testing.T) {
	t.Run("Success placing fortress on own castle", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer, mockFortress, mockCastle := validateFortressOwnCastle(t, ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockPlayer.EXPECT().Castle().Return(mockCastle)
		mockCastle.EXPECT().SetProtection(mockFortress)
		mockFortress.EXPECT().GetID().Return("FW1")
		mockPlayer.EXPECT().RemoveFromHand("FW1").Return(nil, nil)
		mockPlayer.EXPECT().Name().Return("Player1").Times(2)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionFortress, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Success placing fortress on ally castle in 2v2", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockPlayer2, mockFortress, mockCastle := validateFortressAllyCastle(t, ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer2.EXPECT().Castle().Return(mockCastle)
		mockCastle.EXPECT().SetProtection(mockFortress)
		mockFortress.EXPECT().GetID().Return("FW1")
		mockPlayer1.EXPECT().RemoveFromHand("FW1").Return(nil, nil)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockPlayer2.EXPECT().Name().Return("Player2")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer1).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionFortress, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("History message contains player name for own castle", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer, mockFortress, mockCastle := validateFortressOwnCastle(t, ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockPlayer.EXPECT().Castle().Return(mockCastle)
		mockCastle.EXPECT().SetProtection(mockFortress)
		mockFortress.EXPECT().GetID().Return("FW1")
		mockPlayer.EXPECT().RemoveFromHand("FW1").Return(nil, nil)
		mockPlayer.EXPECT().Name().Return("Player1").Times(2)

		var capturedMsg string
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Do(func(msg string, _ types.Category) {
			capturedMsg = msg
		})

		_, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		assert.Contains(t, capturedMsg, "Player1")
		assert.Contains(t, capturedMsg, "fortified")
	})

	t.Run("History message contains ally name for ally castle", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockPlayer2, mockFortress, mockCastle := validateFortressAllyCastle(t, ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer2.EXPECT().Castle().Return(mockCastle)
		mockCastle.EXPECT().SetProtection(mockFortress)
		mockFortress.EXPECT().GetID().Return("FW1")
		mockPlayer1.EXPECT().RemoveFromHand("FW1").Return(nil, nil)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockPlayer2.EXPECT().Name().Return("Player2")

		var capturedMsg string
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Do(func(msg string, _ types.Category) {
			capturedMsg = msg
		})

		_, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		assert.Contains(t, capturedMsg, "Player2's castle")
	})
}

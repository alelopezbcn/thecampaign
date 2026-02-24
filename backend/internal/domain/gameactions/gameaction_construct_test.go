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

// validateConstructOwnCastle runs Validate for own castle construction.
func validateConstructOwnCastle(t *testing.T, ctrl *gomock.Controller) (
	gameactions.GameAction, *mocks.MockGame, *mocks.MockPlayer, *mocks.MockResource,
) {
	t.Helper()
	mockGame := mocks.NewMockGame(ctrl)
	mockPlayer1 := mocks.NewMockPlayer(ctrl)
	mockResource := mocks.NewMockResource(ctrl)

	mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeConstruct)
	mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
	mockPlayer1.EXPECT().GetCardFromHand("G1").Return(mockResource, true)

	action := gameactions.NewConstructAction("Player1", "G1", "")
	if err := action.Validate(mockGame); err != nil {
		t.Fatalf("validateConstructOwnCastle: unexpected error: %v", err)
	}
	return action, mockGame, mockPlayer1, mockResource
}

// validateConstructAllyCastle runs Validate for ally castle construction (2v2).
func validateConstructAllyCastle(t *testing.T, ctrl *gomock.Controller) (
	gameactions.GameAction, *mocks.MockGame, *mocks.MockPlayer, *mocks.MockPlayer, *mocks.MockResource,
) {
	t.Helper()
	mockGame := mocks.NewMockGame(ctrl)
	mockPlayer1 := mocks.NewMockPlayer(ctrl)
	mockPlayer2 := mocks.NewMockPlayer(ctrl)
	mockResource := mocks.NewMockResource(ctrl)

	mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeConstruct)
	mockGame.EXPECT().GetPlayer("Player2").Return(mockPlayer2)
	mockGame.EXPECT().PlayerIndex("Player1").Return(0)
	mockGame.EXPECT().PlayerIndex("Player2").Return(1)
	mockGame.EXPECT().SameTeam(0, 1).Return(true)
	mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
	mockPlayer1.EXPECT().GetCardFromHand("G1").Return(mockResource, true)

	action := gameactions.NewConstructAction("Player1", "G1", "Player2")
	if err := action.Validate(mockGame); err != nil {
		t.Fatalf("validateConstructAllyCastle: unexpected error: %v", err)
	}
	return action, mockGame, mockPlayer1, mockPlayer2, mockResource
}

func TestConstructAction_PlayerName(t *testing.T) {
	action := gameactions.NewConstructAction("Player1", "G1", "")
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestConstructAction_NextPhase(t *testing.T) {
	action := gameactions.NewConstructAction("Player1", "G1", "")
	assert.Equal(t, types.PhaseTypeEndTurn, action.NextPhase())
}

func TestConstructAction_Validate(t *testing.T) {
	t.Run("Error when not in Construct phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack).Times(2)

		action := gameactions.NewConstructAction("Player1", "G1", "")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot construct in the")
	})

	t.Run("Success for own castle construction", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, _, _, _ := validateConstructOwnCastle(t, ctrl)
		assert.NotNil(t, action)
	})

	t.Run("Error when constructing on non-ally", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeConstruct)
		mockGame.EXPECT().GetPlayer("Player2").Return(mockPlayer2)
		mockGame.EXPECT().PlayerIndex("Player1").Return(0)
		mockGame.EXPECT().PlayerIndex("Player2").Return(1)
		mockGame.EXPECT().SameTeam(0, 1).Return(false)

		action := gameactions.NewConstructAction("Player1", "G1", "Player2")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can only construct on ally's castle")
	})

	t.Run("Error when card not in hand for ally construct", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeConstruct)
		mockGame.EXPECT().GetPlayer("Player2").Return(mockPlayer2)
		mockGame.EXPECT().PlayerIndex("Player1").Return(0)
		mockGame.EXPECT().PlayerIndex("Player2").Return(1)
		mockGame.EXPECT().SameTeam(0, 1).Return(true)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("G1").Return(nil, false)

		action := gameactions.NewConstructAction("Player1", "G1", "Player2")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "card not in hand")
	})

	t.Run("Success stores target player for ally construct", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, _, _, _, _ := validateConstructAllyCastle(t, ctrl)
		assert.NotNil(t, action)
	})
}

func TestConstructAction_Execute(t *testing.T) {
	t.Run("Error when castle Construct fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockResource := validateConstructOwnCastle(t, ctrl)
		mockCastle := mocks.NewMockCastle(ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().Castle().Return(mockCastle)
		mockCastle.EXPECT().Construct(mockResource).Return(errors.New("invalid card"))

		result, _, err := action.Execute(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "constructing castle failed")
		assert.NotNil(t, result)
	})

	t.Run("Success constructing own castle", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockResource := validateConstructOwnCastle(t, ctrl)
		mockCastle := mocks.NewMockCastle(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().Castle().Return(mockCastle)
		mockCastle.EXPECT().Construct(mockResource).Return(nil)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockResource.EXPECT().GetID().Return("G1")
		mockPlayer1.EXPECT().RemoveFromHand("G1").Return(nil, nil)
		mockGame.EXPECT().Status(mockPlayer1).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionConstruct, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Success constructing on ally castle in 2v2", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockPlayer2, mockResource := validateConstructAllyCastle(t, ctrl)
		mockCastle := mocks.NewMockCastle(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer2.EXPECT().Castle().Return(mockCastle)
		mockCastle.EXPECT().Construct(mockResource).Return(nil)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockPlayer2.EXPECT().Name().Return("Player2")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockResource.EXPECT().GetID().Return("G1")
		mockPlayer1.EXPECT().RemoveFromHand("G1").Return(nil, nil)
		mockGame.EXPECT().Status(mockPlayer1).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionConstruct, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Error when ally castle construct fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockPlayer2, mockResource := validateConstructAllyCastle(t, ctrl)
		mockCastle := mocks.NewMockCastle(ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer2.EXPECT().Castle().Return(mockCastle)
		mockCastle.EXPECT().Construct(mockResource).Return(errors.New("castle full"))

		result, _, err := action.Execute(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "constructing on ally castle failed")
		assert.NotNil(t, result)
	})

	t.Run("History updated on own castle construct", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockResource := validateConstructOwnCastle(t, ctrl)
		mockCastle := mocks.NewMockCastle(ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().Castle().Return(mockCastle)
		mockCastle.EXPECT().Construct(mockResource).Return(nil)
		mockPlayer1.EXPECT().Name().Return("Player1")

		var capturedMsg string
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Do(func(msg string, _ types.Category) {
			capturedMsg = msg
		})
		mockResource.EXPECT().GetID().Return("G1")
		mockPlayer1.EXPECT().RemoveFromHand("G1").Return(nil, nil)

		_, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		assert.Contains(t, capturedMsg, "Player1")
		assert.Contains(t, capturedMsg, "constructed")
	})

	t.Run("History updated on ally castle construct", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockPlayer2, mockResource := validateConstructAllyCastle(t, ctrl)
		mockCastle := mocks.NewMockCastle(ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer2.EXPECT().Castle().Return(mockCastle)
		mockCastle.EXPECT().Construct(mockResource).Return(nil)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockPlayer2.EXPECT().Name().Return("Player2")

		var capturedMsg string
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Do(func(msg string, _ types.Category) {
			capturedMsg = msg
		})
		mockResource.EXPECT().GetID().Return("G1")
		mockPlayer1.EXPECT().RemoveFromHand("G1").Return(nil, nil)

		_, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		assert.Contains(t, capturedMsg, "Player2's castle")
	})
}

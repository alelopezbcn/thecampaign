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

func TestEndTurnPhaseAction_PlayerName(t *testing.T) {
	action := gameactions.NewEndTurnPhaseAction("Player1", false)
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestEndTurnPhaseAction_Validate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGame := mocks.NewMockGame(ctrl)
	action := gameactions.NewEndTurnPhaseAction("Player1", false)
	err := action.Validate(mockGame)
	assert.NoError(t, err)
}

func TestEndTurnPhaseAction_NextPhase(t *testing.T) {
	action := gameactions.NewEndTurnPhaseAction("Player1", false)
	assert.Equal(t, types.PhaseTypeDrawCard, action.NextPhase())
}

func TestEndTurnPhaseAction_Execute(t *testing.T) {
	t.Run("Switches turn and returns next player status", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player2"}

		mockPlayer1.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1) // called in execute body
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().SwitchTurn()
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer2) // called in statusFn
		mockGame.EXPECT().Status(mockPlayer2).Return(expectedStatus)

		action := gameactions.NewEndTurnPhaseAction("Player1", false)
		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionEndTurn, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Turn wraps from last player to first", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockPlayer2.EXPECT().Name().Return("Player2")
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer2) // in execute body
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().SwitchTurn()
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1) // in statusFn
		mockGame.EXPECT().Status(mockPlayer1).Return(expectedStatus)

		action := gameactions.NewEndTurnPhaseAction("Player2", false)
		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionEndTurn, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("History records normal end turn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)

		var capturedMsg string
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Do(func(msg string, _ types.Category) {
			capturedMsg = msg
		})
		mockGame.EXPECT().SwitchTurn()
		// statusFn is not called in this test; no second CurrentPlayer/Status expectation

		action := gameactions.NewEndTurnPhaseAction("Player1", false)
		_, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		assert.Contains(t, capturedMsg, "Player1")
		assert.Contains(t, capturedMsg, "ended their turn")
	})

	t.Run("History records expired turn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)

		var capturedMsg string
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Do(func(msg string, _ types.Category) {
			capturedMsg = msg
		})
		mockGame.EXPECT().SwitchTurn()

		action := gameactions.NewEndTurnPhaseAction("Player1", true)
		_, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		assert.Contains(t, capturedMsg, "Player1")
		assert.Contains(t, capturedMsg, "turn expired")
	})
}

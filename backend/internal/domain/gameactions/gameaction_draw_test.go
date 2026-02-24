package gameactions_test

import (
	"errors"
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gameactions"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDrawCardAction_PlayerName(t *testing.T) {
	action := gameactions.NewDrawCardAction("Player1")
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestDrawCardAction_Validate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGame := mocks.NewMockGame(ctrl)
	action := gameactions.NewDrawCardAction("Player1")
	err := action.Validate(mockGame)
	assert.NoError(t, err)
}

func TestDrawCardAction_NextPhase(t *testing.T) {
	action := gameactions.NewDrawCardAction("Player1")
	assert.Equal(t, types.PhaseTypeAttack, action.NextPhase())
}

func TestDrawCardAction_Execute(t *testing.T) {
	t.Run("Error when drawing fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockGame.EXPECT().DrawCards(mockPlayer1, 1).Return(nil, errors.New("no cards left to draw"))

		action := gameactions.NewDrawCardAction("Player1")
		result, statusFn, err := action.Execute(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no cards left to draw")
		assert.Nil(t, result)
		assert.Nil(t, statusFn)
	})

	t.Run("Success drawing card", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockDrawnCard := mocks.NewMockCard(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockGame.EXPECT().DrawCards(mockPlayer1, 1).Return([]cards.Card{mockDrawnCard}, nil)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard).Return(true)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer1, mockDrawnCard).Return(expectedStatus)

		action := gameactions.NewDrawCardAction("Player1")
		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionDraw, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Hand limit exceeded returns result without error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockGame.EXPECT().DrawCards(mockPlayer1, 1).Return(nil, board.ErrHandLimitExceeded)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer1).Return(expectedStatus)

		action := gameactions.NewDrawCardAction("Player1")
		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionType(""), result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("History updated on successful draw", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockDrawnCard := mocks.NewMockCard(ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockGame.EXPECT().DrawCards(mockPlayer1, 1).Return([]cards.Card{mockDrawnCard}, nil)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard).Return(true)
		mockPlayer1.EXPECT().Name().Return("Player1")

		var capturedMsg string
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Do(func(msg string, _ types.Category) {
			capturedMsg = msg
		})

		action := gameactions.NewDrawCardAction("Player1")
		_, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		assert.Contains(t, capturedMsg, "drew")
		assert.Contains(t, capturedMsg, "Player1")
	})

	t.Run("History updated when hand limit exceeded", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockGame.EXPECT().DrawCards(mockPlayer1, 1).Return(nil, board.ErrHandLimitExceeded)
		mockPlayer1.EXPECT().Name().Return("Player1")

		var capturedMsg string
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Do(func(msg string, _ types.Category) {
			capturedMsg = msg
		})

		action := gameactions.NewDrawCardAction("Player1")
		_, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		assert.Contains(t, capturedMsg, "can't take more cards")
	})
}

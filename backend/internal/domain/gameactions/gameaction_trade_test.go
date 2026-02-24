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

func TestTradeAction_PlayerName(t *testing.T) {
	action := gameactions.NewTradeAction("Player1", []string{"C1", "C2", "C3"})
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestTradeAction_Validate(t *testing.T) {
	t.Run("Error when already traded this turn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockGame.EXPECT().TurnState().Return(types.TurnState{HasTraded: true})

		action := gameactions.NewTradeAction("Player1", []string{"C1", "C2", "C3"})
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already traded this turn")
	})

	t.Run("Error when not exactly 3 cards", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockGame.EXPECT().TurnState().Return(types.TurnState{})

		action := gameactions.NewTradeAction("Player1", []string{"C1", "C2"})
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must trade exactly 3 cards")
	})

	t.Run("Success with 3 cards", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockGame.EXPECT().TurnState().Return(types.TurnState{})

		action := gameactions.NewTradeAction("Player1", []string{"C1", "C2", "C3"})
		err := action.Validate(mockGame)

		assert.NoError(t, err)
	})
}

func TestTradeAction_Execute(t *testing.T) {
	t.Run("Error when RemoveFromHand fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().RemoveFromHand("C1", "C2", "C3").Return(nil, errors.New("card not found"))

		action := gameactions.NewTradeAction("Player1", []string{"C1", "C2", "C3"})
		result, _, err := action.Execute(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "giving cards for trading failed")
		assert.NotNil(t, result)
	})

	t.Run("Success trading 3 cards for 1", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockCard1 := mocks.NewMockCard(ctrl)
		mockCard2 := mocks.NewMockCard(ctrl)
		mockCard3 := mocks.NewMockCard(ctrl)
		mockDrawnCard := mocks.NewMockCard(ctrl)
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().RemoveFromHand("C1", "C2", "C3").Return(
			[]cards.Card{mockCard1, mockCard2, mockCard3}, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockCard1)
		mockGame.EXPECT().OnCardMovedToPile(mockCard2)
		mockGame.EXPECT().OnCardMovedToPile(mockCard3)
		mockGame.EXPECT().DrawCards(mockPlayer1, 1).Return([]cards.Card{mockDrawnCard}, nil)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().SetHasTraded(true)
		mockGame.EXPECT().SetCanTrade(false)
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeBuy)
		mockGame.EXPECT().Status(mockPlayer1, mockDrawnCard).Return(expectedStatus)

		action := gameactions.NewTradeAction("Player1", []string{"C1", "C2", "C3"})
		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionTrade, result.Action)
		assert.Equal(t, types.PhaseTypeBuy, action.NextPhase())
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("History updated on successful trade", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockCard1 := mocks.NewMockCard(ctrl)
		mockCard2 := mocks.NewMockCard(ctrl)
		mockCard3 := mocks.NewMockCard(ctrl)
		mockDrawnCard := mocks.NewMockCard(ctrl)

		var capturedMsg string
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().RemoveFromHand("C1", "C2", "C3").Return(
			[]cards.Card{mockCard1, mockCard2, mockCard3}, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockCard1)
		mockGame.EXPECT().OnCardMovedToPile(mockCard2)
		mockGame.EXPECT().OnCardMovedToPile(mockCard3)
		mockGame.EXPECT().DrawCards(mockPlayer1, 1).Return([]cards.Card{mockDrawnCard}, nil)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Do(func(msg string, _ types.Category) {
			capturedMsg = msg
		})
		mockGame.EXPECT().SetHasTraded(true)
		mockGame.EXPECT().SetCanTrade(false)
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeBuy)

		action := gameactions.NewTradeAction("Player1", []string{"C1", "C2", "C3"})
		_, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		assert.Contains(t, capturedMsg, "Player1")
		assert.Contains(t, capturedMsg, "traded")
	})
}

package domain

import (
	"errors"
	"strings"
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestTradeAction_PlayerName(t *testing.T) {
	action := NewTradeAction("Player1", []string{"C1", "C2", "C3"})
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestTradeAction_Validate(t *testing.T) {
	t.Run("Error when already traded this turn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:     []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn: 0,
			hasTraded:   true,
		}

		action := NewTradeAction("Player1", []string{"C1", "C2", "C3"})
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already traded this turn")
	})

	t.Run("Error when not exactly 3 cards", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:     []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn: 0,
		}

		action := NewTradeAction("Player1", []string{"C1", "C2"})
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must trade exactly 3 cards")
	})

	t.Run("Success with 3 cards", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:     []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn: 0,
		}

		action := NewTradeAction("Player1", []string{"C1", "C2", "C3"})
		err := action.Validate(g)

		assert.NoError(t, err)
	})
}

func TestTradeAction_Execute(t *testing.T) {
	t.Run("Error when GiveCards fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GiveCards("C1", "C2", "C3").Return(nil, errors.New("card not found"))

		g := &Game{
			Players:     []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn: 0,
		}

		action := NewTradeAction("Player1", []string{"C1", "C2", "C3"})
		result, _, err := action.Execute(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "giving cards for trading failed")
		assert.NotNil(t, result)
	})

	t.Run("Success trading 3 cards for 1", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)
		mockCard1 := mocks.NewMockCard(ctrl)
		mockCard2 := mocks.NewMockCard(ctrl)
		mockCard3 := mocks.NewMockCard(ctrl)
		mockDrawnCard := mocks.NewMockCard(ctrl)

		expectedStatus := GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GiveCards("C1", "C2", "C3").Return(
			[]ports.Card{mockCard1, mockCard2, mockCard3}, nil)

		// Traded cards go to discard pile
		mockDiscardPile.EXPECT().Discard(mockCard1)
		mockDiscardPile.EXPECT().Discard(mockCard2)
		mockDiscardPile.EXPECT().Discard(mockCard3)

		// Draw 1 card
		mockPlayer1.EXPECT().CanTakeCards(1).Return(true)
		mockDeck.EXPECT().DrawCard().Return(mockDrawnCard, true)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.PhaseTypeBuy,
			deck:               mockDeck,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
			history:            []historyLine{},
		}

		mockProvider.EXPECT().Get(mockPlayer1, g, mockDrawnCard).Return(expectedStatus)

		action := NewTradeAction("Player1", []string{"C1", "C2", "C3"})
		result, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionTrade, result.Action)
		assert.True(t, g.hasTraded)
		assert.False(t, g.CanTrade)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("NextPhase returns current game phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)
		mockCard1 := mocks.NewMockCard(ctrl)
		mockCard2 := mocks.NewMockCard(ctrl)
		mockCard3 := mocks.NewMockCard(ctrl)
		mockDrawnCard := mocks.NewMockCard(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GiveCards("C1", "C2", "C3").Return(
			[]ports.Card{mockCard1, mockCard2, mockCard3}, nil)
		mockDiscardPile.EXPECT().Discard(mockCard1)
		mockDiscardPile.EXPECT().Discard(mockCard2)
		mockDiscardPile.EXPECT().Discard(mockCard3)
		mockPlayer1.EXPECT().CanTakeCards(1).Return(true)
		mockDeck.EXPECT().DrawCard().Return(mockDrawnCard, true)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard)

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeBuy,
			deck:          mockDeck,
			discardPile:   mockDiscardPile,
			history:       []historyLine{},
		}

		action := NewTradeAction("Player1", []string{"C1", "C2", "C3"})
		_, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		assert.Equal(t, types.PhaseTypeBuy, action.NextPhase())
	})

	t.Run("History updated on successful trade", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)
		mockCard1 := mocks.NewMockCard(ctrl)
		mockCard2 := mocks.NewMockCard(ctrl)
		mockCard3 := mocks.NewMockCard(ctrl)
		mockDrawnCard := mocks.NewMockCard(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GiveCards("C1", "C2", "C3").Return(
			[]ports.Card{mockCard1, mockCard2, mockCard3}, nil)
		mockDiscardPile.EXPECT().Discard(mockCard1)
		mockDiscardPile.EXPECT().Discard(mockCard2)
		mockDiscardPile.EXPECT().Discard(mockCard3)
		mockPlayer1.EXPECT().CanTakeCards(1).Return(true)
		mockDeck.EXPECT().DrawCard().Return(mockDrawnCard, true)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard)

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeBuy,
			deck:          mockDeck,
			discardPile:   mockDiscardPile,
			history:       []historyLine{},
		}

		action := NewTradeAction("Player1", []string{"C1", "C2", "C3"})
		_, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		found := false
		for _, h := range g.history {
			if strings.Contains(h.Msg, "Player1") && strings.Contains(h.Msg, "traded") {
				found = true
				break
			}
		}
		assert.True(t, found, "History should contain trade action")
	})
}

package gameactions_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDrawCardAction_PlayerName(t *testing.T) {
	action := NewDrawCardAction("Player1")
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestDrawCardAction_Validate(t *testing.T) {
	action := NewDrawCardAction("Player1")
	err := action.Validate(&game{})
	assert.NoError(t, err)
}

func TestDrawCardAction_NextPhase(t *testing.T) {
	action := NewDrawCardAction("Player1")
	assert.Equal(t, types.PhaseTypeAttack, action.NextPhase())
}

func TestDrawCardAction_Execute(t *testing.T) {
	t.Run("Error when deck is empty and discard pile is also empty", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().CanTakeCards(1).Return(true)
		mockDeck.EXPECT().DrawCards(1, mockDiscardPile).Return(nil, errors.New("no cards left to draw"))

		g := &game{
			players:     []board.Player{mockPlayer1, mockPlayer2},
			currentTurn: 0,
			deck:        mockDeck,
			discardPile: mockDiscardPile,
		}

		action := NewDrawCardAction("Player1")
		result, statusFn, err := action.Execute(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no cards left to draw")
		assert.Nil(t, result)
		assert.Nil(t, statusFn)
	})

	t.Run("Success drawing card normally", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDrawnCard := mocks.NewMockCard(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().CanTakeCards(1).Return(true)
		mockDeck.EXPECT().DrawCards(1, mockDiscardPile).Return([]cards.Card{mockDrawnCard}, nil)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard).Return(true)

		g := &game{
			players:            []board.Player{mockPlayer1, mockPlayer2},
			currentTurn:        0,
			deck:               mockDeck,
			discardPile:        mockDiscardPile,
			gameStatusProvider: mockProvider,
		}

		mockProvider.EXPECT().Get(mockPlayer1, g, mockDrawnCard).Return(expectedStatus)

		action := NewDrawCardAction("Player1")
		result, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionDraw, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Hand limit exceeded - returns result without error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().CanTakeCards(1).Return(false)

		g := &game{
			players:            []board.Player{mockPlayer1, mockPlayer2},
			currentTurn:        0,
			deck:               mockDeck,
			discardPile:        mockDiscardPile,
			gameStatusProvider: mockProvider,
		}

		mockProvider.EXPECT().Get(mockPlayer1, g).Return(expectedStatus)

		action := NewDrawCardAction("Player1")
		result, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionType(""), result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Deck replenishes from discard pile", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDrawnCard := mocks.NewMockCard(ctrl)
		mockDiscardedCard := mocks.NewMockCard(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().CanTakeCards(1).Return(true)
		mockDeck.EXPECT().DrawCards(1, mockDiscardPile).Return(nil, errors.New("no cards left to draw"))
		mockDiscardPile.EXPECT().Empty().Return([]cards.Card{mockDiscardedCard})
		mockDeck.EXPECT().DrawCards(1, mockDiscardPile).Return([]cards.Card{mockDrawnCard}, nil)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard).Return(true)

		g := &game{
			players:            []board.Player{mockPlayer1, mockPlayer2},
			currentTurn:        0,
			deck:               mockDeck,
			discardPile:        mockDiscardPile,
			gameStatusProvider: mockProvider,
		}

		mockProvider.EXPECT().Get(mockPlayer1, g, mockDrawnCard).Return(expectedStatus)

		action := NewDrawCardAction("Player1")
		result, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionDraw, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("History is updated on successful draw", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockDrawnCard := mocks.NewMockCard(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().CanTakeCards(1).Return(true)
		mockDeck.EXPECT().DrawCards(1, mockDiscardPile).Return([]cards.Card{mockDrawnCard}, nil)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard).Return(true)

		g := &game{
			players:     []board.Player{mockPlayer1, mockPlayer2},
			currentTurn: 0,
			deck:        mockDeck,
			discardPile: mockDiscardPile,
			history:     []types.HistoryLine{},
		}

		action := NewDrawCardAction("Player1")
		_, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		found := false
		for _, h := range g.history {
			if strings.Contains(h.Msg, "drew") && strings.Contains(h.Msg, "Player1") {
				found = true
				break
			}
		}
		assert.True(t, found, "History should contain the draw action")
	})

	t.Run("History is updated when hand limit exceeded", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().CanTakeCards(1).Return(false)

		g := &game{
			players:     []board.Player{mockPlayer1, mockPlayer2},
			currentTurn: 0,
			deck:        mockDeck,
			discardPile: mockDiscardPile,
			history:     []types.HistoryLine{},
		}

		action := NewDrawCardAction("Player1")
		_, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		found := false
		for _, h := range g.history {
			if strings.Contains(h.Msg, "can't take more cards") {
				found = true
				break
			}
		}
		assert.True(t, found, "History should contain hand limit exceeded message")
	})
}

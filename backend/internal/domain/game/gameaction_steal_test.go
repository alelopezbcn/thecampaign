package game

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

func TestStealAction_PlayerName(t *testing.T) {
	action := NewStealAction("Player1", "Player2", 0)
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestStealAction_NextPhase(t *testing.T) {
	action := NewStealAction("Player1", "Player2", 0)
	assert.Equal(t, types.PhaseTypeBuy, action.NextPhase())
}

func TestStealAction_Validate(t *testing.T) {
	t.Run("Error when not in SpySteal phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeBuy,
		}

		action := NewStealAction("Player1", "Player2", 0)
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot steal in the")
	})

	t.Run("Error when player has no thief", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().HasThief().Return(false)

		g := &Game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeSpySteal,
		}

		action := NewStealAction("Player1", "Player2", 1)
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "player does not have a thief")
	})

	t.Run("Success stores target player", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer1.EXPECT().HasThief().Return(true)

		g := &Game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeSpySteal,
		}

		action := NewStealAction("Player1", "Player2", 2)
		err := action.Validate(g)

		assert.NoError(t, err)
		assert.Equal(t, mockPlayer2, action.targetPlayer)
	})
}

func TestStealAction_Execute(t *testing.T) {
	t.Run("Error when stealing fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().CardStolenFromHand(0).Return(nil, errors.New("invalid position"))

		g := &Game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeSpySteal,
		}

		action := NewStealAction("Player1", "Player2", 0)
		action.targetPlayer = mockPlayer2

		result, _, err := action.Execute(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stealing card failed")
		assert.NotNil(t, result)
	})

	t.Run("Error when Thief returns nil", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockStolenCard := mocks.NewMockCard(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().CardStolenFromHand(0).Return(mockStolenCard, nil)
		mockPlayer1.EXPECT().Thief().Return(nil)

		g := &Game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeSpySteal,
		}

		action := NewStealAction("Player1", "Player2", 0)
		action.targetPlayer = mockPlayer2

		result, _, err := action.Execute(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to retrieve thief card")
		assert.NotNil(t, result)
	})

	t.Run("Success returns result with stolen card info", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)
		mockThief := mocks.NewMockThief(ctrl)
		mockStolenCard := mocks.NewMockCard(ctrl)

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer2.EXPECT().CardStolenFromHand(2).Return(mockStolenCard, nil)
		mockPlayer1.EXPECT().Thief().Return(mockThief)
		mockDiscardPile.EXPECT().Discard(mockThief)
		mockPlayer1.EXPECT().TakeCards(mockStolenCard)

		g := &Game{
			players:            []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.PhaseTypeSpySteal,
			discardPile:        mockDiscardPile,
			gameStatusProvider: mockProvider,
			history:            []types.HistoryLine{},
		}

		mockProvider.EXPECT().GetWithModal(
			mockPlayer1, g, []cards.Card{mockStolenCard},
		).Return(expectedStatus)

		action := NewStealAction("Player1", "Player2", 2)
		action.targetPlayer = mockPlayer2

		result, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionSteal, result.Action)
		assert.Equal(t, "Player2", result.StolenFrom)
		assert.Equal(t, mockStolenCard, result.StolenCard)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("History updated on successful steal", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)
		mockThief := mocks.NewMockThief(ctrl)
		mockStolenCard := mocks.NewMockCard(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer2.EXPECT().CardStolenFromHand(1).Return(mockStolenCard, nil)
		mockPlayer1.EXPECT().Thief().Return(mockThief)
		mockDiscardPile.EXPECT().Discard(mockThief)
		mockPlayer1.EXPECT().TakeCards(mockStolenCard)

		g := &Game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeSpySteal,
			discardPile:   mockDiscardPile,
			history:       []types.HistoryLine{},
		}

		action := NewStealAction("Player1", "Player2", 1)
		action.targetPlayer = mockPlayer2

		_, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		found := false
		for _, h := range g.history {
			if strings.Contains(h.Msg, "Player1") && strings.Contains(h.Msg, "stole") {
				found = true
				break
			}
		}
		assert.True(t, found, "History should contain steal action")
	})
}

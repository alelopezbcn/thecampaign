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

func TestCatapultAction_PlayerName(t *testing.T) {
	action := NewCatapultAction("Player1", "Player2", 0)
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestCatapultAction_NextPhase(t *testing.T) {
	action := NewCatapultAction("Player1", "Player2", 0)
	assert.Equal(t, types.PhaseTypeSpySteal, action.NextPhase())
}

func TestCatapultAction_Validate(t *testing.T) {
	t.Run("Error when not in Attack phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeBuy,
		}

		action := NewCatapultAction("Player1", "Player2", 0)
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot use catapult in the")
	})

	t.Run("Error when player has no catapult", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().HasCatapult().Return(false)

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeAttack,
		}

		action := NewCatapultAction("Player1", "Player2", 0)
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "player does not have a catapult")
	})

	t.Run("Error when Catapult returns nil", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer1.EXPECT().HasCatapult().Return(true)
		mockPlayer1.EXPECT().Catapult().Return(nil)

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeAttack,
		}

		action := NewCatapultAction("Player1", "Player2", 0)
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "player does not have a catapult to attack")
	})

	t.Run("Success stores catapult and target player", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockCatapult := mocks.NewMockCatapult(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer1.EXPECT().HasCatapult().Return(true)
		mockPlayer1.EXPECT().Catapult().Return(mockCatapult)

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeAttack,
		}

		action := NewCatapultAction("Player1", "Player2", 2)
		err := action.Validate(g)

		assert.NoError(t, err)
		assert.Equal(t, mockCatapult, action.catapult)
		assert.Equal(t, mockPlayer2, action.targetPlayer)
	})
}

func TestCatapultAction_Execute(t *testing.T) {
	t.Run("Error when castle attack fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockCatapult := mocks.NewMockCatapult(ctrl)
		mockCastle := mocks.NewMockCastle(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer2.EXPECT().Castle().Return(mockCastle)
		mockCatapult.EXPECT().Attack(mockCastle, 0).Return(nil, errors.New("invalid position"))

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeAttack,
		}

		action := NewCatapultAction("Player1", "Player2", 0)
		action.catapult = mockCatapult
		action.targetPlayer = mockPlayer2

		result, _, err := action.Execute(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "attacking castle failed")
		assert.NotNil(t, result)
	})

	t.Run("Success returns result and discards stolen gold", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)
		mockCatapult := mocks.NewMockCatapult(ctrl)
		mockCastle := mocks.NewMockCastle(ctrl)
		mockStolenGold := mocks.NewMockResource(ctrl)

		expectedStatus := GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer2.EXPECT().Castle().Return(mockCastle)
		mockCatapult.EXPECT().Attack(mockCastle, 2).Return(mockStolenGold, nil)
		mockStolenGold.EXPECT().Value().Return(3)
		mockDiscardPile.EXPECT().Discard(mockStolenGold)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.PhaseTypeAttack,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
			history:            []types.HistoryLine{},
		}

		mockProvider.EXPECT().Get(mockPlayer1, g).Return(expectedStatus)

		action := NewCatapultAction("Player1", "Player2", 2)
		action.catapult = mockCatapult
		action.targetPlayer = mockPlayer2

		result, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionCatapult, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("History is updated on success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)
		mockCatapult := mocks.NewMockCatapult(ctrl)
		mockCastle := mocks.NewMockCastle(ctrl)
		mockStolenGold := mocks.NewMockResource(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer2.EXPECT().Castle().Return(mockCastle)
		mockCatapult.EXPECT().Attack(mockCastle, 1).Return(mockStolenGold, nil)
		mockStolenGold.EXPECT().Value().Return(5)
		mockDiscardPile.EXPECT().Discard(mockStolenGold)

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeAttack,
			discardPile:   mockDiscardPile,
			history:       []types.HistoryLine{},
		}

		action := NewCatapultAction("Player1", "Player2", 1)
		action.catapult = mockCatapult
		action.targetPlayer = mockPlayer2

		_, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		found := false
		for _, h := range g.history {
			if strings.Contains(h.Msg, "Player1") && strings.Contains(h.Msg, "gold") && strings.Contains(h.Msg, "Player2") {
				found = true
				break
			}
		}
		assert.True(t, found, "History should contain catapult action")
	})
}

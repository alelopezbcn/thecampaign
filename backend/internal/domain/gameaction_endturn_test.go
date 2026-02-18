package domain

import (
	"strings"
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestEndTurnPhaseAction_PlayerName(t *testing.T) {
	action := NewEndTurnPhaseAction("Player1", false)
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestEndTurnPhaseAction_Validate(t *testing.T) {
	action := NewEndTurnPhaseAction("Player1", false)
	err := action.Validate(&Game{})
	assert.NoError(t, err)
}

func TestEndTurnPhaseAction_NextPhase(t *testing.T) {
	action := NewEndTurnPhaseAction("Player1", false)
	assert.Equal(t, types.PhaseTypeDrawCard, action.NextPhase())
}

func TestEndTurnPhaseAction_Execute(t *testing.T) {
	t.Run("Switches to next player", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		expectedStatus := GameStatus{CurrentPlayer: "Player2"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.PhaseTypeEndTurn,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
			turnState: TurnState{HasMovedWarrior: true, HasTraded: true},
		}

		mockProvider.EXPECT().Get(mockPlayer2, g).Return(expectedStatus)

		action := NewEndTurnPhaseAction("Player1", false)
		result, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionEndTurn, result.Action)
		assert.Equal(t, 1, g.CurrentTurn)
		assert.False(t, g.turnState.HasMovedWarrior)
		assert.False(t, g.turnState.HasTraded)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Turn wraps around from last player to first", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		expectedStatus := GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        1,
			currentAction:      types.PhaseTypeEndTurn,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
		}

		mockProvider.EXPECT().Get(mockPlayer1, g).Return(expectedStatus)

		action := NewEndTurnPhaseAction("Player2", false)
		result, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionEndTurn, result.Action)
		assert.Equal(t, 0, g.CurrentTurn)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("History records normal end turn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		g := &Game{
			Players:     []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn: 0,
			discardPile: mockDiscardPile,
			history:     []historyLine{},
		}

		action := NewEndTurnPhaseAction("Player1", false)
		_, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		found := false
		for _, h := range g.history {
			if strings.Contains(h.Msg, "Player1") && strings.Contains(h.Msg, "ended their turn") {
				found = true
				break
			}
		}
		assert.True(t, found, "History should contain turn end message")
	})

	t.Run("History records expired turn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		g := &Game{
			Players:     []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn: 0,
			discardPile: mockDiscardPile,
			history:     []historyLine{},
		}

		action := NewEndTurnPhaseAction("Player1", true)
		_, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		found := false
		for _, h := range g.history {
			if strings.Contains(h.Msg, "Player1") && strings.Contains(h.Msg, "turn expired") {
				found = true
				break
			}
		}
		assert.True(t, found, "History should contain turn expired message")
	})
}

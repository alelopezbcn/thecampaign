package game

import (
	"errors"
	"strings"
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestMoveWarriorAction_PlayerName(t *testing.T) {
	action := NewMoveWarriorAction("Player1", "K1", "")
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestMoveWarriorAction_Validate(t *testing.T) {
	t.Run("Error when already moved warrior this turn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:     []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn: 0,
			turnState:   TurnState{HasMovedWarrior: true},
		}

		action := NewMoveWarriorAction("Player1", "K1", "")
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already moved a warrior this turn")
	})

	t.Run("Success for own field move", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:     []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn: 0,
		}

		action := NewMoveWarriorAction("Player1", "K1", "")
		err := action.Validate(g)

		assert.NoError(t, err)
		assert.Nil(t, action.targetPlayer)
	})

	t.Run("Error when target player not found for ally move", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		g := &Game{
			Players:     []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn: 0,
		}

		action := NewMoveWarriorAction("Player1", "K1", "Unknown")
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target player Unknown not found")
	})

	t.Run("Error when target is not ally", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		g := &Game{
			Players:     []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn: 0,
			Mode:        types.GameMode1v1,
		}

		action := NewMoveWarriorAction("Player1", "K1", "Player2")
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can only move warriors to ally's field")
	})

	t.Run("Error when card not in hand for ally move", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer4 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromHand("K1").Return(nil, false)

		g := &Game{
			Players:     []ports.Player{mockPlayer1, mockPlayer2, mockPlayer3, mockPlayer4},
			CurrentTurn: 0,
			Mode:        types.GameMode2v2,
			Teams:       map[int][]int{0: {0, 1}, 1: {2, 3}},
		}

		action := NewMoveWarriorAction("Player1", "K1", "Player2")
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "card with ID K1 not found in hand")
	})

	t.Run("Error when card is not a warrior for ally move", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer4 := mocks.NewMockPlayer(ctrl)
		mockCard := mocks.NewMockCard(ctrl) // Not a Warrior

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromHand("G1").Return(mockCard, true)

		g := &Game{
			Players:     []ports.Player{mockPlayer1, mockPlayer2, mockPlayer3, mockPlayer4},
			CurrentTurn: 0,
			Mode:        types.GameMode2v2,
			Teams:       map[int][]int{0: {0, 1}, 1: {2, 3}},
		}

		action := NewMoveWarriorAction("Player1", "G1", "Player2")
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "only warrior cards can be moved to field")
	})

	t.Run("Success stores target player and warrior for ally move", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer4 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromHand("K1").Return(mockWarrior, true)

		g := &Game{
			Players:     []ports.Player{mockPlayer1, mockPlayer2, mockPlayer3, mockPlayer4},
			CurrentTurn: 0,
			Mode:        types.GameMode2v2,
			Teams:       map[int][]int{0: {0, 1}, 1: {2, 3}},
		}

		action := NewMoveWarriorAction("Player1", "K1", "Player2")
		err := action.Validate(g)

		assert.NoError(t, err)
		assert.Equal(t, mockPlayer2, action.targetPlayer)
		assert.Equal(t, mockWarrior, action.warrior)
	})
}

func TestMoveWarriorAction_Execute(t *testing.T) {
	t.Run("Error when MoveCardToField fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().MoveCardToField("K1").Return(errors.New("card not found"))

		g := &Game{
			Players:     []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn: 0,
			history:     []types.HistoryLine{},
		}

		action := NewMoveWarriorAction("Player1", "K1", "")
		result, _, err := action.Execute(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "moving warrior to field failed")
		assert.NotNil(t, result)
	})

	t.Run("Success moving warrior to own field", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().MoveCardToField("K1").Return(nil)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.PhaseTypeAttack,
			gameStatusProvider: mockProvider,
			history:            []types.HistoryLine{},
		}

		mockProvider.EXPECT().Get(mockPlayer1, g).Return(expectedStatus)

		action := NewMoveWarriorAction("Player1", "K1", "")
		result, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionMoveWarrior, result.Action)
		assert.Equal(t, "K1", result.MovedWarriorID)
		assert.True(t, g.turnState.HasMovedWarrior)
		assert.False(t, g.turnState.CanMoveWarrior)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Success moving warrior to ally field in 2v2", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer4 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockField := mocks.NewMockField(ctrl)
		mockHand := mocks.NewMockHand(ctrl)

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer2.EXPECT().Field().Return(mockField)
		mockField.EXPECT().AddWarriors(mockWarrior)
		mockPlayer1.EXPECT().Hand().Return(mockHand)
		mockHand.EXPECT().RemoveCard(mockWarrior).Return(true)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2, mockPlayer3, mockPlayer4},
			CurrentTurn:        0,
			currentAction:      types.PhaseTypeAttack,
			Mode:               types.GameMode2v2,
			Teams:              map[int][]int{0: {0, 1}, 1: {2, 3}},
			gameStatusProvider: mockProvider,
			history:            []types.HistoryLine{},
		}

		mockProvider.EXPECT().Get(mockPlayer1, g).Return(expectedStatus)

		action := NewMoveWarriorAction("Player1", "K1", "Player2")
		action.targetPlayer = mockPlayer2
		action.warrior = mockWarrior

		result, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionMoveWarrior, result.Action)
		assert.Equal(t, "K1", result.MovedWarriorID)
		assert.True(t, g.turnState.HasMovedWarrior)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("NextPhase returns current game phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().MoveCardToField("K1").Return(nil)

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeBuy,
			history:       []types.HistoryLine{},
		}

		action := NewMoveWarriorAction("Player1", "K1", "")
		_, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		assert.Equal(t, types.PhaseTypeBuy, action.NextPhase())
	})

	t.Run("History updated on own field move", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().MoveCardToField("K1").Return(nil)

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeAttack,
			history:       []types.HistoryLine{},
		}

		action := NewMoveWarriorAction("Player1", "K1", "")
		_, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		found := false
		for _, h := range g.history {
			if strings.Contains(h.Msg, "Player1") && strings.Contains(h.Msg, "moved warrior") {
				found = true
				break
			}
		}
		assert.True(t, found, "History should contain move warrior action")
	})

	t.Run("History updated on ally field move", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer4 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockField := mocks.NewMockField(ctrl)
		mockHand := mocks.NewMockHand(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer2.EXPECT().Field().Return(mockField)
		mockField.EXPECT().AddWarriors(mockWarrior)
		mockPlayer1.EXPECT().Hand().Return(mockHand)
		mockHand.EXPECT().RemoveCard(mockWarrior).Return(true)

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2, mockPlayer3, mockPlayer4},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeAttack,
			history:       []types.HistoryLine{},
		}

		action := NewMoveWarriorAction("Player1", "K1", "Player2")
		action.targetPlayer = mockPlayer2
		action.warrior = mockWarrior

		_, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		found := false
		for _, h := range g.history {
			if strings.Contains(h.Msg, "Player1") && strings.Contains(h.Msg, "Player2's field") {
				found = true
				break
			}
		}
		assert.True(t, found, "History should contain ally field move action")
	})
}

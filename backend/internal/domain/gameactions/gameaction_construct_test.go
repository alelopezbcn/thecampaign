package gameactions

import (
	"errors"
	"strings"
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestConstructAction_PlayerName(t *testing.T) {
	action := NewConstructAction("Player1", "G1", "")
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestConstructAction_NextPhase(t *testing.T) {
	action := NewConstructAction("Player1", "G1", "")
	assert.Equal(t, types.PhaseTypeEndTurn, action.NextPhase())
}

func TestConstructAction_Validate(t *testing.T) {
	t.Run("Error when not in Construct phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			currentTurn:   0,
			currentAction: types.PhaseTypeAttack,
		}

		action := NewConstructAction("Player1", "G1", "")
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot construct in the")
	})

	t.Run("Success for own castle construction", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			currentTurn:   0,
			currentAction: types.PhaseTypeConstruct,
		}

		action := NewConstructAction("Player1", "G1", "")
		err := action.Validate(g)

		assert.NoError(t, err)
		assert.Nil(t, action.targetPlayer)
	})

	t.Run("Error when constructing on non-ally in 2v2", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		g := &game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			currentTurn:   0,
			currentAction: types.PhaseTypeConstruct,
			mode:          types.GameMode1v1,
		}

		action := NewConstructAction("Player1", "G1", "Player2")
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can only construct on ally's castle")
	})

	t.Run("Error when card not in hand for ally construct", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer4 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromHand("G1").Return(nil, false)

		g := &game{
			players:       []board.Player{mockPlayer1, mockPlayer2, mockPlayer3, mockPlayer4},
			currentTurn:   0,
			currentAction: types.PhaseTypeConstruct,
			mode:          types.GameMode2v2,
			teams:         map[int][]int{0: {0, 1}, 1: {2, 3}},
		}

		action := NewConstructAction("Player1", "G1", "Player2")
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "card not in hand")
	})

	t.Run("Success stores target player for ally construct", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer4 := mocks.NewMockPlayer(ctrl)
		mockResource := mocks.NewMockResource(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromHand("G1").Return(mockResource, true)

		g := &game{
			players:       []board.Player{mockPlayer1, mockPlayer2, mockPlayer3, mockPlayer4},
			currentTurn:   0,
			currentAction: types.PhaseTypeConstruct,
			mode:          types.GameMode2v2,
			teams:         map[int][]int{0: {0, 1}, 1: {2, 3}},
		}

		action := NewConstructAction("Player1", "G1", "Player2")
		err := action.Validate(g)

		assert.NoError(t, err)
		assert.Equal(t, mockPlayer2, action.targetPlayer)
	})
}

func TestConstructAction_Execute(t *testing.T) {
	t.Run("Error when player Construct fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().Construct("G1").Return(errors.New("invalid card"))

		g := &game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			currentTurn:   0,
			currentAction: types.PhaseTypeConstruct,
			history:       []types.HistoryLine{},
		}

		action := NewConstructAction("Player1", "G1", "")
		result, _, err := action.Execute(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "constructing card failed")
		assert.NotNil(t, result)
	})

	t.Run("Success constructing own castle", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().Construct("G1").Return(nil)

		g := &game{
			players:            []board.Player{mockPlayer1, mockPlayer2},
			currentTurn:        0,
			currentAction:      types.PhaseTypeConstruct,
			gameStatusProvider: mockProvider,
			history:            []types.HistoryLine{},
		}

		mockProvider.EXPECT().Get(mockPlayer1, g).Return(expectedStatus)

		action := NewConstructAction("Player1", "G1", "")
		result, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionConstruct, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Success constructing on ally castle in 2v2", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer4 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockResource := mocks.NewMockResource(ctrl)
		mockCastle := mocks.NewMockCastle(ctrl)
		mockHand := mocks.NewMockHand(ctrl)

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromHand("G1").Return(mockResource, true)
		mockPlayer2.EXPECT().Castle().Return(mockCastle).AnyTimes()
		mockCastle.EXPECT().IsConstructed().Return(true).AnyTimes()
		mockCastle.EXPECT().Construct(mockResource).Return(nil)
		mockPlayer1.EXPECT().Hand().Return(mockHand)
		mockHand.EXPECT().RemoveCard(mockResource).Return(true)

		g := &game{
			players:            []board.Player{mockPlayer1, mockPlayer2, mockPlayer3, mockPlayer4},
			currentTurn:        0,
			currentAction:      types.PhaseTypeConstruct,
			mode:               types.GameMode2v2,
			teams:              map[int][]int{0: {0, 1}, 1: {2, 3}},
			gameStatusProvider: mockProvider,
			history:            []types.HistoryLine{},
		}

		mockProvider.EXPECT().Get(mockPlayer1, g).Return(expectedStatus)

		action := NewConstructAction("Player1", "G1", "Player2")
		action.targetPlayer = mockPlayer2

		result, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionConstruct, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Error when ally castle construct fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer4 := mocks.NewMockPlayer(ctrl)
		mockResource := mocks.NewMockResource(ctrl)
		mockCastle := mocks.NewMockCastle(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromHand("G1").Return(mockResource, true)
		mockPlayer2.EXPECT().Castle().Return(mockCastle).AnyTimes()
		mockCastle.EXPECT().IsConstructed().Return(true).AnyTimes()
		mockCastle.EXPECT().Construct(mockResource).Return(errors.New("castle full"))

		g := &game{
			players:       []board.Player{mockPlayer1, mockPlayer2, mockPlayer3, mockPlayer4},
			currentTurn:   0,
			currentAction: types.PhaseTypeConstruct,
			mode:          types.GameMode2v2,
			teams:         map[int][]int{0: {0, 1}, 1: {2, 3}},
			history:       []types.HistoryLine{},
		}

		action := NewConstructAction("Player1", "G1", "Player2")
		action.targetPlayer = mockPlayer2

		result, _, err := action.Execute(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "constructing on ally castle failed")
		assert.NotNil(t, result)
	})

	t.Run("History updated on own castle construct", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().Construct("G1").Return(nil)

		g := &game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			currentTurn:   0,
			currentAction: types.PhaseTypeConstruct,
			history:       []types.HistoryLine{},
		}

		action := NewConstructAction("Player1", "G1", "")
		_, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		found := false
		for _, h := range g.history {
			if strings.Contains(h.Msg, "Player1") && strings.Contains(h.Msg, "constructed") {
				found = true
				break
			}
		}
		assert.True(t, found, "History should contain construct action")
	})

	t.Run("History updated on ally castle construct", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer4 := mocks.NewMockPlayer(ctrl)
		mockResource := mocks.NewMockResource(ctrl)
		mockCastle := mocks.NewMockCastle(ctrl)
		mockHand := mocks.NewMockHand(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromHand("G1").Return(mockResource, true)
		mockPlayer2.EXPECT().Castle().Return(mockCastle).AnyTimes()
		mockCastle.EXPECT().IsConstructed().Return(true).AnyTimes()
		mockCastle.EXPECT().Construct(mockResource).Return(nil)
		mockPlayer1.EXPECT().Hand().Return(mockHand)
		mockHand.EXPECT().RemoveCard(mockResource).Return(true)

		g := &game{
			players:       []board.Player{mockPlayer1, mockPlayer2, mockPlayer3, mockPlayer4},
			currentTurn:   0,
			currentAction: types.PhaseTypeConstruct,
			mode:          types.GameMode2v2,
			teams:         map[int][]int{0: {0, 1}, 1: {2, 3}},
			history:       []types.HistoryLine{},
		}

		action := NewConstructAction("Player1", "G1", "Player2")
		action.targetPlayer = mockPlayer2

		_, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		found := false
		for _, h := range g.history {
			if strings.Contains(h.Msg, "Player2's castle") {
				found = true
				break
			}
		}
		assert.True(t, found, "History should mention ally's castle")
	})
}

package game

import (
	"strings"
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSpyAction_PlayerName(t *testing.T) {
	action := NewSpyAction("Player1", "Player2", 1)
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestSpyAction_NextPhase(t *testing.T) {
	action := NewSpyAction("Player1", "Player2", 1)
	assert.Equal(t, types.PhaseTypeBuy, action.NextPhase())
}

func TestSpyAction_Validate(t *testing.T) {
	t.Run("Error when not in SpySteal phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeAttack,
		}

		action := NewSpyAction("Player1", "Player2", 1)
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot use spy in the")
	})

	t.Run("Error when player has no spy", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().HasSpy().Return(false)

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeSpySteal,
		}

		action := NewSpyAction("Player1", "Player2", 1)
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "player does not have a spy")
	})

	t.Run("Success when player has spy", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().HasSpy().Return(true)

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeSpySteal,
		}

		action := NewSpyAction("Player1", "Player2", 1)
		err := action.Validate(g)

		assert.NoError(t, err)
	})
}

func TestSpyAction_Execute(t *testing.T) {
	t.Run("Error when invalid spy option", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeSpySteal,
		}

		action := NewSpyAction("Player1", "Player2", 3)
		result, _, err := action.Execute(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid Spy option")
		assert.NotNil(t, result)
	})

	t.Run("Error when Spy returns nil", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockDeck.EXPECT().Reveal(5).Return([]ports.Card{})
		mockPlayer1.EXPECT().Spy().Return(nil)

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeSpySteal,
			deck:          mockDeck,
			history:       []types.HistoryLine{},
		}

		action := NewSpyAction("Player1", "Player2", 1)
		result, _, err := action.Execute(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to retrieve spy card")
		assert.NotNil(t, result)
	})

	t.Run("Success spying top 5 cards from deck (option 1)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)
		mockSpy := mocks.NewMockSpy(ctrl)
		mockRevealedCard := mocks.NewMockCard(ctrl)

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}
		revealedCards := []ports.Card{mockRevealedCard}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().Spy().Return(mockSpy)
		mockDiscardPile.EXPECT().Discard(mockSpy)
		mockDeck.EXPECT().Reveal(5).Return(revealedCards)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.PhaseTypeSpySteal,
			deck:               mockDeck,
			discardPile:        mockDiscardPile,
			gameStatusProvider: mockProvider,
			history:            []types.HistoryLine{},
		}

		mockProvider.EXPECT().GetWithModal(mockPlayer1, g, revealedCards).Return(expectedStatus)

		action := NewSpyAction("Player1", "Player2", 1)
		result, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionSpy, result.Action)
		assert.Equal(t, types.SpyTargetDeck, result.Spy.Target)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Success spying enemy hand (option 2)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)
		mockSpy := mocks.NewMockSpy(ctrl)
		mockEnemyHand := mocks.NewMockHand(ctrl)
		mockEnemyCard := mocks.NewMockCard(ctrl)

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}
		enemyCards := []ports.Card{mockEnemyCard}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer1.EXPECT().Spy().Return(mockSpy)
		mockDiscardPile.EXPECT().Discard(mockSpy)
		mockPlayer2.EXPECT().Hand().Return(mockEnemyHand)
		mockEnemyHand.EXPECT().ShowCards().Return(enemyCards)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.PhaseTypeSpySteal,
			discardPile:        mockDiscardPile,
			gameStatusProvider: mockProvider,
			history:            []types.HistoryLine{},
		}

		mockProvider.EXPECT().GetWithModal(mockPlayer1, g, enemyCards).Return(expectedStatus)

		action := NewSpyAction("Player1", "Player2", 2)
		result, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionSpy, result.Action)
		assert.Equal(t, types.SpyTargetPlayer, result.Spy.Target)
		assert.Equal(t, "Player2", result.Spy.TargetPlayer)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("History updated on deck spy", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)
		mockSpy := mocks.NewMockSpy(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().Spy().Return(mockSpy)
		mockDiscardPile.EXPECT().Discard(mockSpy)
		mockDeck.EXPECT().Reveal(5).Return([]ports.Card{})

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeSpySteal,
			deck:          mockDeck,
			discardPile:   mockDiscardPile,
			history:       []types.HistoryLine{},
		}

		action := NewSpyAction("Player1", "Player2", 1)
		_, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		found := false
		for _, h := range g.history {
			if strings.Contains(h.Msg, "Player1") && strings.Contains(h.Msg, "spied top 5") {
				found = true
				break
			}
		}
		assert.True(t, found, "History should contain deck spy action")
	})

	t.Run("History updated on player spy", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)
		mockSpy := mocks.NewMockSpy(ctrl)
		mockEnemyHand := mocks.NewMockHand(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer1.EXPECT().Spy().Return(mockSpy)
		mockDiscardPile.EXPECT().Discard(mockSpy)
		mockPlayer2.EXPECT().Hand().Return(mockEnemyHand)
		mockEnemyHand.EXPECT().ShowCards().Return([]ports.Card{})

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeSpySteal,
			discardPile:   mockDiscardPile,
			history:       []types.HistoryLine{},
		}

		action := NewSpyAction("Player1", "Player2", 2)
		_, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		found := false
		for _, h := range g.history {
			if strings.Contains(h.Msg, "Player1") && strings.Contains(h.Msg, "spied on") && strings.Contains(h.Msg, "Player2") {
				found = true
				break
			}
		}
		assert.True(t, found, "History should contain player spy action")
	})
}

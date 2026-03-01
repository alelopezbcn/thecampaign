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

// validateSpyActionSetup calls Validate on a new spy action and returns the action and mocks.
// mockSpy is in player1's hand (looked up by "spy-id"), so a.spy is populated after Validate.
func validateSpyActionSetup(t *testing.T, ctrl *gomock.Controller, option int) (gameactions.GameAction, *mocks.MockGame, *mocks.MockPlayer, *mocks.MockSpy) {
	t.Helper()
	mockGame := mocks.NewMockGame(ctrl)
	mockPlayer1 := mocks.NewMockPlayer(ctrl)
	mockSpy := mocks.NewMockSpy(ctrl)

	mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeSpySteal)
	mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
	mockPlayer1.EXPECT().GetCardFromHand("spy-id").Return(mockSpy, true)

	action := gameactions.NewSpyAction("Player1", "Player2", option, "spy-id")
	if err := action.Validate(mockGame); err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	return action, mockGame, mockPlayer1, mockSpy
}

func TestSpyAction_PlayerName(t *testing.T) {
	action := gameactions.NewSpyAction("Player1", "Player2", 1, "spy-id")
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestSpyAction_NextPhase(t *testing.T) {
	action := gameactions.NewSpyAction("Player1", "Player2", 1, "spy-id")
	assert.Equal(t, types.PhaseTypeBuy, action.NextPhase())
}

func TestSpyAction_Validate(t *testing.T) {
	t.Run("Error when not in SpySteal phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack).Times(2)

		action := gameactions.NewSpyAction("Player1", "Player2", 1, "spy-id")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot use spy in the")
	})

	t.Run("Error when card not found in hand", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeSpySteal)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("spy-id").Return(nil, false)

		action := gameactions.NewSpyAction("Player1", "Player2", 1, "spy-id")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found in hand")
	})

	t.Run("Error when card found but wrong type", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockCard := mocks.NewMockCard(ctrl) // not a Spy

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeSpySteal)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("spy-id").Return(mockCard, true)

		action := gameactions.NewSpyAction("Player1", "Player2", 1, "spy-id")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is not a spy card")
	})

	t.Run("Success when player has spy", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockSpy := mocks.NewMockSpy(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeSpySteal)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("spy-id").Return(mockSpy, true)

		action := gameactions.NewSpyAction("Player1", "Player2", 1, "spy-id")
		err := action.Validate(mockGame)

		assert.NoError(t, err)
	})
}

func TestSpyAction_Execute(t *testing.T) {
	t.Run("Error when invalid spy option", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)

		action := gameactions.NewSpyAction("Player1", "Player2", 3, "spy-id")
		result, _, err := action.Execute(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid Spy option")
		assert.NotNil(t, result)
	})

	t.Run("Error when RemoveFromHand fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockSpy := validateSpyActionSetup(t, ctrl, 1)

		mockBoard := mocks.NewMockBoard(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Board().Return(mockBoard)
		mockBoard.EXPECT().Deck().Return(mockDeck)
		mockDeck.EXPECT().Reveal(5).Return([]cards.Card{})
		mockSpy.EXPECT().GetID().Return("spy-id")
		mockPlayer1.EXPECT().RemoveFromHand("spy-id").Return(nil, errors.New("not found"))

		result, _, err := action.Execute(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "removing spy from hand failed")
		assert.NotNil(t, result)
	})

	t.Run("Success spying top 5 cards from deck (option 1)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockSpy := validateSpyActionSetup(t, ctrl, 1)

		mockBoard := mocks.NewMockBoard(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockRevealedCard := mocks.NewMockCard(ctrl)
		revealedCards := []cards.Card{mockRevealedCard}
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Board().Return(mockBoard)
		mockBoard.EXPECT().Deck().Return(mockDeck)
		mockDeck.EXPECT().Reveal(5).Return(revealedCards)
		mockSpy.EXPECT().GetID().Return("spy-id")
		mockPlayer1.EXPECT().RemoveFromHand("spy-id").Return([]cards.Card{mockSpy}, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockSpy)
		mockGame.EXPECT().StatusWithModal(mockPlayer1, revealedCards).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionSpy, result.Action)
		assert.Equal(t, types.SpyTargetDeck, result.Spy.Target)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Success spying enemy hand (option 2)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockSpy := validateSpyActionSetup(t, ctrl, 2)

		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockEnemyHand := mocks.NewMockHand(ctrl)
		mockEnemyCard := mocks.NewMockCard(ctrl)
		enemyCards := []cards.Card{mockEnemyCard}
		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockGame.EXPECT().GetTargetPlayer("Player1", "Player2").Return(mockPlayer2, nil)
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockPlayer2.EXPECT().Hand().Return(mockEnemyHand)
		mockEnemyHand.EXPECT().ShowCards().Return(enemyCards)
		mockSpy.EXPECT().GetID().Return("spy-id")
		mockPlayer1.EXPECT().RemoveFromHand("spy-id").Return([]cards.Card{mockSpy}, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockSpy)
		mockGame.EXPECT().StatusWithModal(mockPlayer1, enemyCards).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

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

		action, mockGame, mockPlayer1, mockSpy := validateSpyActionSetup(t, ctrl, 1)

		mockBoard := mocks.NewMockBoard(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)

		var capturedMsg string
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Do(func(msg string, _ types.Category) {
			capturedMsg = msg
		})
		mockGame.EXPECT().Board().Return(mockBoard)
		mockBoard.EXPECT().Deck().Return(mockDeck)
		mockDeck.EXPECT().Reveal(5).Return([]cards.Card{})
		mockSpy.EXPECT().GetID().Return("spy-id")
		mockPlayer1.EXPECT().RemoveFromHand("spy-id").Return([]cards.Card{mockSpy}, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockSpy)

		_, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		assert.Contains(t, capturedMsg, "Player1")
		assert.Contains(t, capturedMsg, "spied top 5")
	})

	t.Run("History updated on player spy", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockSpy := validateSpyActionSetup(t, ctrl, 2)

		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockEnemyHand := mocks.NewMockHand(ctrl)

		var capturedMsg string
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockGame.EXPECT().GetTargetPlayer("Player1", "Player2").Return(mockPlayer2, nil)
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Do(func(msg string, _ types.Category) {
			capturedMsg = msg
		})
		mockPlayer2.EXPECT().Hand().Return(mockEnemyHand)
		mockEnemyHand.EXPECT().ShowCards().Return([]cards.Card{})
		mockSpy.EXPECT().GetID().Return("spy-id")
		mockPlayer1.EXPECT().RemoveFromHand("spy-id").Return([]cards.Card{mockSpy}, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockSpy)

		_, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		assert.Contains(t, capturedMsg, "Player1")
		assert.Contains(t, capturedMsg, "spied on")
		assert.Contains(t, capturedMsg, "Player2")
	})
}

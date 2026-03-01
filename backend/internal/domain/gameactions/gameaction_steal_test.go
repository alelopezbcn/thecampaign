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

// validateStealAction calls Validate on a new steal action and returns key mocks.
// player1 has mockThief (looked up by "thief-id"); player2 has mockStolenCard (position 1).
func validateStealAction(t *testing.T, ctrl *gomock.Controller, cardPosition int) (
	gameactions.GameAction, *mocks.MockGame, *mocks.MockPlayer, *mocks.MockPlayer, *mocks.MockHand, *mocks.MockCard, *mocks.MockThief,
) {
	t.Helper()
	mockGame := mocks.NewMockGame(ctrl)
	mockPlayer1 := mocks.NewMockPlayer(ctrl)
	mockPlayer2 := mocks.NewMockPlayer(ctrl)
	mockHand2 := mocks.NewMockHand(ctrl)
	mockStolenCard := mocks.NewMockCard(ctrl)
	mockThief := mocks.NewMockThief(ctrl)

	mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeSpySteal)
	mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
	mockPlayer1.EXPECT().GetCardFromHand("thief-id").Return(mockThief, true)
	mockGame.EXPECT().GetTargetPlayer("Player1", "Player2").Return(mockPlayer2, nil)
	mockPlayer2.EXPECT().Hand().Return(mockHand2)
	mockHand2.EXPECT().ShowCards().Return([]cards.Card{mockStolenCard})

	action := gameactions.NewStealAction("Player1", "Player2", cardPosition, "thief-id")
	if err := action.Validate(mockGame); err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	return action, mockGame, mockPlayer1, mockPlayer2, mockHand2, mockStolenCard, mockThief
}

func TestStealAction_PlayerName(t *testing.T) {
	action := gameactions.NewStealAction("Player1", "Player2", 0, "thief-id")
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestStealAction_NextPhase(t *testing.T) {
	action := gameactions.NewStealAction("Player1", "Player2", 0, "thief-id")
	assert.Equal(t, types.PhaseTypeBuy, action.NextPhase())
}

func TestStealAction_Validate(t *testing.T) {
	t.Run("Error when not in SpySteal phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeBuy).Times(2)

		action := gameactions.NewStealAction("Player1", "Player2", 0, "thief-id")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot steal in the")
	})

	t.Run("Error when card not found in hand", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeSpySteal)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("thief-id").Return(nil, false)

		action := gameactions.NewStealAction("Player1", "Player2", 1, "thief-id")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found in hand")
	})

	t.Run("Error when card found but wrong type", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockCard := mocks.NewMockCard(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeSpySteal)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("thief-id").Return(mockCard, true)

		action := gameactions.NewStealAction("Player1", "Player2", 1, "thief-id")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is not a thief card")
	})

	t.Run("Error when card position is invalid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockHand2 := mocks.NewMockHand(ctrl)
		mockThief := mocks.NewMockThief(ctrl)
		mockCard := mocks.NewMockCard(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeSpySteal)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("thief-id").Return(mockThief, true)
		mockGame.EXPECT().GetTargetPlayer("Player1", "Player2").Return(mockPlayer2, nil)
		mockPlayer2.EXPECT().Hand().Return(mockHand2)
		mockHand2.EXPECT().ShowCards().Return([]cards.Card{mockCard}) // only 1 card

		action := gameactions.NewStealAction("Player1", "Player2", 5, "thief-id") // position 5 is invalid
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid position")
	})

	t.Run("Success with valid target and position", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockHand2 := mocks.NewMockHand(ctrl)
		mockThief := mocks.NewMockThief(ctrl)
		mockCard := mocks.NewMockCard(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeSpySteal)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("thief-id").Return(mockThief, true)
		mockGame.EXPECT().GetTargetPlayer("Player1", "Player2").Return(mockPlayer2, nil)
		mockPlayer2.EXPECT().Hand().Return(mockHand2)
		mockHand2.EXPECT().ShowCards().Return([]cards.Card{mockCard})

		action := gameactions.NewStealAction("Player1", "Player2", 1, "thief-id")
		err := action.Validate(mockGame)

		assert.NoError(t, err)
	})
}

func TestStealAction_Execute(t *testing.T) {
	t.Run("Error when RemoveFromHand (thief) fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockPlayer2, mockHand2, mockStolenCard, mockThief :=
			validateStealAction(t, ctrl, 1)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		// steal() calls Hand().RemoveCard on player2
		mockPlayer2.EXPECT().Hand().Return(mockHand2)
		mockHand2.EXPECT().RemoveCard(mockStolenCard)
		// RemoveFromHand(thief) fails
		mockThief.EXPECT().GetID().Return("thief-id")
		mockPlayer1.EXPECT().RemoveFromHand("thief-id").Return(nil, errors.New("thief not found"))

		result, _, err := action.Execute(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "removing thief from hand failed")
		assert.NotNil(t, result)
	})

	t.Run("Success returns result with stolen card info", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockPlayer2, mockHand2, mockStolenCard, mockThief :=
			validateStealAction(t, ctrl, 1)

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer2.EXPECT().Hand().Return(mockHand2)
		mockHand2.EXPECT().RemoveCard(mockStolenCard)
		mockThief.EXPECT().GetID().Return("thief-id")
		mockPlayer1.EXPECT().RemoveFromHand("thief-id").Return([]cards.Card{mockThief}, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockThief)
		mockPlayer1.EXPECT().TakeCards(mockStolenCard)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().StatusWithModal(mockPlayer1, []cards.Card{mockStolenCard}).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

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

		action, mockGame, mockPlayer1, mockPlayer2, mockHand2, mockStolenCard, mockThief :=
			validateStealAction(t, ctrl, 1)

		var capturedMsg string
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer2.EXPECT().Hand().Return(mockHand2)
		mockHand2.EXPECT().RemoveCard(mockStolenCard)
		mockThief.EXPECT().GetID().Return("thief-id")
		mockPlayer1.EXPECT().RemoveFromHand("thief-id").Return([]cards.Card{mockThief}, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockThief)
		mockPlayer1.EXPECT().TakeCards(mockStolenCard)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Do(func(msg string, _ types.Category) {
			capturedMsg = msg
		})

		_, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		assert.Contains(t, capturedMsg, "Player1")
		assert.Contains(t, capturedMsg, "stole")
	})
}

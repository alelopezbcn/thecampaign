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

// validateSabotageAction calls Validate on a new sabotage action and returns key mocks.
// player1 has mockSabotage (looked up by "sabotage-id"); player2 has mockTargetCard in hand.
func validateSabotageAction(t *testing.T, ctrl *gomock.Controller) (
	gameactions.GameAction, *mocks.MockGame, *mocks.MockPlayer, *mocks.MockPlayer, *mocks.MockHand, *mocks.MockCard, *mocks.MockSabotage,
) {
	t.Helper()
	mockGame := mocks.NewMockGame(ctrl)
	mockPlayer1 := mocks.NewMockPlayer(ctrl)
	mockPlayer2 := mocks.NewMockPlayer(ctrl)
	mockHand2 := mocks.NewMockHand(ctrl)
	mockTargetCard := mocks.NewMockCard(ctrl)
	mockSabotage := mocks.NewMockSabotage(ctrl)

	mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeSpySteal)
	mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
	mockPlayer1.EXPECT().GetCardFromHand("sabotage-id").Return(mockSabotage, true)
	mockGame.EXPECT().GetTargetPlayer("Player1", "Player2").Return(mockPlayer2, nil)
	mockPlayer2.EXPECT().Hand().Return(mockHand2)
	mockHand2.EXPECT().ShowCards().Return([]cards.Card{mockTargetCard})

	action := gameactions.NewSabotageAction("Player1", "Player2", "sabotage-id")
	if err := action.Validate(mockGame); err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	return action, mockGame, mockPlayer1, mockPlayer2, mockHand2, mockTargetCard, mockSabotage
}

func TestSabotageAction_PlayerName(t *testing.T) {
	action := gameactions.NewSabotageAction("Player1", "Player2", "sabotage-id")
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestSabotageAction_NextPhase(t *testing.T) {
	action := gameactions.NewSabotageAction("Player1", "Player2", "sabotage-id")
	assert.Equal(t, types.PhaseTypeBuy, action.NextPhase())
}

func TestSabotageAction_Validate(t *testing.T) {
	t.Run("Error when not in SpySteal phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeBuy).Times(2)

		action := gameactions.NewSabotageAction("Player1", "Player2", "sabotage-id")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot use sabotage in the")
	})

	t.Run("Error when card not found in hand", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeSpySteal)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("sabotage-id").Return(nil, false)

		action := gameactions.NewSabotageAction("Player1", "Player2", "sabotage-id")
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
		mockPlayer1.EXPECT().GetCardFromHand("sabotage-id").Return(mockCard, true)

		action := gameactions.NewSabotageAction("Player1", "Player2", "sabotage-id")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a sabotage card")
	})

	t.Run("Error when target player hand is empty", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockHand2 := mocks.NewMockHand(ctrl)
		mockSabotage := mocks.NewMockSabotage(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeSpySteal)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("sabotage-id").Return(mockSabotage, true)
		mockGame.EXPECT().GetTargetPlayer("Player1", "Player2").Return(mockPlayer2, nil)
		mockPlayer2.EXPECT().Hand().Return(mockHand2)
		mockHand2.EXPECT().ShowCards().Return([]cards.Card{})

		action := gameactions.NewSabotageAction("Player1", "Player2", "sabotage-id")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target player has no cards to destroy")
	})

	t.Run("Success with valid target and non-empty hand", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockHand2 := mocks.NewMockHand(ctrl)
		mockSabotage := mocks.NewMockSabotage(ctrl)
		mockCard := mocks.NewMockCard(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeSpySteal)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("sabotage-id").Return(mockSabotage, true)
		mockGame.EXPECT().GetTargetPlayer("Player1", "Player2").Return(mockPlayer2, nil)
		mockPlayer2.EXPECT().Hand().Return(mockHand2)
		mockHand2.EXPECT().ShowCards().Return([]cards.Card{mockCard})

		action := gameactions.NewSabotageAction("Player1", "Player2", "sabotage-id")
		err := action.Validate(mockGame)

		assert.NoError(t, err)
	})
}

func TestSabotageAction_Execute(t *testing.T) {
	t.Run("Error when RemoveFromHand (sabotage card) fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockPlayer2, mockHand2, mockTargetCard, mockSabotage :=
			validateSabotageAction(t, ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		// destroyRandomCard calls Hand().RemoveCard on player2
		mockPlayer2.EXPECT().Hand().Return(mockHand2)
		mockHand2.EXPECT().RemoveCard(mockTargetCard)
		// discard the destroyed card
		mockGame.EXPECT().OnCardMovedToPile(mockTargetCard)
		// RemoveFromHand(sabotage) fails
		mockSabotage.EXPECT().GetID().Return("SAB1")
		mockPlayer1.EXPECT().RemoveFromHand("SAB1").Return(nil, errors.New("sabotage not found"))

		result, _, err := action.Execute(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "removing sabotage card from hand failed")
		assert.NotNil(t, result)
	})

	t.Run("Success returns result with destroyed card info", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockPlayer2, mockHand2, mockTargetCard, mockSabotage :=
			validateSabotageAction(t, ctrl)

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer2.EXPECT().Hand().Return(mockHand2)
		mockHand2.EXPECT().RemoveCard(mockTargetCard)
		mockGame.EXPECT().OnCardMovedToPile(mockTargetCard)
		mockSabotage.EXPECT().GetID().Return("SAB1")
		mockPlayer1.EXPECT().RemoveFromHand("SAB1").Return([]cards.Card{mockSabotage}, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockSabotage)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().StatusWithModal(mockPlayer1, []cards.Card{mockTargetCard}).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionSabotage, result.Action)
		assert.NotNil(t, result.Sabotage)
		assert.Equal(t, "Player2", result.Sabotage.From)
		assert.Equal(t, mockTargetCard, result.Sabotage.Card)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("History updated on successful sabotage", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockPlayer2, mockHand2, mockTargetCard, mockSabotage :=
			validateSabotageAction(t, ctrl)

		var capturedMsg string
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer2.EXPECT().Hand().Return(mockHand2)
		mockHand2.EXPECT().RemoveCard(mockTargetCard)
		mockGame.EXPECT().OnCardMovedToPile(mockTargetCard)
		mockSabotage.EXPECT().GetID().Return("SAB1")
		mockPlayer1.EXPECT().RemoveFromHand("SAB1").Return([]cards.Card{mockSabotage}, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockSabotage)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Do(func(msg string, _ types.Category) {
			capturedMsg = msg
		})

		_, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		assert.Contains(t, capturedMsg, "Player1")
		assert.Contains(t, capturedMsg, "destroyed")
	})
}

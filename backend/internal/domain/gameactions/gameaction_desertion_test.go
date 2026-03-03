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

// validateDesertionAction sets up valid mocks through Validate and returns the action and all mocks.
// player1 has mockDesertion (looked up by "desertion-id"); player2 has mockWarrior on their field.
func validateDesertionAction(
	t *testing.T, ctrl *gomock.Controller,
) (
	gameactions.GameAction,
	*mocks.MockGame,
	*mocks.MockPlayer, // player1 – current player
	*mocks.MockPlayer, // player2 – target player
	*mocks.MockField, // player2's field
	*mocks.MockWarrior, // the weak warrior on player2's field
	*mocks.MockDesertion, // the desertion card in player1's hand
) {
	t.Helper()
	mockGame := mocks.NewMockGame(ctrl)
	mockPlayer1 := mocks.NewMockPlayer(ctrl)
	mockDesertion := mocks.NewMockDesertion(ctrl)
	mockPlayer2 := mocks.NewMockPlayer(ctrl)
	mockField2 := mocks.NewMockField(ctrl)
	mockWarrior := mocks.NewMockWarrior(ctrl)

	mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeSpySteal)
	mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
	mockPlayer1.EXPECT().GetCardFromHand("desertion-id").Return(mockDesertion, true)
	mockGame.EXPECT().GetTargetPlayer("Player1", "Player2").Return(mockPlayer2, nil)
	mockPlayer2.EXPECT().Field().Return(mockField2)
	mockField2.EXPECT().GetWarrior("W1").Return(mockWarrior, true)
	mockWarrior.EXPECT().Health().Return(3) // ≤ DesertionMaxHP (5)

	action := gameactions.NewDesertionAction("Player1", "Player2", "W1", "desertion-id")
	if err := action.Validate(mockGame); err != nil {
		t.Fatalf("validateDesertionAction: unexpected Validate error: %v", err)
	}
	return action, mockGame, mockPlayer1, mockPlayer2, mockField2, mockWarrior, mockDesertion
}

// ──────────────────────────────────────────────────────────────────────────────
// PlayerName / NextPhase
// ──────────────────────────────────────────────────────────────────────────────

func TestDesertionAction_PlayerName(t *testing.T) {
	action := gameactions.NewDesertionAction("Player1", "Player2", "W1", "desertion-id")
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestDesertionAction_NextPhase(t *testing.T) {
	action := gameactions.NewDesertionAction("Player1", "Player2", "W1", "desertion-id")
	assert.Equal(t, types.PhaseTypeBuy, action.NextPhase())
}

// ──────────────────────────────────────────────────────────────────────────────
// Validate
// ──────────────────────────────────────────────────────────────────────────────

func TestDesertionAction_Validate(t *testing.T) {
	t.Run("Error when not in SpySteal phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack).Times(2)

		action := gameactions.NewDesertionAction("Player1", "Player2", "W1", "desertion-id")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot use desertion in the")
	})

	t.Run("Error when card not found in hand", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeSpySteal)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("desertion-id").Return(nil, false)

		action := gameactions.NewDesertionAction("Player1", "Player2", "W1", "desertion-id")
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
		mockPlayer1.EXPECT().GetCardFromHand("desertion-id").Return(mockCard, true)

		action := gameactions.NewDesertionAction("Player1", "Player2", "W1", "desertion-id")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a desertion card")
	})

	t.Run("Error when GetTargetPlayer fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockDesertion := mocks.NewMockDesertion(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeSpySteal)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("desertion-id").Return(mockDesertion, true)
		mockGame.EXPECT().GetTargetPlayer("Player1", "Unknown").Return(nil, errors.New("player not found"))

		action := gameactions.NewDesertionAction("Player1", "Unknown", "W1", "desertion-id")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "player not found")
	})

	t.Run("Error when warrior not found on enemy field", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockDesertion := mocks.NewMockDesertion(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockField2 := mocks.NewMockField(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeSpySteal)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("desertion-id").Return(mockDesertion, true)
		mockGame.EXPECT().GetTargetPlayer("Player1", "Player2").Return(mockPlayer2, nil)
		mockPlayer2.EXPECT().Field().Return(mockField2)
		mockField2.EXPECT().GetWarrior("MISSING").Return(nil, false)

		action := gameactions.NewDesertionAction("Player1", "Player2", "MISSING", "desertion-id")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("Error when warrior HP exceeds DesertionMaxHP", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockDesertion := mocks.NewMockDesertion(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockField2 := mocks.NewMockField(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeSpySteal)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("desertion-id").Return(mockDesertion, true)
		mockGame.EXPECT().GetTargetPlayer("Player1", "Player2").Return(mockPlayer2, nil)
		mockPlayer2.EXPECT().Field().Return(mockField2)
		mockField2.EXPECT().GetWarrior("W1").Return(mockWarrior, true)
		mockWarrior.EXPECT().Health().Return(cards.DesertionMaxHP + 1) // too healthy

		action := gameactions.NewDesertionAction("Player1", "Player2", "W1", "desertion-id")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "HP")
	})

	t.Run("Success with warrior at exactly DesertionMaxHP", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockDesertion := mocks.NewMockDesertion(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockField2 := mocks.NewMockField(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)

		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeSpySteal)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("desertion-id").Return(mockDesertion, true)
		mockGame.EXPECT().GetTargetPlayer("Player1", "Player2").Return(mockPlayer2, nil)
		mockPlayer2.EXPECT().Field().Return(mockField2)
		mockField2.EXPECT().GetWarrior("W1").Return(mockWarrior, true)
		mockWarrior.EXPECT().Health().Return(cards.DesertionMaxHP) // exactly at limit

		action := gameactions.NewDesertionAction("Player1", "Player2", "W1", "desertion-id")
		err := action.Validate(mockGame)

		assert.NoError(t, err)
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// Execute
// ──────────────────────────────────────────────────────────────────────────────

func TestDesertionAction_Execute(t *testing.T) {
	t.Run("Error when RemoveWarrior returns false", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, _, mockPlayer2, mockField2, mockWarrior, _ :=
			validateDesertionAction(t, ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mocks.NewMockPlayer(ctrl))
		mockPlayer2.EXPECT().Field().Return(mockField2)
		mockField2.EXPECT().RemoveWarrior(mockWarrior).Return(false) // removal fails

		result, statusFn, err := action.Execute(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to remove warrior")
		assert.NotNil(t, result)
		assert.Nil(t, statusFn)
	})

	t.Run("Error when RemoveFromHand fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockPlayer2, mockField2, mockWarrior, mockDesertion :=
			validateDesertionAction(t, ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer2.EXPECT().Field().Return(mockField2)
		mockField2.EXPECT().RemoveWarrior(mockWarrior).Return(true)
		mockPlayer1.EXPECT().PlaceWarriorOnField(mockWarrior)
		mockDesertion.EXPECT().GetID().Return("DES1")
		mockPlayer1.EXPECT().RemoveFromHand("DES1").Return(nil, errors.New("card not found"))

		result, statusFn, err := action.Execute(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "removing desertion card from hand failed")
		assert.NotNil(t, result)
		assert.Nil(t, statusFn)
	})

	t.Run("Success: warrior moved to player field and card discarded", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockPlayer2, mockField2, mockWarrior, mockDesertion :=
			validateDesertionAction(t, ctrl)

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer2.EXPECT().Field().Return(mockField2)
		mockField2.EXPECT().RemoveWarrior(mockWarrior).Return(true)
		mockPlayer1.EXPECT().PlaceWarriorOnField(mockWarrior)
		mockDesertion.EXPECT().GetID().Return("DES1")
		mockPlayer1.EXPECT().RemoveFromHand("DES1").Return([]cards.Card{mockDesertion}, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockDesertion)
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer1).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, statusFn)
		assert.Equal(t, types.LastActionDesertion, result.Action)
		assert.NotNil(t, result.Desertion)
		assert.Equal(t, "Player2", result.Desertion.FromPlayer)
		assert.Equal(t, mockWarrior, result.Desertion.Warrior)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("History is recorded on successful desertion", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockPlayer2, mockField2, mockWarrior, mockDesertion :=
			validateDesertionAction(t, ctrl)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer2.EXPECT().Field().Return(mockField2)
		mockField2.EXPECT().RemoveWarrior(mockWarrior).Return(true)
		mockPlayer1.EXPECT().PlaceWarriorOnField(mockWarrior)
		mockDesertion.EXPECT().GetID().Return("DES1")
		mockPlayer1.EXPECT().RemoveFromHand("DES1").Return([]cards.Card{mockDesertion}, nil)
		mockGame.EXPECT().OnCardMovedToPile(mockDesertion)
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockGame.EXPECT().AddHistory(
			gomock.Eq("Player2's warrior deserted to Player1's ranks"),
			types.CategoryAction,
		)
		mockGame.EXPECT().Status(mockPlayer1).Return(gamestatus.GameStatus{}).AnyTimes()

		_, _, err := action.Execute(mockGame)
		assert.NoError(t, err)
	})
}

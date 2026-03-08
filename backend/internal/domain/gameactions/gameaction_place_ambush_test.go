package gameactions_test

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gameactions"
)

func TestPlaceAmbushAction_PlayerName(t *testing.T) {
	action := gameactions.NewPlaceAmbushAction("Player1", "", "amb1")
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestPlaceAmbushAction_NextPhase(t *testing.T) {
	action := gameactions.NewPlaceAmbushAction("Player1", "", "amb1")
	assert.Equal(t, types.PhaseTypeBuy, action.NextPhase())
}

func TestPlaceAmbushAction_Validate(t *testing.T) {
	t.Run("Error when not in Attack phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeBuy).Times(2)

		action := gameactions.NewPlaceAmbushAction("Player1", "", "amb1")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot place ambush in the")
	})

	t.Run("Error when card not in hand", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer := mocks.NewMockPlayer(ctrl)
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockPlayer.EXPECT().GetCardFromHand("amb1").Return(nil, false)

		action := gameactions.NewPlaceAmbushAction("Player1", "", "amb1")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found in hand")
	})

	t.Run("Error when card is not an ambush", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer := mocks.NewMockPlayer(ctrl)
		mockCard := mocks.NewMockCard(ctrl) // not an Ambush
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockPlayer.EXPECT().GetCardFromHand("amb1").Return(mockCard, true)

		action := gameactions.NewPlaceAmbushAction("Player1", "", "amb1")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not an ambush card")
	})

	t.Run("Error when field already has an ambush", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer := mocks.NewMockPlayer(ctrl)
		mockField := mocks.NewMockField(ctrl)
		ambushCard := cards.NewAmbush("amb1")
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockPlayer.EXPECT().GetCardFromHand("AMB1").Return(ambushCard, true)
		mockPlayer.EXPECT().Field().Return(mockField)
		mockField.EXPECT().SlotCards().Return([]cards.Card{cards.NewAmbush("existing")})

		action := gameactions.NewPlaceAmbushAction("Player1", "", "AMB1")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already has an ambush")
	})

	t.Run("Success validates without error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer := mocks.NewMockPlayer(ctrl)
		mockField := mocks.NewMockField(ctrl)
		ambushCard := cards.NewAmbush("amb1")
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockPlayer.EXPECT().GetCardFromHand("AMB1").Return(ambushCard, true)
		mockPlayer.EXPECT().Field().Return(mockField)
		mockField.EXPECT().SlotCards().Return(nil)

		action := gameactions.NewPlaceAmbushAction("Player1", "", "AMB1")
		err := action.Validate(mockGame)

		assert.NoError(t, err)
	})

	t.Run("Error when target player not found (ally field)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer := mocks.NewMockPlayer(ctrl)
		ambushCard := cards.NewAmbush("AMB1")
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockPlayer.EXPECT().GetCardFromHand("AMB1").Return(ambushCard, true)
		mockPlayer.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().GetPlayer("Ally").Return(nil)

		action := gameactions.NewPlaceAmbushAction("Player1", "Ally", "AMB1")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("Error when target is not an ally", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		ambushCard := cards.NewAmbush("AMB1")
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("AMB1").Return(ambushCard, true)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().GetPlayer("Player2").Return(mockPlayer2)
		mockGame.EXPECT().PlayerIndex("Player1").Return(0)
		mockGame.EXPECT().PlayerIndex("Player2").Return(1)
		mockGame.EXPECT().SameTeam(0, 1).Return(false)

		action := gameactions.NewPlaceAmbushAction("Player1", "Player2", "AMB1")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not an ally")
	})

	t.Run("Success validates for ally field", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockAllyField := mocks.NewMockField(ctrl)
		ambushCard := cards.NewAmbush("AMB1")
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("AMB1").Return(ambushCard, true)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().GetPlayer("Player2").Return(mockPlayer2)
		mockGame.EXPECT().PlayerIndex("Player1").Return(0)
		mockGame.EXPECT().PlayerIndex("Player2").Return(1)
		mockGame.EXPECT().SameTeam(0, 1).Return(true)
		mockPlayer2.EXPECT().Field().Return(mockAllyField)
		mockAllyField.EXPECT().SlotCards().Return(nil)

		action := gameactions.NewPlaceAmbushAction("Player1", "Player2", "AMB1")
		err := action.Validate(mockGame)

		assert.NoError(t, err)
	})
}

func TestPlaceAmbushAction_Execute(t *testing.T) {
	t.Run("Ambush removed from hand and placed in field", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}
		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer := mocks.NewMockPlayer(ctrl)
		mockField := mocks.NewMockField(ctrl)
		ambushCard := cards.NewAmbush("amb1")

		// Validate first
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockPlayer.EXPECT().GetCardFromHand("AMB1").Return(ambushCard, true)
		mockPlayer.EXPECT().Field().Return(mockField).AnyTimes()
		mockField.EXPECT().SlotCards().Return(nil)

		action := gameactions.NewPlaceAmbushAction("Player1", "", "AMB1")
		_ = action.Validate(mockGame)

		// Execute
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockPlayer.EXPECT().RemoveFromHand("AMB1").Return(nil, nil)
		mockField.EXPECT().SetSlotCard(ambushCard)
		mockPlayer.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionPlaceAmbush, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Ambush placed on ally field in 2v2", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}
		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockAllyField := mocks.NewMockField(ctrl)
		ambushCard := cards.NewAmbush("AMB1")

		// Validate
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("AMB1").Return(ambushCard, true)
		mockPlayer1.EXPECT().Name().Return("Player1")
		mockGame.EXPECT().GetPlayer("Player2").Return(mockPlayer2)
		mockGame.EXPECT().PlayerIndex("Player1").Return(0)
		mockGame.EXPECT().PlayerIndex("Player2").Return(1)
		mockGame.EXPECT().SameTeam(0, 1).Return(true)
		mockPlayer2.EXPECT().Field().Return(mockAllyField).AnyTimes()
		mockAllyField.EXPECT().SlotCards().Return(nil)

		action := gameactions.NewPlaceAmbushAction("Player1", "Player2", "AMB1")
		_ = action.Validate(mockGame)

		// Execute
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().RemoveFromHand("AMB1").Return(nil, nil)
		mockAllyField.EXPECT().SetSlotCard(ambushCard)
		mockPlayer1.EXPECT().Name().Return("Player1").Times(2) // condition check + format string
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer1).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionPlaceAmbush, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})
}

package gameactions_test

import (
	"strings"
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/gameactions"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestBuyMercenaryAction_PlayerName(t *testing.T) {
	action := gameactions.NewBuyMercenaryAction("Player1", "gold-123")
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestBuyMercenaryAction_NextPhase(t *testing.T) {
	action := gameactions.NewBuyMercenaryAction("Player1", "gold-123")
	assert.Equal(t, types.PhaseTypeConstruct, action.NextPhase())
}

func TestBuyMercenaryAction_Validate(t *testing.T) {
	t.Run("Error when not in Buy phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack).Times(2)

		action := gameactions.NewBuyMercenaryAction("Player1", "gold-123")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot buy mercenary in the")
	})

	t.Run("Error when card not in hand", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer := mocks.NewMockPlayer(ctrl)
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeBuy)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockPlayer.EXPECT().GetCardFromHand("card-123").Return(nil, false)

		action := gameactions.NewBuyMercenaryAction("Player1", "card-123")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resource card not in hand")
	})

	t.Run("Error when card is not a Resource type", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer := mocks.NewMockPlayer(ctrl)
		mockCard := mocks.NewMockCard(ctrl)
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeBuy)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockPlayer.EXPECT().GetCardFromHand("card-123").Return(mockCard, true)

		action := gameactions.NewBuyMercenaryAction("Player1", "card-123")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "only gold cards can be used to hire a mercenary")
	})

	t.Run("Error when resource value is less than 8", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer := mocks.NewMockPlayer(ctrl)
		mockResource := mocks.NewMockResource(ctrl)
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeBuy)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockPlayer.EXPECT().GetCardFromHand("gold-4").Return(mockResource, true)
		mockResource.EXPECT().Value().Return(4)

		action := gameactions.NewBuyMercenaryAction("Player1", "gold-4")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "need at least 8 gold")
	})

	t.Run("Success with gold value exactly 8", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer := mocks.NewMockPlayer(ctrl)
		mockResource := mocks.NewMockResource(ctrl)
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeBuy)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockPlayer.EXPECT().GetCardFromHand("gold-8").Return(mockResource, true)
		mockResource.EXPECT().Value().Return(8)

		action := gameactions.NewBuyMercenaryAction("Player1", "gold-8")
		err := action.Validate(mockGame)

		assert.NoError(t, err)
	})
}

// validateBuyMercenaryAction sets up a successful Validate call and returns the populated action.
func validateBuyMercenaryAction(t *testing.T, ctrl *gomock.Controller, cardID string, goldValue int) (
	action gameactions.GameAction,
	mockGame *mocks.MockGame,
	mockPlayer *mocks.MockPlayer,
	mockResource *mocks.MockResource,
) {
	t.Helper()
	mockGame = mocks.NewMockGame(ctrl)
	mockPlayer = mocks.NewMockPlayer(ctrl)
	mockResource = mocks.NewMockResource(ctrl)

	mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeBuy)
	mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
	mockPlayer.EXPECT().GetCardFromHand(cardID).Return(mockResource, true)
	mockResource.EXPECT().Value().Return(goldValue)

	a := gameactions.NewBuyMercenaryAction("Player1", cardID)
	if err := a.Validate(mockGame); err != nil {
		t.Fatalf("validateBuyMercenaryAction: unexpected error: %v", err)
	}
	return a, mockGame, mockPlayer, mockResource
}

func TestBuyMercenaryAction_Execute(t *testing.T) {
	t.Run("Success: gold removed, mercenary placed on field, history logged", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		action, mockGame, mockPlayer, mockResource := validateBuyMercenaryAction(t, ctrl, "gold-8", 8)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockResource.EXPECT().GetID().Return("gold-8")
		mockPlayer.EXPECT().RemoveFromHand("gold-8").Return(nil, nil)
		mockPlayer.EXPECT().TakeCards(gomock.Any()).Return(true)      // mercenary (any ID)
		mockPlayer.EXPECT().MoveCardToField(gomock.Any()).Return(nil) // mercenary ID (UUID)
		mockGame.EXPECT().OnCardMovedToPile(mockResource)
		mockPlayer.EXPECT().Name().Return("Player1").AnyTimes()
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionBuyMercenary, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("History message contains player name and hired", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer, mockResource := validateBuyMercenaryAction(t, ctrl, "gold-8", 8)

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer)
		mockResource.EXPECT().GetID().Return("gold-8")
		mockPlayer.EXPECT().RemoveFromHand("gold-8").Return(nil, nil)
		mockPlayer.EXPECT().TakeCards(gomock.Any()).Return(true)
		mockPlayer.EXPECT().MoveCardToField(gomock.Any()).Return(nil)
		mockGame.EXPECT().OnCardMovedToPile(mockResource)
		mockPlayer.EXPECT().Name().Return("Player1").AnyTimes()
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Do(func(msg string, _ types.Category) {
			assert.True(t, strings.Contains(msg, "Player1") && strings.Contains(msg, "Mercenary"),
				"History should contain player name and Mercenary")
		})
		mockGame.EXPECT().Status(mockPlayer).Return(gamestatus.GameStatus{})

		_, statusFn, err := action.Execute(mockGame)
		assert.NoError(t, err)
		statusFn() // invoke closure to satisfy Status mock expectation
	})
}

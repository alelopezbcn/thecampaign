package gameactions_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gameactions"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestBuyAction_PlayerName(t *testing.T) {
	action := gameactions.NewBuyAction("Player1", "gold-123")
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestBuyAction_NextPhase(t *testing.T) {
	action := gameactions.NewBuyAction("Player1", "gold-123")
	assert.Equal(t, types.PhaseTypeConstruct, action.NextPhase())
}

func TestBuyAction_Validate(t *testing.T) {
	t.Run("Error when not in Buy phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeAttack).Times(2)

		action := gameactions.NewBuyAction("Player1", "gold-123")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot buy in the")
	})

	t.Run("Error when card not in hand", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeBuy)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("card-123").Return(nil, false)

		action := gameactions.NewBuyAction("Player1", "card-123")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Resource card not in hand: card-123")
	})

	t.Run("Error when card is not a Resource type", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockCard := mocks.NewMockCard(ctrl)
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeBuy)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("card-123").Return(mockCard, true)

		action := gameactions.NewBuyAction("Player1", "card-123")
		err := action.Validate(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "only gold cards can be used to buy")
	})

	t.Run("Success validates without error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGame := mocks.NewMockGame(ctrl)
		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockResource := mocks.NewMockResource(ctrl)
		mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeBuy)
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockPlayer1.EXPECT().GetCardFromHand("gold-123").Return(mockResource, true)

		action := gameactions.NewBuyAction("Player1", "gold-123")
		err := action.Validate(mockGame)

		assert.NoError(t, err)
	})
}

// validateBuyAction sets up a successful Validate call and returns the populated action.
func validateBuyAction(t *testing.T, ctrl *gomock.Controller, cardID string) (
	action gameactions.GameAction,
	mockGame *mocks.MockGame,
	mockPlayer1 *mocks.MockPlayer,
	mockResource *mocks.MockResource,
) {
	t.Helper()
	mockGame = mocks.NewMockGame(ctrl)
	mockPlayer1 = mocks.NewMockPlayer(ctrl)
	mockResource = mocks.NewMockResource(ctrl)

	mockGame.EXPECT().CurrentAction().Return(types.PhaseTypeBuy)
	mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
	mockPlayer1.EXPECT().GetCardFromHand(cardID).Return(mockResource, true)

	a := gameactions.NewBuyAction("Player1", cardID)
	if err := a.Validate(mockGame); err != nil {
		t.Fatalf("validateBuyAction: unexpected error: %v", err)
	}
	return a, mockGame, mockPlayer1, mockResource
}

func TestBuyAction_Execute(t *testing.T) {
	t.Run("Error when DrawCards fails (non-hand-limit)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockResource := validateBuyAction(t, ctrl, "gold-123")

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockResource.EXPECT().Value().Return(2)
		mockResource.EXPECT().GetID().Return("gold-123")
		mockPlayer1.EXPECT().RemoveFromHand("gold-123").Return(nil, nil)
		mockGame.EXPECT().DrawCards(mockPlayer1, 1).Return(nil, errors.New("deck empty"))
		mockPlayer1.EXPECT().TakeCards(mockResource).Return(true)

		result, _, err := action.Execute(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "drawing card for buying failed")
		assert.NotNil(t, result)
	})

	t.Run("Error when hand limit exceeded", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		action, mockGame, mockPlayer1, mockResource := validateBuyAction(t, ctrl, "gold-123")

		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockResource.EXPECT().Value().Return(4)
		mockResource.EXPECT().GetID().Return("gold-123")
		mockPlayer1.EXPECT().RemoveFromHand("gold-123").Return(nil, nil)
		mockGame.EXPECT().DrawCards(mockPlayer1, 2).Return(nil, board.ErrHandLimitExceeded)
		mockPlayer1.EXPECT().TakeCards(mockResource).Return(true)

		result, _, err := action.Execute(mockGame)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cards in hand limit exceeded")
		assert.NotNil(t, result)
	})

	t.Run("Success buying with gold value 2 draws 1 card", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}
		mockDrawnCard := mocks.NewMockCard(ctrl)

		action, mockGame, mockPlayer1, mockResource := validateBuyAction(t, ctrl, "gold-123")

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockResource.EXPECT().Value().Return(2)
		mockResource.EXPECT().GetID().Return("gold-123")
		mockPlayer1.EXPECT().RemoveFromHand("gold-123").Return(nil, nil)
		mockGame.EXPECT().DrawCards(mockPlayer1, 1).Return([]cards.Card{mockDrawnCard}, nil)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard).Return(true)
		mockGame.EXPECT().OnCardMovedToPile(mockResource)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer1, mockDrawnCard).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionBuy, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Success buying with gold value 4 draws 2 cards", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}
		mockDrawnCard1 := mocks.NewMockCard(ctrl)
		mockDrawnCard2 := mocks.NewMockCard(ctrl)

		action, mockGame, mockPlayer1, mockResource := validateBuyAction(t, ctrl, "gold-456")

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockResource.EXPECT().Value().Return(4)
		mockResource.EXPECT().GetID().Return("gold-456")
		mockPlayer1.EXPECT().RemoveFromHand("gold-456").Return(nil, nil)
		mockGame.EXPECT().DrawCards(mockPlayer1, 2).Return([]cards.Card{mockDrawnCard1, mockDrawnCard2}, nil)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard1, mockDrawnCard2).Return(true)
		mockGame.EXPECT().OnCardMovedToPile(mockResource)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer1, mockDrawnCard1, mockDrawnCard2).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionBuy, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Gold value 5 draws 2 cards (integer division)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}
		mockDrawnCard1 := mocks.NewMockCard(ctrl)
		mockDrawnCard2 := mocks.NewMockCard(ctrl)

		action, mockGame, mockPlayer1, mockResource := validateBuyAction(t, ctrl, "gold-5")

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockResource.EXPECT().Value().Return(5)
		mockResource.EXPECT().GetID().Return("gold-5")
		mockPlayer1.EXPECT().RemoveFromHand("gold-5").Return(nil, nil)
		mockGame.EXPECT().DrawCards(mockPlayer1, 2).Return([]cards.Card{mockDrawnCard1, mockDrawnCard2}, nil)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard1, mockDrawnCard2).Return(true)
		mockGame.EXPECT().OnCardMovedToPile(mockResource)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any())
		mockGame.EXPECT().Status(mockPlayer1, mockDrawnCard1, mockDrawnCard2).Return(expectedStatus)

		result, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionBuy, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("History is updated on successful buy", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDrawnCard := mocks.NewMockCard(ctrl)

		action, mockGame, mockPlayer1, mockResource := validateBuyAction(t, ctrl, "gold-123")

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockGame.EXPECT().CurrentPlayer().Return(mockPlayer1)
		mockResource.EXPECT().Value().Return(2)
		mockResource.EXPECT().GetID().Return("gold-123")
		mockPlayer1.EXPECT().RemoveFromHand("gold-123").Return(nil, nil)
		mockGame.EXPECT().DrawCards(mockPlayer1, 1).Return([]cards.Card{mockDrawnCard}, nil)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard).Return(true)
		mockGame.EXPECT().OnCardMovedToPile(mockResource)
		mockGame.EXPECT().AddHistory(gomock.Any(), gomock.Any()).Do(func(msg string, _ types.Category) {
			assert.True(t, strings.Contains(msg, "Player1") && strings.Contains(msg, "bought"),
				"History should contain buy action")
		})
		_, statusFn, err := action.Execute(mockGame)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
	})
}

package game

import (
	"errors"
	"strings"
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestBuyAction_PlayerName(t *testing.T) {
	action := NewBuyAction("Player1", "gold-123")
	assert.Equal(t, "Player1", action.PlayerName())
}

func TestBuyAction_NextPhase(t *testing.T) {
	action := NewBuyAction("Player1", "gold-123")
	assert.Equal(t, types.PhaseTypeConstruct, action.NextPhase())
}

func TestBuyAction_Validate(t *testing.T) {
	t.Run("Error when not in Buy phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeAttack,
		}

		action := NewBuyAction("Player1", "gold-123")
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot buy in the")
	})

	t.Run("Error when card not in hand", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromHand("card-123").Return(nil, false)

		g := &Game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeBuy,
		}

		action := NewBuyAction("Player1", "card-123")
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Resource card not in hand: card-123")
	})

	t.Run("Error when card is not a Resource type", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockCard := mocks.NewMockCard(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromHand("card-123").Return(mockCard, true)

		g := &Game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeBuy,
		}

		action := NewBuyAction("Player1", "card-123")
		err := action.Validate(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "only gold cards can be used to buy")
	})

	t.Run("Success stores resource", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockResource := mocks.NewMockResource(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromHand("gold-123").Return(mockResource, true)

		g := &Game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeBuy,
		}

		action := NewBuyAction("Player1", "gold-123")
		err := action.Validate(g)

		assert.NoError(t, err)
		assert.Equal(t, mockResource, action.resource)
	})
}

func TestBuyAction_Execute(t *testing.T) {
	t.Run("Error when hand limit exceeded", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockResource := mocks.NewMockResource(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockResource.EXPECT().Value().Return(4)
		mockResource.EXPECT().GetID().Return("gold-123")
		mockPlayer1.EXPECT().GiveCards("gold-123").Return(nil, nil)
		mockPlayer1.EXPECT().CanTakeCards(2).Return(false)
		mockPlayer1.EXPECT().TakeCards(mockResource).Return(true) // resource returned to hand

		g := &Game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeBuy,
		}

		action := NewBuyAction("Player1", "gold-123")
		action.resource = mockResource

		result, _, err := action.Execute(g)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cards in hand limit exceeded")
		assert.NotNil(t, result)
	})

	t.Run("Success buying with gold value 2 draws 1 card", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockResource := mocks.NewMockResource(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDrawnCard := mocks.NewMockCard(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockResource.EXPECT().Value().Return(2)
		mockResource.EXPECT().GetID().Return("gold-123")
		mockPlayer1.EXPECT().GiveCards("gold-123").Return(nil, nil)
		mockPlayer1.EXPECT().CanTakeCards(1).Return(true)
		mockDeck.EXPECT().DrawCards(1, mockDiscardPile).Return([]cards.Card{mockDrawnCard}, nil)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard).Return(true)
		mockDiscardPile.EXPECT().Discard(mockResource)

		g := &Game{
			players:            []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.PhaseTypeBuy,
			deck:               mockDeck,
			discardPile:        mockDiscardPile,
			gameStatusProvider: mockProvider,
			history:            []types.HistoryLine{},
		}

		mockProvider.EXPECT().Get(mockPlayer1, g, mockDrawnCard).Return(expectedStatus)

		action := NewBuyAction("Player1", "gold-123")
		action.resource = mockResource

		result, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionBuy, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Success buying with gold value 4 draws 2 cards", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockResource := mocks.NewMockResource(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDrawnCard1 := mocks.NewMockCard(ctrl)
		mockDrawnCard2 := mocks.NewMockCard(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockResource.EXPECT().Value().Return(4)
		mockResource.EXPECT().GetID().Return("gold-456")
		mockPlayer1.EXPECT().GiveCards("gold-456").Return(nil, nil)
		mockPlayer1.EXPECT().CanTakeCards(2).Return(true)
		mockDeck.EXPECT().DrawCards(2, mockDiscardPile).Return([]cards.Card{mockDrawnCard1, mockDrawnCard2}, nil)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard1, mockDrawnCard2).Return(true)
		mockDiscardPile.EXPECT().Discard(mockResource)

		g := &Game{
			players:            []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.PhaseTypeBuy,
			deck:               mockDeck,
			discardPile:        mockDiscardPile,
			gameStatusProvider: mockProvider,
			history:            []types.HistoryLine{},
		}

		mockProvider.EXPECT().Get(mockPlayer1, g, mockDrawnCard1, mockDrawnCard2).Return(expectedStatus)

		action := NewBuyAction("Player1", "gold-456")
		action.resource = mockResource

		result, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionBuy, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Gold value 5 draws 2 cards (integer division)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockResource := mocks.NewMockResource(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDrawnCard1 := mocks.NewMockCard(ctrl)
		mockDrawnCard2 := mocks.NewMockCard(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockResource.EXPECT().Value().Return(5)
		mockResource.EXPECT().GetID().Return("gold-5")
		mockPlayer1.EXPECT().GiveCards("gold-5").Return(nil, nil)
		mockPlayer1.EXPECT().CanTakeCards(2).Return(true)
		mockDeck.EXPECT().DrawCards(2, mockDiscardPile).Return([]cards.Card{mockDrawnCard1, mockDrawnCard2}, nil)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard1, mockDrawnCard2).Return(true)
		mockDiscardPile.EXPECT().Discard(mockResource)

		g := &Game{
			players:            []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.PhaseTypeBuy,
			deck:               mockDeck,
			discardPile:        mockDiscardPile,
			gameStatusProvider: mockProvider,
			history:            []types.HistoryLine{},
		}

		mockProvider.EXPECT().Get(mockPlayer1, g, mockDrawnCard1, mockDrawnCard2).Return(expectedStatus)

		action := NewBuyAction("Player1", "gold-5")
		action.resource = mockResource

		result, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionBuy, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("Success when deck needs replenishing from discard pile", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockResource := mocks.NewMockResource(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDrawnCard := mocks.NewMockCard(ctrl)
		mockDiscardedCard := mocks.NewMockCard(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		expectedStatus := gamestatus.GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockResource.EXPECT().Value().Return(2)
		mockResource.EXPECT().GetID().Return("gold-123")
		mockPlayer1.EXPECT().GiveCards("gold-123").Return(nil, nil)
		mockPlayer1.EXPECT().CanTakeCards(1).Return(true)
		// First draw fails - deck is empty
		mockDeck.EXPECT().DrawCards(1, mockDiscardPile).Return(nil, errors.New("no cards left to draw"))
		// Replenish from discard pile
		mockDiscardPile.EXPECT().Empty().Return([]cards.Card{mockDiscardedCard})
		mockDeck.EXPECT().DrawCards(1, mockDiscardPile).Return([]cards.Card{mockDrawnCard}, nil)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard).Return(true)
		mockDiscardPile.EXPECT().Discard(mockResource)

		g := &Game{
			players:            []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.PhaseTypeBuy,
			deck:               mockDeck,
			discardPile:        mockDiscardPile,
			gameStatusProvider: mockProvider,
			history:            []types.HistoryLine{},
		}

		mockProvider.EXPECT().Get(mockPlayer1, g, mockDrawnCard).Return(expectedStatus)

		action := NewBuyAction("Player1", "gold-123")
		action.resource = mockResource

		result, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, types.LastActionBuy, result.Action)
		assert.Equal(t, expectedStatus, statusFn())
	})

	t.Run("History updated on successful buy", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockResource := mocks.NewMockResource(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)
		mockDrawnCard := mocks.NewMockCard(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockResource.EXPECT().Value().Return(2)
		mockResource.EXPECT().GetID().Return("gold-123")
		mockPlayer1.EXPECT().GiveCards("gold-123").Return(nil, nil)
		mockPlayer1.EXPECT().CanTakeCards(1).Return(true)
		mockDeck.EXPECT().DrawCards(1, mockDiscardPile).Return([]cards.Card{mockDrawnCard}, nil)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard).Return(true)
		mockDiscardPile.EXPECT().Discard(mockResource)

		g := &Game{
			players:       []board.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.PhaseTypeBuy,
			deck:          mockDeck,
			discardPile:   mockDiscardPile,
			history:       []types.HistoryLine{},
		}

		action := NewBuyAction("Player1", "gold-123")
		action.resource = mockResource

		_, statusFn, err := action.Execute(g)

		assert.NoError(t, err)
		assert.NotNil(t, statusFn)
		found := false
		for _, h := range g.history {
			if strings.Contains(h.Msg, "Player1") && strings.Contains(h.Msg, "bought") {
				found = true
				break
			}
		}
		assert.True(t, found, "History should contain buy action")
	})
}

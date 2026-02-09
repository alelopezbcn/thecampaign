package domain

import (
	"errors"
	"strings"
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGame_Buy(t *testing.T) {
	t.Run("Error when not current player's turn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:     []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn: 0,
		}

		status, err := g.Buy("Player2", "card-123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Player2 not your turn")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when card not in hand", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromHand("card-123").Return(nil, false)

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeBuy,
		}

		status, err := g.Buy("Player1", "card-123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Resource card not in hand: card-123")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when card is not a Resource type", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockCard := mocks.NewMockCard(ctrl) // Not a Resource

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromHand("card-123").Return(mockCard, true)

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeBuy,
		}

		status, err := g.Buy("Player1", "card-123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "only gold cards can be used to buy")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when not in Buy phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeAttack, // Not in Buy phase
		}

		status, err := g.Buy("Player1", "gold-123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot buy in the")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when player cannot take more cards (hand limit exceeded)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockResource := mocks.NewMockResource(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromHand("gold-123").Return(mockResource, true)
		mockResource.EXPECT().Value().Return(4)
		mockResource.EXPECT().GetID().Return("gold-123")
		mockPlayer1.EXPECT().GiveCards("gold-123").Return(nil, nil)
		mockPlayer1.EXPECT().TakeCards(mockResource).Return(true)
		mockPlayer1.EXPECT().CanTakeCards(2).Return(false) // Hand limit exceeded

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeBuy,
			deck:          mockDeck,
		}

		status, err := g.Buy("Player1", "gold-123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cards in hand limit exceeded")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Success buying with gold value 2 (draws 1 card)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockResource := mocks.NewMockResource(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDrawnCard := mocks.NewMockCard(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		expectedStatus := GameStatus{
			CurrentPlayer: "Player1",
		}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromHand("gold-123").Return(mockResource, true)
		mockResource.EXPECT().Value().Return(2)
		mockResource.EXPECT().GetID().Return("gold-123")
		mockPlayer1.EXPECT().GiveCards("gold-123").Return(nil, nil)
		mockPlayer1.EXPECT().CanTakeCards(1).Return(true)
		mockDeck.EXPECT().DrawCard().Return(mockDrawnCard, true)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard).Return(true)
		mockDiscardPile.EXPECT().Discard(mockResource)
		// nextAction expectations
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)
		mockPlayer1.EXPECT().CanConstruct().Return(true)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.ActionTypeBuy,
			deck:               mockDeck,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
		}

		mockProvider.EXPECT().Get(
			mockPlayer1,
			g, mockDrawnCard,
		).Return(expectedStatus)
		status, err := g.Buy("Player1", "gold-123")

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
		assert.Equal(t, types.ActionTypeConstruct, g.currentAction)
	})

	t.Run("Success buying with gold value 4 (draws 2 cards)", func(t *testing.T) {
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

		expectedStatus := GameStatus{
			CurrentPlayer: "Player1",
		}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromHand("gold-456").Return(mockResource, true)
		mockResource.EXPECT().Value().Return(4)
		mockResource.EXPECT().GetID().Return("gold-456")
		mockPlayer1.EXPECT().GiveCards("gold-456").Return(nil, nil)
		mockPlayer1.EXPECT().CanTakeCards(2).Return(true)
		mockDeck.EXPECT().DrawCard().Return(mockDrawnCard1, true)
		mockDeck.EXPECT().DrawCard().Return(mockDrawnCard2, true)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard1, mockDrawnCard2).Return(true)
		mockDiscardPile.EXPECT().Discard(mockResource)
		// nextAction expectations
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)
		mockPlayer1.EXPECT().CanConstruct().Return(true)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.ActionTypeBuy,
			deck:               mockDeck,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
		}
		mockProvider.EXPECT().Get(
			mockPlayer1,
			g, mockDrawnCard1, mockDrawnCard2,
		).Return(expectedStatus)

		status, err := g.Buy("Player1", "gold-456")

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
		assert.Equal(t, types.ActionTypeConstruct, g.currentAction)
	})

	t.Run("Success buying with gold value 6 (draws 3 cards)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockResource := mocks.NewMockResource(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDrawnCard1 := mocks.NewMockCard(ctrl)
		mockDrawnCard2 := mocks.NewMockCard(ctrl)
		mockDrawnCard3 := mocks.NewMockCard(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		expectedStatus := GameStatus{
			CurrentPlayer: "Player1",
		}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromHand("gold-789").Return(mockResource, true)
		mockResource.EXPECT().Value().Return(6)
		mockResource.EXPECT().GetID().Return("gold-789")
		mockPlayer1.EXPECT().GiveCards("gold-789").Return(nil, nil)
		mockPlayer1.EXPECT().CanTakeCards(3).Return(true)
		mockDeck.EXPECT().DrawCard().Return(mockDrawnCard1, true)
		mockDeck.EXPECT().DrawCard().Return(mockDrawnCard2, true)
		mockDeck.EXPECT().DrawCard().Return(mockDrawnCard3, true)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard1, mockDrawnCard2, mockDrawnCard3).Return(true)
		mockDiscardPile.EXPECT().Discard(mockResource)
		// nextAction expectations
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)
		mockPlayer1.EXPECT().CanConstruct().Return(true)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.ActionTypeBuy,
			deck:               mockDeck,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
		}
		mockProvider.EXPECT().Get(
			mockPlayer1,
			g, mockDrawnCard1, mockDrawnCard2, mockDrawnCard3,
		).Return(expectedStatus)

		status, err := g.Buy("Player1", "gold-789")

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
		assert.Equal(t, types.ActionTypeConstruct, g.currentAction)
	})

	t.Run("Success buying when deck needs replenishing from discard pile", func(t *testing.T) {
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

		expectedStatus := GameStatus{
			CurrentPlayer: "Player1",
		}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromHand("gold-123").Return(mockResource, true)
		mockResource.EXPECT().Value().Return(2)
		mockResource.EXPECT().GetID().Return("gold-123")
		mockPlayer1.EXPECT().GiveCards("gold-123").Return(nil, nil)
		mockPlayer1.EXPECT().CanTakeCards(1).Return(true)
		// First draw fails - deck is empty
		mockDeck.EXPECT().DrawCard().Return(nil, false)
		// Replenish is called with cards from discard pile
		mockDiscardPile.EXPECT().Empty().Return([]ports.Card{mockDiscardedCard})
		mockDeck.EXPECT().Replenish([]ports.Card{mockDiscardedCard})
		// Second draw succeeds after replenish
		mockDeck.EXPECT().DrawCard().Return(mockDrawnCard, true)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard).Return(true)
		mockDiscardPile.EXPECT().Discard(mockResource)
		// nextAction expectations
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)
		mockPlayer1.EXPECT().CanConstruct().Return(true)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.ActionTypeBuy,
			deck:               mockDeck,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
		}

		mockProvider.EXPECT().Get(
			mockPlayer1,
			g, mockDrawnCard,
		).Return(expectedStatus)

		status, err := g.Buy("Player1", "gold-123")

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
	})

	t.Run("Success buying with CanMoveWarrior and CanTrade flags set", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockResource := mocks.NewMockResource(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDrawnCard := mocks.NewMockCard(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		expectedStatus := GameStatus{
			CurrentPlayer: "Player1",
		}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromHand("gold-123").Return(mockResource, true)
		mockResource.EXPECT().Value().Return(2)
		mockResource.EXPECT().GetID().Return("gold-123")
		mockPlayer1.EXPECT().GiveCards("gold-123").Return(nil, nil)
		mockPlayer1.EXPECT().CanTakeCards(1).Return(true)
		mockDeck.EXPECT().DrawCard().Return(mockDrawnCard, true)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard).Return(true)
		mockDiscardPile.EXPECT().Discard(mockResource)
		// nextAction expectations
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)
		mockPlayer1.EXPECT().CanConstruct().Return(true)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.ActionTypeBuy,
			deck:               mockDeck,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
			CanMoveWarrior:     true,
			CanTrade:           true,
		}

		mockProvider.EXPECT().Get(
			mockPlayer1,
			g, mockDrawnCard,
		).Return(expectedStatus)

		status, err := g.Buy("Player1", "gold-123")

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
	})

	t.Run("Buy with gold value 5 draws 2 cards (integer division)", func(t *testing.T) {
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

		expectedStatus := GameStatus{
			CurrentPlayer: "Player1",
		}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromHand("gold-5").Return(mockResource, true)
		mockResource.EXPECT().Value().Return(5) // 5/2 = 2 cards
		mockResource.EXPECT().GetID().Return("gold-5")
		mockPlayer1.EXPECT().GiveCards("gold-5").Return(nil, nil)
		mockPlayer1.EXPECT().CanTakeCards(2).Return(true)
		mockDeck.EXPECT().DrawCard().Return(mockDrawnCard1, true)
		mockDeck.EXPECT().DrawCard().Return(mockDrawnCard2, true)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard1, mockDrawnCard2).Return(true)
		mockDiscardPile.EXPECT().Discard(mockResource)
		// nextAction expectations
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)
		mockPlayer1.EXPECT().CanConstruct().Return(true)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.ActionTypeBuy,
			deck:               mockDeck,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
		}

		mockProvider.EXPECT().Get(
			mockPlayer1,
			g, mockDrawnCard1, mockDrawnCard2,
		).Return(expectedStatus)

		status, err := g.Buy("Player1", "gold-5")

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
	})

	t.Run("Player 2 can buy on their turn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockResource := mocks.NewMockResource(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDrawnCard := mocks.NewMockCard(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		expectedStatus := GameStatus{
			CurrentPlayer: "Player2",
		}

		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer2.EXPECT().GetCardFromHand("gold-p2").Return(mockResource, true)
		mockResource.EXPECT().Value().Return(2)
		mockResource.EXPECT().GetID().Return("gold-p2")
		mockPlayer2.EXPECT().GiveCards("gold-p2").Return(nil, nil)
		mockPlayer2.EXPECT().CanTakeCards(1).Return(true)
		mockDeck.EXPECT().DrawCard().Return(mockDrawnCard, true)
		mockPlayer2.EXPECT().TakeCards(mockDrawnCard).Return(true)
		mockDiscardPile.EXPECT().Discard(mockResource)
		// nextAction expectations
		mockPlayer2.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer2.EXPECT().CanTradeCards().Return(false)
		mockPlayer2.EXPECT().CanConstruct().Return(true)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        1, // Player 2's turn
			currentAction:      types.ActionTypeBuy,
			deck:               mockDeck,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
		}

		mockProvider.EXPECT().Get(
			mockPlayer2,
			g,
			mockDrawnCard,
		).Return(expectedStatus)

		status, err := g.Buy("Player2", "gold-p2")

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
	})
}

func TestGame_DrawCard(t *testing.T) {
	t.Run("Error when not current player's turn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:     []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn: 0,
		}

		status, err := g.DrawCard("Player2")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Player2 not your turn")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when deck is empty and discard pile is also empty", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().CanTakeCards(1).Return(true)
		// First draw attempt fails
		mockDeck.EXPECT().DrawCard().Return(nil, false)
		// Replenish is called with empty discard pile
		mockDiscardPile.EXPECT().Empty().Return([]ports.Card{})
		mockDeck.EXPECT().Replenish([]ports.Card{})
		// Second draw attempt also fails
		mockDeck.EXPECT().DrawCard().Return(nil, false)

		g := &Game{
			Players:     []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn: 0,
			deck:        mockDeck,
			discardPile: mockDiscardPile,
		}

		status, err := g.DrawCard("Player1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no cards left to draw")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Success drawing card normally", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDrawnCard := mocks.NewMockCard(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		expectedStatus := GameStatus{
			CurrentPlayer: "Player1",
		}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().CanTakeCards(1).Return(true)
		mockDeck.EXPECT().DrawCard().Return(mockDrawnCard, true)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard).Return(true)
		// nextAction expectations for ActionTypeAttack
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)
		mockPlayer1.EXPECT().HasCatapult().Return(false)
		// Castle().CanBeAttacked() not called because HasCatapult() returned false (short-circuit)
		mockPlayer1.EXPECT().CanAttack().Return(true) // Can attack

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			deck:               mockDeck,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
		}

		mockProvider.EXPECT().Get(
			mockPlayer1,
			g,
			mockDrawnCard,
		).Return(expectedStatus)

		status, err := g.DrawCard("Player1")

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
		assert.Equal(t, types.ActionTypeAttack, g.currentAction)
	})

	t.Run("Success when hand limit exceeded - continues but doesn't take card", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		expectedStatus := GameStatus{
			CurrentPlayer: "Player1",
		}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().CanTakeCards(1).Return(false) // Hand limit exceeded
		// Note: TakeCards should NOT be called when hand limit exceeded
		// nextAction expectations for ActionTypeAttack
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)
		mockPlayer1.EXPECT().HasCatapult().Return(false)
		// Castle().CanBeAttacked() not called because HasCatapult() returned false (short-circuit)
		mockPlayer1.EXPECT().CanAttack().Return(true)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			deck:               mockDeck,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
		}

		mockProvider.EXPECT().Get(
			mockPlayer1,
			g,
		).Return(expectedStatus)

		status, err := g.DrawCard("Player1")

		assert.NoError(t, err) // No error returned - game continues
		assert.Equal(t, expectedStatus, status)
		assert.Equal(t, types.ActionTypeAttack, g.currentAction)
	})

	t.Run("Success with deck replenishing from discard pile", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDrawnCard := mocks.NewMockCard(ctrl)
		mockDiscardedCard := mocks.NewMockCard(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		expectedStatus := GameStatus{
			CurrentPlayer: "Player1",
		}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().CanTakeCards(1).Return(true)
		// First draw fails - deck is empty
		mockDeck.EXPECT().DrawCard().Return(nil, false)
		// Replenish is called
		mockDiscardPile.EXPECT().Empty().Return([]ports.Card{mockDiscardedCard})
		mockDeck.EXPECT().Replenish([]ports.Card{mockDiscardedCard})
		// Second draw succeeds after replenish
		mockDeck.EXPECT().DrawCard().Return(mockDrawnCard, true)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard).Return(true)
		// nextAction expectations for ActionTypeAttack
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)
		mockPlayer1.EXPECT().HasCatapult().Return(false)
		// Castle().CanBeAttacked() not called because HasCatapult() returned false (short-circuit)
		mockPlayer1.EXPECT().CanAttack().Return(true)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			deck:               mockDeck,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
		}

		mockProvider.EXPECT().Get(
			mockPlayer1,
			g,
			mockDrawnCard,
		).Return(expectedStatus)

		status, err := g.DrawCard("Player1")

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
	})

	t.Run("Verify history is updated on successful draw", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDrawnCard := mocks.NewMockCard(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().CanTakeCards(1).Return(true)
		mockDeck.EXPECT().DrawCard().Return(mockDrawnCard, true)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard).Return(true)
		// nextAction expectations for ActionTypeAttack
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)
		mockPlayer1.EXPECT().HasCatapult().Return(false)
		// Castle().CanBeAttacked() not called because HasCatapult() returned false (short-circuit)
		mockPlayer1.EXPECT().CanAttack().Return(true)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			deck:               mockDeck,
			discardPile:        mockDiscardPile,
			history:            []historyLine{},
			GameStatusProvider: mockProvider,
		}

		mockProvider.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(GameStatus{})

		_, err := g.DrawCard("Player1")

		assert.NoError(t, err)
		found := false
		for _, h := range g.history {
			if strings.Contains(h.Msg,"drew") && strings.Contains(h.Msg,"Player1") {
				found = true
				break
			}
		}
		assert.True(t, found, "History should contain the draw action")
	})

	t.Run("Verify history is updated when hand limit exceeded", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().CanTakeCards(1).Return(false) // Hand limit exceeded
		// nextAction expectations for ActionTypeAttack
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)
		mockPlayer1.EXPECT().HasCatapult().Return(false)
		// Castle().CanBeAttacked() not called because HasCatapult() returned false (short-circuit)
		mockPlayer1.EXPECT().CanAttack().Return(true)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			deck:               mockDeck,
			discardPile:        mockDiscardPile,
			history:            []historyLine{},
			GameStatusProvider: mockProvider,
		}

		mockProvider.EXPECT().Get(gomock.Any(), gomock.Any()).Return(GameStatus{})

		_, err := g.DrawCard("Player1")

		assert.NoError(t, err)
		found := false
		for _, h := range g.history {
			if strings.Contains(h.Msg,"can't take more cards") {
				found = true
				break
			}
		}
		assert.True(t, found, "History should contain hand limit exceeded message")
	})

	t.Run("Success with CanMoveWarrior and CanTrade flags set", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDrawnCard := mocks.NewMockCard(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		expectedStatus := GameStatus{
			CurrentPlayer: "Player1",
		}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().CanTakeCards(1).Return(true)
		mockDeck.EXPECT().DrawCard().Return(mockDrawnCard, true)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard).Return(true)
		// nextAction expectations for ActionTypeAttack
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(true) // Has warriors
		mockPlayer1.EXPECT().CanTradeCards().Return(true)     // Can trade
		mockPlayer1.EXPECT().HasCatapult().Return(false)
		// Castle().CanBeAttacked() not called because HasCatapult() returned false (short-circuit)
		mockPlayer1.EXPECT().CanAttack().Return(true)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			deck:               mockDeck,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
			CanMoveWarrior:     true,
			CanTrade:           true,
		}

		mockProvider.EXPECT().Get(
			mockPlayer1,
			g,
			mockDrawnCard,
		).Return(expectedStatus)

		status, err := g.DrawCard("Player1")

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
	})

	t.Run("Player 2 can draw on their turn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDrawnCard := mocks.NewMockCard(ctrl)
		mockCemetery := mocks.NewMockCemetery(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		expectedStatus := GameStatus{
			CurrentPlayer: "Player2",
		}

		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer2.EXPECT().CanTakeCards(1).Return(true)
		mockDeck.EXPECT().DrawCard().Return(mockDrawnCard, true)
		mockPlayer2.EXPECT().TakeCards(mockDrawnCard).Return(true)
		// nextAction expectations for ActionTypeAttack
		mockPlayer2.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer2.EXPECT().CanTradeCards().Return(false)
		mockPlayer2.EXPECT().HasCatapult().Return(false)
		// Castle().CanBeAttacked() not called because HasCatapult() returned false (short-circuit)
		mockPlayer2.EXPECT().CanAttack().Return(true)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        1, // Player 2's turn
			deck:               mockDeck,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
			cemetery:           mockCemetery,
		}

		mockProvider.EXPECT().Get(
			mockPlayer2,
			g,
			mockDrawnCard,
		).Return(expectedStatus)

		status, err := g.DrawCard("Player2")

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
		assert.Equal(t, types.ActionTypeAttack, g.currentAction)
	})
}

// func TestAttacks(t *testing.T) {
// 	t.Run("Knight attacks Archer causing double damage", func(t *testing.T) {
// 		dmgAmnt := 4
// 		k := cards.NewKnight("k1")
// 		a := cards.NewArcher("a1")
// 		sword := cards.NewSword("s1", dmgAmnt)

// 		p1 := newPlayerWithCards("Player1",
// 			[]ports.Card{sword},
// 			[]ports.Warrior{k},
// 		)
// 		p2 := newPlayerWithCards("Player2",
// 			[]ports.Card{},
// 			[]ports.Warrior{a},
// 		)
// 		g := &Game{
// 			Players: []ports.Player{p1, p2},
// 		}

// 		err := g.Attack(p1.Name(), a.GetID(), sword.GetID())
// 		assert.NoError(t, err)
// 		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*2, a.Health())
// 		assert.NotContains(t, p1.Hand().ShowCards(), sword)
// 	})
// 	t.Run("Knight attacks Mage causing normal damage", func(t *testing.T) {
// 		dmgAmnt := 4
// 		k := cards.NewKnight("k1")
// 		m := cards.NewMage("m1")
// 		sword := cards.NewSword("s1", dmgAmnt)

// 		p1 := newPlayerWithCards("Player1",
// 			[]ports.Card{sword},
// 			[]ports.Warrior{k},
// 		)
// 		p2 := newPlayerWithCards("Player2",
// 			[]ports.Card{},
// 			[]ports.Warrior{m},
// 		)
// 		g := &Game{
// 			Players: []ports.Player{p1, p2},
// 		}

// 		err := g.Attack(p1.Name(), m.GetID(), sword.GetID())
// 		assert.NoError(t, err)
// 		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*1, m.Health())
// 		assert.NotContains(t, p1.Hand().ShowCards(), sword)
// 	})
// 	t.Run("Knight attacks Knight causing normal damage", func(t *testing.T) {
// 		dmgAmnt := 4
// 		k := cards.NewKnight("k1")
// 		k2 := cards.NewKnight("k2")
// 		sword := cards.NewSword("s1", dmgAmnt)

// 		p1 := newPlayerWithCards("Player1",
// 			[]ports.Card{sword},
// 			[]ports.Warrior{k},
// 		)
// 		p2 := newPlayerWithCards("Player2",
// 			[]ports.Card{},
// 			[]ports.Warrior{k2},
// 		)
// 		g := &Game{
// 			Players: []ports.Player{p1, p2},
// 		}

// 		err := g.Attack(p1.Name(), k2.GetID(), sword.GetID())
// 		assert.NoError(t, err)
// 		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*1, k2.Health())
// 		assert.NotContains(t, p1.Hand().ShowCards(), sword)
// 	})
// 	t.Run("Knight attacks Dragon causing normal damage", func(t *testing.T) {
// 		dmgAmnt := 4
// 		k := cards.NewKnight("k1")
// 		d := cards.NewDragon("d1")
// 		sword := cards.NewSword("s1", dmgAmnt)

// 		p1 := newPlayerWithCards("Player1",
// 			[]ports.Card{sword},
// 			[]ports.Warrior{k},
// 		)
// 		p2 := newPlayerWithCards("Player2",
// 			[]ports.Card{},
// 			[]ports.Warrior{d},
// 		)
// 		g := &Game{
// 			Players: []ports.Player{p1, p2},
// 		}

// 		err := g.Attack(p1.Name(), d.GetID(), sword.GetID())
// 		assert.NoError(t, err)
// 		assert.Equal(t, cards.DragonMaxHealth-dmgAmnt*1, d.Health())
// 		assert.NotContains(t, p1.Hand().ShowCards(), sword)
// 	})
// 	t.Run("Knight cant attack with wrong weapon", func(t *testing.T) {
// 		dmgAmnt := 4
// 		k := cards.NewKnight("k1")
// 		a := cards.NewArcher("a1")
// 		poison := cards.NewPoison("s1", dmgAmnt)

// 		p1 := newPlayerWithCards("Player1",
// 			[]ports.Card{poison},
// 			[]ports.Warrior{k},
// 		)
// 		p2 := newPlayerWithCards("Player2",
// 			[]ports.Card{},
// 			[]ports.Warrior{a},
// 		)
// 		g := &Game{
// 			Players: []ports.Player{p1, p2},
// 		}

// 		err := g.Attack(p1.Name(), a.GetID(), poison.GetID())
// 		assert.Error(t, err)
// 		assert.Equal(t, cards.WarriorMaxHealth, a.Health())
// 		assert.Contains(t, p1.Hand().ShowCards(), poison)
// 	})

// 	t.Run("Archer attacks Mage causing double damage", func(t *testing.T) {
// 		dmgAmnt := 4
// 		attacker := cards.NewArcher("a1")
// 		target := cards.NewMage("a1")
// 		weapon := cards.NewArrow("s1", dmgAmnt)

// 		p1 := newPlayerWithCards("Player1",
// 			[]ports.Card{weapon},
// 			[]ports.Warrior{attacker},
// 		)
// 		p2 := newPlayerWithCards("Player2",
// 			[]ports.Card{},
// 			[]ports.Warrior{target},
// 		)
// 		g := &Game{
// 			Players: []ports.Player{p1, p2},
// 		}

// 		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
// 		assert.NoError(t, err)
// 		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*2, target.Health())
// 		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
// 	})
// 	t.Run("Archer attacks Knight causing normal damage", func(t *testing.T) {
// 		dmgAmnt := 4
// 		attacker := cards.NewArcher("a1")
// 		target := cards.NewKnight("k1")
// 		weapon := cards.NewArrow("s1", dmgAmnt)

// 		p1 := newPlayerWithCards("Player1",
// 			[]ports.Card{weapon},
// 			[]ports.Warrior{attacker},
// 		)
// 		p2 := newPlayerWithCards("Player2",
// 			[]ports.Card{},
// 			[]ports.Warrior{target},
// 		)
// 		g := &Game{
// 			Players: []ports.Player{p1, p2},
// 		}

// 		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
// 		assert.NoError(t, err)
// 		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*1, target.Health())
// 		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
// 	})
// 	t.Run("Archer attacks Archer causing normal damage", func(t *testing.T) {
// 		dmgAmnt := 4
// 		attacker := cards.NewArcher("a1")
// 		target := cards.NewArcher("a2")
// 		weapon := cards.NewArrow("s1", dmgAmnt)

// 		p1 := newPlayerWithCards("Player1",
// 			[]ports.Card{weapon},
// 			[]ports.Warrior{attacker},
// 		)
// 		p2 := newPlayerWithCards("Player2",
// 			[]ports.Card{},
// 			[]ports.Warrior{target},
// 		)
// 		g := &Game{
// 			Players: []ports.Player{p1, p2},
// 		}

// 		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
// 		assert.NoError(t, err)
// 		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*1, target.Health())
// 		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
// 	})
// 	t.Run("Archer attacks Dragon causing normal damage", func(t *testing.T) {
// 		dmgAmnt := 4
// 		attacker := cards.NewArcher("a1")
// 		target := cards.NewDragon("d1")
// 		weapon := cards.NewArrow("s1", dmgAmnt)

// 		p1 := newPlayerWithCards("Player1",
// 			[]ports.Card{weapon},
// 			[]ports.Warrior{attacker},
// 		)
// 		p2 := newPlayerWithCards("Player2",
// 			[]ports.Card{},
// 			[]ports.Warrior{target},
// 		)
// 		g := &Game{
// 			Players: []ports.Player{p1, p2},
// 		}

// 		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
// 		assert.NoError(t, err)
// 		assert.Equal(t, cards.DragonMaxHealth-dmgAmnt*1, target.Health())
// 		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
// 	})
// 	t.Run("Archer cant attack with wrong weapon", func(t *testing.T) {
// 		dmgAmnt := 4
// 		attacker := cards.NewArcher("a1")
// 		target := cards.NewMage("m1")
// 		weapon := cards.NewSword("s1", dmgAmnt)

// 		p1 := newPlayerWithCards("Player1",
// 			[]ports.Card{weapon},
// 			[]ports.Warrior{attacker},
// 		)
// 		p2 := newPlayerWithCards("Player2",
// 			[]ports.Card{},
// 			[]ports.Warrior{target},
// 		)
// 		g := &Game{
// 			Players: []ports.Player{p1, p2},
// 		}

// 		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
// 		assert.Error(t, err)
// 		assert.Equal(t, cards.WarriorMaxHealth, target.Health())
// 		assert.Contains(t, p1.Hand().ShowCards(), weapon)
// 	})

// 	t.Run("Mage attacks Knight causing double damage", func(t *testing.T) {
// 		dmgAmnt := 4
// 		attacker := cards.NewMage("m1")
// 		target := cards.NewKnight("k1")
// 		weapon := cards.NewPoison("s1", dmgAmnt)

// 		p1 := newPlayerWithCards("Player1",
// 			[]ports.Card{weapon},
// 			[]ports.Warrior{attacker},
// 		)
// 		p2 := newPlayerWithCards("Player2",
// 			[]ports.Card{},
// 			[]ports.Warrior{target},
// 		)
// 		g := &Game{
// 			Players: []ports.Player{p1, p2},
// 		}

// 		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
// 		assert.NoError(t, err)
// 		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*2, target.Health())
// 		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
// 	})
// 	t.Run("Mage attacks Archer causing normal damage", func(t *testing.T) {
// 		dmgAmnt := 4
// 		attacker := cards.NewMage("m1")
// 		target := cards.NewArcher("a1")
// 		weapon := cards.NewPoison("s1", dmgAmnt)

// 		p1 := newPlayerWithCards("Player1",
// 			[]ports.Card{weapon},
// 			[]ports.Warrior{attacker},
// 		)
// 		p2 := newPlayerWithCards("Player2",
// 			[]ports.Card{},
// 			[]ports.Warrior{target},
// 		)
// 		g := &Game{
// 			Players: []ports.Player{p1, p2},
// 		}

// 		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
// 		assert.NoError(t, err)
// 		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*1, target.Health())
// 		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
// 	})
// 	t.Run("Mage attacks Mage causing normal damage", func(t *testing.T) {
// 		dmgAmnt := 4
// 		attacker := cards.NewMage("m1")
// 		target := cards.NewMage("m2")
// 		weapon := cards.NewPoison("s1", dmgAmnt)

// 		p1 := newPlayerWithCards("Player1",
// 			[]ports.Card{weapon},
// 			[]ports.Warrior{attacker},
// 		)
// 		p2 := newPlayerWithCards("Player2",
// 			[]ports.Card{},
// 			[]ports.Warrior{target},
// 		)
// 		g := &Game{
// 			Players: []ports.Player{p1, p2},
// 		}

// 		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
// 		assert.NoError(t, err)
// 		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*1, target.Health())
// 		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
// 	})
// 	t.Run("Mage attacks Dragon causing normal damage", func(t *testing.T) {
// 		dmgAmnt := 4
// 		attacker := cards.NewMage("m1")
// 		target := cards.NewDragon("d1")
// 		weapon := cards.NewPoison("s1", dmgAmnt)

// 		p1 := newPlayerWithCards("Player1",
// 			[]ports.Card{weapon},
// 			[]ports.Warrior{attacker},
// 		)
// 		p2 := newPlayerWithCards("Player2",
// 			[]ports.Card{},
// 			[]ports.Warrior{target},
// 		)
// 		g := &Game{
// 			Players: []ports.Player{p1, p2},
// 		}

// 		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
// 		assert.NoError(t, err)
// 		assert.Equal(t, cards.DragonMaxHealth-dmgAmnt*1, target.Health())
// 		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
// 	})
// 	t.Run("Mage cant attack with wrong weapon", func(t *testing.T) {
// 		dmgAmnt := 4
// 		attacker := cards.NewMage("m1")
// 		target := cards.NewKnight("k1")
// 		weapon := cards.NewArrow("s1", dmgAmnt)

// 		p1 := newPlayerWithCards("Player1",
// 			[]ports.Card{weapon},
// 			[]ports.Warrior{attacker},
// 		)
// 		p2 := newPlayerWithCards("Player2",
// 			[]ports.Card{},
// 			[]ports.Warrior{target},
// 		)
// 		g := &Game{
// 			Players: []ports.Player{p1, p2},
// 		}

// 		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
// 		assert.Error(t, err)
// 		assert.Equal(t, cards.WarriorMaxHealth, target.Health())
// 		assert.Contains(t, p1.Hand().ShowCards(), weapon)
// 	})

// 	t.Run("Player cant attack with non existing cards", func(t *testing.T) {
// 		k := cards.NewKnight("k1")
// 		a := cards.NewArcher("a1")
// 		sword := cards.NewSword("s1", 4)

// 		p1 := newPlayerWithCards("Player1",
// 			[]ports.Card{sword},
// 			[]ports.Warrior{k},
// 		)
// 		p2 := newPlayerWithCards("Player2",
// 			[]ports.Card{},
// 			[]ports.Warrior{a},
// 		)
// 		g := &Game{
// 			Players: []ports.Player{p1, p2},
// 		}

// 		err := g.Attack(p1.Name(), "non-existent-target", sword.GetID())
// 		assert.Error(t, err)

// 		err = g.Attack(p1.Name(), a.GetID(), "non-existent-weapon")
// 		assert.Error(t, err)
// 	})

// 	t.Run("Dragon attacks Knight with Sword causing normal damage", func(t *testing.T) {
// 		dmgAmnt := 6
// 		attacker := cards.NewDragon("d1")
// 		target := cards.NewKnight("k1")
// 		weapon := cards.NewSword("s1", dmgAmnt)

// 		p1 := newPlayerWithCards("Player1",
// 			[]ports.Card{weapon},
// 			[]ports.Warrior{attacker},
// 		)
// 		p2 := newPlayerWithCards("Player2",
// 			[]ports.Card{},
// 			[]ports.Warrior{target},
// 		)
// 		g := &Game{
// 			Players: []ports.Player{p1, p2},
// 		}

// 		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
// 		assert.NoError(t, err)
// 		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*1, target.Health())
// 		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
// 	})
// 	t.Run("Dragon attacks Knight with Arrow causing normal damage", func(t *testing.T) {
// 		dmgAmnt := 6
// 		attacker := cards.NewDragon("d1")
// 		target := cards.NewKnight("k1")
// 		weapon := cards.NewArrow("s1", dmgAmnt)

// 		p1 := newPlayerWithCards("Player1",
// 			[]ports.Card{weapon},
// 			[]ports.Warrior{attacker},
// 		)
// 		p2 := newPlayerWithCards("Player2",
// 			[]ports.Card{},
// 			[]ports.Warrior{target},
// 		)
// 		g := &Game{
// 			Players: []ports.Player{p1, p2},
// 		}

// 		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
// 		assert.NoError(t, err)
// 		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*1, target.Health())
// 		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
// 	})
// 	t.Run("Dragon attacks Knight with Poison causing double damage", func(t *testing.T) {
// 		dmgAmnt := 6
// 		attacker := cards.NewDragon("d1")
// 		target := cards.NewKnight("k1")
// 		weapon := cards.NewPoison("s1", dmgAmnt)

// 		p1 := newPlayerWithCards("Player1",
// 			[]ports.Card{weapon},
// 			[]ports.Warrior{attacker},
// 		)
// 		p2 := newPlayerWithCards("Player2",
// 			[]ports.Card{},
// 			[]ports.Warrior{target},
// 		)
// 		g := &Game{
// 			Players: []ports.Player{p1, p2},
// 		}

// 		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
// 		assert.NoError(t, err)
// 		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*2, target.Health())
// 		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
// 	})
// 	t.Run("Dragon attacks Archer with Sword causing double damage", func(t *testing.T) {
// 		dmgAmnt := 6
// 		attacker := cards.NewDragon("d1")
// 		target := cards.NewArcher("a1")
// 		weapon := cards.NewSword("s1", dmgAmnt)

// 		p1 := newPlayerWithCards("Player1",
// 			[]ports.Card{weapon},
// 			[]ports.Warrior{attacker},
// 		)
// 		p2 := newPlayerWithCards("Player2",
// 			[]ports.Card{},
// 			[]ports.Warrior{target},
// 		)
// 		g := &Game{
// 			Players: []ports.Player{p1, p2},
// 		}

// 		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
// 		assert.NoError(t, err)
// 		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*2, target.Health())
// 		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
// 	})
// 	t.Run("Dragon attacks Archer with Arrow causing normal damage", func(t *testing.T) {
// 		dmgAmnt := 6
// 		attacker := cards.NewDragon("d1")
// 		target := cards.NewArcher("a1")
// 		weapon := cards.NewArrow("s1", dmgAmnt)

// 		p1 := newPlayerWithCards("Player1",
// 			[]ports.Card{weapon},
// 			[]ports.Warrior{attacker},
// 		)
// 		p2 := newPlayerWithCards("Player2",
// 			[]ports.Card{},
// 			[]ports.Warrior{target},
// 		)
// 		g := &Game{
// 			Players: []ports.Player{p1, p2},
// 		}

// 		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
// 		assert.NoError(t, err)
// 		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*1, target.Health())
// 		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
// 	})
// 	t.Run("Dragon attacks Archer with Poison causing normal damage", func(t *testing.T) {
// 		dmgAmnt := 6
// 		attacker := cards.NewDragon("d1")
// 		target := cards.NewArcher("a1")
// 		weapon := cards.NewPoison("s1", dmgAmnt)

// 		p1 := newPlayerWithCards("Player1",
// 			[]ports.Card{weapon},
// 			[]ports.Warrior{attacker},
// 		)
// 		p2 := newPlayerWithCards("Player2",
// 			[]ports.Card{},
// 			[]ports.Warrior{target},
// 		)
// 		g := &Game{
// 			Players: []ports.Player{p1, p2},
// 		}

// 		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
// 		assert.NoError(t, err)
// 		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*1, target.Health())
// 		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
// 	})
// 	t.Run("Dragon attacks Mage with Sword causing normal damage", func(t *testing.T) {
// 		dmgAmnt := 6
// 		attacker := cards.NewDragon("d1")
// 		target := cards.NewMage("m1")
// 		weapon := cards.NewSword("s1", dmgAmnt)

// 		p1 := newPlayerWithCards("Player1",
// 			[]ports.Card{weapon},
// 			[]ports.Warrior{attacker},
// 		)
// 		p2 := newPlayerWithCards("Player2",
// 			[]ports.Card{},
// 			[]ports.Warrior{target},
// 		)
// 		g := &Game{
// 			Players: []ports.Player{p1, p2},
// 		}

// 		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
// 		assert.NoError(t, err)
// 		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*1, target.Health())
// 		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
// 	})
// 	t.Run("Dragon attacks Mage with Arrow causing double damage", func(t *testing.T) {
// 		dmgAmnt := 6
// 		attacker := cards.NewDragon("d1")
// 		target := cards.NewMage("m1")
// 		weapon := cards.NewArrow("s1", dmgAmnt)

// 		p1 := newPlayerWithCards("Player1",
// 			[]ports.Card{weapon},
// 			[]ports.Warrior{attacker},
// 		)
// 		p2 := newPlayerWithCards("Player2",
// 			[]ports.Card{},
// 			[]ports.Warrior{target},
// 		)
// 		g := &Game{
// 			Players: []ports.Player{p1, p2},
// 		}

// 		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
// 		assert.NoError(t, err)
// 		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*2, target.Health())
// 		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
// 	})
// 	t.Run("Dragon attacks Mage with Poison causing normal damage", func(t *testing.T) {
// 		dmgAmnt := 6
// 		attacker := cards.NewDragon("d1")
// 		target := cards.NewMage("m1")
// 		weapon := cards.NewPoison("s1", dmgAmnt)

// 		p1 := newPlayerWithCards("Player1",
// 			[]ports.Card{weapon},
// 			[]ports.Warrior{attacker},
// 		)
// 		p2 := newPlayerWithCards("Player2",
// 			[]ports.Card{},
// 			[]ports.Warrior{target},
// 		)
// 		g := &Game{
// 			Players: []ports.Player{p1, p2},
// 		}

// 		err := g.Attack(p1.Name(), target.GetID(), weapon.GetID())
// 		assert.NoError(t, err)
// 		assert.Equal(t, cards.WarriorMaxHealth-dmgAmnt*1, target.Health())
// 		assert.NotContains(t, p1.Hand().ShowCards(), weapon)
// 	})
// 	t.Run("Warrior dead on second attack", func(t *testing.T) {
// 		dmgAmnt := 5
// 		k := cards.NewKnight("k1")
// 		a := cards.NewArcher("a1")
// 		a2 := cards.NewArcher("a2")
// 		sword1 := cards.NewSword("s1", dmgAmnt)
// 		sword2 := cards.NewSword("s2", dmgAmnt)
// 		g := &Game{}
// 		p1 := newPlayerWithCardAndObserver("Player1",
// 			[]ports.Card{sword1, sword2},
// 			[]ports.Warrior{k},
// 			g,
// 		)
// 		p2 := newPlayerWithCardAndObserver("Player2",
// 			[]ports.Card{},
// 			[]ports.Warrior{a, a2},
// 			g,
// 		)

// 		g.Players = []ports.Player{p1, p2}

// 		err := g.Attack(p1.Name(), a.GetID(), sword1.GetID())
// 		assert.NoError(t, err)
// 		assert.NotContains(t, p1.Hand().ShowCards(), sword1)

// 		err = g.Attack(p1.Name(), a.GetID(), sword2.GetID())
// 		assert.NoError(t, err)
// 		assert.NotContains(t, p1.Hand().ShowCards(), sword2)

// 		assert.Equal(t, 0, a.Health())
// 		_, ok := p2.GetCardFromField(a.GetID())
// 		assert.False(t, ok, "Archer should have been removed from field after death")
// 		_, ok = p2.GetCardFromField(a2.GetID())
// 		assert.True(t, ok, "Second Archer should still be on the field")
// 		assert.True(t, foundInCemetery(g, a), "Cemetery should contain the dead archer")
// 		assert.True(t, foundInDiscardPile(g, sword1), "Discard pile should contain the used sword")
// 		assert.True(t, foundInDiscardPile(g, sword2), "Discard pile should contain the used sword")
// 	})
// 	t.Run("Dragon dead on multiple attacks", func(t *testing.T) {
// 		dmgAmnt := 5
// 		m1 := cards.NewMage("m1")
// 		k2 := cards.NewKnight("k2")
// 		a3 := cards.NewArcher("a3")

// 		target := cards.NewDragon("d1")
// 		a2 := cards.NewArcher("a2")

// 		poison1 := cards.NewPoison("p1", dmgAmnt)
// 		sword2 := cards.NewSword("s2", dmgAmnt)
// 		arrow3 := cards.NewArrow("a3", dmgAmnt)
// 		sword4 := cards.NewSword("s4", dmgAmnt)

// 		g := &Game{}
// 		p1 := newPlayerWithCardAndObserver("Player1",
// 			[]ports.Card{poison1, sword2, arrow3, sword4},
// 			[]ports.Warrior{m1, k2, a3},
// 			g,
// 		)
// 		p2 := newPlayerWithCardAndObserver("Player2",
// 			[]ports.Card{},
// 			[]ports.Warrior{target, a2},
// 			g,
// 		)

// 		g.Players = []ports.Player{p1, p2}

// 		err := g.Attack(p1.Name(), target.GetID(), poison1.GetID())
// 		assert.NoError(t, err)
// 		err = g.Attack(p1.Name(), target.GetID(), sword2.GetID())
// 		assert.NoError(t, err)
// 		err = g.Attack(p1.Name(), target.GetID(), arrow3.GetID())
// 		assert.NoError(t, err)
// 		err = g.Attack(p1.Name(), target.GetID(), sword4.GetID())
// 		assert.NoError(t, err)

// 		assert.Equal(t, 0, target.Health())
// 		_, ok := p1.GetCardFromField(m1.GetID())
// 		assert.True(t, ok, "Mage should still be on the field")
// 		_, ok = p1.GetCardFromField(k2.GetID())
// 		assert.True(t, ok, "Knight should still be on the field")
// 		_, ok = p1.GetCardFromField(a3.GetID())
// 		assert.True(t, ok, "Archer should still be on the field")

// 		_, ok = p2.GetCardFromField(target.GetID())
// 		assert.False(t, ok, "Dragon should have been removed from field after death")
// 		_, ok = p2.GetCardFromField(a2.GetID())
// 		assert.True(t, ok, "Archer should still be on the field")

// 		_, ok = p1.GetCardFromHand(poison1.GetID())
// 		assert.False(t, ok, "Poison should have been discarded after attack")
// 		_, ok = p1.GetCardFromHand(sword2.GetID())
// 		assert.False(t, ok, "Sword should have been discarded after attack")
// 		_, ok = p1.GetCardFromHand(arrow3.GetID())
// 		assert.False(t, ok, "Arrow should have been discarded after attack")
// 		_, ok = p1.GetCardFromHand(sword4.GetID())
// 		assert.False(t, ok, "Sword should have been discarded after attack")

// 		assert.True(t, foundInCemetery(g, target), "Cemetery should contain the dead dragon")
// 		assert.True(t, foundInDiscardPile(g, poison1), "Discard pile should contain the used poison")
// 		assert.True(t, foundInDiscardPile(g, sword2), "Discard pile should contain the used sword")
// 		assert.True(t, foundInDiscardPile(g, arrow3), "Discard pile should contain the used arrow")
// 		assert.True(t, foundInDiscardPile(g, sword4), "Discard pile should contain the used sword")
// 	})
// }

// func TestGame_SpecialPower(t *testing.T) {
// 	t.Run("Use special power of Archer (Instant Kill) on warrior", func(t *testing.T) {
// 		a := cards.NewArcher("a1")
// 		target := cards.NewArcher("a2")
// 		sp := cards.NewSpecialPower("sp")
// 		g := &Game{}
// 		p1 := newPlayerWithCardAndObserver("Player1",
// 			[]ports.Card{sp},
// 			[]ports.Warrior{a},
// 			g,
// 		)
// 		p2 := newPlayerWithCardAndObserver("Player2",
// 			[]ports.Card{},
// 			[]ports.Warrior{target},
// 			g,
// 		)

// 		g.Players = []ports.Player{p1, p2}

// 		err := g.SpecialPower(p1.Name(), a.GetID(), target.GetID(), sp.GetID())
// 		assert.NoError(t, err)

// 		assert.Equal(t, 0, target.Health())
// 		_, ok := p2.GetCardFromField(target.GetID())
// 		assert.False(t, ok, "Target should have been removed from field after death")
// 		_, ok = p1.GetCardFromHand(sp.GetID())
// 		assert.False(t, ok, "Special Power should have been discarded after attack")
// 		assert.True(t, foundInCemetery(g, target), "Cemetery should contain the dead target")
// 		assert.True(t, foundInDiscardPile(g, sp), "Discard pile should contain the used special power")
// 	})
// 	t.Run("Use special power of Archer (Instant Kill) on dragon", func(t *testing.T) {
// 		a := cards.NewArcher("a1")
// 		target := cards.NewDragon("dr")
// 		sp := cards.NewSpecialPower("sp")
// 		g := &Game{}
// 		p1 := newPlayerWithCardAndObserver("Player1",
// 			[]ports.Card{sp},
// 			[]ports.Warrior{a},
// 			g,
// 		)
// 		p2 := newPlayerWithCardAndObserver("Player2",
// 			[]ports.Card{},
// 			[]ports.Warrior{target},
// 			g,
// 		)

// 		g.Players = []ports.Player{p1, p2}

// 		err := g.SpecialPower(p1.Name(), a.GetID(), target.GetID(), sp.GetID())
// 		assert.NoError(t, err)

// 		assert.Equal(t, cards.DragonMaxHealth-cards.SpecialPowerDamage, target.Health())
// 		_, ok := p2.GetCardFromField(target.GetID())
// 		assert.True(t, ok, "Dragon should still be on the field")

// 		assert.True(t, findInAttackedBy(target.AttackedBy(), sp.GetID()), "Target should have been marked as attacked by special power")
// 	})
// 	t.Run("Use special power of Mage (Heal) on warrior", func(t *testing.T) {
// 		m := cards.NewMage("m1")
// 		target := cards.NewKnight("k1")
// 		attacker := cards.NewArcher("a1")
// 		arrow := cards.NewArrow("s1", 4)

// 		sp := cards.NewSpecialPower("sp")
// 		g := &Game{}
// 		p1 := newPlayerWithCardAndObserver("Player1",
// 			[]ports.Card{arrow},
// 			[]ports.Warrior{attacker},
// 			g,
// 		)
// 		p2 := newPlayerWithCardAndObserver("Player2",
// 			[]ports.Card{sp},
// 			[]ports.Warrior{m, target},
// 			g,
// 		)

// 		g.Players = []ports.Player{p1, p2}

// 		_ = g.Attack(p1.Name(), target.GetID(), arrow.GetID())
// 		assert.Equal(t, cards.WarriorMaxHealth-4, target.Health())
// 		err := g.EndTurn(p1.Name())
// 		assert.NoError(t, err)
// 		err = g.SpecialPower(p2.Name(), m.GetID(), target.GetID(), sp.GetID())
// 		assert.NoError(t, err)

// 		assert.Equal(t, cards.WarriorMaxHealth, target.Health())
// 		_, ok := p2.GetCardFromHand(sp.GetID())
// 		assert.False(t, ok, "Special Power should have been discarded after use")
// 		assert.True(t, foundInDiscardPile(g, sp), "Discard pile should contain the used special power")

// 	})
// 	t.Run("Use special power of Knight (Protection) on warrior", func(t *testing.T) {
// 		user := cards.NewKnight("k1")
// 		target := cards.NewKnight("k2")
// 		attacker := cards.NewArcher("a1")
// 		arrow := cards.NewArrow("a1", 4)
// 		arrow2 := cards.NewArrow("a2", 8)

// 		sp := cards.NewSpecialPower("sp")
// 		g := &Game{}
// 		p1 := newPlayerWithCardAndObserver("Player1",
// 			[]ports.Card{sp},
// 			[]ports.Warrior{user, target},
// 			g,
// 		)
// 		p2 := newPlayerWithCardAndObserver("Player2",
// 			[]ports.Card{arrow, arrow2},
// 			[]ports.Warrior{attacker},
// 			g,
// 		)

// 		g.Players = []ports.Player{p1, p2}

// 		err := g.SpecialPower(p1.Name(), user.GetID(), target.GetID(), sp.GetID())
// 		assert.NoError(t, err)
// 		assert.NotContains(t, p1.Hand().ShowCards(), sp)
// 		isProtected, card := target.IsProtected()
// 		assert.True(t, isProtected)
// 		assert.Equal(t, card, sp)
// 		_ = g.EndTurn(p1.Name())

// 		_ = g.Attack(p2.Name(), target.GetID(), arrow.GetID())
// 		assert.Equal(t, cards.SpecialPowerMaxHealth-4, sp.Health())
// 		assert.Equal(t, cards.WarriorMaxHealth, target.Health())

// 		_ = g.Attack(p2.Name(), target.GetID(), arrow2.GetID())
// 		assert.Equal(t, cards.SpecialPowerMaxHealth-4-8, sp.Health())
// 		assert.Equal(t, cards.WarriorMaxHealth, target.Health())

// 		assert.True(t, foundInDiscardPile(g, sp), "Discard pile should contain the used special power")

// 	})
// }

// func TestDrawCards(t *testing.T) {
// 	t.Run("Take card when deck is empty", func(t *testing.T) {
// 		p := newPlayerWithCards("Player1", []ports.Card{}, []ports.Warrior{})
// 		g := &Game{
// 			Players: []ports.Player{p},
// 			deck:    NewDeck([]ports.Card{}),
// 			discardPile: []ports.Card{
// 				cards.NewSword("s1", 4),
// 				cards.NewArrow("a1", 3),
// 				cards.NewPoison("p1", 4),
// 			},
// 			cemetery: []ports.Warrior{},
// 		}

// 		err := g.DrawCards(p.Name(), 1)
// 		assert.NoError(t, err)
// 		assert.Equal(t, 1, p.CardsInHand(), "Player should have drawn one card from reshuffled deck")
// 		assert.Equal(t, 2, len(g.deck.(*deck).cards), "Deck should have two cards remaining after drawing one")
// 		assert.Equal(t, 0, len(g.discardPile), "Discard pile should be empty after reshuffling into deck")
// 	})
// 	t.Run("Take card from deck to hand", func(t *testing.T) {
// 		p := newPlayerWithCards("Player1",
// 			[]ports.Card{cards.NewGold("g1", 5)},
// 			[]ports.Warrior{})
// 		g := &Game{
// 			Players: []ports.Player{p},
// 			deck: NewDeck([]ports.Card{
// 				cards.NewSword("s1", 4),
// 				cards.NewArrow("a1", 3),
// 				cards.NewPoison("p1", 4),
// 			}),
// 			discardPile: []ports.Card{},
// 			cemetery:    []ports.Warrior{},
// 		}

// 		err := g.DrawCards(p.Name(), 1)
// 		assert.NoError(t, err)
// 		assert.Equal(t, 2, p.CardsInHand(), "Player should have drawn two cards from deck")
// 		assert.Equal(t, 2, len(g.deck.(*deck).cards), "Deck should have one card remaining after drawing two")
// 	})
// }

// func TestNewGame(t *testing.T) {
// 	t.Run("Create new game with two players getting expected number of cards", func(t *testing.T) {
// 		p1 := "Alice"
// 		p2 := "Bob"
// 		g := NewGame(p1, p2, cards.NewDealer())

// 		assert.Equal(t, 2, len(g.Players), "Game should have two players")
// 		assert.Equal(t, 7, g.Players[0].CardsInHand(), "Each player should start with 7 cards in hand")
// 		assert.Equal(t, 7, g.Players[1].CardsInHand(), "Each player should start with 7 cards in hand")
// 		assert.Equal(t, 46, len(g.deck.(*deck).cards), "Deck should start with 40 cards")
// 		assert.Equal(t, g.state, StateSettingInitialWarriors)
// 	})
// 	t.Run("Set initial warriors for players", func(t *testing.T) {
// 		p1 := "Alice"
// 		p2 := "Bob"
// 		g := NewGame(p1, p2, cards.NewDealer())

// 		current, _ := g.WhoIsCurrent()
// 		cont := 0
// 		var warriors1 []string
// 		for _, card := range current.Hand().ShowCards() {
// 			if _, ok := card.(ports.Warrior); ok {
// 				cont++
// 				warriors1 = append(warriors1, card.GetID())
// 				if cont == 3 {
// 					break
// 				}
// 			}
// 		}

// 		err := g.SetInitialWarriors(current.Name(), warriors1)
// 		assert.NoError(t, err)
// 		assert.Equal(t, len(current.Field().Warriors()), len(warriors1))
// 		assert.True(t, fieldContainsCardWithID(current.Field(), warriors1[0]), "Field should contain the warrior with the given ID")
// 		assert.True(t, fieldContainsCardWithID(current.Field(), warriors1[1]), "Field should contain the warrior with the given ID")
// 		assert.True(t, fieldContainsCardWithID(current.Field(), warriors1[2]), "Field should contain the warrior with the given ID")
// 		assert.False(t, handContainsCardWithID(current.Hand(), warriors1[0]), "Hand should not contain the warrior with the given ID")
// 		assert.False(t, handContainsCardWithID(current.Hand(), warriors1[1]), "Hand should not contain the warrior with the given ID")
// 		assert.False(t, handContainsCardWithID(current.Hand(), warriors1[2]), "Hand should not contain the warrior with the given ID")
// 		assert.Equal(t, 4, current.CardsInHand(), "Player should have 4 cards left in hand after setting 3 warriors")

// 		current, _ = g.WhoIsCurrent()
// 		cont = 0
// 		var warriors2 []string
// 		for _, card := range current.Hand().ShowCards() {
// 			if _, ok := card.(ports.Warrior); ok {
// 				cont++
// 				warriors2 = append(warriors2, card.GetID())
// 				if cont == 2 {
// 					break
// 				}
// 			}
// 		}

// 		err = g.SetInitialWarriors(current.Name(), warriors2)
// 		assert.NoError(t, err)
// 		assert.Equal(t, len(current.Field().Warriors()), len(warriors2))
// 		assert.True(t, fieldContainsCardWithID(current.Field(), warriors2[0]), "Field should contain the warrior with the given ID")
// 		assert.True(t, fieldContainsCardWithID(current.Field(), warriors2[1]), "Field should contain the warrior with the given ID")
// 		assert.False(t, handContainsCardWithID(current.Hand(), warriors2[0]), "Hand should not contain the warrior with the given ID")
// 		assert.False(t, handContainsCardWithID(current.Hand(), warriors2[1]), "Hand should not contain the warrior with the given ID")
// 		assert.Equal(t, 5, current.CardsInHand(), "Player should have 5 cards left in hand after setting 2 warriors")
// 		assert.Equal(t, StateWaitingDraw, g.state)
// 	})
// }

// func TestGame_Spy(t *testing.T) {
// 	t.Run("Spy reveals top cards of deck", func(t *testing.T) {
// 		spy := cards.NewSpy("s1")

// 		p1 := newPlayerWithCards("Player1",
// 			[]ports.Card{spy},
// 			[]ports.Warrior{},
// 		)
// 		p2 := newPlayerWithCards("Player2", nil, nil)
// 		deckCards := []ports.Card{
// 			cards.NewGold("g1", 5),
// 			cards.NewSword("sw1", 4),
// 			cards.NewArrow("a1", 3),
// 			cards.NewArrow("a2", 5),
// 			cards.NewDragon("d1"),
// 		}
// 		g := &Game{
// 			Players: []ports.Player{p1, p2},
// 			deck:    NewDeck(deckCards),
// 		}

// 		revealedCards, err := g.Spy(p1.Name(), 1)
// 		assert.NoError(t, err)
// 		assert.Len(t, revealedCards, 5, "Spy should reveal five cards from the top of the deck")
// 		assert.Contains(t, revealedCards, deckCards[0], "Revealed cards should contain the first top card of the deck")
// 		assert.Contains(t, revealedCards, deckCards[1], "Revealed cards should contain the second top card of the deck")
// 		assert.Contains(t, revealedCards, deckCards[2], "Revealed cards should contain the third top card of the deck")
// 		assert.Contains(t, revealedCards, deckCards[3], "Revealed cards should contain the fourth top card of the deck")
// 		assert.Contains(t, revealedCards, deckCards[4], "Revealed cards should contain the fifth top card of the deck")
// 		_, hasSpy := p1.Hand().GetCard(spy.GetID())
// 		assert.False(t, hasSpy)
// 	})
// 	t.Run("Spy reveals opponent's hand", func(t *testing.T) {
// 		spy := cards.NewSpy("s1")
// 		enemyCard1 := cards.NewGold("g1", 5)
// 		enemyCard2 := cards.NewSword("sw1", 4)

// 		p1 := newPlayerWithCards("Player1",
// 			[]ports.Card{spy},
// 			[]ports.Warrior{},
// 		)
// 		p2 := newPlayerWithCards("Player2",
// 			[]ports.Card{enemyCard1, enemyCard2},
// 			[]ports.Warrior{},
// 		)
// 		g := &Game{
// 			Players: []ports.Player{p1, p2},
// 		}

// 		revealedCards, err := g.Spy(p1.Name(), 2)
// 		assert.NoError(t, err)
// 		assert.Len(t, revealedCards, 2, "Spy should reveal two cards from opponent's hand")
// 		assert.Contains(t, revealedCards, enemyCard1, "Revealed cards should contain the first enemy card")
// 		assert.Contains(t, revealedCards, enemyCard2, "Revealed cards should contain the second enemy card")
// 		_, hasSpy := p1.Hand().GetCard(spy.GetID())
// 		assert.False(t, hasSpy)
// 	})
// }

// func TestGame_Steal(t *testing.T) {
// 	t.Run("Steal a card from opponent's hand", func(t *testing.T) {
// 		stealCard := cards.NewThief("t1")
// 		enemyCard1 := cards.NewGold("g1", 5)
// 		enemyCard2 := cards.NewSword("sw1", 4)
// 		enemyCard3 := cards.NewSword("sw2", 7)

// 		p1 := newPlayerWithCards("Player1",
// 			[]ports.Card{stealCard},
// 			[]ports.Warrior{},
// 		)
// 		p2 := newPlayerWithCards("Player2",
// 			[]ports.Card{enemyCard1, enemyCard2, enemyCard3},
// 			[]ports.Warrior{},
// 		)
// 		g := &Game{
// 			Players: []ports.Player{p1, p2},
// 		}

// 		err := g.Steal(p1.Name(), 2)
// 		assert.NoError(t, err)

// 		_, hasSteal := p1.Hand().GetCard(stealCard.GetID())
// 		assert.False(t, hasSteal, "Steal card should be discarded after use")
// 		assert.Len(t, p1.Hand().ShowCards(), 1, "Player should have one more card in hand after stealing")
// 		assert.Len(t, p2.Hand().ShowCards(), 2, "Enemy should have one less card in hand after being stolen from")
// 	})
// }

func findInAttackedBy(cards []ports.Weapon, id string) bool {
	for _, c := range cards {
		if c != nil && c.GetID() == id {
			return true
		}
	}
	return false
}

// func foundInCemetery(g *Game, a ports.Warrior) bool {
// 	for _, w := range g.cemetery {
// 		if w == a || (w != nil && w.GetID() == a.GetID()) {
// 			return true
// 		}
// 	}
// 	return false
// }

// func foundInDiscardPile(g *Game, a ports.Card) bool {
// 	// discardPile is now an interface, cannot range over it
// 	return false
// }

func fieldContainsCardWithID(field ports.Field, id string) bool {
	for _, c := range field.Warriors() {
		if c != nil && c.GetID() == id {
			return true
		}
	}
	return false
}

func handContainsCardWithID(hand ports.Hand, id string) bool {
	for _, c := range hand.ShowCards() {
		if c != nil && c.GetID() == id {
			return true
		}
	}
	return false
}

func newPlayerWithCards(name string, cardsInHand []ports.Card,
	cardsInField []ports.Warrior) ports.Player {
	return newPlayerWithCardAndObserver(name, cardsInHand, cardsInField, nil)
}

func newPlayerWithCardAndObserver(name string, cardsInHand []ports.Card,
	cardsInField []ports.Warrior, game *Game) ports.Player {
	p := &player{
		name:                           name,
		cardMovedToPileObserver:        game,
		warriorMovedToCemeteryObserver: game,
		hand: &hand{
			cards: cardsInHand,
		},
		field: &field{
			cards: cardsInField,
		},
		castle: &castle{},
	}

	for _, card := range cardsInField {
		card.AddCardMovedToPileObserver(p)
		targ, ok := card.(ports.Warrior)
		if ok {
			targ.AddWarriorDeadObserver(p)
		}
	}
	for _, card := range cardsInHand {
		card.AddCardMovedToPileObserver(p)
	}

	return p
}

func TestGame_EndTurn(t *testing.T) {
	t.Run("Error when not current player's turn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:     []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn: 0, // Player1's turn
		}

		status, err := g.EndTurn("Player2") // Player2 tries to end turn

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Player2 not your turn")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Success ending turn switches to next player", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		expectedStatus := GameStatus{
			CurrentPlayer: "Player2",
		}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		// nextAction expectations for ActionTypeDrawCard -> checks capabilities
		mockPlayer2.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer2.EXPECT().CanTradeCards().Return(false)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0, // Player1's turn
			currentAction:      types.ActionTypeEndTurn,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
			hasMovedWarrior:    true, // These should be reset
			hasTraded:          true,
		}

		mockProvider.EXPECT().Get(mockPlayer2, g).Return(expectedStatus)

		status, err := g.EndTurn("Player1")

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
		assert.Equal(t, 1, g.CurrentTurn) // Turn switched to Player2
		assert.False(t, g.hasMovedWarrior)
		assert.False(t, g.hasTraded)
		assert.Equal(t, types.ActionTypeDrawCard, g.currentAction)
	})

	t.Run("Turn wraps around from last player to first", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		expectedStatus := GameStatus{
			CurrentPlayer: "Player1",
		}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		// nextAction expectations
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        1, // Player2's turn
			currentAction:      types.ActionTypeEndTurn,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
		}

		mockProvider.EXPECT().Get(mockPlayer1, g).Return(expectedStatus)

		status, err := g.EndTurn("Player2")

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
		assert.Equal(t, 0, g.CurrentTurn) // Wrapped back to Player1
	})

	t.Run("History is updated when turn ends", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer2.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer2.EXPECT().CanTradeCards().Return(false)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
			history:            []historyLine{},
		}

		mockProvider.EXPECT().Get(gomock.Any(), gomock.Any()).Return(GameStatus{})

		_, err := g.EndTurn("Player1")

		assert.NoError(t, err)
		found := false
		for _, h := range g.history {
			if strings.Contains(h.Msg,"Player1") && strings.Contains(h.Msg,"turn ended") {
				found = true
				break
			}
		}
		assert.True(t, found, "History should contain turn change message")
	})
}

func TestGame_Attack(t *testing.T) {
	t.Run("Error when not current player's turn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeAttack,
		}

		status, err := g.Attack("Player2", "Player1", "targetID", "weaponID")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Player2 not your turn")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when not in Attack phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeBuy, // Not Attack phase
		}

		status, err := g.Attack("Player1", "Player2", "targetID", "weaponID")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot attack in the")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when target card not in enemy field", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer2.EXPECT().GetCardFromField("targetID").Return(nil, false)

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeAttack,
		}

		status, err := g.Attack("Player1", "Player2", "targetID", "weaponID")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target card not in enemy field")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when weapon card not in hand", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer2.EXPECT().GetCardFromField("targetID").Return(mockWarrior, true)
		mockPlayer1.EXPECT().GetCardFromHand("weaponID").Return(nil, false)

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeAttack,
		}

		status, err := g.Attack("Player1", "Player2", "targetID", "weaponID")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "weapon card not in hand")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when card is not a weapon", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockResource := mocks.NewMockResource(ctrl) // Not a weapon

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer2.EXPECT().GetCardFromField("targetID").Return(mockWarrior, true)
		mockPlayer1.EXPECT().GetCardFromHand("resourceID").Return(mockResource, true)

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeAttack,
		}

		status, err := g.Attack("Player1", "Player2", "targetID", "resourceID")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "the card is not a weapon")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when attack action fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockWeapon := mocks.NewMockWeapon(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer2.EXPECT().GetCardFromField("targetID").Return(mockWarrior, true)
		mockPlayer1.EXPECT().GetCardFromHand("weaponID").Return(mockWeapon, true)
		mockPlayer1.EXPECT().Attack(mockWarrior, mockWeapon).Return(errors.New("attack failed"))

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeAttack,
		}

		status, err := g.Attack("Player1", "Player2", "targetID", "weaponID")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "attack action failed")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Success attacking enemy warrior", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockWeapon := mocks.NewMockWeapon(ctrl)

		expectedStatus := GameStatus{
			CurrentPlayer: "Player1",
		}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer2.EXPECT().GetCardFromField("K1").Return(mockWarrior, true)
		mockPlayer1.EXPECT().GetCardFromHand("S1").Return(mockWeapon, true)
		mockPlayer1.EXPECT().Attack(mockWarrior, mockWeapon).Return(nil)
		mockWarrior.EXPECT().String().Return("Knight (20)")
		mockWeapon.EXPECT().String().Return("Sword (5)")

		// nextAction expectations for ActionTypeSpySteal
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)
		mockPlayer1.EXPECT().HasSpy().Return(false)
		mockPlayer1.EXPECT().HasThief().Return(false)
		mockPlayer1.EXPECT().CanBuy().Return(true)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.ActionTypeAttack,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
			history:            []historyLine{},
		}

		mockProvider.EXPECT().Get(mockPlayer1, g).Return(expectedStatus)

		status, err := g.Attack("Player1", "Player2", "K1", "S1")

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
		assert.Equal(t, types.ActionTypeBuy, g.currentAction) // Moved to Buy phase
	})

	t.Run("History is updated on successful attack", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockWeapon := mocks.NewMockWeapon(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer2.EXPECT().GetCardFromField("K1").Return(mockWarrior, true)
		mockPlayer1.EXPECT().GetCardFromHand("S1").Return(mockWeapon, true)
		mockPlayer1.EXPECT().Attack(mockWarrior, mockWeapon).Return(nil)
		mockWarrior.EXPECT().String().Return("Knight (20)")
		mockWeapon.EXPECT().String().Return("Sword (5)")

		// nextAction expectations
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)
		mockPlayer1.EXPECT().HasSpy().Return(false)
		mockPlayer1.EXPECT().HasThief().Return(false)
		mockPlayer1.EXPECT().CanBuy().Return(true)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.ActionTypeAttack,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
			history:            []historyLine{},
		}

		mockProvider.EXPECT().Get(gomock.Any(), gomock.Any()).Return(GameStatus{})

		_, err := g.Attack("Player1", "Player2", "K1", "S1")

		assert.NoError(t, err)
		found := false
		for _, h := range g.history {
			if strings.Contains(h.Msg,"Player1") && strings.Contains(h.Msg,"attacked") {
				found = true
				break
			}
		}
		assert.True(t, found, "History should contain attack action")
	})
}

func TestGame_SkipPhase(t *testing.T) {
	t.Run("Error when not current player's turn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeAttack,
		}

		status, err := g.SkipPhase("Player2")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Player2 not your turn")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when trying to skip DrawCard phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeDrawCard,
		}

		status, err := g.SkipPhase("Player1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot skip this phase")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when trying to skip EndTurn phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeEndTurn,
		}

		status, err := g.SkipPhase("Player1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot skip this phase")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Success skipping Attack phase moves to SpySteal", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		expectedStatus := GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		// nextAction expectations for ActionTypeSpySteal
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)
		mockPlayer1.EXPECT().HasSpy().Return(true) // Has spy, stays in SpySteal

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.ActionTypeAttack,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
			history:            []historyLine{},
		}

		mockProvider.EXPECT().Get(mockPlayer1, g).Return(expectedStatus)

		status, err := g.SkipPhase("Player1")

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
		assert.Equal(t, types.ActionTypeSpySteal, g.currentAction)
	})

	t.Run("Success skipping SpySteal phase moves to Buy", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		expectedStatus := GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		// nextAction expectations for ActionTypeBuy
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)
		mockPlayer1.EXPECT().CanBuy().Return(true) // Can buy, stays in Buy

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.ActionTypeSpySteal,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
			history:            []historyLine{},
		}

		mockProvider.EXPECT().Get(mockPlayer1, g).Return(expectedStatus)

		status, err := g.SkipPhase("Player1")

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
		assert.Equal(t, types.ActionTypeBuy, g.currentAction)
	})

	t.Run("Success skipping Buy phase moves to Construct", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		expectedStatus := GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		// nextAction expectations for ActionTypeConstruct
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)
		mockPlayer1.EXPECT().CanConstruct().Return(true) // Can construct

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.ActionTypeBuy,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
			history:            []historyLine{},
		}

		mockProvider.EXPECT().Get(mockPlayer1, g).Return(expectedStatus)

		status, err := g.SkipPhase("Player1")

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
		assert.Equal(t, types.ActionTypeConstruct, g.currentAction)
	})

	t.Run("Success skipping Construct phase moves to EndTurn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		expectedStatus := GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		// nextAction expectations for ActionTypeEndTurn
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.ActionTypeConstruct,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
			history:            []historyLine{},
		}

		mockProvider.EXPECT().Get(mockPlayer1, g).Return(expectedStatus)

		status, err := g.SkipPhase("Player1")

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
		assert.Equal(t, types.ActionTypeEndTurn, g.currentAction)
	})

	t.Run("History is updated when phase is skipped", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)
		mockPlayer1.EXPECT().HasSpy().Return(false)
		mockPlayer1.EXPECT().HasThief().Return(false)
		mockPlayer1.EXPECT().CanBuy().Return(true)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.ActionTypeAttack,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
			history:            []historyLine{},
		}

		mockProvider.EXPECT().Get(gomock.Any(), gomock.Any()).Return(GameStatus{})

		_, err := g.SkipPhase("Player1")

		assert.NoError(t, err)
		found := false
		for _, h := range g.history {
			if strings.Contains(h.Msg,"Player1") && strings.Contains(h.Msg,"skipped") {
				found = true
				break
			}
		}
		assert.True(t, found, "History should contain skip phase action")
	})
}

func TestGame_MoveWarriorToField(t *testing.T) {
	t.Run("Error when not current player's turn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:     []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn: 0,
		}

		status, err := g.MoveWarriorToField("Player2", "K1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Player2 not your turn")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when MoveCardToField fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().MoveCardToField("K1").Return(errors.New("card not found"))

		g := &Game{
			Players:     []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn: 0,
		}

		status, err := g.MoveWarriorToField("Player1", "K1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "moving warrior to field failed")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Success moving warrior to field", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		expectedStatus := GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().MoveCardToField("K1").Return(nil)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
			history:            []historyLine{},
		}

		mockProvider.EXPECT().Get(mockPlayer1, g).Return(expectedStatus)

		status, err := g.MoveWarriorToField("Player1", "K1")

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
		assert.True(t, g.hasMovedWarrior)
	})

	t.Run("History is updated on success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().MoveCardToField("K1").Return(nil)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
			history:            []historyLine{},
		}

		mockProvider.EXPECT().Get(gomock.Any(), gomock.Any()).Return(GameStatus{})

		_, err := g.MoveWarriorToField("Player1", "K1")

		assert.NoError(t, err)
		found := false
		for _, h := range g.history {
			if strings.Contains(h.Msg,"Player1") && strings.Contains(h.Msg,"moved warrior") {
				found = true
				break
			}
		}
		assert.True(t, found, "History should contain move warrior action")
	})

	t.Run("Player 2 can move warrior on their turn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		expectedStatus := GameStatus{CurrentPlayer: "Player2"}

		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer2.EXPECT().MoveCardToField("A1").Return(nil)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        1,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
			history:            []historyLine{},
		}

		mockProvider.EXPECT().Get(mockPlayer2, g).Return(expectedStatus)

		status, err := g.MoveWarriorToField("Player2", "A1")

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
		assert.True(t, g.hasMovedWarrior)
	})
}

func TestGame_Trade(t *testing.T) {
	t.Run("Error when not current player's turn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:     []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn: 0,
		}

		status, err := g.Trade("Player2", []string{"C1", "C2", "C3"})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Player2 not your turn")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when already traded this turn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:     []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn: 0,
			hasTraded:   true,
		}

		status, err := g.Trade("Player1", []string{"C1", "C2", "C3"})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already traded this turn")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when not exactly 3 cards", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:     []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn: 0,
		}

		status, err := g.Trade("Player1", []string{"C1", "C2"})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must trade exactly 3 cards")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when GiveCards fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GiveCards("C1", "C2", "C3").Return(nil, errors.New("card not found"))

		g := &Game{
			Players:     []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn: 0,
		}

		status, err := g.Trade("Player1", []string{"C1", "C2", "C3"})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "giving cards for trading failed")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Success trading 3 cards for 1", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDeck := mocks.NewMockDeck(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)
		mockCard1 := mocks.NewMockCard(ctrl)
		mockCard2 := mocks.NewMockCard(ctrl)
		mockCard3 := mocks.NewMockCard(ctrl)
		mockDrawnCard := mocks.NewMockCard(ctrl)

		expectedStatus := GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GiveCards("C1", "C2", "C3").Return(
			[]ports.Card{mockCard1, mockCard2, mockCard3}, nil)

		// Traded cards go to discard pile
		mockDiscardPile.EXPECT().Discard(mockCard1)
		mockDiscardPile.EXPECT().Discard(mockCard2)
		mockDiscardPile.EXPECT().Discard(mockCard3)

		// Draw 1 card
		mockPlayer1.EXPECT().CanTakeCards(1).Return(true)
		mockDeck.EXPECT().DrawCard().Return(mockDrawnCard, true)
		mockPlayer1.EXPECT().TakeCards(mockDrawnCard)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			deck:               mockDeck,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
			history:            []historyLine{},
		}

		mockProvider.EXPECT().Get(mockPlayer1, g, mockDrawnCard).Return(expectedStatus)

		status, err := g.Trade("Player1", []string{"C1", "C2", "C3"})

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
		assert.True(t, g.hasTraded)
		assert.False(t, g.CanTrade)
	})
}

func TestGame_Construct(t *testing.T) {
	t.Run("Error when not current player's turn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeConstruct,
		}

		status, err := g.Construct("Player2", "G1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Player2 not your turn")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when not in Construct phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeAttack,
		}

		status, err := g.Construct("Player1", "G1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot construct in the")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when player Construct fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().Construct("G1").Return(errors.New("invalid card"))

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeConstruct,
		}

		status, err := g.Construct("Player1", "G1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "constructing card failed")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Success constructing transitions to EndTurn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		expectedStatus := GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().Construct("G1").Return(nil)

		// nextAction expectations for ActionTypeEndTurn
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.ActionTypeConstruct,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
			history:            []historyLine{},
		}

		mockProvider.EXPECT().Get(mockPlayer1, g).Return(expectedStatus)

		status, err := g.Construct("Player1", "G1")

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
		assert.Equal(t, types.ActionTypeEndTurn, g.currentAction)
	})

	t.Run("History is updated on success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().Construct("G1").Return(nil)
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.ActionTypeConstruct,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
			history:            []historyLine{},
		}

		mockProvider.EXPECT().Get(gomock.Any(), gomock.Any()).Return(GameStatus{})

		_, err := g.Construct("Player1", "G1")

		assert.NoError(t, err)
		found := false
		for _, h := range g.history {
			if strings.Contains(h.Msg,"Player1") && strings.Contains(h.Msg,"constructed") {
				found = true
				break
			}
		}
		assert.True(t, found, "History should contain construct action")
	})

	t.Run("Error when constructing on non-ally in 2v2", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeConstruct,
			Mode:          types.GameMode1v1,
		}

		_, err := g.Construct("Player1", "G1", "Player2")

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

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2, mockPlayer3, mockPlayer4},
			CurrentTurn:   0,
			currentAction: types.ActionTypeConstruct,
			Mode:          types.GameMode2v2,
			Teams:         map[int][]int{0: {0, 1}, 1: {2, 3}},
		}

		_, err := g.Construct("Player1", "G1", "Player2")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "card not in hand")
	})

	t.Run("Success constructing on ally castle in 2v2", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer4 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)
		mockResource := mocks.NewMockResource(ctrl)
		mockCastle := mocks.NewMockCastle(ctrl)
		mockHand := mocks.NewMockHand(ctrl)

		expectedStatus := GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromHand("G1").Return(mockResource, true)
		mockPlayer2.EXPECT().Castle().Return(mockCastle).AnyTimes()
		mockCastle.EXPECT().IsConstructed().Return(true).AnyTimes()
		mockCastle.EXPECT().Construct(mockResource).Return(nil)
		mockPlayer1.EXPECT().Hand().Return(mockHand)
		mockHand.EXPECT().RemoveCard(mockResource).Return(true)

		// nextAction expectations for ActionTypeEndTurn
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2, mockPlayer3, mockPlayer4},
			CurrentTurn:        0,
			currentAction:      types.ActionTypeConstruct,
			Mode:               types.GameMode2v2,
			Teams:              map[int][]int{0: {0, 1}, 1: {2, 3}},
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
			history:            []historyLine{},
		}

		mockProvider.EXPECT().Get(mockPlayer1, g).Return(expectedStatus)

		status, err := g.Construct("Player1", "G1", "Player2")

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
		assert.Equal(t, types.ActionTypeEndTurn, g.currentAction)
		// Check history mentions ally castle
		found := false
		for _, h := range g.history {
			if strings.Contains(h.Msg,"Player2's castle") {
				found = true
				break
			}
		}
		assert.True(t, found, "History should mention ally's castle")
	})
}

func TestGame_Spy(t *testing.T) {
	t.Run("Error when not current player's turn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeSpySteal,
		}

		status, err := g.Spy("Player2", "Player1", 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Player2 not your turn")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when not in SpySteal phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeAttack,
		}

		status, err := g.Spy("Player1", "Player2", 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot use Spy in the")
		assert.Equal(t, GameStatus{}, status)
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
			currentAction: types.ActionTypeSpySteal,
		}

		status, err := g.Spy("Player1", "Player2", 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "player does not have a spy")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when invalid spy option", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().HasSpy().Return(true)

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeSpySteal,
		}

		status, err := g.Spy("Player1", "Player2", 3) // Invalid option

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid Spy option")
		assert.Equal(t, GameStatus{}, status)
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

		expectedStatus := GameStatus{CurrentPlayer: "Player1"}
		revealedCards := []ports.Card{mockRevealedCard}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().HasSpy().Return(true)
		mockPlayer1.EXPECT().Spy().Return(mockSpy)
		mockDiscardPile.EXPECT().Discard(mockSpy)
		mockDeck.EXPECT().Reveal(5).Return(revealedCards)

		// nextAction expectations for ActionTypeBuy
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)
		mockPlayer1.EXPECT().CanBuy().Return(true)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.ActionTypeSpySteal,
			deck:               mockDeck,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
			history:            []historyLine{},
		}

		mockProvider.EXPECT().GetWithModal(mockPlayer1, g, revealedCards).Return(expectedStatus)

		status, err := g.Spy("Player1", "Player2", 1)

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
		assert.Equal(t, types.ActionTypeBuy, g.currentAction)
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

		expectedStatus := GameStatus{CurrentPlayer: "Player1"}
		enemyCards := []ports.Card{mockEnemyCard}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer1.EXPECT().HasSpy().Return(true)
		mockPlayer1.EXPECT().Spy().Return(mockSpy)
		mockDiscardPile.EXPECT().Discard(mockSpy)
		mockPlayer2.EXPECT().Hand().Return(mockEnemyHand)
		mockEnemyHand.EXPECT().ShowCards().Return(enemyCards)

		// nextAction expectations for ActionTypeBuy
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)
		mockPlayer1.EXPECT().CanBuy().Return(true)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.ActionTypeSpySteal,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
			history:            []historyLine{},
		}

		mockProvider.EXPECT().GetWithModal(mockPlayer1, g, enemyCards).Return(expectedStatus)

		status, err := g.Spy("Player1", "Player2", 2)

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
	})
}

func TestGame_Steal(t *testing.T) {
	t.Run("Error when not current player's turn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeSpySteal,
		}

		status, err := g.Steal("Player2", "Player1", 0)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Player2 not your turn")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when not in SpySteal phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeBuy,
		}

		status, err := g.Steal("Player1", "Player2", 0)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot use Steal in the")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when player has no thief", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer1.EXPECT().HasThief().Return(false)

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeSpySteal,
		}

		status, err := g.Steal("Player1", "Player2", 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "player does not have a thief")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when stealing fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer1.EXPECT().HasThief().Return(true)
		mockPlayer2.EXPECT().CardStolenFromHand(0).Return(nil, errors.New("invalid position"))

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeSpySteal,
		}

		status, err := g.Steal("Player1", "Player2", 0)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stealing card failed")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Success stealing a card from enemy", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)
		mockThief := mocks.NewMockThief(ctrl)
		mockStolenCard := mocks.NewMockCard(ctrl)

		expectedStatus := GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer1.EXPECT().HasThief().Return(true)
		mockPlayer2.EXPECT().CardStolenFromHand(2).Return(mockStolenCard, nil)
		mockPlayer1.EXPECT().Thief().Return(mockThief)
		mockDiscardPile.EXPECT().Discard(mockThief) // Thief goes to discard
		mockPlayer1.EXPECT().TakeCards(mockStolenCard)

		// nextAction expectations for ActionTypeBuy
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)
		mockPlayer1.EXPECT().CanBuy().Return(true)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.ActionTypeSpySteal,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
			history:            []historyLine{},
		}

		mockProvider.EXPECT().GetWithModal(
			mockPlayer1, g, []ports.Card{mockStolenCard},
		).Return(expectedStatus)

		status, err := g.Steal("Player1", "Player2", 2)

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
		assert.Equal(t, types.ActionTypeBuy, g.currentAction)
	})

	t.Run("History is updated on successful steal", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)
		mockThief := mocks.NewMockThief(ctrl)
		mockStolenCard := mocks.NewMockCard(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer1.EXPECT().HasThief().Return(true)
		mockPlayer2.EXPECT().CardStolenFromHand(1).Return(mockStolenCard, nil)
		mockPlayer1.EXPECT().Thief().Return(mockThief)
		mockDiscardPile.EXPECT().Discard(mockThief)
		mockPlayer1.EXPECT().TakeCards(mockStolenCard)
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)
		mockPlayer1.EXPECT().CanBuy().Return(false)
		mockPlayer1.EXPECT().CanConstruct().Return(false)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.ActionTypeSpySteal,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
			history:            []historyLine{},
		}

		mockProvider.EXPECT().GetWithModal(gomock.Any(), gomock.Any(), gomock.Any()).Return(GameStatus{})

		_, err := g.Steal("Player1", "Player2", 1)

		assert.NoError(t, err)
		found := false
		for _, h := range g.history {
			if strings.Contains(h.Msg,"Player1") && strings.Contains(h.Msg,"stole") {
				found = true
				break
			}
		}
		assert.True(t, found, "History should contain steal action")
	})
}

func TestGame_Catapult(t *testing.T) {
	t.Run("Error when not current player's turn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeAttack,
		}

		status, err := g.Catapult("Player2", "Player1", 0)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Player2 not your turn")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when not in Attack phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeBuy,
		}

		status, err := g.Catapult("Player1", "Player2", 0)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot use catapult in the")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when player has no catapult", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer1.EXPECT().HasCatapult().Return(false)

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeAttack,
		}

		status, err := g.Catapult("Player1", "Player2", 0)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "player does not have a catapult")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when castle attack fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockCatapult := mocks.NewMockCatapult(ctrl)
		mockCastle := mocks.NewMockCastle(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer1.EXPECT().HasCatapult().Return(true)
		mockPlayer1.EXPECT().Catapult().Return(mockCatapult)
		mockPlayer2.EXPECT().Castle().Return(mockCastle)
		mockCatapult.EXPECT().Attack(mockCastle, 0).Return(nil, errors.New("invalid position"))

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeAttack,
		}

		status, err := g.Catapult("Player1", "Player2", 0)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "attacking castle failed")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Success attacking castle with catapult", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)
		mockCatapult := mocks.NewMockCatapult(ctrl)
		mockCastle := mocks.NewMockCastle(ctrl)
		mockStolenGold := mocks.NewMockResource(ctrl)

		expectedStatus := GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer1.EXPECT().HasCatapult().Return(true)
		mockPlayer1.EXPECT().Catapult().Return(mockCatapult)
		mockPlayer2.EXPECT().Castle().Return(mockCastle)
		mockCatapult.EXPECT().Attack(mockCastle, 2).Return(mockStolenGold, nil)
		mockStolenGold.EXPECT().Value().Return(3)
		mockDiscardPile.EXPECT().Discard(mockStolenGold)

		// nextAction expectations for ActionTypeSpySteal
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)
		mockPlayer1.EXPECT().HasSpy().Return(false)
		mockPlayer1.EXPECT().HasThief().Return(false)
		mockPlayer1.EXPECT().CanBuy().Return(true)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.ActionTypeAttack,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
			history:            []historyLine{},
		}

		mockProvider.EXPECT().Get(mockPlayer1, g).Return(expectedStatus)

		status, err := g.Catapult("Player1", "Player2", 2)

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
		assert.Equal(t, types.ActionTypeBuy, g.currentAction)
	})
}

func TestGame_SpecialPower(t *testing.T) {
	t.Run("Error when not current player's turn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeAttack,
		}

		status, err := g.SpecialPower("Player2", "K1", "EK1", "SP1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Player2 not your turn")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when not in Attack phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeBuy,
		}

		status, err := g.SpecialPower("Player1", "K1", "EK1", "SP1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot use special power in the")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when warrior not in field", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromField("K1").Return(nil, false)

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeAttack,
		}

		status, err := g.SpecialPower("Player1", "K1", "EK1", "SP1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "warrior card not in field")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when target not in any field", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromField("K1").Return(mockWarrior, true)
		mockWarrior.EXPECT().Type().Return(types.ArcherWarriorType)
		// Target not in player's field
		mockPlayer1.EXPECT().GetCardFromField("EK1").Return(nil, false)
		// Target not in enemy's field either
		mockPlayer2.EXPECT().GetCardFromField("EK1").Return(nil, false)

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeAttack,
		}

		status, err := g.SpecialPower("Player1", "K1", "EK1", "SP1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target card not valid")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when weapon not in hand", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockTarget := mocks.NewMockWarrior(ctrl)

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromField("K1").Return(mockWarrior, true)
		mockWarrior.EXPECT().Type().Return(types.ArcherWarriorType)
		mockPlayer1.EXPECT().GetCardFromField("EK1").Return(nil, false)
		mockPlayer2.EXPECT().GetCardFromField("EK1").Return(mockTarget, true)
		mockPlayer1.EXPECT().GetCardFromHand("SP1").Return(nil, false)

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeAttack,
		}

		status, err := g.SpecialPower("Player1", "K1", "EK1", "SP1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "weapon card not in hand")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Error when card is not a special power", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockTarget := mocks.NewMockWarrior(ctrl)
		mockResource := mocks.NewMockResource(ctrl) // Not a special power

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromField("K1").Return(mockWarrior, true)
		mockWarrior.EXPECT().Type().Return(types.ArcherWarriorType)
		mockPlayer1.EXPECT().GetCardFromField("EK1").Return(nil, false)
		mockPlayer2.EXPECT().GetCardFromField("EK1").Return(mockTarget, true)
		mockPlayer1.EXPECT().GetCardFromHand("SP1").Return(mockResource, true)

		g := &Game{
			Players:       []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:   0,
			currentAction: types.ActionTypeAttack,
		}

		status, err := g.SpecialPower("Player1", "K1", "EK1", "SP1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "the card is not a special power")
		assert.Equal(t, GameStatus{}, status)
	})

	t.Run("Success using special power on enemy target", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockTarget := mocks.NewMockWarrior(ctrl)
		mockSP := mocks.NewMockSpecialPower(ctrl)

		expectedStatus := GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromField("K1").Return(mockWarrior, true)
		mockWarrior.EXPECT().Type().Return(types.ArcherWarriorType)
		mockPlayer1.EXPECT().GetCardFromField("EK1").Return(nil, false)
		mockPlayer2.EXPECT().GetCardFromField("EK1").Return(mockTarget, true)
		mockPlayer1.EXPECT().GetCardFromHand("SP1").Return(mockSP, true)
		mockPlayer1.EXPECT().UseSpecialPower(mockWarrior, mockTarget, mockSP).Return(nil)
		mockTarget.EXPECT().String().Return("Knight (20)")

		// nextAction expectations for ActionTypeSpySteal
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)
		mockPlayer1.EXPECT().HasSpy().Return(false)
		mockPlayer1.EXPECT().HasThief().Return(false)
		mockPlayer1.EXPECT().CanBuy().Return(true)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.ActionTypeAttack,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
			history:            []historyLine{},
		}

		mockProvider.EXPECT().Get(mockPlayer1, g).Return(expectedStatus)

		status, err := g.SpecialPower("Player1", "K1", "EK1", "SP1")

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
		assert.Equal(t, types.ActionTypeBuy, g.currentAction)
	})

	t.Run("Success using special power on own target (protect/heal)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockProvider := NewMockGameStatusProvider(ctrl)
		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockTarget := mocks.NewMockWarrior(ctrl)
		mockSP := mocks.NewMockSpecialPower(ctrl)

		expectedStatus := GameStatus{CurrentPlayer: "Player1"}

		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().GetCardFromField("K1").Return(mockWarrior, true)
		mockWarrior.EXPECT().Type().Return(types.KnightWarriorType)
		// Target found in own field
		mockPlayer1.EXPECT().GetCardFromField("A1").Return(mockTarget, true)
		mockPlayer1.EXPECT().GetCardFromHand("SP1").Return(mockSP, true)
		mockPlayer1.EXPECT().UseSpecialPower(mockWarrior, mockTarget, mockSP).Return(nil)
		mockTarget.EXPECT().String().Return("Archer (20)")

		// nextAction expectations for ActionTypeSpySteal
		mockPlayer1.EXPECT().HasWarriorsInHand().Return(false)
		mockPlayer1.EXPECT().CanTradeCards().Return(false)
		mockPlayer1.EXPECT().HasSpy().Return(false)
		mockPlayer1.EXPECT().HasThief().Return(false)
		mockPlayer1.EXPECT().CanBuy().Return(true)

		g := &Game{
			Players:            []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:        0,
			currentAction:      types.ActionTypeAttack,
			discardPile:        mockDiscardPile,
			GameStatusProvider: mockProvider,
			history:            []historyLine{},
		}

		mockProvider.EXPECT().Get(mockPlayer1, g).Return(expectedStatus)

		status, err := g.SpecialPower("Player1", "K1", "A1", "SP1")

		assert.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
	})
}

func TestGame_OnCastleCompletion(t *testing.T) {
	t.Run("1v1 sets individual winner", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Mode: types.GameMode1v1,
		}

		g.OnCastleCompletion(mockPlayer1)

		assert.True(t, g.gameOver)
		assert.Equal(t, "Player1", g.winner)
	})

	t.Run("2v2 sets team winner", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Mode: types.GameMode2v2,
		}

		g.OnCastleCompletion(mockPlayer1)

		assert.True(t, g.gameOver)
		assert.Equal(t, "Player1's team", g.winner)
	})

	t.Run("FFA3 sets individual winner", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Mode: types.GameModeFFA3,
		}

		g.OnCastleCompletion(mockPlayer1)

		assert.True(t, g.gameOver)
		assert.Equal(t, "Player1", g.winner)
	})
}

func TestGame_OnFieldWithoutWarriors(t *testing.T) {
	t.Run("1v1 current player wins immediately", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		g := &Game{
			Players:           []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:       0,
			Mode:              types.GameMode1v1,
			EliminatedPlayers: make(map[int]bool),
			history:           []historyLine{},
		}

		g.OnFieldWithoutWarriors("Player2")

		assert.True(t, g.gameOver)
		assert.Equal(t, "Player1", g.winner)
	})

	t.Run("FFA3 eliminates player, game continues with 2 remaining", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer3.EXPECT().Name().Return("Player3").AnyTimes()

		mockHand2 := mocks.NewMockHand(ctrl)
		mockCastle2 := mocks.NewMockCastle(ctrl)
		mockPlayer2.EXPECT().Hand().Return(mockHand2)
		mockHand2.EXPECT().ShowCards().Return([]ports.Card{})
		mockPlayer2.EXPECT().Castle().Return(mockCastle2)
		mockCastle2.EXPECT().ResourceCards().Return([]ports.Resource{})

		g := &Game{
			Players:           []ports.Player{mockPlayer1, mockPlayer2, mockPlayer3},
			CurrentTurn:       0,
			Mode:              types.GameModeFFA3,
			EliminatedPlayers: make(map[int]bool),
			history:           []historyLine{},
		}

		g.OnFieldWithoutWarriors("Player2")

		assert.False(t, g.gameOver)
		assert.True(t, g.EliminatedPlayers[1])
		assert.Contains(t, g.history, historyLine{Msg: "Player2 has been eliminated!", Category: types.CategoryElimination})
	})

	t.Run("FFA3 last player standing wins", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer3.EXPECT().Name().Return("Player3").AnyTimes()

		mockHand3 := mocks.NewMockHand(ctrl)
		mockCastle3 := mocks.NewMockCastle(ctrl)
		mockPlayer3.EXPECT().Hand().Return(mockHand3)
		mockHand3.EXPECT().ShowCards().Return([]ports.Card{})
		mockPlayer3.EXPECT().Castle().Return(mockCastle3)
		mockCastle3.EXPECT().ResourceCards().Return([]ports.Resource{})

		g := &Game{
			Players:           []ports.Player{mockPlayer1, mockPlayer2, mockPlayer3},
			CurrentTurn:       0,
			Mode:              types.GameModeFFA3,
			EliminatedPlayers: map[int]bool{1: true}, // Player2 already eliminated
			history:           []historyLine{},
		}

		g.OnFieldWithoutWarriors("Player3")

		assert.True(t, g.gameOver)
		assert.Equal(t, "Player1", g.winner)
		assert.True(t, g.EliminatedPlayers[2])
	})

	t.Run("FFA5 eliminates player, game continues", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayers := make([]*mocks.MockPlayer, 5)
		players := make([]ports.Player, 5)
		for i := 0; i < 5; i++ {
			mp := mocks.NewMockPlayer(ctrl)
			mp.EXPECT().Name().Return(
				"Player" + string(rune('1'+i))).AnyTimes()
			mockPlayers[i] = mp
			players[i] = mp
		}

		// Mock Hand/Castle for eliminated player (Player2 = index 1)
		mockHand := mocks.NewMockHand(ctrl)
		mockCastle := mocks.NewMockCastle(ctrl)
		mockPlayers[1].EXPECT().Hand().Return(mockHand)
		mockHand.EXPECT().ShowCards().Return([]ports.Card{})
		mockPlayers[1].EXPECT().Castle().Return(mockCastle)
		mockCastle.EXPECT().ResourceCards().Return([]ports.Resource{})

		g := &Game{
			Players:           players,
			CurrentTurn:       0,
			Mode:              types.GameModeFFA5,
			EliminatedPlayers: make(map[int]bool),
			history:           []historyLine{},
		}

		g.OnFieldWithoutWarriors("Player2")

		assert.False(t, g.gameOver)
		assert.True(t, g.EliminatedPlayers[1])
	})

	t.Run("2v2 eliminates one enemy, game continues", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer4 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer3.EXPECT().Name().Return("Player3").AnyTimes()
		mockPlayer4.EXPECT().Name().Return("Player4").AnyTimes()

		mockHand2 := mocks.NewMockHand(ctrl)
		mockCastle2 := mocks.NewMockCastle(ctrl)
		mockPlayer2.EXPECT().Hand().Return(mockHand2)
		mockHand2.EXPECT().ShowCards().Return([]ports.Card{})
		mockPlayer2.EXPECT().Castle().Return(mockCastle2)
		mockCastle2.EXPECT().ResourceCards().Return([]ports.Resource{})

		g := &Game{
			Players:           []ports.Player{mockPlayer1, mockPlayer2, mockPlayer3, mockPlayer4},
			CurrentTurn:       0, // Player1's turn (Team 1)
			Mode:              types.GameMode2v2,
			Teams:             map[int][]int{1: {0, 2}, 2: {1, 3}},
			EliminatedPlayers: make(map[int]bool),
			history:           []historyLine{},
		}

		// Player2 (Team 2) loses warriors, but Player4 (Team 2) is still alive
		g.OnFieldWithoutWarriors("Player2")

		assert.False(t, g.gameOver)
		assert.True(t, g.EliminatedPlayers[1])
		assert.Contains(t, g.history, historyLine{Msg: "Player2 has been eliminated!", Category: types.CategoryElimination})
	})

	t.Run("2v2 both enemies eliminated, team wins", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer4 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer3.EXPECT().Name().Return("Player3").AnyTimes()
		mockPlayer4.EXPECT().Name().Return("Player4").AnyTimes()

		mockHand4 := mocks.NewMockHand(ctrl)
		mockCastle4 := mocks.NewMockCastle(ctrl)
		mockPlayer4.EXPECT().Hand().Return(mockHand4)
		mockHand4.EXPECT().ShowCards().Return([]ports.Card{})
		mockPlayer4.EXPECT().Castle().Return(mockCastle4)
		mockCastle4.EXPECT().ResourceCards().Return([]ports.Resource{})

		g := &Game{
			Players:           []ports.Player{mockPlayer1, mockPlayer2, mockPlayer3, mockPlayer4},
			CurrentTurn:       0, // Player1's turn (Team 1)
			Mode:              types.GameMode2v2,
			Teams:             map[int][]int{1: {0, 2}, 2: {1, 3}},
			EliminatedPlayers: map[int]bool{1: true}, // Player2 already eliminated
			history:           []historyLine{},
		}

		// Player4 (last of Team 2) loses warriors
		g.OnFieldWithoutWarriors("Player4")

		assert.True(t, g.gameOver)
		assert.Equal(t, "Player1's team", g.winner)
		assert.True(t, g.EliminatedPlayers[3])
	})
}

func TestGame_IsGameOver(t *testing.T) {
	t.Run("Returns false initially", func(t *testing.T) {
		g := &Game{}

		gameOver, winner := g.IsGameOver()

		assert.False(t, gameOver)
		assert.Empty(t, winner)
	})

	t.Run("Returns true after game ends", func(t *testing.T) {
		g := &Game{
			gameOver: true,
			winner:   "Player1",
		}

		gameOver, winner := g.IsGameOver()

		assert.True(t, gameOver)
		assert.Equal(t, "Player1", winner)
	})
}

func TestGame_GetHistory(t *testing.T) {
	t.Run("Returns all history on first call", func(t *testing.T) {
		g := &Game{
			history: []historyLine{
				{Msg: "msg1", Category: types.CategoryInfo},
				{Msg: "msg2", Category: types.CategoryInfo},
				{Msg: "msg3", Category: types.CategoryInfo},
			},
		}

		result := g.GetHistory()

		assert.Len(t, result, 3)
		assert.Equal(t, "msg1", result[0].Msg)
		assert.Equal(t, "msg3", result[2].Msg)
	})

	t.Run("Returns only new messages on subsequent calls", func(t *testing.T) {
		g := &Game{
			history: []historyLine{
				{Msg: "msg1", Category: types.CategoryInfo},
				{Msg: "msg2", Category: types.CategoryInfo},
			},
		}

		_ = g.GetHistory() // First call reads all

		g.history = append(g.history,
			historyLine{Msg: "msg3", Category: types.CategoryInfo},
			historyLine{Msg: "msg4", Category: types.CategoryInfo},
		)
		result := g.GetHistory()

		assert.Len(t, result, 2)
		assert.Equal(t, "msg3", result[0].Msg)
		assert.Equal(t, "msg4", result[1].Msg)
	})

	t.Run("Returns empty slice when no new messages", func(t *testing.T) {
		g := &Game{
			history: []historyLine{
				{Msg: "msg1", Category: types.CategoryInfo},
			},
		}

		_ = g.GetHistory()
		result := g.GetHistory()

		assert.Empty(t, result)
	})
}

func TestGame_OnWarriorMovedToCemetery(t *testing.T) {
	t.Run("Adds warrior to cemetery and records history", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockCemetery := mocks.NewMockCemetery(ctrl)
		mockWarrior := mocks.NewMockWarrior(ctrl)
		mockCemetery.EXPECT().AddCorp(mockWarrior)

		g := &Game{
			cemetery: mockCemetery,
			history:  []historyLine{},
		}

		g.OnWarriorMovedToCemetery(mockWarrior)

		assert.Contains(t, g.history, historyLine{Msg: "warrior buried in cemetery", Category: types.CategoryInfo})
	})
}

func TestGame_AutoMoveWarriorToField(t *testing.T) {
	t.Run("Success moving warrior", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().MoveCardToField("W1").Return(nil)

		g := &Game{
			Players: []ports.Player{mockPlayer1},
		}

		err := g.AutoMoveWarriorToField("Player1", "W1")

		assert.NoError(t, err)
	})

	t.Run("Error when player not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players: []ports.Player{mockPlayer1},
		}

		err := g.AutoMoveWarriorToField("Unknown", "W1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "player Unknown not found")
	})

	t.Run("Error when MoveCardToField fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer1.EXPECT().MoveCardToField("W1").Return(errors.New("field full"))

		g := &Game{
			Players: []ports.Player{mockPlayer1},
		}

		err := g.AutoMoveWarriorToField("Player1", "W1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "field full")
	})
}

func TestGame_SameTeam(t *testing.T) {
	t.Run("Returns false for non-2v2 mode", func(t *testing.T) {
		g := &Game{Mode: types.GameMode1v1}
		assert.False(t, g.SameTeam(0, 1))
	})

	t.Run("Returns true for same team in 2v2", func(t *testing.T) {
		g := &Game{
			Mode:  types.GameMode2v2,
			Teams: map[int][]int{1: {0, 2}, 2: {1, 3}},
		}

		assert.True(t, g.SameTeam(0, 2))  // Team 1
		assert.True(t, g.SameTeam(2, 0))  // Symmetric
		assert.True(t, g.SameTeam(1, 3))  // Team 2
		assert.True(t, g.SameTeam(3, 1))  // Symmetric
	})

	t.Run("Returns false for different teams in 2v2", func(t *testing.T) {
		g := &Game{
			Mode:  types.GameMode2v2,
			Teams: map[int][]int{1: {0, 2}, 2: {1, 3}},
		}

		assert.False(t, g.SameTeam(0, 1))
		assert.False(t, g.SameTeam(0, 3))
		assert.False(t, g.SameTeam(2, 1))
		assert.False(t, g.SameTeam(2, 3))
	})
}

func TestGame_Allies(t *testing.T) {
	t.Run("Returns nil for 1v1", func(t *testing.T) {
		g := &Game{Mode: types.GameMode1v1}
		assert.Nil(t, g.Allies(0))
	})

	t.Run("Returns nil for FFA3", func(t *testing.T) {
		g := &Game{Mode: types.GameModeFFA3}
		assert.Nil(t, g.Allies(0))
	})

	t.Run("Returns teammate for 2v2", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer4 := mocks.NewMockPlayer(ctrl)

		g := &Game{
			Players: []ports.Player{mockPlayer1, mockPlayer2, mockPlayer3, mockPlayer4},
			Mode:    types.GameMode2v2,
			Teams:   map[int][]int{1: {0, 2}, 2: {1, 3}},
		}

		allies0 := g.Allies(0)
		assert.Len(t, allies0, 1)
		assert.Equal(t, mockPlayer3, allies0[0])

		allies1 := g.Allies(1)
		assert.Len(t, allies1, 1)
		assert.Equal(t, mockPlayer4, allies1[0])
	})
}

func TestGame_Enemies(t *testing.T) {
	t.Run("1v1 returns opponent", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		g := &Game{
			Players:           []ports.Player{mockPlayer1, mockPlayer2},
			Mode:              types.GameMode1v1,
			EliminatedPlayers: make(map[int]bool),
		}

		enemies := g.Enemies(0)
		assert.Len(t, enemies, 1)
		assert.Equal(t, mockPlayer2, enemies[0])
	})

	t.Run("2v2 excludes teammates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer4 := mocks.NewMockPlayer(ctrl)

		g := &Game{
			Players:           []ports.Player{mockPlayer1, mockPlayer2, mockPlayer3, mockPlayer4},
			Mode:              types.GameMode2v2,
			Teams:             map[int][]int{1: {0, 2}, 2: {1, 3}},
			EliminatedPlayers: make(map[int]bool),
		}

		enemies := g.Enemies(0) // Player1 (Team 1)
		assert.Len(t, enemies, 2)
		assert.Equal(t, mockPlayer2, enemies[0])
		assert.Equal(t, mockPlayer4, enemies[1])
	})

	t.Run("Excludes eliminated players", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)

		g := &Game{
			Players:           []ports.Player{mockPlayer1, mockPlayer2, mockPlayer3},
			Mode:              types.GameModeFFA3,
			EliminatedPlayers: map[int]bool{1: true},
		}

		enemies := g.Enemies(0)
		assert.Len(t, enemies, 1)
		assert.Equal(t, mockPlayer3, enemies[0])
	})
}

func TestGame_getTargetPlayer(t *testing.T) {
	t.Run("Error when target not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players: []ports.Player{mockPlayer1},
		}

		_, err := g.getTargetPlayer("Player1", "Unknown")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target player Unknown not found")
	})

	t.Run("Error when targeting self", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players:           []ports.Player{mockPlayer1},
			EliminatedPlayers: make(map[int]bool),
		}

		_, err := g.getTargetPlayer("Player1", "Player1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot attack yourself")
	})

	t.Run("Error when targeting ally in 2v2", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer4 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer3.EXPECT().Name().Return("Player3").AnyTimes()
		mockPlayer4.EXPECT().Name().Return("Player4").AnyTimes()

		g := &Game{
			Players:           []ports.Player{mockPlayer1, mockPlayer2, mockPlayer3, mockPlayer4},
			Mode:              types.GameMode2v2,
			Teams:             map[int][]int{1: {0, 2}, 2: {1, 3}},
			EliminatedPlayers: make(map[int]bool),
		}

		_, err := g.getTargetPlayer("Player1", "Player3") // Player3 is teammate
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot attack your ally")
	})

	t.Run("Error when targeting eliminated player", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer3.EXPECT().Name().Return("Player3").AnyTimes()

		g := &Game{
			Players:           []ports.Player{mockPlayer1, mockPlayer2, mockPlayer3},
			Mode:              types.GameModeFFA3,
			EliminatedPlayers: map[int]bool{1: true},
		}

		_, err := g.getTargetPlayer("Player1", "Player2")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot attack eliminated player")
	})

	t.Run("Success targeting valid enemy", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		g := &Game{
			Players:           []ports.Player{mockPlayer1, mockPlayer2},
			Mode:              types.GameMode1v1,
			EliminatedPlayers: make(map[int]bool),
		}

		target, err := g.getTargetPlayer("Player1", "Player2")
		assert.NoError(t, err)
		assert.Equal(t, mockPlayer2, target)
	})

	t.Run("Success targeting valid enemy in 2v2", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)
		mockPlayer4 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()
		mockPlayer3.EXPECT().Name().Return("Player3").AnyTimes()
		mockPlayer4.EXPECT().Name().Return("Player4").AnyTimes()

		g := &Game{
			Players:           []ports.Player{mockPlayer1, mockPlayer2, mockPlayer3, mockPlayer4},
			Mode:              types.GameMode2v2,
			Teams:             map[int][]int{1: {0, 2}, 2: {1, 3}},
			EliminatedPlayers: make(map[int]bool),
		}

		target, err := g.getTargetPlayer("Player1", "Player2") // Player2 is enemy
		assert.NoError(t, err)
		assert.Equal(t, mockPlayer2, target)
	})
}

func TestGame_switchTurn(t *testing.T) {
	t.Run("Switches to next player", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		g := &Game{
			Players:           []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:       0,
			hasMovedWarrior:   true,
			hasTraded:         true,
			currentAction:     types.ActionTypeEndTurn,
			EliminatedPlayers: make(map[int]bool),
		}

		g.switchTurn()

		assert.Equal(t, 1, g.CurrentTurn)
		assert.False(t, g.hasMovedWarrior)
		assert.False(t, g.hasTraded)
		assert.Equal(t, types.ActionTypeDrawCard, g.currentAction)
	})

	t.Run("Wraps around to first player", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)

		g := &Game{
			Players:           []ports.Player{mockPlayer1, mockPlayer2},
			CurrentTurn:       1,
			EliminatedPlayers: make(map[int]bool),
		}

		g.switchTurn()

		assert.Equal(t, 0, g.CurrentTurn)
	})

	t.Run("Skips eliminated players", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer3 := mocks.NewMockPlayer(ctrl)

		g := &Game{
			Players:           []ports.Player{mockPlayer1, mockPlayer2, mockPlayer3},
			CurrentTurn:       0,
			EliminatedPlayers: map[int]bool{1: true}, // Player2 eliminated
		}

		g.switchTurn()

		assert.Equal(t, 2, g.CurrentTurn) // Skips Player2
	})

	t.Run("Skips multiple eliminated players", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		players := make([]ports.Player, 5)
		for i := 0; i < 5; i++ {
			players[i] = mocks.NewMockPlayer(ctrl)
		}

		g := &Game{
			Players:           players,
			CurrentTurn:       0,
			EliminatedPlayers: map[int]bool{1: true, 2: true, 3: true},
		}

		g.switchTurn()

		assert.Equal(t, 4, g.CurrentTurn)
	})
}

func TestGame_PlayerIndex(t *testing.T) {
	t.Run("Returns correct index", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		g := &Game{
			Players: []ports.Player{mockPlayer1, mockPlayer2},
		}

		assert.Equal(t, 0, g.PlayerIndex("Player1"))
		assert.Equal(t, 1, g.PlayerIndex("Player2"))
	})

	t.Run("Returns -1 for unknown player", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players: []ports.Player{mockPlayer1},
		}

		assert.Equal(t, -1, g.PlayerIndex("Unknown"))
	})
}

func TestGame_GetPlayer(t *testing.T) {
	t.Run("Returns player by name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer2 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()
		mockPlayer2.EXPECT().Name().Return("Player2").AnyTimes()

		g := &Game{
			Players: []ports.Player{mockPlayer1, mockPlayer2},
		}

		assert.Equal(t, mockPlayer1, g.GetPlayer("Player1"))
		assert.Equal(t, mockPlayer2, g.GetPlayer("Player2"))
	})

	t.Run("Returns nil for unknown player", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPlayer1 := mocks.NewMockPlayer(ctrl)
		mockPlayer1.EXPECT().Name().Return("Player1").AnyTimes()

		g := &Game{
			Players: []ports.Player{mockPlayer1},
		}

		assert.Nil(t, g.GetPlayer("Unknown"))
	})
}

func TestGame_OnCardMovedToPile(t *testing.T) {
	t.Run("Discards card to pile", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDiscardPile := mocks.NewMockDiscardPile(ctrl)
		mockCard := mocks.NewMockCard(ctrl)
		mockDiscardPile.EXPECT().Discard(mockCard)

		g := &Game{
			discardPile: mockDiscardPile,
		}

		g.OnCardMovedToPile(mockCard)
	})
}

func TestGame_addToHistory(t *testing.T) {
	t.Run("Adds message to history", func(t *testing.T) {
		g := &Game{history: []historyLine{}}
		g.addToHistory("test message", types.CategoryInfo)
		assert.Len(t, g.history, 1)
		assert.Equal(t, "test message", g.history[0].Msg)
		assert.Equal(t, types.CategoryInfo, g.history[0].Category)
	})

	t.Run("Does not add empty message", func(t *testing.T) {
		g := &Game{history: []historyLine{}}
		g.addToHistory("", types.CategoryInfo)
		assert.Empty(t, g.history)
	})
}

func TestGame_validatePlayers(t *testing.T) {
	t.Run("1v1 requires 2 players", func(t *testing.T) {
		assert.NoError(t, validatePlayers([]string{"A", "B"}, types.GameMode1v1))
		assert.Error(t, validatePlayers([]string{"A"}, types.GameMode1v1))
		assert.Error(t, validatePlayers([]string{"A", "B", "C"}, types.GameMode1v1))
	})

	t.Run("2v2 requires 4 players", func(t *testing.T) {
		assert.NoError(t, validatePlayers([]string{"A", "B", "C", "D"}, types.GameMode2v2))
		assert.Error(t, validatePlayers([]string{"A", "B"}, types.GameMode2v2))
		assert.Error(t, validatePlayers([]string{"A", "B", "C", "D", "E"}, types.GameMode2v2))
	})

	t.Run("FFA3 requires 3 players", func(t *testing.T) {
		assert.NoError(t, validatePlayers([]string{"A", "B", "C"}, types.GameModeFFA3))
		assert.Error(t, validatePlayers([]string{"A", "B"}, types.GameModeFFA3))
	})

	t.Run("FFA5 requires 5 players", func(t *testing.T) {
		assert.NoError(t, validatePlayers([]string{"A", "B", "C", "D", "E"}, types.GameModeFFA5))
		assert.Error(t, validatePlayers([]string{"A", "B", "C"}, types.GameModeFFA5))
	})

	t.Run("Invalid game mode", func(t *testing.T) {
		err := validatePlayers([]string{"A", "B"}, "invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid game mode")
	})
}

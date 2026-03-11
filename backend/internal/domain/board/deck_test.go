package board_test

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// makeMockCards returns n MockCard instances as a []cards.Card slice.
func makeMockCards(ctrl *gomock.Controller, n int) []cards.Card {
	result := make([]cards.Card, n)
	for i := range result {
		result[i] = mocks.NewMockCard(ctrl)
	}
	return result
}

func TestDeck_Count_EmptyAfterCreation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	d := board.NewDeck(mocks.NewMockDealer(ctrl))
	assert.Equal(t, 0, d.Count())
}

func TestDeck_DrawCards_ErrorWhenDeckAndDiscardBothEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	d := board.NewDeck(mocks.NewMockDealer(ctrl))
	dp := board.NewDiscardPile()

	result, err := d.DrawCards(1, dp)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestDeck_DrawCards_ReshufflesDiscardWhenDeckIsEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c1 := mocks.NewMockCard(ctrl)
	c2 := mocks.NewMockCard(ctrl)

	d := board.NewDeck(mocks.NewMockDealer(ctrl))
	dp := board.NewDiscardPile()
	dp.Discard(c1)
	dp.Discard(c2)

	result, err := d.DrawCards(1, dp)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	// One card was drawn; the other stays in the (now-reshuffled) deck.
	assert.Equal(t, 1, d.Count())
	// Discard pile was emptied during reshuffle.
	assert.Equal(t, 0, dp.Count())
}

func TestDeck_DrawCards_DrawsMultipleCardsFromReshuffle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	discarded := makeMockCards(ctrl, 5)

	d := board.NewDeck(mocks.NewMockDealer(ctrl))
	dp := board.NewDiscardPile()
	for _, c := range discarded {
		dp.Discard(c)
	}

	result, err := d.DrawCards(3, dp)

	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, 2, d.Count())
}

func TestDeck_DrawCards_ErrorWhenNotEnoughCardsEvenAfterReshuffle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Only 1 card in discard — asking for 2 should fail after the reshuffle draw.
	c1 := mocks.NewMockCard(ctrl)

	d := board.NewDeck(mocks.NewMockDealer(ctrl))
	dp := board.NewDiscardPile()
	dp.Discard(c1)

	// First iteration: deck empty → reshuffle discard (1 card) → draw 1 → OK.
	// Second iteration: deck empty again, discard empty → error.
	result, err := d.DrawCards(2, dp)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestDeck_Deal_EachPlayerReceivesWarriorsAndOtherCards(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 2 players: 3 warriors each (6 total), 4 others each + 2 extra → 10 total.
	warriors := makeMockCards(ctrl, 6)
	others := makeMockCards(ctrl, 10)

	dealer := mocks.NewMockDealer(ctrl)
	dealer.EXPECT().WarriorsCards(2).Return(warriors)
	dealer.EXPECT().OtherCards(2).Return(others)

	p1 := mocks.NewMockPlayer(ctrl)
	p2 := mocks.NewMockPlayer(ctrl)

	// Each player gets TakeCards called twice: once with 3 warriors, once with 4 others.
	p1.EXPECT().TakeCards(gomock.Any(), gomock.Any(), gomock.Any()).Return(true)
	p1.EXPECT().TakeCards(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true)
	p2.EXPECT().TakeCards(gomock.Any(), gomock.Any(), gomock.Any()).Return(true)
	p2.EXPECT().TakeCards(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true)

	d := board.NewDeck(dealer)
	d.Deal([]board.Player{p1, p2})

	// 6 warriors all assigned + 10 others, 8 assigned → 2 remaining in deck.
	assert.Equal(t, 2, d.Count())
}

func TestDeck_Reveal_ReturnsRequestedCount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Populate deck via reshuffle: add 4 cards to discard, draw 1 → 3 remain.
	cards4 := makeMockCards(ctrl, 4)
	d := board.NewDeck(mocks.NewMockDealer(ctrl))
	dp := board.NewDiscardPile()
	for _, c := range cards4 {
		dp.Discard(c)
	}
	_, _ = d.DrawCards(1, dp)
	// Deck now has 3 cards.

	assert.Len(t, d.Reveal(2), 2)
}

func TestDeck_Reveal_ClampsToAvailableCards(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cards3 := makeMockCards(ctrl, 3)
	d := board.NewDeck(mocks.NewMockDealer(ctrl))
	dp := board.NewDiscardPile()
	for _, c := range cards3 {
		dp.Discard(c)
	}
	// Draw 0 to trigger reshuffle and populate deck with all 3 cards.
	_, _ = d.DrawCards(1, dp)
	// Deck has 2 cards.

	// Reveal(100) should return all available cards, not panic.
	assert.Len(t, d.Reveal(100), d.Count())
}

func TestDeck_Reveal_ReturnsEmptyForZero(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	d := board.NewDeck(mocks.NewMockDealer(ctrl))
	assert.Empty(t, d.Reveal(0))
}

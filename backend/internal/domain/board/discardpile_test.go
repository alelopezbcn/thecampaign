package board_test

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDiscardPile_NewDiscardPile_IsEmpty(t *testing.T) {
	dp := board.NewDiscardPile()
	assert.Equal(t, 0, dp.Count())
	assert.Empty(t, dp.Cards())
	assert.Nil(t, dp.GetLast())
}

func TestDiscardPile_Discard_IncreasesCount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dp := board.NewDiscardPile()
	c := mocks.NewMockCard(ctrl)
	dp.Discard(c)
	assert.Equal(t, 1, dp.Count())
}

func TestDiscardPile_GetLast_ReturnsNilWhenEmpty(t *testing.T) {
	dp := board.NewDiscardPile()
	assert.Nil(t, dp.GetLast())
}

func TestDiscardPile_GetLast_ReturnsMostRecentCard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dp := board.NewDiscardPile()
	c1 := mocks.NewMockCard(ctrl)
	c2 := mocks.NewMockCard(ctrl)
	dp.Discard(c1)
	dp.Discard(c2)
	assert.Equal(t, c2, dp.GetLast())
}

func TestDiscardPile_Cards_ReturnsAllCards(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dp := board.NewDiscardPile()
	c1 := mocks.NewMockCard(ctrl)
	c2 := mocks.NewMockCard(ctrl)
	dp.Discard(c1)
	dp.Discard(c2)

	cards := dp.Cards()
	assert.Len(t, cards, 2)
	assert.Contains(t, cards, c1)
	assert.Contains(t, cards, c2)
}

func TestDiscardPile_Empty_ReturnsCardsAndClearsPile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dp := board.NewDiscardPile()
	c1 := mocks.NewMockCard(ctrl)
	c2 := mocks.NewMockCard(ctrl)
	dp.Discard(c1)
	dp.Discard(c2)

	returned := dp.Empty()
	assert.Len(t, returned, 2)
	assert.Contains(t, returned, c1)
	assert.Contains(t, returned, c2)
	// Pile must be cleared after Empty()
	assert.Equal(t, 0, dp.Count())
	assert.Nil(t, dp.GetLast())
}

func TestDiscardPile_Empty_OnEmptyPile_ReturnsEmptySlice(t *testing.T) {
	dp := board.NewDiscardPile()
	returned := dp.Empty()
	assert.Empty(t, returned)
	assert.Equal(t, 0, dp.Count())
}

func TestDiscardPile_Empty_SubsequentDiscardWorksAfterEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dp := board.NewDiscardPile()
	c1 := mocks.NewMockCard(ctrl)
	c2 := mocks.NewMockCard(ctrl)
	dp.Discard(c1)
	dp.Empty()

	dp.Discard(c2)
	assert.Equal(t, 1, dp.Count())
	assert.Equal(t, c2, dp.GetLast())
}

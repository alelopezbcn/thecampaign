package domain

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHand_NewHand(t *testing.T) {
	h := NewHand()
	assert.Equal(t, 0, h.Count())
	assert.Empty(t, h.ShowCards())
}

func TestHand_AddCards(t *testing.T) {
	t.Run("Adds cards successfully", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		h := NewHand()

		c := mocks.NewMockCard(ctrl)
		c.EXPECT().GetID().Return("c1").AnyTimes()

		err := h.AddCards(c)
		assert.NoError(t, err)
		assert.Equal(t, 1, h.Count())
	})

	t.Run("Adds multiple cards", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		h := NewHand()

		cards := make([]*mocks.MockCard, 3)
		for i := range cards {
			cards[i] = mocks.NewMockCard(ctrl)
			cards[i].EXPECT().GetID().Return("c" + string(rune('0'+i))).AnyTimes()
		}

		err := h.AddCards(cards[0], cards[1], cards[2])
		assert.NoError(t, err)
		assert.Equal(t, 3, h.Count())
	})

	t.Run("Error when exceeding max cards", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		h := NewHand()

		// Add 7 cards (max)
		for i := 0; i < 7; i++ {
			c := mocks.NewMockCard(ctrl)
			c.EXPECT().GetID().Return("c" + string(rune('0'+i))).AnyTimes()
			h.AddCards(c)
		}

		extra := mocks.NewMockCard(ctrl)
		extra.EXPECT().GetID().Return("extra").AnyTimes()
		err := h.AddCards(extra)
		assert.Error(t, err)
		assert.Equal(t, 7, h.Count())
	})
}

func TestHand_GetCard(t *testing.T) {
	t.Run("Finds card by ID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		h := NewHand()

		c := mocks.NewMockCard(ctrl)
		c.EXPECT().GetID().Return("c1").AnyTimes()
		h.AddCards(c)

		found, ok := h.GetCard("c1")
		assert.True(t, ok)
		assert.Equal(t, c, found)
	})

	t.Run("Case insensitive search", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		h := NewHand()

		c := mocks.NewMockCard(ctrl)
		c.EXPECT().GetID().Return("Card1").AnyTimes()
		h.AddCards(c)

		found, ok := h.GetCard("card1")
		assert.True(t, ok)
		assert.Equal(t, c, found)
	})

	t.Run("Returns false when not found", func(t *testing.T) {
		h := NewHand()
		_, ok := h.GetCard("nonexistent")
		assert.False(t, ok)
	})
}

func TestHand_RemoveCard(t *testing.T) {
	t.Run("Removes card successfully", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		h := NewHand()

		c := mocks.NewMockCard(ctrl)
		c.EXPECT().GetID().Return("c1").AnyTimes()
		h.AddCards(c)

		ok := h.RemoveCard(c)
		assert.True(t, ok)
		assert.Equal(t, 0, h.Count())
	})

	t.Run("Returns false when card not in hand", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		h := NewHand()

		c := mocks.NewMockCard(ctrl)
		c.EXPECT().GetID().Return("c1").AnyTimes()

		ok := h.RemoveCard(c)
		assert.False(t, ok)
	})
}

func TestHand_CanAddCards(t *testing.T) {
	h := NewHand()
	assert.True(t, h.CanAddCards(7))
	assert.True(t, h.CanAddCards(1))
	assert.False(t, h.CanAddCards(8))
}

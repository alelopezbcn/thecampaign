package board

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func newTestCastle(ctrl *gomock.Controller) (
	Castle,
	*mocks.MockPlayer,
	*mocks.MockCastleCompletionObserver,
) {
	mockPlayer := mocks.NewMockPlayer(ctrl)
	mockPlayer.EXPECT().Name().Return("TestPlayer").AnyTimes()
	castleObs := mocks.NewMockCastleCompletionObserver(ctrl)
	c := NewCastle(25, mockPlayer, castleObs)
	return c, mockPlayer, castleObs
}

func constructCastle(t *testing.T, c Castle, ctrl *gomock.Controller) {
	t.Helper()
	r := mocks.NewMockResource(ctrl)
	r.EXPECT().Value().Return(1).AnyTimes()
	r.EXPECT().GetID().Return("init").AnyTimes()
	err := c.Construct(r)
	assert.NoError(t, err)
}

func TestCastle_NewCastle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c, _, _ := newTestCastle(ctrl)

	assert.Equal(t, "castle_TestPlayer", c.GetID())
	assert.False(t, c.IsConstructed())
	assert.Equal(t, 0, c.Value())
	assert.Equal(t, 0, c.ResourceCardsCount())
}

func TestCastle_Construct(t *testing.T) {
	t.Run("Construct with resource value 1", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c, _, _ := newTestCastle(ctrl)

		r := mocks.NewMockResource(ctrl)
		r.EXPECT().Value().Return(1).AnyTimes()

		err := c.Construct(r)
		assert.NoError(t, err)
		assert.True(t, c.IsConstructed())
	})

	t.Run("Construct with weapon damage 1", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c, _, _ := newTestCastle(ctrl)

		w := mocks.NewMockWeapon(ctrl)
		w.EXPECT().DamageAmount().Return(1).AnyTimes()

		err := c.Construct(w)
		assert.NoError(t, err)
		assert.True(t, c.IsConstructed())
	})

	t.Run("Error with resource value != 1", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c, _, _ := newTestCastle(ctrl)

		r := mocks.NewMockResource(ctrl)
		r.EXPECT().Value().Return(4).AnyTimes()

		err := c.Construct(r)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid card for constructing")
		assert.False(t, c.IsConstructed())
	})

	t.Run("Error with weapon damage != 1", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c, _, _ := newTestCastle(ctrl)

		w := mocks.NewMockWeapon(ctrl)
		w.EXPECT().DamageAmount().Return(3).AnyTimes()

		err := c.Construct(w)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid card for constructing")
	})

	t.Run("Error with invalid card type", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c, _, _ := newTestCastle(ctrl)

		card := mocks.NewMockCard(ctrl)

		err := c.Construct(card)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid card type for constructing")
	})

	t.Run("Add resource to already constructed castle", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c, _, _ := newTestCastle(ctrl)
		constructCastle(t, c, ctrl)

		gold := mocks.NewMockResource(ctrl)
		gold.EXPECT().Value().Return(4).AnyTimes()

		err := c.Construct(gold)
		assert.NoError(t, err)
		assert.Equal(t, 4, c.Value())
		assert.Equal(t, 1, c.ResourceCardsCount())
	})

	t.Run("Error adding non-resource to constructed castle", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c, _, _ := newTestCastle(ctrl)
		constructCastle(t, c, ctrl)

		card := mocks.NewMockCard(ctrl)

		err := c.Construct(card)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cardBase is not gold")
	})

	t.Run("Triggers completion when value reaches 25", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c, mockPlayer, castleObs := newTestCastle(ctrl)
		constructCastle(t, c, ctrl)

		castleObs.EXPECT().OnCastleCompletion(mockPlayer)

		gold := mocks.NewMockResource(ctrl)
		gold.EXPECT().Value().Return(25).AnyTimes()

		err := c.Construct(gold)
		assert.NoError(t, err)
	})

	t.Run("No completion when value below 25", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c, _, _ := newTestCastle(ctrl)
		constructCastle(t, c, ctrl)

		gold := mocks.NewMockResource(ctrl)
		gold.EXPECT().Value().Return(23).AnyTimes()

		err := c.Construct(gold)
		assert.NoError(t, err)
		assert.Equal(t, 23, c.Value())
	})
}

func TestCastle_Value(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c, _, _ := newTestCastle(ctrl)
	constructCastle(t, c, ctrl)

	g1 := mocks.NewMockResource(ctrl)
	g1.EXPECT().Value().Return(3).AnyTimes()
	g2 := mocks.NewMockResource(ctrl)
	g2.EXPECT().Value().Return(5).AnyTimes()

	c.Construct(g1)
	c.Construct(g2)

	assert.Equal(t, 8, c.Value())
	assert.Equal(t, 2, c.ResourceCardsCount())
}

func TestCastle_RemoveGold(t *testing.T) {
	t.Run("Remove gold from castle", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c, _, _ := newTestCastle(ctrl)
		constructCastle(t, c, ctrl)

		gold := mocks.NewMockResource(ctrl)
		gold.EXPECT().Value().Return(4).AnyTimes()
		gold.EXPECT().GetID().Return("g1").AnyTimes()
		c.Construct(gold)

		removed, err := c.RemoveGold(1)
		assert.NoError(t, err)
		assert.NotNil(t, removed)
		assert.Equal(t, 0, c.ResourceCardsCount())
	})

	t.Run("Error when no resources", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c, _, _ := newTestCastle(ctrl)
		constructCastle(t, c, ctrl)

		_, err := c.RemoveGold(1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no Resource cards to remove")
	})

	t.Run("Error with position 0", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c, _, _ := newTestCastle(ctrl)
		constructCastle(t, c, ctrl)

		gold := mocks.NewMockResource(ctrl)
		gold.EXPECT().Value().Return(2).AnyTimes()
		gold.EXPECT().GetID().Return("g1").AnyTimes()
		c.Construct(gold)

		_, err := c.RemoveGold(0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid position")
	})

	t.Run("Error with position exceeding count", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c, _, _ := newTestCastle(ctrl)
		constructCastle(t, c, ctrl)

		gold := mocks.NewMockResource(ctrl)
		gold.EXPECT().Value().Return(2).AnyTimes()
		gold.EXPECT().GetID().Return("g1").AnyTimes()
		c.Construct(gold)

		_, err := c.RemoveGold(2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid position")
	})
}

func TestCastle_CanBeAttacked(t *testing.T) {
	t.Run("False when not constructed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c, _, _ := newTestCastle(ctrl)

		assert.False(t, c.CanBeAttacked())
	})

	t.Run("False when constructed but no resources", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c, _, _ := newTestCastle(ctrl)
		constructCastle(t, c, ctrl)

		assert.False(t, c.CanBeAttacked())
	})

	t.Run("True when constructed with resources", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c, _, _ := newTestCastle(ctrl)
		constructCastle(t, c, ctrl)

		gold := mocks.NewMockResource(ctrl)
		gold.EXPECT().Value().Return(4).AnyTimes()
		c.Construct(gold)

		assert.True(t, c.CanBeAttacked())
	})
}

func TestCastle_String(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c, _, _ := newTestCastle(ctrl)
	constructCastle(t, c, ctrl)

	gold := mocks.NewMockResource(ctrl)
	gold.EXPECT().Value().Return(5).AnyTimes()
	c.Construct(gold)

	assert.Equal(t, "Castle: 5 Gold coins (1 cards)", c.String())
}

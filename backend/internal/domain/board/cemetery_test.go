package board_test

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/board"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCemetery_NewCemetery_IsEmpty(t *testing.T) {
	c := board.NewCemetery()
	assert.Equal(t, 0, c.Count())
	assert.Nil(t, c.GetLast())
	assert.Empty(t, c.Corps())
}

func TestCemetery_AddCorp_IncreasesCount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := board.NewCemetery()
	w := mocks.NewMockWarrior(ctrl)
	c.AddCorp(w)
	assert.Equal(t, 1, c.Count())
}

func TestCemetery_GetLast_ReturnsNilWhenEmpty(t *testing.T) {
	c := board.NewCemetery()
	assert.Nil(t, c.GetLast())
}

func TestCemetery_GetLast_ReturnsMostRecentCorp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := board.NewCemetery()
	w1 := mocks.NewMockWarrior(ctrl)
	w2 := mocks.NewMockWarrior(ctrl)
	c.AddCorp(w1)
	c.AddCorp(w2)
	assert.Equal(t, w2, c.GetLast())
}

func TestCemetery_Corps_ReturnsAllCorps(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := board.NewCemetery()
	w1 := mocks.NewMockWarrior(ctrl)
	w2 := mocks.NewMockWarrior(ctrl)
	c.AddCorp(w1)
	c.AddCorp(w2)

	corps := c.Corps()
	assert.Len(t, corps, 2)
	assert.Contains(t, corps, w1)
	assert.Contains(t, corps, w2)
}

func TestCemetery_RemoveRandom_ReturnsNilWhenEmpty(t *testing.T) {
	c := board.NewCemetery()
	assert.Nil(t, c.RemoveRandom())
}

func TestCemetery_RemoveRandom_SingleWarrior_RemovesAndReturnsIt(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := board.NewCemetery()
	w := mocks.NewMockWarrior(ctrl)
	c.AddCorp(w)

	result := c.RemoveRandom()
	assert.Equal(t, w, result)
	assert.Equal(t, 0, c.Count())
	assert.Empty(t, c.Corps())
}

func TestCemetery_RemoveRandom_DecreasesCountAndRemovesWarrior(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := board.NewCemetery()
	w1 := mocks.NewMockWarrior(ctrl)
	w2 := mocks.NewMockWarrior(ctrl)
	w3 := mocks.NewMockWarrior(ctrl)
	c.AddCorp(w1)
	c.AddCorp(w2)
	c.AddCorp(w3)

	removed := c.RemoveRandom()
	assert.Equal(t, 2, c.Count())

	// The removed warrior must not appear in the remaining corps.
	for _, corp := range c.Corps() {
		assert.False(t, corp == removed, "removed warrior should not appear in remaining corps")
	}
}

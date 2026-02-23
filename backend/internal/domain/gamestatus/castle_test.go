package gamestatus_test

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewCastle_NotConstructed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	castle := mocks.NewMockCastle(ctrl)
	castle.EXPECT().IsConstructed().Return(false)
	castle.EXPECT().ResourceCardsCount().Return(2)
	castle.EXPECT().Value().Return(5)

	c := NewCastle(castle)

	assert.False(t, c.IsConstructed)
	assert.Equal(t, 2, c.ResourceCards)
	assert.Equal(t, 5, c.Value)
}

func TestNewCastle_Constructed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	castle := mocks.NewMockCastle(ctrl)
	castle.EXPECT().IsConstructed().Return(true)
	castle.EXPECT().ResourceCardsCount().Return(5)
	castle.EXPECT().Value().Return(15)

	c := NewCastle(castle)

	assert.True(t, c.IsConstructed)
	assert.Equal(t, 5, c.ResourceCards)
	assert.Equal(t, 15, c.Value)
}

func TestNewCastle_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	castle := mocks.NewMockCastle(ctrl)
	castle.EXPECT().IsConstructed().Return(false)
	castle.EXPECT().ResourceCardsCount().Return(0)
	castle.EXPECT().Value().Return(0)

	c := NewCastle(castle)

	assert.False(t, c.IsConstructed)
	assert.Equal(t, 0, c.ResourceCards)
	assert.Equal(t, 0, c.Value)
}

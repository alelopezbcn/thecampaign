package gamestatus

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewCemetery_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cemetery := mocks.NewMockCemetery(ctrl)
	cemetery.EXPECT().Count().Return(0)
	cemetery.EXPECT().GetLast().Return(nil)

	c := NewCemetery(cemetery)

	assert.Equal(t, 0, c.Corps)
	assert.Equal(t, Card{}, c.LastCorp)
}

func TestNewCemetery_WithCorps(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lastWarrior := mocks.NewMockWarrior(ctrl)
	lastWarrior.EXPECT().GetID().Return("K1")
	lastWarrior.EXPECT().Type().Return(types.KnightWarriorType)
	lastWarrior.EXPECT().Health().Return(0)

	cemetery := mocks.NewMockCemetery(ctrl)
	cemetery.EXPECT().Count().Return(3)
	cemetery.EXPECT().GetLast().Return(lastWarrior)

	c := NewCemetery(cemetery)

	assert.Equal(t, 3, c.Corps)
	assert.Equal(t, "K1", c.LastCorp.CardID)
	assert.Equal(t, CardTypeKnight, c.LastCorp.CardType)
	assert.Equal(t, 0, c.LastCorp.Value)
}

package gamestatus_test

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewCemetery_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := gamestatus.NewCemetery(0, nil)

	assert.Equal(t, 0, c.Corps)
	assert.Equal(t, gamestatus.Card{}, c.LastCorp)
}

func TestNewCemetery_WithCorps(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lastWarrior := mocks.NewMockWarrior(ctrl)
	lastWarrior.EXPECT().GetID().Return("K1")
	lastWarrior.EXPECT().Type().Return(types.KnightWarriorType)
	lastWarrior.EXPECT().Health().Return(0)

	c := gamestatus.NewCemetery(3, lastWarrior)

	assert.Equal(t, 3, c.Corps)
	assert.Equal(t, "K1", c.LastCorp.CardID)
	assert.Equal(t, gamestatus.CardTypeKnight, c.LastCorp.CardType)
	assert.Equal(t, 0, c.LastCorp.Value)
}

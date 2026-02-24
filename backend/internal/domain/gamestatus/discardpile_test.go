package gamestatus_test

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewDiscardPile_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	d := gamestatus.NewDiscardPile(0, nil)

	assert.Equal(t, 0, d.Cards)
	assert.Equal(t, gamestatus.Card{}, d.LastCard)
}

func TestNewDiscardPile_WithWeaponCard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lastCard := mocks.NewMockWeapon(ctrl)
	lastCard.EXPECT().GetID().Return("S1")
	lastCard.EXPECT().Type().Return(types.SwordWeaponType)
	lastCard.EXPECT().DamageAmount().Return(7)

	d := gamestatus.NewDiscardPile(5, lastCard)

	assert.Equal(t, 5, d.Cards)
	assert.Equal(t, "S1", d.LastCard.CardID)
	assert.Equal(t, gamestatus.CardTypeSword, d.LastCard.CardType)
	assert.Equal(t, 7, d.LastCard.Value)
}

func TestNewDiscardPile_WithResourceCard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lastCard := mocks.NewMockResource(ctrl)
	lastCard.EXPECT().GetID().Return("G1")
	lastCard.EXPECT().Value().Return(3)

	d := gamestatus.NewDiscardPile(2, lastCard)

	assert.Equal(t, 2, d.Cards)
	assert.Equal(t, "G1", d.LastCard.CardID)
	assert.Equal(t, gamestatus.CardTypeResource, d.LastCard.CardType)
	assert.Equal(t, 3, d.LastCard.Value)
}

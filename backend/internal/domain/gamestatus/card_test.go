package gamestatus

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestFromDomainCard_Warrior(t *testing.T) {
	tests := []struct {
		name        string
		warriorType types.WarriorType
		wantType    CardType
	}{
		{"Knight", types.KnightWarriorType, CardTypeKnight},
		{"Archer", types.ArcherWarriorType, CardTypeArcher},
		{"Mage", types.MageWarriorType, CardTypeMage},
		{"Dragon", types.DragonWarriorType, CardTypeDragon},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			warrior := mocks.NewMockWarrior(ctrl)
			warrior.EXPECT().GetID().Return("W1")
			warrior.EXPECT().Type().Return(tt.warriorType)
			warrior.EXPECT().Health().Return(20)

			c := fromDomainCard(warrior)

			assert.Equal(t, "W1", c.CardID)
			assert.Equal(t, tt.wantType, c.CardType)
			assert.Equal(t, 20, c.Value)
		})
	}
}

func TestFromDomainCard_Weapon(t *testing.T) {
	tests := []struct {
		name       string
		weaponType types.WeaponType
		wantType   CardType
	}{
		{"Sword", types.SwordWeaponType, CardTypeSword},
		{"Arrow", types.ArrowWeaponType, CardTypeArrow},
		{"Poison", types.PoisonWeaponType, CardTypePoison},
		{"SpecialPower", types.SpecialPowerWeaponType, CardTypeSpecialPower},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			weapon := mocks.NewMockWeapon(ctrl)
			weapon.EXPECT().GetID().Return("WP1")
			weapon.EXPECT().Type().Return(tt.weaponType)
			weapon.EXPECT().DamageAmount().Return(7)

			c := fromDomainCard(weapon)

			assert.Equal(t, "WP1", c.CardID)
			assert.Equal(t, tt.wantType, c.CardType)
			assert.Equal(t, 7, c.Value)
		})
	}
}

func TestFromDomainCard_Resource(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resource := mocks.NewMockResource(ctrl)
	resource.EXPECT().GetID().Return("G1")
	resource.EXPECT().Value().Return(5)

	c := fromDomainCard(resource)

	assert.Equal(t, "G1", c.CardID)
	assert.Equal(t, CardTypeResource, c.CardType)
	assert.Equal(t, 5, c.Value)
}

func TestFromDomainCard_Spy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	spy := mocks.NewMockSpy(ctrl)
	spy.EXPECT().GetID().Return("SPY1")

	c := fromDomainCard(spy)

	assert.Equal(t, "SPY1", c.CardID)
	assert.Equal(t, CardTypeSpy, c.CardType)
	assert.Equal(t, 0, c.Value)
}

func TestFromDomainCard_Thief(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	thief := mocks.NewMockThief(ctrl)
	thief.EXPECT().GetID().Return("THIEF1")

	c := fromDomainCard(thief)

	assert.Equal(t, "THIEF1", c.CardID)
	assert.Equal(t, CardTypeThief, c.CardType)
	assert.Equal(t, 0, c.Value)
}

func TestFromDomainCard_Catapult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	catapult := mocks.NewMockCatapult(ctrl)
	catapult.EXPECT().GetID().Return("CAT1")

	c := fromDomainCard(catapult)

	assert.Equal(t, "CAT1", c.CardID)
	assert.Equal(t, CardTypeCatapult, c.CardType)
	assert.Equal(t, 0, c.Value)
}

func TestFromDomainCards(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	warrior := mocks.NewMockWarrior(ctrl)
	warrior.EXPECT().GetID().Return("K1")
	warrior.EXPECT().Type().Return(types.KnightWarriorType)
	warrior.EXPECT().Health().Return(20)

	resource := mocks.NewMockResource(ctrl)
	resource.EXPECT().GetID().Return("G1")
	resource.EXPECT().Value().Return(3)

	cards := fromDomainCards([]cards.Card{warrior, resource})

	assert.Len(t, cards, 2)
	assert.Equal(t, "K1", cards[0].CardID)
	assert.Equal(t, CardTypeKnight, cards[0].CardType)
	assert.Equal(t, "G1", cards[1].CardID)
	assert.Equal(t, CardTypeResource, cards[1].CardType)
}

func TestFromDomainCards_Empty(t *testing.T) {
	cards := fromDomainCards([]cards.Card{})

	assert.Empty(t, cards)
}

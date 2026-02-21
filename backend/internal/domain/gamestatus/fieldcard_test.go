package gamestatus

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewFieldCard_BasicWarriorTypes(t *testing.T) {
	tests := []struct {
		name       string
		warriorType types.WarriorType
		wantType   CardType
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
			warrior.EXPECT().Type().Return(tt.warriorType)
			warrior.EXPECT().GetID().Return("W1")
			warrior.EXPECT().Health().Return(18)
			warrior.EXPECT().AttackedBy().Return([]cards.Weapon{})
			warrior.EXPECT().IsProtected().Return(false, nil)

			fc := NewFieldCard(warrior)

			assert.Equal(t, "W1", fc.CardID)
			assert.Equal(t, tt.wantType, fc.CardType)
			assert.Equal(t, 18, fc.Value)
			assert.Empty(t, fc.AttackedBy)
			assert.Equal(t, Card{}, fc.ProtectedBy)
		})
	}
}

func TestNewFieldCard_WithAttackers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sword := mocks.NewMockWeapon(ctrl)
	arrow := mocks.NewMockWeapon(ctrl)

	sword.EXPECT().Type().Return(types.SwordWeaponType)
	sword.EXPECT().GetID().Return("S1")
	sword.EXPECT().DamageAmount().Return(7)

	arrow.EXPECT().Type().Return(types.ArrowWeaponType)
	arrow.EXPECT().GetID().Return("A1")
	arrow.EXPECT().DamageAmount().Return(5)

	warrior := mocks.NewMockWarrior(ctrl)
	warrior.EXPECT().Type().Return(types.KnightWarriorType)
	warrior.EXPECT().GetID().Return("K1")
	warrior.EXPECT().Health().Return(8)
	warrior.EXPECT().AttackedBy().Return([]cards.Weapon{sword, arrow})
	warrior.EXPECT().IsProtected().Return(false, nil)

	fc := NewFieldCard(warrior)

	assert.Equal(t, "K1", fc.CardID)
	assert.Equal(t, CardTypeKnight, fc.CardType)
	assert.Equal(t, 8, fc.Value)
	assert.Len(t, fc.AttackedBy, 2)
	assert.Equal(t, "S1", fc.AttackedBy[0].CardID)
	assert.Equal(t, CardTypeSword, fc.AttackedBy[0].CardType)
	assert.Equal(t, 7, fc.AttackedBy[0].Value)
	assert.Equal(t, "A1", fc.AttackedBy[1].CardID)
	assert.Equal(t, CardTypeArrow, fc.AttackedBy[1].CardType)
	assert.Equal(t, 5, fc.AttackedBy[1].Value)
}

func TestNewFieldCard_WithPoisonAttacker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	poison := mocks.NewMockWeapon(ctrl)
	poison.EXPECT().Type().Return(types.PoisonWeaponType)
	poison.EXPECT().GetID().Return("P1")
	poison.EXPECT().DamageAmount().Return(4)

	warrior := mocks.NewMockWarrior(ctrl)
	warrior.EXPECT().Type().Return(types.MageWarriorType)
	warrior.EXPECT().GetID().Return("M1")
	warrior.EXPECT().Health().Return(16)
	warrior.EXPECT().AttackedBy().Return([]cards.Weapon{poison})
	warrior.EXPECT().IsProtected().Return(false, nil)

	fc := NewFieldCard(warrior)

	assert.Len(t, fc.AttackedBy, 1)
	assert.Equal(t, CardTypePoison, fc.AttackedBy[0].CardType)
}

func TestNewFieldCard_WithProtection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sp := mocks.NewMockSpecialPower(ctrl)
	sp.EXPECT().GetID().Return("SP1")
	sp.EXPECT().Health().Return(10)

	warrior := mocks.NewMockWarrior(ctrl)
	warrior.EXPECT().Type().Return(types.ArcherWarriorType)
	warrior.EXPECT().GetID().Return("A1")
	warrior.EXPECT().Health().Return(20)
	warrior.EXPECT().AttackedBy().Return([]cards.Weapon{})
	warrior.EXPECT().IsProtected().Return(true, sp)

	fc := NewFieldCard(warrior)

	assert.Equal(t, "A1", fc.CardID)
	assert.Equal(t, CardTypeArcher, fc.CardType)
	assert.Equal(t, "SP1", fc.ProtectedBy.CardID)
	assert.Equal(t, CardTypeSpecialPower, fc.ProtectedBy.CardType)
	assert.Equal(t, 10, fc.ProtectedBy.Value)
}

package gamestatus

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewWarriorHandCard(t *testing.T) {
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
			warrior.EXPECT().Type().Return(tt.warriorType)
			warrior.EXPECT().GetID().Return("W1")
			warrior.EXPECT().Health().Return(20)

			hc := NewWarriorHandCard(warrior)

			assert.Equal(t, "W1", hc.CardID)
			assert.Equal(t, tt.wantType, hc.CardType)
			assert.Equal(t, 20, hc.Value)
			assert.True(t, hc.CanBeUsed)
			assert.Empty(t, hc.CanBeUsedOnIDs)
		})
	}
}

func TestNewWeaponHandCard(t *testing.T) {
	t.Run("Sword not usable outside Attack and Construct phases", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		weapon := mocks.NewMockWeapon(ctrl)
		myField := mocks.NewMockField(ctrl)
		enemyField := mocks.NewMockField(ctrl)

		weapon.EXPECT().Type().Return(types.SwordWeaponType)
		weapon.EXPECT().GetID().Return("S1")
		weapon.EXPECT().DamageAmount().Return(7)
		// HasKnight/HasDragon are called before action check
		myField.EXPECT().HasKnight().Return(true)

		hc := NewWeaponHandCard(weapon, myField, []ports.Field{enemyField}, false, types.PhaseTypeBuy)

		assert.Equal(t, "S1", hc.CardID)
		assert.Equal(t, CardTypeSword, hc.CardType)
		assert.Equal(t, 7, hc.Value)
		assert.False(t, hc.CanBeUsed)
		assert.True(t, hc.CanBeTraded)
	})

	t.Run("Arrow not usable in SpySteal phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		weapon := mocks.NewMockWeapon(ctrl)
		myField := mocks.NewMockField(ctrl)
		enemyField := mocks.NewMockField(ctrl)

		weapon.EXPECT().Type().Return(types.ArrowWeaponType)
		weapon.EXPECT().GetID().Return("A1")
		weapon.EXPECT().DamageAmount().Return(5)
		// HasArcher/HasDragon called before action check
		myField.EXPECT().HasArcher().Return(false)
		myField.EXPECT().HasDragon().Return(false)

		hc := NewWeaponHandCard(weapon, myField, []ports.Field{enemyField}, false, types.PhaseTypeSpySteal)

		assert.Equal(t, "A1", hc.CardID)
		assert.Equal(t, CardTypeArrow, hc.CardType)
		assert.False(t, hc.CanBeUsed)
		assert.True(t, hc.CanBeTraded)
	})

	t.Run("Poison maps to CardTypePoison", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		weapon := mocks.NewMockWeapon(ctrl)
		myField := mocks.NewMockField(ctrl)
		enemyField := mocks.NewMockField(ctrl)

		weapon.EXPECT().Type().Return(types.PoisonWeaponType)
		weapon.EXPECT().GetID().Return("P1")
		weapon.EXPECT().DamageAmount().Return(4)
		// HasMage/HasDragon called before action check
		myField.EXPECT().HasMage().Return(false)
		myField.EXPECT().HasDragon().Return(false)

		hc := NewWeaponHandCard(weapon, myField, []ports.Field{enemyField}, false, types.PhaseTypeBuy)

		assert.Equal(t, CardTypePoison, hc.CardType)
	})

	t.Run("Weapon can construct when CanConstruct and in Construct phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		weapon := mocks.NewMockWeapon(ctrl)
		myField := mocks.NewMockField(ctrl)
		enemyField := mocks.NewMockField(ctrl)

		weapon.EXPECT().Type().Return(types.PoisonWeaponType)
		weapon.EXPECT().GetID().Return("P1")
		weapon.EXPECT().DamageAmount().Return(1)
		weapon.EXPECT().CanConstruct().Return(true)
		// HasMage/HasDragon called before action check
		myField.EXPECT().HasMage().Return(false)
		myField.EXPECT().HasDragon().Return(false)

		hc := NewWeaponHandCard(weapon, myField, []ports.Field{enemyField}, false, types.PhaseTypeConstruct)

		assert.Equal(t, "P1", hc.CardID)
		assert.True(t, hc.CanBeUsed)
		assert.True(t, hc.CanBeTraded)
	})

	t.Run("Weapon cannot construct when castle already constructed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		weapon := mocks.NewMockWeapon(ctrl)
		myField := mocks.NewMockField(ctrl)
		enemyField := mocks.NewMockField(ctrl)

		weapon.EXPECT().Type().Return(types.SwordWeaponType)
		weapon.EXPECT().GetID().Return("S1")
		weapon.EXPECT().DamageAmount().Return(1)
		// Short-circuit: !castleConstructed is false, so CanConstruct() is NOT called
		// HasKnight/HasDragon called before action check
		myField.EXPECT().HasKnight().Return(false)
		myField.EXPECT().HasDragon().Return(false)

		hc := NewWeaponHandCard(weapon, myField, []ports.Field{enemyField}, true, types.PhaseTypeConstruct)

		assert.False(t, hc.CanBeUsed)
	})

	t.Run("Weapon cannot construct when CanConstruct is false", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		weapon := mocks.NewMockWeapon(ctrl)
		myField := mocks.NewMockField(ctrl)
		enemyField := mocks.NewMockField(ctrl)

		weapon.EXPECT().Type().Return(types.SwordWeaponType)
		weapon.EXPECT().GetID().Return("S1")
		weapon.EXPECT().DamageAmount().Return(5)
		weapon.EXPECT().CanConstruct().Return(false)
		// HasKnight/HasDragon called before action check
		myField.EXPECT().HasKnight().Return(true)

		hc := NewWeaponHandCard(weapon, myField, []ports.Field{enemyField}, false, types.PhaseTypeConstruct)

		assert.False(t, hc.CanBeUsed)
	})

	t.Run("Sword usable with Knight in Attack phase lists enemy warriors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		weapon := mocks.NewMockWeapon(ctrl)
		myField := mocks.NewMockField(ctrl)
		enemyField := mocks.NewMockField(ctrl)
		enemy1 := mocks.NewMockWarrior(ctrl)

		weapon.EXPECT().Type().Return(types.SwordWeaponType)
		weapon.EXPECT().GetID().Return("S1")
		weapon.EXPECT().DamageAmount().Return(7)
		weapon.EXPECT().MultiplierFactor(enemy1).Return(2)
		myField.EXPECT().HasKnight().Return(true)
		enemy1.EXPECT().GetID().Return("EK1").Times(2)
		enemyField.EXPECT().Warriors().Return([]ports.Warrior{enemy1})

		hc := NewWeaponHandCard(weapon, myField, []ports.Field{enemyField}, false, types.PhaseTypeAttack)

		assert.True(t, hc.CanBeUsed)
		assert.Equal(t, []string{"EK1"}, hc.CanBeUsedOnIDs)
		assert.Equal(t, 2, hc.DmgMultiplier["EK1"])
	})

	t.Run("Sword not usable without Knight or Dragon in Attack phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		weapon := mocks.NewMockWeapon(ctrl)
		myField := mocks.NewMockField(ctrl)
		enemyField := mocks.NewMockField(ctrl)

		weapon.EXPECT().Type().Return(types.SwordWeaponType)
		weapon.EXPECT().GetID().Return("S1")
		weapon.EXPECT().DamageAmount().Return(7)
		myField.EXPECT().HasKnight().Return(false)
		myField.EXPECT().HasDragon().Return(false)
		enemyField.EXPECT().Warriors().Return([]ports.Warrior{})

		hc := NewWeaponHandCard(weapon, myField, []ports.Field{enemyField}, false, types.PhaseTypeAttack)

		assert.False(t, hc.CanBeUsed)
	})

	t.Run("Arrow usable with Dragon in Attack phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		weapon := mocks.NewMockWeapon(ctrl)
		myField := mocks.NewMockField(ctrl)
		enemyField := mocks.NewMockField(ctrl)

		weapon.EXPECT().Type().Return(types.ArrowWeaponType)
		weapon.EXPECT().GetID().Return("A1")
		weapon.EXPECT().DamageAmount().Return(5)
		myField.EXPECT().HasArcher().Return(false)
		myField.EXPECT().HasDragon().Return(true)
		enemyField.EXPECT().Warriors().Return([]ports.Warrior{})

		hc := NewWeaponHandCard(weapon, myField, []ports.Field{enemyField}, false, types.PhaseTypeAttack)

		assert.True(t, hc.CanBeUsed)
	})
}

func TestNewResourceHandCard(t *testing.T) {
	t.Run("Resource not usable outside Buy and Construct phases", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		resource := mocks.NewMockResource(ctrl)
		resource.EXPECT().GetID().Return("G1")
		resource.EXPECT().Value().Return(5)

		hc := NewResourceHandCard(resource, false, false, true, types.PhaseTypeAttack)

		assert.Equal(t, "G1", hc.CardID)
		assert.Equal(t, CardTypeResource, hc.CardType)
		assert.Equal(t, 5, hc.Value)
		assert.False(t, hc.CanBeUsed)
	})

	t.Run("Resource usable when canBuy in Buy phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		resource := mocks.NewMockResource(ctrl)
		resource.EXPECT().GetID().Return("G1")
		resource.EXPECT().Value().Return(5)

		hc := NewResourceHandCard(resource, false, false, true, types.PhaseTypeBuy)

		assert.Equal(t, "G1", hc.CardID)
		assert.True(t, hc.CanBeUsed)
	})

	t.Run("Resource not usable when canBuy is false in Buy phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		resource := mocks.NewMockResource(ctrl)
		resource.EXPECT().GetID().Return("G1")
		resource.EXPECT().Value().Return(1)

		hc := NewResourceHandCard(resource, false, false, false, types.PhaseTypeBuy)

		assert.False(t, hc.CanBeUsed)
	})

	t.Run("Resource usable in Construct phase when castle constructed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		resource := mocks.NewMockResource(ctrl)
		resource.EXPECT().GetID().Return("G1")
		resource.EXPECT().Value().Return(5)

		hc := NewResourceHandCard(resource, true, false, false, types.PhaseTypeConstruct)

		assert.True(t, hc.CanBeUsed) // Castle constructed, any resource can be added
	})

	t.Run("Resource usable in Construct phase when CanConstruct and castle not started", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		resource := mocks.NewMockResource(ctrl)
		resource.EXPECT().GetID().Return("G1")
		resource.EXPECT().Value().Return(1)
		resource.EXPECT().CanConstruct().Return(true)

		hc := NewResourceHandCard(resource, false, false, false, types.PhaseTypeConstruct)

		assert.True(t, hc.CanBeUsed) // CanConstruct is true
	})

	t.Run("Resource not usable in Construct phase when cannot construct and castle not started", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		resource := mocks.NewMockResource(ctrl)
		resource.EXPECT().GetID().Return("G1")
		resource.EXPECT().Value().Return(5)
		resource.EXPECT().CanConstruct().Return(false)

		hc := NewResourceHandCard(resource, false, false, false, types.PhaseTypeConstruct)

		assert.False(t, hc.CanBeUsed)
	})

	t.Run("Resource usable in Construct phase when ally castle is constructed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		resource := mocks.NewMockResource(ctrl)
		resource.EXPECT().GetID().Return("G1")
		resource.EXPECT().Value().Return(5)

		// Player castle NOT constructed, but ally castle IS
		hc := NewResourceHandCard(resource, false, true, false, types.PhaseTypeConstruct)

		assert.True(t, hc.CanBeUsed) // Ally castle constructed, any resource can be added
	})
}

func TestNewSpecialPowerHandCard(t *testing.T) {
	t.Run("Not usable outside Attack phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		sp := mocks.NewMockSpecialPower(ctrl)
		myField := mocks.NewMockField(ctrl)

		sp.EXPECT().GetID().Return("SP1")

		hc := NewSpecialPowerHandCard(sp, myField, []ports.Field{}, []ports.Field{}, types.PhaseTypeBuy)

		assert.Equal(t, "SP1", hc.CardID)
		assert.Equal(t, CardTypeSpecialPower, hc.CardType)
		assert.False(t, hc.CanBeUsed)
		assert.Empty(t, hc.CanBeUsedOnIDs)
	})

	t.Run("Archer can target enemy warriors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		sp := mocks.NewMockSpecialPower(ctrl)
		myField := mocks.NewMockField(ctrl)
		enemyField := mocks.NewMockField(ctrl)
		enemy1 := mocks.NewMockWarrior(ctrl)

		sp.EXPECT().GetID().Return("SP1")
		myField.EXPECT().HasArcher().Return(true)
		myField.EXPECT().HasKnight().Return(false)
		myField.EXPECT().HasMage().Return(false)
		enemy1.EXPECT().GetID().Return("EK1")
		enemyField.EXPECT().Warriors().Return([]ports.Warrior{enemy1})

		hc := NewSpecialPowerHandCard(sp, myField, []ports.Field{}, []ports.Field{enemyField}, types.PhaseTypeAttack)

		assert.True(t, hc.CanBeUsed)
		assert.Equal(t, []string{"EK1"}, hc.CanBeUsedOnIDs)
	})

	t.Run("Knight can protect own unprotected non-dragon warriors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		sp := mocks.NewMockSpecialPower(ctrl)
		myField := mocks.NewMockField(ctrl)
		myWarrior := mocks.NewMockWarrior(ctrl)

		sp.EXPECT().GetID().Return("SP1")
		myField.EXPECT().HasArcher().Return(false)
		myField.EXPECT().HasKnight().Return(true)
		myField.EXPECT().HasMage().Return(false)
		myField.EXPECT().Warriors().Return([]ports.Warrior{myWarrior})
		myWarrior.EXPECT().Type().Return(types.KnightWarriorType)
		myWarrior.EXPECT().IsProtected().Return(false, nil)
		myWarrior.EXPECT().GetID().Return("K1")

		hc := NewSpecialPowerHandCard(sp, myField, []ports.Field{}, []ports.Field{}, types.PhaseTypeAttack)

		assert.True(t, hc.CanBeUsed)
		assert.Equal(t, []string{"K1"}, hc.CanBeUsedOnIDs)
	})

	t.Run("Knight cannot protect dragon", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		sp := mocks.NewMockSpecialPower(ctrl)
		myField := mocks.NewMockField(ctrl)
		dragon := mocks.NewMockWarrior(ctrl)

		sp.EXPECT().GetID().Return("SP1")
		myField.EXPECT().HasArcher().Return(false)
		myField.EXPECT().HasKnight().Return(true)
		myField.EXPECT().HasMage().Return(false)
		myField.EXPECT().Warriors().Return([]ports.Warrior{dragon})
		// IsProtected() is called before Type() check
		dragon.EXPECT().IsProtected().Return(false, nil)
		dragon.EXPECT().Type().Return(types.DragonWarriorType)

		hc := NewSpecialPowerHandCard(sp, myField, []ports.Field{}, []ports.Field{}, types.PhaseTypeAttack)

		assert.True(t, hc.CanBeUsed)
		assert.Empty(t, hc.CanBeUsedOnIDs)
	})

	t.Run("Knight cannot protect already protected warrior", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		sp := mocks.NewMockSpecialPower(ctrl)
		myField := mocks.NewMockField(ctrl)
		protectedWarrior := mocks.NewMockWarrior(ctrl)
		existingSP := mocks.NewMockSpecialPower(ctrl)

		sp.EXPECT().GetID().Return("SP1")
		myField.EXPECT().HasArcher().Return(false)
		myField.EXPECT().HasKnight().Return(true)
		myField.EXPECT().HasMage().Return(false)
		myField.EXPECT().Warriors().Return([]ports.Warrior{protectedWarrior})
		protectedWarrior.EXPECT().Type().Return(types.KnightWarriorType)
		protectedWarrior.EXPECT().IsProtected().Return(true, existingSP)

		hc := NewSpecialPowerHandCard(sp, myField, []ports.Field{}, []ports.Field{}, types.PhaseTypeAttack)

		assert.True(t, hc.CanBeUsed)
		assert.Empty(t, hc.CanBeUsedOnIDs)
	})

	t.Run("Mage can heal damaged non-dragon warriors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		sp := mocks.NewMockSpecialPower(ctrl)
		myField := mocks.NewMockField(ctrl)
		damagedWarrior := mocks.NewMockWarrior(ctrl)

		sp.EXPECT().GetID().Return("SP1")
		myField.EXPECT().HasArcher().Return(false)
		myField.EXPECT().HasKnight().Return(false)
		myField.EXPECT().HasMage().Return(true)
		myField.EXPECT().Warriors().Return([]ports.Warrior{damagedWarrior})
		damagedWarrior.EXPECT().Type().Return(types.MageWarriorType)
		damagedWarrior.EXPECT().IsDamaged().Return(true)
		damagedWarrior.EXPECT().GetID().Return("M1")

		hc := NewSpecialPowerHandCard(sp, myField, []ports.Field{}, []ports.Field{}, types.PhaseTypeAttack)

		assert.True(t, hc.CanBeUsed)
		assert.Equal(t, []string{"M1"}, hc.CanBeUsedOnIDs)
	})

	t.Run("Mage cannot heal undamaged warrior", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		sp := mocks.NewMockSpecialPower(ctrl)
		myField := mocks.NewMockField(ctrl)
		healthyWarrior := mocks.NewMockWarrior(ctrl)

		sp.EXPECT().GetID().Return("SP1")
		myField.EXPECT().HasArcher().Return(false)
		myField.EXPECT().HasKnight().Return(false)
		myField.EXPECT().HasMage().Return(true)
		myField.EXPECT().Warriors().Return([]ports.Warrior{healthyWarrior})
		healthyWarrior.EXPECT().Type().Return(types.ArcherWarriorType)
		healthyWarrior.EXPECT().IsDamaged().Return(false)

		hc := NewSpecialPowerHandCard(sp, myField, []ports.Field{}, []ports.Field{}, types.PhaseTypeAttack)

		assert.True(t, hc.CanBeUsed)
		assert.Empty(t, hc.CanBeUsedOnIDs)
	})

	t.Run("Mage cannot heal dragon", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		sp := mocks.NewMockSpecialPower(ctrl)
		myField := mocks.NewMockField(ctrl)
		dragon := mocks.NewMockWarrior(ctrl)

		sp.EXPECT().GetID().Return("SP1")
		myField.EXPECT().HasArcher().Return(false)
		myField.EXPECT().HasKnight().Return(false)
		myField.EXPECT().HasMage().Return(true)
		myField.EXPECT().Warriors().Return([]ports.Warrior{dragon})
		dragon.EXPECT().Type().Return(types.DragonWarriorType)

		hc := NewSpecialPowerHandCard(sp, myField, []ports.Field{}, []ports.Field{}, types.PhaseTypeAttack)

		assert.True(t, hc.CanBeUsed)
		assert.Empty(t, hc.CanBeUsedOnIDs)
	})

	t.Run("No field warriors means no targets", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		sp := mocks.NewMockSpecialPower(ctrl)
		myField := mocks.NewMockField(ctrl)

		sp.EXPECT().GetID().Return("SP1")
		myField.EXPECT().HasArcher().Return(false)
		myField.EXPECT().HasKnight().Return(false)
		myField.EXPECT().HasMage().Return(false)

		hc := NewSpecialPowerHandCard(sp, myField, []ports.Field{}, []ports.Field{}, types.PhaseTypeAttack)

		assert.True(t, hc.CanBeUsed)
		assert.Empty(t, hc.CanBeUsedOnIDs)
	})
}

func TestNewSpyHandCard(t *testing.T) {
	tests := []struct {
		name     string
		cardID   string
		action   types.PhaseType
		wantUsed bool
		wantType CardType
	}{
		{
			name:     "Spy can be used during SpySteal phase",
			cardID:   "SPY1",
			action:   types.PhaseTypeSpySteal,
			wantUsed: true,
			wantType: CardTypeSpy,
		},
		{
			name:     "Spy cannot be used during Attack phase",
			cardID:   "SPY2",
			action:   types.PhaseTypeAttack,
			wantUsed: false,
			wantType: CardTypeSpy,
		},
		{
			name:     "Spy cannot be used during Buy phase",
			cardID:   "SPY3",
			action:   types.PhaseTypeBuy,
			wantUsed: false,
			wantType: CardTypeSpy,
		},
		{
			name:     "Spy cannot be used during Construct phase",
			cardID:   "SPY4",
			action:   types.PhaseTypeConstruct,
			wantUsed: false,
			wantType: CardTypeSpy,
		},
		{
			name:     "Spy cannot be used during DrawCard phase",
			cardID:   "SPY5",
			action:   types.PhaseTypeDrawCard,
			wantUsed: false,
			wantType: CardTypeSpy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc := NewSpyHandCard(tt.cardID, tt.action)

			assert.Equal(t, tt.cardID, hc.CardID)
			assert.Equal(t, tt.wantType, hc.CardType)
			assert.Equal(t, tt.wantUsed, hc.CanBeUsed)
			assert.Equal(t, 0, hc.Value)
			assert.Empty(t, hc.CanBeUsedOnIDs)
		})
	}
}

func TestNewThiefHandCard(t *testing.T) {
	tests := []struct {
		name     string
		cardID   string
		action   types.PhaseType
		wantUsed bool
		wantType CardType
	}{
		{
			name:     "Thief can be used during SpySteal phase",
			cardID:   "THIEF1",
			action:   types.PhaseTypeSpySteal,
			wantUsed: true,
			wantType: CardTypeThief,
		},
		{
			name:     "Thief cannot be used during Attack phase",
			cardID:   "THIEF2",
			action:   types.PhaseTypeAttack,
			wantUsed: false,
			wantType: CardTypeThief,
		},
		{
			name:     "Thief cannot be used during Buy phase",
			cardID:   "THIEF3",
			action:   types.PhaseTypeBuy,
			wantUsed: false,
			wantType: CardTypeThief,
		},
		{
			name:     "Thief cannot be used during Construct phase",
			cardID:   "THIEF4",
			action:   types.PhaseTypeConstruct,
			wantUsed: false,
			wantType: CardTypeThief,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc := NewThiefHandCard(tt.cardID, tt.action)

			assert.Equal(t, tt.cardID, hc.CardID)
			assert.Equal(t, tt.wantType, hc.CardType)
			assert.Equal(t, tt.wantUsed, hc.CanBeUsed)
			assert.Equal(t, 0, hc.Value)
			assert.Empty(t, hc.CanBeUsedOnIDs)
		})
	}
}

func TestNewCatapultHandCard(t *testing.T) {
	tests := []struct {
		name      string
		cardID    string
		canBeUsed bool
		action    types.PhaseType
		wantUsed  bool
		wantType  CardType
	}{
		{
			name:      "Catapult can be used when enemy castle can be attacked in Attack phase",
			cardID:    "CAT1",
			canBeUsed: true,
			action:    types.PhaseTypeAttack,
			wantUsed:  true,
			wantType:  CardTypeCatapult,
		},
		{
			name:      "Catapult cannot be used when enemy castle cannot be attacked in Attack phase",
			cardID:    "CAT2",
			canBeUsed: false,
			action:    types.PhaseTypeAttack,
			wantUsed:  false,
			wantType:  CardTypeCatapult,
		},
		{
			name:      "Catapult cannot be used during Buy phase even if castle can be attacked",
			cardID:    "CAT3",
			canBeUsed: true,
			action:    types.PhaseTypeBuy,
			wantUsed:  false,
			wantType:  CardTypeCatapult,
		},
		{
			name:      "Catapult cannot be used during SpySteal phase",
			cardID:    "CAT4",
			canBeUsed: true,
			action:    types.PhaseTypeSpySteal,
			wantUsed:  false,
			wantType:  CardTypeCatapult,
		},
		{
			name:      "Catapult cannot be used during Construct phase",
			cardID:    "CAT5",
			canBeUsed: true,
			action:    types.PhaseTypeConstruct,
			wantUsed:  false,
			wantType:  CardTypeCatapult,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc := NewCatapultHandCard(tt.cardID, tt.canBeUsed, tt.action)

			assert.Equal(t, tt.cardID, hc.CardID)
			assert.Equal(t, tt.wantType, hc.CardType)
			assert.Equal(t, tt.wantUsed, hc.CanBeUsed)
			assert.Equal(t, 0, hc.Value)
			assert.Empty(t, hc.CanBeUsedOnIDs)
		})
	}
}

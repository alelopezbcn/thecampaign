package gamestatus_test

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/gamestatus"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/alelopezbcn/thecampaign/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewWarriorHandCard(t *testing.T) {
	tests := []struct {
		name        string
		warriorType types.WarriorType
		wantType    gamestatus.CardType
	}{
		{"Knight", types.KnightWarriorType, gamestatus.CardTypeKnight},
		{"Archer", types.ArcherWarriorType, gamestatus.CardTypeArcher},
		{"Mage", types.MageWarriorType, gamestatus.CardTypeMage},
		{"Dragon", types.DragonWarriorType, gamestatus.CardTypeDragon},
		{"Mercenary", types.MercenaryWarriorType, gamestatus.CardTypeMercenary},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			warrior := mocks.NewMockWarrior(ctrl)
			warrior.EXPECT().Type().Return(tt.warriorType)
			warrior.EXPECT().GetID().Return("W1")
			warrior.EXPECT().Health().Return(20)

			hc := gamestatus.NewWarriorHandCard(warrior)

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
		weapon.EXPECT().Type().Return(types.SwordWeaponType)
		weapon.EXPECT().GetID().Return("S1")
		weapon.EXPECT().DamageAmount().Return(7)
		weapon.EXPECT().CanBeTraded().Return(true)

		myField := gamestatus.FieldInput{HasKnight: true}

		hc := gamestatus.NewWeaponHandCard(weapon, myField, []gamestatus.FieldInput{{}}, false, types.PhaseTypeBuy)

		assert.Equal(t, "S1", hc.CardID)
		assert.Equal(t, gamestatus.CardTypeSword, hc.CardType)
		assert.Equal(t, 7, hc.Value)
		assert.False(t, hc.CanBeUsed)
		assert.True(t, hc.CanBeTraded)
	})

	t.Run("Arrow not usable in SpySteal phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		weapon := mocks.NewMockWeapon(ctrl)
		weapon.EXPECT().Type().Return(types.ArrowWeaponType)
		weapon.EXPECT().GetID().Return("A1")
		weapon.EXPECT().DamageAmount().Return(5)
		weapon.EXPECT().CanBeTraded().Return(true)

		myField := gamestatus.FieldInput{HasArcher: false, HasDragon: false}

		hc := gamestatus.NewWeaponHandCard(weapon, myField, []gamestatus.FieldInput{{}}, false, types.PhaseTypeSpySteal)

		assert.Equal(t, "A1", hc.CardID)
		assert.Equal(t, gamestatus.CardTypeArrow, hc.CardType)
		assert.False(t, hc.CanBeUsed)
		assert.True(t, hc.CanBeTraded)
	})

	t.Run("Poison maps to CardTypePoison", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		weapon := mocks.NewMockWeapon(ctrl)
		weapon.EXPECT().Type().Return(types.PoisonWeaponType)
		weapon.EXPECT().GetID().Return("P1")
		weapon.EXPECT().DamageAmount().Return(4)
		weapon.EXPECT().CanBeTraded().Return(true)

		myField := gamestatus.FieldInput{HasMage: false, HasDragon: false}

		hc := gamestatus.NewWeaponHandCard(weapon, myField, []gamestatus.FieldInput{{}}, false, types.PhaseTypeBuy)

		assert.Equal(t, gamestatus.CardTypePoison, hc.CardType)
	})

	t.Run("Weapon can construct when CanConstruct and in Construct phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		weapon := mocks.NewMockWeapon(ctrl)
		weapon.EXPECT().Type().Return(types.PoisonWeaponType)
		weapon.EXPECT().GetID().Return("P1")
		weapon.EXPECT().DamageAmount().Return(1)
		weapon.EXPECT().CanConstruct().Return(true)
		weapon.EXPECT().CanBeTraded().Return(true)

		myField := gamestatus.FieldInput{HasMage: false, HasDragon: false}

		hc := gamestatus.NewWeaponHandCard(weapon, myField, []gamestatus.FieldInput{{}}, false, types.PhaseTypeConstruct)

		assert.Equal(t, "P1", hc.CardID)
		assert.True(t, hc.CanBeUsed)
		assert.True(t, hc.CanBeTraded)
	})

	t.Run("Weapon cannot construct when castle already constructed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		weapon := mocks.NewMockWeapon(ctrl)
		weapon.EXPECT().Type().Return(types.SwordWeaponType)
		weapon.EXPECT().GetID().Return("S1")
		weapon.EXPECT().DamageAmount().Return(1)
		// Short-circuit: castleConstructed=true, so CanConstruct() is NOT called

		myField := gamestatus.FieldInput{HasKnight: false, HasDragon: false}

		weapon.EXPECT().CanBeTraded().Return(true)

		hc := gamestatus.NewWeaponHandCard(weapon, myField, []gamestatus.FieldInput{{}}, true, types.PhaseTypeConstruct)

		assert.False(t, hc.CanBeUsed)
	})

	t.Run("Weapon cannot construct when CanConstruct is false", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		weapon := mocks.NewMockWeapon(ctrl)
		weapon.EXPECT().Type().Return(types.SwordWeaponType)
		weapon.EXPECT().GetID().Return("S1")
		weapon.EXPECT().DamageAmount().Return(5)
		weapon.EXPECT().CanConstruct().Return(false)
		weapon.EXPECT().CanBeTraded().Return(true)

		myField := gamestatus.FieldInput{HasKnight: true}

		hc := gamestatus.NewWeaponHandCard(weapon, myField, []gamestatus.FieldInput{{}}, false, types.PhaseTypeConstruct)

		assert.False(t, hc.CanBeUsed)
	})

	t.Run("Sword usable with Knight in Attack phase lists enemy warriors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		weapon := mocks.NewMockWeapon(ctrl)
		enemy1 := mocks.NewMockWarrior(ctrl)

		weapon.EXPECT().Type().Return(types.SwordWeaponType)
		weapon.EXPECT().GetID().Return("S1")
		weapon.EXPECT().DamageAmount().Return(7)
		weapon.EXPECT().MultiplierFactor(enemy1).Return(2)
		weapon.EXPECT().CanBeTraded().Return(true)
		enemy1.EXPECT().GetID().Return("EK1").Times(2)

		myField := gamestatus.FieldInput{HasKnight: true}
		enemyField := gamestatus.FieldInput{Warriors: []cards.Warrior{enemy1}}

		hc := gamestatus.NewWeaponHandCard(weapon, myField, []gamestatus.FieldInput{enemyField}, false, types.PhaseTypeAttack)

		assert.True(t, hc.CanBeUsed)
		assert.Equal(t, []string{"EK1"}, hc.CanBeUsedOnIDs)
		assert.Equal(t, 2, hc.DmgMultiplier["EK1"])
	})

	t.Run("Sword not usable without Knight or Dragon in Attack phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		weapon := mocks.NewMockWeapon(ctrl)
		weapon.EXPECT().Type().Return(types.SwordWeaponType)
		weapon.EXPECT().GetID().Return("S1")
		weapon.EXPECT().DamageAmount().Return(7)
		weapon.EXPECT().CanBeTraded().Return(true)

		myField := gamestatus.FieldInput{HasKnight: false, HasDragon: false}
		enemyField := gamestatus.FieldInput{Warriors: []cards.Warrior{}}

		hc := gamestatus.NewWeaponHandCard(weapon, myField, []gamestatus.FieldInput{enemyField}, false, types.PhaseTypeAttack)

		assert.False(t, hc.CanBeUsed)
	})

	t.Run("Arrow usable with Dragon in Attack phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		weapon := mocks.NewMockWeapon(ctrl)
		weapon.EXPECT().Type().Return(types.ArrowWeaponType)
		weapon.EXPECT().GetID().Return("A1")
		weapon.EXPECT().DamageAmount().Return(5)
		weapon.EXPECT().CanBeTraded().Return(true)

		myField := gamestatus.FieldInput{HasArcher: false, HasDragon: true}
		enemyField := gamestatus.FieldInput{Warriors: []cards.Warrior{}}

		hc := gamestatus.NewWeaponHandCard(weapon, myField, []gamestatus.FieldInput{enemyField}, false, types.PhaseTypeAttack)

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

		hc := gamestatus.NewResourceHandCard(resource, false, false, true, types.PhaseTypeAttack)

		assert.Equal(t, "G1", hc.CardID)
		assert.Equal(t, gamestatus.CardTypeResource, hc.CardType)
		assert.Equal(t, 5, hc.Value)
		assert.False(t, hc.CanBeUsed)
	})

	t.Run("Resource usable when canBuy in Buy phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		resource := mocks.NewMockResource(ctrl)
		resource.EXPECT().GetID().Return("G1")
		resource.EXPECT().Value().Return(5)

		hc := gamestatus.NewResourceHandCard(resource, false, false, true, types.PhaseTypeBuy)

		assert.Equal(t, "G1", hc.CardID)
		assert.True(t, hc.CanBeUsed)
	})

	t.Run("Resource not usable when canBuy is false in Buy phase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		resource := mocks.NewMockResource(ctrl)
		resource.EXPECT().GetID().Return("G1")
		resource.EXPECT().Value().Return(1)

		hc := gamestatus.NewResourceHandCard(resource, false, false, false, types.PhaseTypeBuy)

		assert.False(t, hc.CanBeUsed)
	})

	t.Run("Resource usable in Construct phase when castle constructed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		resource := mocks.NewMockResource(ctrl)
		resource.EXPECT().GetID().Return("G1")
		resource.EXPECT().Value().Return(5)

		hc := gamestatus.NewResourceHandCard(resource, true, false, false, types.PhaseTypeConstruct)

		assert.True(t, hc.CanBeUsed) // Castle constructed, any resource can be added
	})

	t.Run("Resource usable in Construct phase when CanConstruct and castle not started", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		resource := mocks.NewMockResource(ctrl)
		resource.EXPECT().GetID().Return("G1")
		resource.EXPECT().Value().Return(1)
		resource.EXPECT().CanConstruct().Return(true)

		hc := gamestatus.NewResourceHandCard(resource, false, false, false, types.PhaseTypeConstruct)

		assert.True(t, hc.CanBeUsed) // CanConstruct is true
	})

	t.Run("Resource not usable in Construct phase when cannot construct and castle not started", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		resource := mocks.NewMockResource(ctrl)
		resource.EXPECT().GetID().Return("G1")
		resource.EXPECT().Value().Return(5)
		resource.EXPECT().CanConstruct().Return(false)

		hc := gamestatus.NewResourceHandCard(resource, false, false, false, types.PhaseTypeConstruct)

		assert.False(t, hc.CanBeUsed)
	})

	t.Run("Resource usable in Construct phase when ally castle is constructed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		resource := mocks.NewMockResource(ctrl)
		resource.EXPECT().GetID().Return("G1")
		resource.EXPECT().Value().Return(5)

		// Player castle NOT constructed, but ally castle IS
		hc := gamestatus.NewResourceHandCard(resource, false, true, false, types.PhaseTypeConstruct)

		assert.True(t, hc.CanBeUsed) // Ally castle constructed, any resource can be added
	})
}

func TestNewSpecialPowerHandCard(t *testing.T) {
	t.Run("Not usable outside Attack phase", func(t *testing.T) {
		hc := gamestatus.NewSpecialPowerHandCard("SP1",
			gamestatus.FieldInput{}, []gamestatus.FieldInput{}, []gamestatus.FieldInput{},
			types.PhaseTypeBuy)

		assert.Equal(t, "SP1", hc.CardID)
		assert.Equal(t, gamestatus.CardTypeSpecialPower, hc.CardType)
		assert.False(t, hc.CanBeUsed)
		assert.Empty(t, hc.CanBeUsedOnIDs)
	})

	t.Run("Archer can target enemy warriors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		enemy1 := mocks.NewMockWarrior(ctrl)
		enemy1.EXPECT().GetID().Return("EK1")

		myField := gamestatus.FieldInput{HasArcher: true, HasKnight: false, HasMage: false}
		enemyField := gamestatus.FieldInput{Warriors: []cards.Warrior{enemy1}}

		hc := gamestatus.NewSpecialPowerHandCard("SP1",
			myField, []gamestatus.FieldInput{}, []gamestatus.FieldInput{enemyField},
			types.PhaseTypeAttack)

		assert.True(t, hc.CanBeUsed)
		assert.Equal(t, []string{"EK1"}, hc.CanBeUsedOnIDs)
	})

	t.Run("Knight can protect own unprotected non-dragon warriors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		myWarrior := mocks.NewMockWarrior(ctrl)
		myWarrior.EXPECT().Type().Return(types.KnightWarriorType)
		myWarrior.EXPECT().IsProtected().Return(false, nil)
		myWarrior.EXPECT().GetID().Return("K1")

		myField := gamestatus.FieldInput{
			HasArcher: false, HasKnight: true, HasMage: false,
			Warriors: []cards.Warrior{myWarrior},
		}

		hc := gamestatus.NewSpecialPowerHandCard("SP1",
			myField, []gamestatus.FieldInput{}, []gamestatus.FieldInput{},
			types.PhaseTypeAttack)

		assert.True(t, hc.CanBeUsed)
		assert.Equal(t, []string{"K1"}, hc.CanBeUsedOnIDs)
	})

	t.Run("Knight cannot protect dragon", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dragon := mocks.NewMockWarrior(ctrl)
		dragon.EXPECT().IsProtected().Return(false, nil)
		dragon.EXPECT().Type().Return(types.DragonWarriorType)

		myField := gamestatus.FieldInput{
			HasArcher: false, HasKnight: true, HasMage: false,
			Warriors: []cards.Warrior{dragon},
		}

		hc := gamestatus.NewSpecialPowerHandCard("SP1",
			myField, []gamestatus.FieldInput{}, []gamestatus.FieldInput{},
			types.PhaseTypeAttack)

		assert.True(t, hc.CanBeUsed)
		assert.Empty(t, hc.CanBeUsedOnIDs)
	})

	t.Run("Knight cannot protect already protected warrior", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		protectedWarrior := mocks.NewMockWarrior(ctrl)
		existingSP := mocks.NewMockSpecialPower(ctrl)
		protectedWarrior.EXPECT().Type().Return(types.KnightWarriorType)
		protectedWarrior.EXPECT().IsProtected().Return(true, existingSP)

		myField := gamestatus.FieldInput{
			HasArcher: false, HasKnight: true, HasMage: false,
			Warriors: []cards.Warrior{protectedWarrior},
		}

		hc := gamestatus.NewSpecialPowerHandCard("SP1",
			myField, []gamestatus.FieldInput{}, []gamestatus.FieldInput{},
			types.PhaseTypeAttack)

		assert.True(t, hc.CanBeUsed)
		assert.Empty(t, hc.CanBeUsedOnIDs)
	})

	t.Run("Mage can heal damaged non-dragon warriors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		damagedWarrior := mocks.NewMockWarrior(ctrl)
		damagedWarrior.EXPECT().Type().Return(types.MageWarriorType)
		damagedWarrior.EXPECT().IsDamaged().Return(true)
		damagedWarrior.EXPECT().GetID().Return("M1")

		myField := gamestatus.FieldInput{
			HasArcher: false, HasKnight: false, HasMage: true,
			Warriors: []cards.Warrior{damagedWarrior},
		}

		hc := gamestatus.NewSpecialPowerHandCard("SP1",
			myField, []gamestatus.FieldInput{}, []gamestatus.FieldInput{},
			types.PhaseTypeAttack)

		assert.True(t, hc.CanBeUsed)
		assert.Equal(t, []string{"M1"}, hc.CanBeUsedOnIDs)
	})

	t.Run("Mage cannot heal undamaged warrior", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		healthyWarrior := mocks.NewMockWarrior(ctrl)
		healthyWarrior.EXPECT().Type().Return(types.ArcherWarriorType)
		healthyWarrior.EXPECT().IsDamaged().Return(false)

		myField := gamestatus.FieldInput{
			HasArcher: false, HasKnight: false, HasMage: true,
			Warriors: []cards.Warrior{healthyWarrior},
		}

		hc := gamestatus.NewSpecialPowerHandCard("SP1",
			myField, []gamestatus.FieldInput{}, []gamestatus.FieldInput{},
			types.PhaseTypeAttack)

		assert.True(t, hc.CanBeUsed)
		assert.Empty(t, hc.CanBeUsedOnIDs)
	})

	t.Run("Mage cannot heal dragon", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dragon := mocks.NewMockWarrior(ctrl)
		dragon.EXPECT().Type().Return(types.DragonWarriorType)

		myField := gamestatus.FieldInput{
			HasArcher: false, HasKnight: false, HasMage: true,
			Warriors: []cards.Warrior{dragon},
		}

		hc := gamestatus.NewSpecialPowerHandCard("SP1",
			myField, []gamestatus.FieldInput{}, []gamestatus.FieldInput{},
			types.PhaseTypeAttack)

		assert.True(t, hc.CanBeUsed)
		assert.Empty(t, hc.CanBeUsedOnIDs)
	})

	t.Run("No field warriors means no targets", func(t *testing.T) {
		myField := gamestatus.FieldInput{HasArcher: false, HasKnight: false, HasMage: false}

		hc := gamestatus.NewSpecialPowerHandCard("SP1",
			myField, []gamestatus.FieldInput{}, []gamestatus.FieldInput{},
			types.PhaseTypeAttack)

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
		wantType gamestatus.CardType
	}{
		{
			name:     "Spy can be used during SpySteal phase",
			cardID:   "SPY1",
			action:   types.PhaseTypeSpySteal,
			wantUsed: true,
			wantType: gamestatus.CardTypeSpy,
		},
		{
			name:     "Spy cannot be used during Attack phase",
			cardID:   "SPY2",
			action:   types.PhaseTypeAttack,
			wantUsed: false,
			wantType: gamestatus.CardTypeSpy,
		},
		{
			name:     "Spy cannot be used during Buy phase",
			cardID:   "SPY3",
			action:   types.PhaseTypeBuy,
			wantUsed: false,
			wantType: gamestatus.CardTypeSpy,
		},
		{
			name:     "Spy cannot be used during Construct phase",
			cardID:   "SPY4",
			action:   types.PhaseTypeConstruct,
			wantUsed: false,
			wantType: gamestatus.CardTypeSpy,
		},
		{
			name:     "Spy cannot be used during DrawCard phase",
			cardID:   "SPY5",
			action:   types.PhaseTypeDrawCard,
			wantUsed: false,
			wantType: gamestatus.CardTypeSpy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc := gamestatus.NewSpyHandCard(tt.cardID, tt.action)

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
		wantType gamestatus.CardType
	}{
		{
			name:     "Thief can be used during SpySteal phase",
			cardID:   "THIEF1",
			action:   types.PhaseTypeSpySteal,
			wantUsed: true,
			wantType: gamestatus.CardTypeThief,
		},
		{
			name:     "Thief cannot be used during Attack phase",
			cardID:   "THIEF2",
			action:   types.PhaseTypeAttack,
			wantUsed: false,
			wantType: gamestatus.CardTypeThief,
		},
		{
			name:     "Thief cannot be used during Buy phase",
			cardID:   "THIEF3",
			action:   types.PhaseTypeBuy,
			wantUsed: false,
			wantType: gamestatus.CardTypeThief,
		},
		{
			name:     "Thief cannot be used during Construct phase",
			cardID:   "THIEF4",
			action:   types.PhaseTypeConstruct,
			wantUsed: false,
			wantType: gamestatus.CardTypeThief,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc := gamestatus.NewThiefHandCard(tt.cardID, tt.action)

			assert.Equal(t, tt.cardID, hc.CardID)
			assert.Equal(t, tt.wantType, hc.CardType)
			assert.Equal(t, tt.wantUsed, hc.CanBeUsed)
			assert.Equal(t, 0, hc.Value)
			assert.Empty(t, hc.CanBeUsedOnIDs)
		})
	}
}

func TestNewSabotageHandCard(t *testing.T) {
	tests := []struct {
		name             string
		cardID           string
		anyEnemyHasCards bool
		action           types.PhaseType
		wantUsed         bool
		wantType         gamestatus.CardType
	}{
		{
			name:             "Usable during SpySteal when enemy has cards",
			cardID:           "SAB1",
			anyEnemyHasCards: true,
			action:           types.PhaseTypeSpySteal,
			wantUsed:         true,
			wantType:         gamestatus.CardTypeSabotage,
		},
		{
			name:             "Not usable during SpySteal when no enemy has cards",
			cardID:           "SAB2",
			anyEnemyHasCards: false,
			action:           types.PhaseTypeSpySteal,
			wantUsed:         false,
			wantType:         gamestatus.CardTypeSabotage,
		},
		{
			name:             "Not usable during Attack phase",
			cardID:           "SAB3",
			anyEnemyHasCards: true,
			action:           types.PhaseTypeAttack,
			wantUsed:         false,
			wantType:         gamestatus.CardTypeSabotage,
		},
		{
			name:             "Not usable during Buy phase",
			cardID:           "SAB4",
			anyEnemyHasCards: true,
			action:           types.PhaseTypeBuy,
			wantUsed:         false,
			wantType:         gamestatus.CardTypeSabotage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc := gamestatus.NewSabotageHandCard(tt.cardID, tt.anyEnemyHasCards, tt.action)

			assert.Equal(t, tt.cardID, hc.CardID)
			assert.Equal(t, tt.wantType, hc.CardType)
			assert.Equal(t, tt.wantUsed, hc.CanBeUsed)
			assert.Equal(t, 0, hc.Value)
			assert.Empty(t, hc.CanBeUsedOnIDs)
		})
	}
}

func TestNewFortressHandCard(t *testing.T) {
	tests := []struct {
		name                  string
		cardID                string
		castleConstructed     bool
		allyCastleConstructed bool
		action                types.PhaseType
		wantUsed              bool
	}{
		{
			name:              "Usable in Construct phase when own castle is constructed",
			cardID:            "FW1",
			castleConstructed: true,
			action:            types.PhaseTypeConstruct,
			wantUsed:          true,
		},
		{
			name:                  "Usable in Construct phase when ally castle is constructed",
			cardID:                "FW1",
			allyCastleConstructed: true,
			action:                types.PhaseTypeConstruct,
			wantUsed:              true,
		},
		{
			name:     "Not usable in Construct phase when no castle is constructed",
			cardID:   "FW1",
			action:   types.PhaseTypeConstruct,
			wantUsed: false,
		},
		{
			name:              "Not usable outside Construct phase even with castle constructed",
			cardID:            "FW1",
			castleConstructed: true,
			action:            types.PhaseTypeAttack,
			wantUsed:          false,
		},
		{
			name:              "Not usable in Buy phase",
			cardID:            "FW1",
			castleConstructed: true,
			action:            types.PhaseTypeBuy,
			wantUsed:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc := gamestatus.NewFortressHandCard(tt.cardID, tt.castleConstructed, tt.allyCastleConstructed, tt.action)

			assert.Equal(t, tt.cardID, hc.CardID)
			assert.Equal(t, gamestatus.CardTypeFortress, hc.CardType)
			assert.Equal(t, 0, hc.Value)
			assert.Equal(t, tt.wantUsed, hc.CanBeUsed)
			assert.Empty(t, hc.CanBeUsedOnIDs)
		})
	}
}

func TestNewResurrectionHandCard(t *testing.T) {
	tests := []struct {
		name          string
		cardID        string
		cemeteryCount int
		action        types.PhaseType
		wantUsed      bool
	}{
		{
			name:          "Usable in Attack phase when cemetery has warriors",
			cardID:        "RES1",
			cemeteryCount: 2,
			action:        types.PhaseTypeAttack,
			wantUsed:      true,
		},
		{
			name:          "Not usable in Attack phase when cemetery is empty",
			cardID:        "RES1",
			cemeteryCount: 0,
			action:        types.PhaseTypeAttack,
			wantUsed:      false,
		},
		{
			name:          "Not usable outside Attack phase even with warriors in cemetery",
			cardID:        "RES1",
			cemeteryCount: 3,
			action:        types.PhaseTypeBuy,
			wantUsed:      false,
		},
		{
			name:          "Not usable in Construct phase",
			cardID:        "RES1",
			cemeteryCount: 1,
			action:        types.PhaseTypeConstruct,
			wantUsed:      false,
		},
		{
			name:          "Not usable in SpySteal phase",
			cardID:        "RES1",
			cemeteryCount: 1,
			action:        types.PhaseTypeSpySteal,
			wantUsed:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc := gamestatus.NewResurrectionHandCard(tt.cardID, tt.cemeteryCount, tt.action)

			assert.Equal(t, tt.cardID, hc.CardID)
			assert.Equal(t, gamestatus.CardTypeResurrection, hc.CardType)
			assert.Equal(t, 0, hc.Value)
			assert.Equal(t, tt.wantUsed, hc.CanBeUsed)
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
		wantType  gamestatus.CardType
	}{
		{
			name:      "Catapult can be used when enemy castle can be attacked in Attack phase",
			cardID:    "CAT1",
			canBeUsed: true,
			action:    types.PhaseTypeAttack,
			wantUsed:  true,
			wantType:  gamestatus.CardTypeCatapult,
		},
		{
			name:      "Catapult cannot be used when enemy castle cannot be attacked in Attack phase",
			cardID:    "CAT2",
			canBeUsed: false,
			action:    types.PhaseTypeAttack,
			wantUsed:  false,
			wantType:  gamestatus.CardTypeCatapult,
		},
		{
			name:      "Catapult cannot be used during Buy phase even if castle can be attacked",
			cardID:    "CAT3",
			canBeUsed: true,
			action:    types.PhaseTypeBuy,
			wantUsed:  false,
			wantType:  gamestatus.CardTypeCatapult,
		},
		{
			name:      "Catapult cannot be used during SpySteal phase",
			cardID:    "CAT4",
			canBeUsed: true,
			action:    types.PhaseTypeSpySteal,
			wantUsed:  false,
			wantType:  gamestatus.CardTypeCatapult,
		},
		{
			name:      "Catapult cannot be used during Construct phase",
			cardID:    "CAT5",
			canBeUsed: true,
			action:    types.PhaseTypeConstruct,
			wantUsed:  false,
			wantType:  gamestatus.CardTypeCatapult,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc := gamestatus.NewCatapultHandCard(tt.cardID, tt.canBeUsed, tt.action)

			assert.Equal(t, tt.cardID, hc.CardID)
			assert.Equal(t, tt.wantType, hc.CardType)
			assert.Equal(t, tt.wantUsed, hc.CanBeUsed)
			assert.Equal(t, 0, hc.Value)
			assert.Empty(t, hc.CanBeUsedOnIDs)
		})
	}
}

func TestNewWeaponHandCard_MercenaryEnablesAllWeapons(t *testing.T) {
	t.Run("Sword usable with Mercenary on field", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		weapon := mocks.NewMockWeapon(ctrl)
		weapon.EXPECT().Type().Return(types.SwordWeaponType)
		weapon.EXPECT().GetID().Return("S1")
		weapon.EXPECT().DamageAmount().Return(5)
		weapon.EXPECT().CanBeTraded().Return(true)

		myField := gamestatus.FieldInput{HasMercenary: true}
		enemy := gamestatus.FieldInput{Warriors: []cards.Warrior{}}

		hc := gamestatus.NewWeaponHandCard(weapon, myField, []gamestatus.FieldInput{enemy}, false, types.PhaseTypeAttack)

		assert.True(t, hc.CanBeUsed)
	})

	t.Run("Arrow usable with Mercenary on field", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		weapon := mocks.NewMockWeapon(ctrl)
		weapon.EXPECT().Type().Return(types.ArrowWeaponType)
		weapon.EXPECT().GetID().Return("A1")
		weapon.EXPECT().DamageAmount().Return(5)
		weapon.EXPECT().CanBeTraded().Return(true)

		myField := gamestatus.FieldInput{HasMercenary: true}
		enemy := gamestatus.FieldInput{Warriors: []cards.Warrior{}}

		hc := gamestatus.NewWeaponHandCard(weapon, myField, []gamestatus.FieldInput{enemy}, false, types.PhaseTypeAttack)

		assert.True(t, hc.CanBeUsed)
	})

	t.Run("Poison usable with Mercenary on field", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		weapon := mocks.NewMockWeapon(ctrl)
		weapon.EXPECT().Type().Return(types.PoisonWeaponType)
		weapon.EXPECT().GetID().Return("P1")
		weapon.EXPECT().DamageAmount().Return(5)
		weapon.EXPECT().CanBeTraded().Return(true)

		myField := gamestatus.FieldInput{HasMercenary: true}
		enemy := gamestatus.FieldInput{Warriors: []cards.Warrior{}}

		hc := gamestatus.NewWeaponHandCard(weapon, myField, []gamestatus.FieldInput{enemy}, false, types.PhaseTypeAttack)

		assert.True(t, hc.CanBeUsed)
	})

	t.Run("Sword not usable when only Mercenary missing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		weapon := mocks.NewMockWeapon(ctrl)
		weapon.EXPECT().Type().Return(types.SwordWeaponType)
		weapon.EXPECT().GetID().Return("S1")
		weapon.EXPECT().DamageAmount().Return(5)
		weapon.EXPECT().CanBeTraded().Return(true)

		myField := gamestatus.FieldInput{HasMercenary: false, HasKnight: false, HasDragon: false}

		hc := gamestatus.NewWeaponHandCard(weapon, myField, []gamestatus.FieldInput{{}}, false, types.PhaseTypeAttack)

		assert.False(t, hc.CanBeUsed)
	})
}

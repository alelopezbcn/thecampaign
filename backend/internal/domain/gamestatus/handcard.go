package gamestatus

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type HandCard struct {
	Card
	CanBeUsedOnIDs []string       `json:"use_on"`
	CanBeUsed      bool           `json:"can_be_used"`
	DmgMultiplier  map[string]int `json:"dmg_mult"`
	CanBeTraded    bool           `json:"can_be_traded"`
}

func newHandCard(cardID string, cardType CardType, value int,
	canBeUsedOnIDs []string, canBeUsed bool,
) HandCard {
	return HandCard{
		Card: Card{
			CardID:   cardID,
			CardType: cardType,
			Value:    value,
		},
		CanBeUsedOnIDs: canBeUsedOnIDs,
		CanBeUsed:      canBeUsed,
	}
}

func NewWarriorHandCard(warrior cards.Warrior) HandCard {
	var aCardType CardType
	switch warrior.Type() {
	case types.KnightWarriorType:
		aCardType = CardTypeKnight
	case types.ArcherWarriorType:
		aCardType = CardTypeArcher
	case types.MageWarriorType:
		aCardType = CardTypeMage
	case types.DragonWarriorType:
		aCardType = CardTypeDragon
	}

	return newHandCard(warrior.GetID(), aCardType,
		warrior.Health(), []string{}, true)
}

func NewWeaponHandCard(weapon cards.Weapon, myField FieldInput,
	enemyFields []FieldInput, castleConstructed bool,
	action types.PhaseType,
) HandCard {
	canBeUsed := false
	var aCardType CardType

	switch weapon.Type() {
	case types.SwordWeaponType:
		aCardType = CardTypeSword
		canBeUsed = myField.HasKnight || myField.HasDragon
	case types.ArrowWeaponType:
		aCardType = CardTypeArrow
		canBeUsed = myField.HasArcher || myField.HasDragon
	case types.PoisonWeaponType:
		aCardType = CardTypePoison
		canBeUsed = myField.HasMage || myField.HasDragon
	}

	if action != types.PhaseTypeConstruct &&
		action != types.PhaseTypeAttack {
		hc := newHandCard(weapon.GetID(), aCardType,
			weapon.DamageAmount(), []string{}, false)
		hc.CanBeTraded = true
		return hc
	}

	if action == types.PhaseTypeConstruct {
		canBeUsed = !castleConstructed && weapon.CanConstruct()
		hc := newHandCard(weapon.GetID(), aCardType,
			weapon.DamageAmount(), []string{}, canBeUsed)
		hc.CanBeTraded = true
		return hc
	}

	mults := map[string]int{}
	attackableIDs := []string{}
	// Build attackableIDs from ALL enemy fields
	for _, ef := range enemyFields {
		for _, v := range ef.Warriors {
			mults[v.GetID()] = weapon.MultiplierFactor(v)
			attackableIDs = append(attackableIDs, v.GetID())
		}
	}

	hc := newHandCard(weapon.GetID(), aCardType,
		weapon.DamageAmount(), attackableIDs, canBeUsed)
	hc.DmgMultiplier = mults
	hc.CanBeTraded = true

	return hc
}

func NewSpecialPowerHandCard(cardID string,
	myField FieldInput, allyFields []FieldInput, enemyFields []FieldInput,
	action types.PhaseType,
) HandCard {
	if action != types.PhaseTypeAttack {
		return newHandCard(cardID, CardTypeSpecialPower,
			0, []string{}, false)
	}

	canBeUsedOnIDs := []string{}

	if myField.HasArcher {
		for _, ef := range enemyFields {
			for _, warrior := range ef.Warriors {
				canBeUsedOnIDs = append(canBeUsedOnIDs, warrior.GetID())
			}
		}
	}

	if myField.HasKnight {
		for _, warrior := range myField.Warriors {
			isProtected, _ := warrior.IsProtected()
			if warrior.Type() == types.DragonWarriorType || isProtected {
				continue
			}
			canBeUsedOnIDs = append(canBeUsedOnIDs, warrior.GetID())
		}
		for _, af := range allyFields {
			for _, warrior := range af.Warriors {
				isProtected, _ := warrior.IsProtected()
				if warrior.Type() == types.DragonWarriorType || isProtected {
					continue
				}
				canBeUsedOnIDs = append(canBeUsedOnIDs, warrior.GetID())
			}
		}
	}

	if myField.HasMage {
		for _, warrior := range myField.Warriors {
			if warrior.Type() == types.DragonWarriorType || !warrior.IsDamaged() {
				continue
			}
			canBeUsedOnIDs = append(canBeUsedOnIDs, warrior.GetID())
		}
		for _, af := range allyFields {
			for _, warrior := range af.Warriors {
				if warrior.Type() == types.DragonWarriorType || !warrior.IsDamaged() {
					continue
				}
				canBeUsedOnIDs = append(canBeUsedOnIDs, warrior.GetID())
			}
		}
	}

	return newHandCard(cardID, CardTypeSpecialPower,
		0, canBeUsedOnIDs, true)
}

func NewHarpoonHandCard(cardID string, enemyFields []FieldInput,
	action types.PhaseType,
) HandCard {
	if action != types.PhaseTypeAttack {
		hc := newHandCard(cardID, CardTypeHarpoon, 0, []string{}, false)
		hc.CanBeTraded = true
		return hc
	}

	canBeUsedOnIDs := []string{}

	for _, ef := range enemyFields {
		for _, warrior := range ef.Warriors {
			if warrior.Type() == types.DragonWarriorType {
				canBeUsedOnIDs = append(canBeUsedOnIDs, warrior.GetID())
			}
		}
	}

	hc := newHandCard(cardID, CardTypeHarpoon, 0, canBeUsedOnIDs, len(canBeUsedOnIDs) > 0)
	hc.CanBeTraded = true
	return hc
}

func NewBloodRainHandCard(cardID string, enemyFields []FieldInput,
	action types.PhaseType,
) HandCard {
	if action != types.PhaseTypeAttack {
		hc := newHandCard(cardID, CardTypeBloodRain, 0, []string{}, false)
		hc.CanBeTraded = true
		return hc
	}

	hasTargets := false
	for _, ef := range enemyFields {
		if len(ef.Warriors) > 0 {
			hasTargets = true
			break
		}
	}

	hc := newHandCard(cardID, CardTypeBloodRain, 0, []string{}, hasTargets)
	hc.CanBeTraded = true
	return hc
}

func NewCatapultHandCard(cardID string, canBeUsed bool,
	action types.PhaseType,
) HandCard {
	if action != types.PhaseTypeAttack {
		return newHandCard(cardID, CardTypeCatapult, 0, []string{}, false)
	}

	return newHandCard(cardID, CardTypeCatapult, 0, []string{},
		canBeUsed)
}

func NewSpyHandCard(cardID string, action types.PhaseType) HandCard {
	return newHandCard(cardID, CardTypeSpy, 0, []string{},
		action == types.PhaseTypeSpySteal)
}

func NewThiefHandCard(cardID string, action types.PhaseType) HandCard {
	return newHandCard(cardID, CardTypeThief, 0, []string{},
		action == types.PhaseTypeSpySteal)
}

func NewResourceHandCard(resource cards.Resource, playerCastleConstructed bool,
	allyCastleConstructed bool, canBuy bool, action types.PhaseType,
) HandCard {
	if action != types.PhaseTypeBuy && action != types.PhaseTypeConstruct {
		return newHandCard(resource.GetID(), CardTypeResource,
			resource.Value(), []string{}, false)
	}

	if action == types.PhaseTypeConstruct {
		// If player's or ally's castle is already constructed, any resource can be added
		if playerCastleConstructed || allyCastleConstructed {
			return newHandCard(resource.GetID(), CardTypeResource,
				resource.Value(), []string{}, true)
		}
		// If no castle has been started, only resources that can start construction
		return newHandCard(resource.GetID(), CardTypeResource,
			resource.Value(), []string{}, resource.CanConstruct())
	}

	return newHandCard(resource.GetID(), CardTypeResource,
		resource.Value(), []string{}, canBuy)
}

// specialWeaponHandCardBuilders maps each special WeaponType to its HandCard builder.
// Standard weapons (Sword/Arrow/Poison) fall through to NewWeaponHandCard.
// Adding a new special weapon = one entry here.
var specialWeaponHandCardBuilders = map[types.WeaponType]func(id string, viewer ViewerInput, game GameStatusDTO, action types.PhaseType) HandCard{
	types.SpecialPowerWeaponType: func(id string, viewer ViewerInput, game GameStatusDTO, action types.PhaseType) HandCard {
		return NewSpecialPowerHandCard(id, viewer.Field, game.AllyFields, game.EnemyFields, action)
	},
	types.HarpoonWeaponType: func(id string, viewer ViewerInput, game GameStatusDTO, action types.PhaseType) HandCard {
		return NewHarpoonHandCard(id, game.EnemyFields, action)
	},
	types.BloodRainWeaponType: func(id string, viewer ViewerInput, game GameStatusDTO, action types.PhaseType) HandCard {
		return NewBloodRainHandCard(id, game.EnemyFields, action)
	},
}

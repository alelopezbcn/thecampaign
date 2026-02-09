package gamestatus

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
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
	canBeUsedOnIDs []string, canBeUsed bool) HandCard {

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

func NewWarriorHandCard(warrior ports.Warrior) HandCard {
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

func NewWeaponHandCard(weapon ports.Weapon, myField ports.Field,
	enemyFields []ports.Field, castleConstructed bool,
	action types.ActionType) HandCard {

	canBeUsed := false
	var aCardType CardType

	switch weapon.Type() {
	case types.SwordWeaponType:
		aCardType = CardTypeSword
		canBeUsed = myField.HasKnight() || myField.HasDragon()
	case types.ArrowWeaponType:
		aCardType = CardTypeArrow
		canBeUsed = myField.HasArcher() || myField.HasDragon()
	case types.PoisonWeaponType:
		aCardType = CardTypePoison
		canBeUsed = myField.HasMage() || myField.HasDragon()
	}

	if action != types.ActionTypeConstruct &&
		action != types.ActionTypeAttack {
		hc := newHandCard(weapon.GetID(), aCardType,
			weapon.DamageAmount(), []string{}, false)
		hc.CanBeTraded = true
		return hc
	}

	if action == types.ActionTypeConstruct {
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
		for _, v := range ef.Warriors() {
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

func NewSpecialPowerHandCard(specialPower ports.SpecialPower,
	myField ports.Field, allyFields []ports.Field, enemyFields []ports.Field,
	action types.ActionType) HandCard {

	if action != types.ActionTypeAttack {
		return newHandCard(specialPower.GetID(), CardTypeSpecialPower,
			0, []string{}, false)
	}

	canBeUsedOnIDs := []string{}

	if myField.HasArcher() {
		for _, ef := range enemyFields {
			for _, warrior := range ef.Warriors() {
				canBeUsedOnIDs = append(canBeUsedOnIDs, warrior.GetID())
			}
		}
	}

	if myField.HasKnight() {
		for _, warrior := range myField.Warriors() {
			isProtected, _ := warrior.IsProtected()
			if warrior.Type() == types.DragonWarriorType || isProtected {
				continue
			}
			canBeUsedOnIDs = append(canBeUsedOnIDs, warrior.GetID())
		}
		for _, af := range allyFields {
			for _, warrior := range af.Warriors() {
				isProtected, _ := warrior.IsProtected()
				if warrior.Type() == types.DragonWarriorType || isProtected {
					continue
				}
				canBeUsedOnIDs = append(canBeUsedOnIDs, warrior.GetID())
			}
		}
	}

	if myField.HasMage() {
		for _, warrior := range myField.Warriors() {
			if warrior.Type() == types.DragonWarriorType || !warrior.IsDamaged() {
				continue
			}
			canBeUsedOnIDs = append(canBeUsedOnIDs, warrior.GetID())
		}
		for _, af := range allyFields {
			for _, warrior := range af.Warriors() {
				if warrior.Type() == types.DragonWarriorType || !warrior.IsDamaged() {
					continue
				}
				canBeUsedOnIDs = append(canBeUsedOnIDs, warrior.GetID())
			}
		}
	}

	return newHandCard(specialPower.GetID(), CardTypeSpecialPower,
		0, canBeUsedOnIDs, true)
}

func NewCatapultHandCard(cardID string, canBeUsed bool,
	action types.ActionType) HandCard {

	if action != types.ActionTypeAttack {
		return newHandCard(cardID, CardTypeCatapult, 0, []string{}, false)
	}

	return newHandCard(cardID, CardTypeCatapult, 0, []string{},
		canBeUsed)
}

func NewSpyHandCard(cardID string, action types.ActionType) HandCard {
	return newHandCard(cardID, CardTypeSpy, 0, []string{},
		action == types.ActionTypeSpySteal)
}

func NewThiefHandCard(cardID string, action types.ActionType) HandCard {
	return newHandCard(cardID, CardTypeThief, 0, []string{},
		action == types.ActionTypeSpySteal)
}

func NewResourceHandCard(resource ports.Resource, playerCastleConstructed bool,
	allyCastleConstructed bool, canBuy bool, action types.ActionType) HandCard {

	if action != types.ActionTypeBuy && action != types.ActionTypeConstruct {
		return newHandCard(resource.GetID(), CardTypeResource,
			resource.Value(), []string{}, false)
	}

	if action == types.ActionTypeConstruct {
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

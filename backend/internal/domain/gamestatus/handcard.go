package gamestatus

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type HandCard struct {
	Card
	CanBeUsedOnIDs []string `json:"use_on"`
	CanBeUsed      bool     `json:"can_be_used"`
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

func newWarriorHandCard(warrior ports.Warrior) HandCard {
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

func newWeaponHandCard(weapon ports.Weapon, myField ports.Field,
	enemyField ports.Field, castleConstructed bool,
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
		return newHandCard(weapon.GetID(), aCardType,
			weapon.DamageAmount(), []string{}, false)
	}

	if action == types.ActionTypeConstruct {
		canBeUsed = !castleConstructed && weapon.CanConstruct()
		return newHandCard(weapon.GetID(), aCardType,
			weapon.DamageAmount(), []string{}, canBeUsed)
	}

	return newHandCard(weapon.GetID(), aCardType,
		weapon.DamageAmount(), enemyField.AttackableIDs(), canBeUsed)
}

func newSpecialPowerHandCard(specialPower ports.SpecialPower,
	myField ports.Field, enemyField ports.Field,
	action types.ActionType) HandCard {

	if action != types.ActionTypeAttack {
		return newHandCard(specialPower.GetID(), CardTypeSpecialPower,
			0, []string{}, false)
	}

	canBeUsedOnIDs := []string{}

	if myField.HasArcher() {
		for _, warrior := range enemyField.Warriors() {
			if ok, card := warrior.IsProtected(); ok {
				canBeUsedOnIDs = append(canBeUsedOnIDs, card.GetID())
			} else {
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
	}
	if myField.HasMage() {
		for _, warrior := range myField.Warriors() {
			if warrior.Type() == types.DragonWarriorType || !warrior.IsDamaged() {
				continue
			}
			canBeUsedOnIDs = append(canBeUsedOnIDs, warrior.GetID())
		}
	}

	return newHandCard(specialPower.GetID(), CardTypeSpecialPower,
		0, canBeUsedOnIDs, true)
}

func newCatapultHandCard(cardID string, enemyCastleCanBeAttacked bool,
	action types.ActionType) HandCard {

	if action != types.ActionTypeAttack {
		return newHandCard(cardID, CardTypeCatapult, 0, []string{}, false)
	}

	return newHandCard(cardID, CardTypeCatapult, 0, []string{},
		enemyCastleCanBeAttacked)
}

func newSpyHandCard(cardID string, action types.ActionType) HandCard {
	return newHandCard(cardID, CardTypeSpy, 0, []string{},
		action == types.ActionTypeSpySteal)
}

func newThiefHandCard(cardID string, action types.ActionType) HandCard {
	return newHandCard(cardID, CardTypeThief, 0, []string{},
		action == types.ActionTypeSpySteal)
}

func newResourceHandCard(resource ports.Resource, playerCastleConstructed bool,
	action types.ActionType) HandCard {

	if action != types.ActionTypeBuy && action != types.ActionTypeConstruct {
		return newHandCard(resource.GetID(), CardTypeResource,
			resource.Value(), []string{}, false)
	}

	if action == types.ActionTypeConstruct {
		// If player's castle is already constructed/started, any resource can be added
		if playerCastleConstructed {
			return newHandCard(resource.GetID(), CardTypeResource,
				resource.Value(), []string{}, true)
		}
		// If player's castle hasn't started, only resources that can start construction
		return newHandCard(resource.GetID(), CardTypeResource,
			resource.Value(), []string{}, resource.CanConstruct())
	}

	// Buy action - use resource's CanBuy method
	return newHandCard(resource.GetID(), CardTypeResource,
		resource.Value(), []string{}, resource.CanBuy())
}

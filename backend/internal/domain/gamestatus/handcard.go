package gamestatus

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type HandCard struct {
	Card
	CanBeUsedOnIDs []string
	CanConstruct   bool
}

func newHandCard(cardID string, cardType CardType, value int,
	canBeUsedOnIDs []string, canConstruct bool) HandCard {

	return HandCard{
		Card: Card{
			CardID:   cardID,
			CardType: cardType,
			Value:    value,
		},
		CanBeUsedOnIDs: canBeUsedOnIDs,
		CanConstruct:   canConstruct,
	}
}

func newWarriorHandCard(warrior ports.Warrior) HandCard {
	var aCardType CardType
	switch warrior.Type() {
	case ports.KnightWarriorType:
		aCardType = CardTypeKnight
	case ports.ArcherWarriorType:
		aCardType = CardTypeArcher
	case ports.MageWarriorType:
		aCardType = CardTypeMage
	case ports.DragonWarriorType:
		aCardType = CardTypeDragon
	}

	return newHandCard(warrior.GetID(), aCardType,
		warrior.Health(), []string{}, false)
}

func newWeaponHandCard(weapon ports.Weapon, myField ports.Field,
	attackableIDs []string) HandCard {

	canBeUsed := false
	var aCardType CardType

	switch weapon.Type() {
	case ports.SwordWeaponType:
		aCardType = CardTypeSword
		canBeUsed = myField.HasKnight() || myField.HasDragon()
	case ports.ArrowWeaponType:
		aCardType = CardTypeArrow
		canBeUsed = myField.HasArcher() || myField.HasDragon()
	case ports.PoisonWeaponType:
		aCardType = CardTypePoison
		canBeUsed = myField.HasMage() || myField.HasDragon()
	}

	if !canBeUsed {
		attackableIDs = []string{}
	}

	return newHandCard(weapon.GetID(), aCardType,
		weapon.DamageAmount(), attackableIDs, weapon.CanConstruct())
}

func newSpecialPowerHandCard(specialPower ports.SpecialPower,
	myField ports.Field, enemyField ports.Field) HandCard {

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
			if warrior.Type() == ports.DragonWarriorType || isProtected {
				continue
			}
			canBeUsedOnIDs = append(canBeUsedOnIDs, warrior.GetID())
		}
	}
	if myField.HasMage() {
		for _, warrior := range myField.Warriors() {
			if warrior.Type() == ports.DragonWarriorType || !warrior.IsDamaged() {
				continue
			}
			canBeUsedOnIDs = append(canBeUsedOnIDs, warrior.GetID())
		}
	}

	return newHandCard(specialPower.GetID(), CardTypeSpecialPower,
		0, canBeUsedOnIDs, false)
}

func newCatapultHandCard(cardID string, castleID string) HandCard {
	canBeUsedOnIDs := []string{}
	if castleID != "" {
		canBeUsedOnIDs = append(canBeUsedOnIDs, castleID)
	}
	return newHandCard(cardID, CardTypeCatapult, 0, canBeUsedOnIDs, false)
}

func newSpyHandCard(cardID string) HandCard {
	return newHandCard(cardID, CardTypeSpy, 0, []string{}, false)
}

func newThiefHandCard(cardID string) HandCard {
	return newHandCard(cardID, CardTypeThief, 0, []string{}, false)
}

func newResourceHandCard(resource ports.Resource) HandCard {
	return newHandCard(resource.GetID(), CardTypeResource,
		resource.Value(), []string{}, resource.CanConstruct())
}

func (c HandCard) CanBeUsed() bool {
	return len(c.CanBeUsedOnIDs) > 0
}

func (c HandCard) String() string {
	return fmt.Sprintf("%s | CanAffect: %v", c.Card.String(), c.CanBeUsedOnIDs)
}

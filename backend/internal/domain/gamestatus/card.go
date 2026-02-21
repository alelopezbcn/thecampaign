package gamestatus

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type Card struct {
	CardID   string   `json:"card_id"`
	CardType CardType `json:"card_type"`
	Value    int      `json:"value"`
}

func newCard(cardID string, cardType CardType, value int) Card {
	return Card{
		CardID:   cardID,
		CardType: cardType,
		Value:    value,
	}
}

func fromDomainCards(dcs []cards.Card) []Card {
	cards := []Card{}
	for _, v := range dcs {
		cards = append(cards, fromDomainCard(v))
	}

	return cards
}

func fromDomainCard(dc cards.Card) Card {
	cardID := dc.GetID()
	var aCardType CardType
	var value int

	switch c := dc.(type) {
	case cards.Warrior:
		switch c.Type() {
		case types.ArcherWarriorType:
			aCardType = CardTypeArcher
		case types.MageWarriorType:
			aCardType = CardTypeMage
		case types.KnightWarriorType:
			aCardType = CardTypeKnight
		case types.DragonWarriorType:
			aCardType = CardTypeDragon
		}
		value = c.Health()

	case cards.Weapon:
		switch c.Type() {
		case types.SwordWeaponType:
			aCardType = CardTypeSword
		case types.ArrowWeaponType:
			aCardType = CardTypeArrow
		case types.PoisonWeaponType:
			aCardType = CardTypePoison
		case types.SpecialPowerWeaponType:
			aCardType = CardTypeSpecialPower
		case types.HarpoonWeaponType:
			aCardType = CardTypeHarpoon
		}
		value = c.DamageAmount()

	case cards.Resource:
		aCardType = CardTypeResource
		value = c.Value()

	case cards.Spy:
		aCardType = CardTypeSpy
		value = 0

	case cards.Thief:
		aCardType = CardTypeThief
		value = 0

	case cards.Catapult:
		aCardType = CardTypeCatapult
		value = 0
	}

	return newCard(cardID, aCardType, value)
}

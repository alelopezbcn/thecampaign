package gamestatus

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
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

func (c Card) String() string {
	return fmt.Sprintf("%s - %s (%d)", c.CardID, c.CardType.String(), c.Value)
}

func FromDomainCards(dcs []ports.Card) []Card {
	cards := []Card{}
	for _, v := range dcs {
		cards = append(cards, FromDomainCard(v))
	}

	return cards
}

func FromDomainCard(dc ports.Card) Card {
	cardID := dc.GetID()
	var aCardType CardType
	var value int

	switch c := dc.(type) {
	case ports.Warrior:
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

	case ports.Weapon:
		switch c.Type() {
		case types.SwordWeaponType:
			aCardType = CardTypeSword
		case types.ArrowWeaponType:
			aCardType = CardTypeArrow
		case types.PoisonWeaponType:
			aCardType = CardTypePoison
		case types.SpecialPowerWeaponType:
			aCardType = CardTypeSpecialPower
		}
		value = c.DamageAmount()

	case ports.Resource:
		aCardType = CardTypeResource
		value = c.Value()

	case ports.Spy:
		aCardType = CardTypeSpy
		value = 0

	case ports.Thief:
		aCardType = CardTypeThief
		value = 0

	case ports.Catapult:
		aCardType = CardTypeCatapult
		value = 0
	}

	return newCard(cardID, aCardType, value)

}

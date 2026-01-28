package gamestatus

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type FieldCard struct {
	Card
	AttackedBy  []Card `json:"attacked_by"`
	ProtectedBy Card   `json:"protected_by"`
}

func NewFieldCard(warrior ports.Warrior) FieldCard {

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

	var attackedByCards []Card
	for _, attacker := range warrior.AttackedBy() {
		var weaponCardType CardType
		switch attacker.Type() {
		case types.SwordWeaponType:
			weaponCardType = CardTypeSword
		case types.ArrowWeaponType:
			weaponCardType = CardTypeArrow
		case types.PoisonWeaponType:
			weaponCardType = CardTypePoison
		}
		attackedByCards = append(attackedByCards, newCard(attacker.GetID(),
			weaponCardType, attacker.DamageAmount()))
	}

	var protectedByCard Card
	if ok, shield := warrior.IsProtected(); ok {
		protectedByCard = newCard(shield.GetID(), CardTypeSpecialPower, shield.Health())
	}

	return FieldCard{
		Card:        newCard(warrior.GetID(), aCardType, warrior.Health()),
		AttackedBy:  attackedByCards,
		ProtectedBy: protectedByCard,
	}
}

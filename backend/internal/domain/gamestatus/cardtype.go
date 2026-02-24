package gamestatus

import "github.com/alelopezbcn/thecampaign/internal/domain/types"

type CardType struct {
	Name    string `json:"name"`
	SubName string `json:"sub_name"`
	Color   string `json:"color"`
}

var (
	CardTypeKnight       = CardType{"Warrior", "Knight", "#0969bd"} // Blue
	CardTypeSword        = CardType{"Weapon", "Sword", "#5fd7ff"}   // Light Blue
	CardTypeArcher       = CardType{"Warrior", "Archer", "#348b34"} // Green
	CardTypeArrow        = CardType{"Weapon", "Arrow", "#61dd61"}   // Light Green
	CardTypeMage         = CardType{"Warrior", "Mage", "#892e89"}   // Purple
	CardTypePoison       = CardType{"Weapon", "Poison", "#e571e5"}  // Light Purple (Thistle)
	CardTypeDragon       = CardType{"Warrior", "Dragon", "#FF0000"} // Red
	CardTypeResource     = CardType{"Resource", "", "#fbfbae"}      // Yellow
	CardTypeSpecialPower = CardType{"SpecialPower", "", "#FFA500"}  // Orange
	CardTypeSpy          = CardType{"Spy", "", "#D3D3D3"}           // Light Gray
	CardTypeThief        = CardType{"Thief", "", "#D3D3D3"}         // Light Gray
	CardTypeCatapult     = CardType{"Catapult", "", "#D3D3D3"}      // Light Gray
	CardTypeHarpoon      = CardType{"Harpoon", "", "#c80000"}       // Dark Red
	CardTypeBloodRain    = CardType{"BloodRain", "", "#FFA500"}     // Orange
)

// zeroValueCardTypes maps the card Name() to its CardType for singleton cards
// that always have value 0. Add one entry here when introducing a new such card.
var zeroValueCardTypes = map[string]CardType{
	"Spy":      CardTypeSpy,
	"Thief":    CardTypeThief,
	"Catapult": CardTypeCatapult,
}

// warriorCardTypes maps each WarriorType to its CardType for serialization.
// Add one entry here when introducing a new warrior.
var warriorCardTypes = map[types.WarriorType]CardType{
	types.KnightWarriorType: CardTypeKnight,
	types.ArcherWarriorType: CardTypeArcher,
	types.MageWarriorType:   CardTypeMage,
	types.DragonWarriorType: CardTypeDragon,
}

// weaponCardTypes maps each WeaponType to its CardType for serialization.
// Add one entry here when introducing a new weapon.
var weaponCardTypes = map[types.WeaponType]CardType{
	types.SwordWeaponType:        CardTypeSword,
	types.ArrowWeaponType:        CardTypeArrow,
	types.PoisonWeaponType:       CardTypePoison,
	types.SpecialPowerWeaponType: CardTypeSpecialPower,
	types.HarpoonWeaponType:      CardTypeHarpoon,
	types.BloodRainWeaponType:    CardTypeBloodRain,
}

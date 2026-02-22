package gamestatus

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
	CardTypeBloodRain    = CardType{"Blood Rain", "", "#FFA500"}    // Orange
)

package gamestatus

type CardType struct {
	Name    string `json:"name"`
	SubName string `json:"sub_name"`
	Color   string `json:"color"`
}

var (
	CardTypeKnight       = CardType{"Warrior", "Knight", "#0000FF"} // Blue
	CardTypeSword        = CardType{"Weapon", "Sword", "#ADD8E6"}   // Light Blue
	CardTypeArcher       = CardType{"Warrior", "Archer", "#008000"} // Green
	CardTypeArrow        = CardType{"Weapon", "Arrow", "#90EE90"}   // Light Green
	CardTypeMage         = CardType{"Warrior", "Mage", "#800080"}   // Purple
	CardTypePoison       = CardType{"Weapon", "Poison", "#D8BFD8"}  // Light Purple (Thistle)
	CardTypeDragon       = CardType{"Warrior", "Dragon", "#FF0000"} // Red
	CardTypeResource     = CardType{"Resource", "", "#FFFF00"}      // Yellow
	CardTypeSpecialPower = CardType{"SpecialPower", "", "#FFA500"}  // Orange
	CardTypeSpy          = CardType{"Spy", "", "#D3D3D3"}           // Light Gray
	CardTypeThief        = CardType{"Thief", "", "#D3D3D3"}         // Light Gray
	CardTypeCatapult     = CardType{"Catapult", "", "#D3D3D3"}      // Light Gray
)

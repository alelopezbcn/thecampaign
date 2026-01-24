package gamestatus

type CardType struct {
	Name  string
	Color string
}

func (c CardType) String() string {
	return c.Name
}

var (
	CardTypeKnight       = CardType{"Knight", "#0000FF"}       // Blue
	CardTypeSword        = CardType{"Weapon", "#ADD8E6"}       // Light Blue
	CardTypeArcher       = CardType{"Archer", "#008000"}       // Green
	CardTypeArrow        = CardType{"Weapon", "#90EE90"}       // Light Green
	CardTypeMage         = CardType{"Mage", "#800080"}         // Purple
	CardTypePoison       = CardType{"Weapon", "#D8BFD8"}       // Light Purple (Thistle)
	CardTypeDragon       = CardType{"Dragon", "#FF0000"}       // Red
	CardTypeResource     = CardType{"Resource", "#FFFF00"}     // Yellow
	CardTypeSpecialPower = CardType{"SpecialPower", "#FFA500"} // Orange
	CardTypeSpy          = CardType{"Spy", "#D3D3D3"}          // Light Gray
	CardTypeThief        = CardType{"Thief", "#D3D3D3"}        // Light Gray
	CardTypeCatapult     = CardType{"Catapult", "#D3D3D3"}     // Light Gray
)

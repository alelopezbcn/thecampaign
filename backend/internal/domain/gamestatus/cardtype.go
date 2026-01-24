package gamestatus

type cardType struct {
	Name  string
	Color string
}

func (c cardType) String() string {
	return c.Name
}

var (
	CardTypeKnight       = cardType{"Knight", "#0000FF"}       // Blue
	CardTypeSword        = cardType{"Weapon", "#ADD8E6"}       // Light Blue
	CardTypeArcher       = cardType{"Archer", "#008000"}       // Green
	CardTypeArrow        = cardType{"Weapon", "#90EE90"}       // Light Green
	CardTypeMage         = cardType{"Mage", "#800080"}         // Purple
	CardTypePoison       = cardType{"Weapon", "#D8BFD8"}       // Light Purple (Thistle)
	CardTypeDragon       = cardType{"Dragon", "#FF0000"}       // Red
	CardTypeResource     = cardType{"Resource", "#FFFF00"}     // Yellow
	CardTypeSpecialPower = cardType{"SpecialPower", "#FFA500"} // Orange
	CardTypeSpy          = cardType{"Spy", "#D3D3D3"}          // Light Gray
	CardTypeThief        = cardType{"Thief", "#D3D3D3"}        // Light Gray
	CardTypeCatapult     = cardType{"Catapult", "#D3D3D3"}     // Light Gray
)

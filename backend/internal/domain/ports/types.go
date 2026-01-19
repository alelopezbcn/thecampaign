package ports

type WarriorType string

const (
	ArcherType WarriorType = "Archer"
	KnightType WarriorType = "Knight"
	MageType   WarriorType = "Mage"
	DragonType WarriorType = "Dragon"
)

func (wt WarriorType) String() string {
	return string(wt)
}

type WeaponType string

const (
	SwordType        WeaponType = "Sword"
	ArrowType        WeaponType = "Arrow"
	PoisonType       WeaponType = "Poison"
	SpecialPowerType WeaponType = "Special Power"
)

func (wt WeaponType) String() string {
	return string(wt)
}
